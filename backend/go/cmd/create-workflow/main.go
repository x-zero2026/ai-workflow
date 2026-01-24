package main

import (
	"context"
	"database/sql"
	"encoding/json"
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

	// Parse request body
	var req models.CreateWorkflowRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.BadRequest("Invalid request body"), nil
	}

	// Validate required fields
	if req.WorkflowName == "" || req.Description == "" || req.Source == "" ||
		req.TemplateName == "" || req.HTTPMethod == "" || req.BaseURL == "" ||
		req.BearerToken == "" || req.ExternalWorkflowID == "" || req.ProjectID == "" {
		return response.BadRequest("Missing required fields"), nil
	}

	// Validate source
	if req.Source != "coze" && req.Source != "n8n" {
		return response.BadRequest("Invalid source, must be 'coze' or 'n8n'"), nil
	}

	// Validate template_name
	if req.TemplateName != "workflow" && req.TemplateName != "streamflow" {
		return response.BadRequest("Invalid template_name, must be 'workflow' or 'streamflow'"), nil
	}

	// Validate http_method
	if req.HTTPMethod != "GET" && req.HTTPMethod != "POST" && req.HTTPMethod != "PUT" {
		return response.BadRequest("Invalid http_method, must be 'GET', 'POST', or 'PUT'"), nil
	}

	// Check if user has access to the project
	hasAccess, err := db.CheckProjectAccess(database, claims.DID, req.ProjectID)
	if err != nil {
		log.Printf("Error checking project access: %v", err)
		return response.InternalError("Failed to check project access"), nil
	}
	if !hasAccess {
		return response.Forbidden("Access denied to this project"), nil
	}

	// Set default parameters and headers if not provided
	if req.Parameters == nil {
		req.Parameters = json.RawMessage("{}")
	}
	if req.Headers == nil {
		req.Headers = json.RawMessage("{}")
	}

	// Create workflow
	workflowID, err := createWorkflow(&req, claims.DID)
	if err != nil {
		log.Printf("Error creating workflow: %v", err)
		return response.InternalError("Failed to create workflow"), nil
	}

	return response.Success(map[string]interface{}{
		"workflow_id":   workflowID,
		"workflow_name": req.WorkflowName,
	}), nil
}

func createWorkflow(req *models.CreateWorkflowRequest, creatorDID string) (string, error) {
	query := `
		INSERT INTO workflows (
			workflow_name, description, source, template_name,
			http_method, base_url, bearer_token, external_workflow_id,
			parameters, headers, project_id, creator_did
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING workflow_id
	`

	var workflowID string
	err := database.QueryRow(
		query,
		req.WorkflowName,
		req.Description,
		req.Source,
		req.TemplateName,
		req.HTTPMethod,
		req.BaseURL,
		req.BearerToken,
		req.ExternalWorkflowID,
		req.Parameters,
		req.Headers,
		req.ProjectID,
		creatorDID,
	).Scan(&workflowID)

	return workflowID, err
}

func main() {
	lambda.Start(handler)
}
