package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/xzero/ai-workflow/pkg/auth"
	"github.com/xzero/ai-workflow/pkg/db"
	"github.com/xzero/ai-workflow/pkg/models"
	"github.com/xzero/ai-workflow/pkg/response"
)

var database *sql.DB

func init() {
	var err error
	database, err = db.Connect(
		os.Getenv("SUPABASE_URL"),
		os.Getenv("DB_PASSWORD"),
	)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract and validate JWT token
	token, err := auth.ExtractToken(request.Headers["Authorization"])
	if err != nil {
		return response.Unauthorized("Invalid authorization header"), nil
	}

	claims, err := auth.ValidateToken(token, os.Getenv("JWT_SECRET"))
	if err != nil {
		return response.Unauthorized("Invalid or expired token"), nil
	}

	// Get project_id from path parameters
	projectID := request.PathParameters["projectId"]
	if projectID == "" {
		return response.BadRequest("Missing project_id"), nil
	}

	// Check if user has access to the project
	hasAccess, err := db.CheckProjectAccess(database, claims.DID, projectID)
	if err != nil {
		log.Printf("Error checking project access: %v", err)
		return response.InternalError("Failed to check project access"), nil
	}
	if !hasAccess {
		return response.Forbidden("Access denied to this project"), nil
	}

	// Get visible workflows for the project
	workflows, err := getVisibleWorkflows(projectID, claims.DID)
	if err != nil {
		log.Printf("Error getting workflows: %v", err)
		return response.InternalError("Failed to get workflows"), nil
	}

	return response.Success(workflows), nil
}

func getVisibleWorkflows(projectID, userDID string) ([]models.Workflow, error) {
	query := `
		SELECT 
			w.workflow_id,
			w.workflow_name,
			w.description,
			w.source,
			w.template_name,
			w.http_method,
			w.base_url,
			w.bearer_token,
			w.external_workflow_id,
			w.parameters,
			w.headers,
			w.project_id,
			w.creator_did,
			w.is_shared,
			w.created_at,
			w.updated_at
		FROM workflows w
		LEFT JOIN project_workflow_settings pws 
			ON w.workflow_id = pws.workflow_id 
			AND pws.project_id = $1
		WHERE (
			w.project_id = $1 OR w.is_shared = true
		)
		AND (
			pws.is_hidden IS NULL OR pws.is_hidden = false
		)
		ORDER BY w.created_at DESC
	`

	rows, err := database.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []models.Workflow
	for rows.Next() {
		var w models.Workflow
		err := rows.Scan(
			&w.WorkflowID,
			&w.WorkflowName,
			&w.Description,
			&w.Source,
			&w.TemplateName,
			&w.HTTPMethod,
			&w.BaseURL,
			&w.BearerToken,
			&w.ExternalWorkflowID,
			&w.Parameters,
			&w.Headers,
			&w.ProjectID,
			&w.CreatorDID,
			&w.IsShared,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, w)
	}

	if workflows == nil {
		workflows = []models.Workflow{}
	}

	return workflows, nil
}

func main() {
	lambda.Start(handler)
}
