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

	// Get workflow_id from path parameters
	workflowID := request.PathParameters["id"]
	if workflowID == "" {
		return response.BadRequest("Missing workflow_id"), nil
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
	// Admin can delete any workflow in the project
	// Creator can delete their own workflow
	isAdmin, err := db.CheckProjectAdmin(database, claims.DID, workflow.ProjectID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		return response.InternalError("Failed to check permissions"), nil
	}

	isCreator := workflow.CreatorDID == claims.DID

	if !isAdmin && !isCreator {
		return response.Forbidden("Only admin or creator can delete this workflow"), nil
	}

	// Delete workflow
	if err := deleteWorkflow(workflowID); err != nil {
		log.Printf("Error deleting workflow: %v", err)
		return response.InternalError("Failed to delete workflow"), nil
	}

	return response.Success(map[string]interface{}{
		"message": "Workflow deleted successfully",
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

func deleteWorkflow(workflowID string) error {
	query := `DELETE FROM workflows WHERE workflow_id = $1`
	_, err := database.Exec(query, workflowID)
	return err
}

func main() {
	lambda.Start(handler)
}
