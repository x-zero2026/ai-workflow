package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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
	var req models.ExecuteWorkflowRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.BadRequest("Invalid request body"), nil
	}

	// Get workflow
	workflow, err := getWorkflow(workflowID)
	if err == sql.ErrNoRows {
		return response.NotFound("Workflow not found"), nil
	}
	if err != nil {
		log.Printf("Error getting workflow: %v", err)
		return response.InternalError("Failed to get workflow"), nil
	}

	// Check if user has access to the workflow's project OR if workflow is shared
	hasAccess, err := db.CheckProjectAccess(database, claims.DID, workflow.ProjectID)
	if err != nil {
		log.Printf("Error checking project access: %v", err)
		return response.InternalError("Failed to check project access"), nil
	}
	
	// Allow access if user has project access OR workflow is shared
	if !hasAccess && !workflow.IsShared {
		return response.Forbidden("Access denied to this workflow"), nil
	}

	// Execute workflow
	result, err := executeWorkflow(workflow, &req)
	if err != nil {
		log.Printf("Error executing workflow: %v", err)
		return response.InternalError("Failed to execute workflow: " + err.Error()), nil
	}

	return response.Success(result), nil
}

func getWorkflow(workflowID string) (*models.Workflow, error) {
	query := `
		SELECT 
			workflow_id, workflow_name, description, source, template_name,
			http_method, base_url, bearer_token, external_workflow_id,
			parameters, headers, project_id, creator_did, is_shared,
			created_at, updated_at
		FROM workflows
		WHERE workflow_id = $1
	`

	var w models.Workflow
	err := database.QueryRow(query, workflowID).Scan(
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

	return &w, nil
}

func executeWorkflow(workflow *models.Workflow, req *models.ExecuteWorkflowRequest) (*models.ExecuteWorkflowResponse, error) {
	// Merge parameters: use request parameters if provided, otherwise use workflow defaults
	var parameters map[string]interface{}
	if req.Parameters != nil && len(req.Parameters) > 0 {
		if err := json.Unmarshal(req.Parameters, &parameters); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(workflow.Parameters, &parameters); err != nil {
			return nil, err
		}
	}

	// Build request body based on source
	var requestBody map[string]interface{}
	
	if workflow.Source == "coze" {
		// Coze API format: { "workflow_id": "xxx", "parameters": { ... } }
		requestBody = map[string]interface{}{
			"workflow_id": workflow.ExternalWorkflowID,
			"parameters":  parameters,
		}
	} else {
		// n8n or other formats: add workflow_id to parameters directly
		parameters["workflow_id"] = workflow.ExternalWorkflowID
		requestBody = parameters
	}

	// Merge headers: use request headers if provided, otherwise use workflow defaults
	var customHeaders map[string]string
	if req.Headers != nil && len(req.Headers) > 0 {
		if err := json.Unmarshal(req.Headers, &customHeaders); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(workflow.Headers, &customHeaders); err != nil {
			return nil, err
		}
	}

	// Build request headers with defaults
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + workflow.BearerToken
	headers["Content-Type"] = "application/json"

	// Merge custom headers (ALLOW overriding any header including Authorization)
	// User's input takes precedence
	for k, v := range customHeaders {
		headers[k] = v
	}

	// Build request body
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(workflow.HTTPMethod, workflow.BaseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	// Execute HTTP request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response body as JSON
	var respBodyJSON interface{}
	if err := json.Unmarshal(respBody, &respBodyJSON); err != nil {
		// If not JSON, use raw string
		respBodyJSON = string(respBody)
	}

	// Build response headers map
	respHeaders := make(map[string]string)
	for k, v := range httpResp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	// Build response
	result := &models.ExecuteWorkflowResponse{
		Request: models.ExecuteWorkflowRequestInfo{
			Method:  workflow.HTTPMethod,
			URL:     workflow.BaseURL,
			Headers: headers,
			Body:    requestBody,
		},
		Response: models.ExecuteWorkflowResponseInfo{
			Status:     httpResp.StatusCode,
			StatusText: http.StatusText(httpResp.StatusCode),
			Headers:    respHeaders,
			Body:       respBodyJSON,
		},
	}

	return result, nil
}

func main() {
	lambda.Start(handler)
}
