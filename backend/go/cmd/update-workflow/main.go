package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Get workflow_id from path parameters
	workflowID := request.PathParameters["id"]
	if workflowID == "" {
		return response.BadRequest("Missing workflow_id"), nil
	}

	// Parse request body
	var req models.UpdateWorkflowRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.BadRequest("Invalid request body"), nil
	}

	// Get workflow to check permissions
	workflow, err := getWorkflow(workflowID)
	if err == sql.ErrNoRows {
		return response.NotFound("Workflow not found"), nil
	}
	if err != nil {
		log.Printf("Error getting workflow: %v", err)
		return response.InternalError("Failed to get workflow"), nil
	}

	// Check permissions
	// Admin can modify any workflow in the project
	// Creator can modify their own workflow
	isAdmin, err := db.CheckProjectAdmin(database, claims.DID, workflow.ProjectID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		return response.InternalError("Failed to check permissions"), nil
	}

	isCreator := workflow.CreatorDID == claims.DID

	if !isAdmin && !isCreator {
		return response.Forbidden("Only admin or creator can update this workflow"), nil
	}

	// Validate fields if provided
	if req.Source != nil && *req.Source != "coze" && *req.Source != "n8n" {
		return response.BadRequest("Invalid source, must be 'coze' or 'n8n'"), nil
	}
	if req.TemplateName != nil && *req.TemplateName != "workflow" && *req.TemplateName != "streamflow" {
		return response.BadRequest("Invalid template_name, must be 'workflow' or 'streamflow'"), nil
	}
	if req.HTTPMethod != nil && *req.HTTPMethod != "GET" && *req.HTTPMethod != "POST" && *req.HTTPMethod != "PUT" {
		return response.BadRequest("Invalid http_method, must be 'GET', 'POST', or 'PUT'"), nil
	}

	// Update workflow
	if err := updateWorkflow(workflowID, &req); err != nil {
		log.Printf("Error updating workflow: %v", err)
		return response.InternalError("Failed to update workflow"), nil
	}

	return response.Success(map[string]interface{}{
		"workflow_id": workflowID,
		"message":     "Workflow updated successfully",
	}), nil
}

func getWorkflow(workflowID string) (*models.Workflow, error) {
	query := `
		SELECT workflow_id, project_id, creator_did
		FROM workflows
		WHERE workflow_id = $1
	`

	var w models.Workflow
	err := database.QueryRow(query, workflowID).Scan(
		&w.WorkflowID,
		&w.ProjectID,
		&w.CreatorDID,
	)

	if err != nil {
		return nil, err
	}

	return &w, nil
}

func updateWorkflow(workflowID string, req *models.UpdateWorkflowRequest) error {
	// Build dynamic UPDATE query
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if req.WorkflowName != nil {
		setClauses = append(setClauses, fmt.Sprintf("workflow_name = $%d", argIndex))
		args = append(args, *req.WorkflowName)
		argIndex++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Source != nil {
		setClauses = append(setClauses, fmt.Sprintf("source = $%d", argIndex))
		args = append(args, *req.Source)
		argIndex++
	}
	if req.TemplateName != nil {
		setClauses = append(setClauses, fmt.Sprintf("template_name = $%d", argIndex))
		args = append(args, *req.TemplateName)
		argIndex++
	}
	if req.HTTPMethod != nil {
		setClauses = append(setClauses, fmt.Sprintf("http_method = $%d", argIndex))
		args = append(args, *req.HTTPMethod)
		argIndex++
	}
	if req.BaseURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("base_url = $%d", argIndex))
		args = append(args, *req.BaseURL)
		argIndex++
	}
	if req.BearerToken != nil {
		setClauses = append(setClauses, fmt.Sprintf("bearer_token = $%d", argIndex))
		args = append(args, *req.BearerToken)
		argIndex++
	}
	if req.ExternalWorkflowID != nil {
		setClauses = append(setClauses, fmt.Sprintf("external_workflow_id = $%d", argIndex))
		args = append(args, *req.ExternalWorkflowID)
		argIndex++
	}
	if req.Parameters != nil {
		setClauses = append(setClauses, fmt.Sprintf("parameters = $%d", argIndex))
		args = append(args, *req.Parameters)
		argIndex++
	}
	if req.Headers != nil {
		setClauses = append(setClauses, fmt.Sprintf("headers = $%d", argIndex))
		args = append(args, *req.Headers)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil // Nothing to update
	}

	// Add updated_at timestamp
	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))

	// Add workflow_id to args
	args = append(args, workflowID)

	query := fmt.Sprintf("UPDATE workflows SET %s WHERE workflow_id = $%d", 
		strings.Join(setClauses, ", "), argIndex)

	_, err := database.Exec(query, args...)
	return err
}

func main() {
	lambda.Start(handler)
}
