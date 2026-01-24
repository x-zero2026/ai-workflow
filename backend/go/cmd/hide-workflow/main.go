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

	// Get project_id and workflow_id from path parameters
	projectID := request.PathParameters["projectId"]
	workflowID := request.PathParameters["workflowId"]
	
	if projectID == "" || workflowID == "" {
		return response.BadRequest("Missing project_id or workflow_id"), nil
	}

	// Parse request body
	var req models.HideWorkflowRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.BadRequest("Invalid request body"), nil
	}

	// Check permissions - only project admin can hide workflows
	isAdmin, err := db.CheckProjectAdmin(database, claims.DID, projectID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		return response.InternalError("Failed to check permissions"), nil
	}

	if !isAdmin {
		return response.Forbidden("Only project admin can hide workflows"), nil
	}

	// Check if workflow exists
	exists, err := workflowExists(workflowID)
	if err != nil {
		log.Printf("Error checking workflow existence: %v", err)
		return response.InternalError("Failed to check workflow"), nil
	}
	if !exists {
		return response.NotFound("Workflow not found"), nil
	}

	// Update hide status
	if err := updateHideStatus(projectID, workflowID, req.IsHidden); err != nil {
		log.Printf("Error updating hide status: %v", err)
		return response.InternalError("Failed to update hide status"), nil
	}

	return response.Success(map[string]interface{}{
		"project_id":  projectID,
		"workflow_id": workflowID,
		"is_hidden":   req.IsHidden,
	}), nil
}

func workflowExists(workflowID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM workflows WHERE workflow_id = $1)`
	err := database.QueryRow(query, workflowID).Scan(&exists)
	return exists, err
}

func updateHideStatus(projectID, workflowID string, isHidden bool) error {
	// Use UPSERT to insert or update the hide status
	query := `
		INSERT INTO project_workflow_settings (project_id, workflow_id, is_hidden)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, workflow_id)
		DO UPDATE SET is_hidden = $3, updated_at = CURRENT_TIMESTAMP
	`
	_, err := database.Exec(query, projectID, workflowID, isHidden)
	return err
}

func main() {
	lambda.Start(handler)
}
