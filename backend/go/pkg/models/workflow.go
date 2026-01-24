package models

import (
	"encoding/json"
	"time"
)

// Workflow represents a workflow configuration
type Workflow struct {
	WorkflowID         string          `json:"workflow_id"`
	WorkflowName       string          `json:"workflow_name"`
	Description        string          `json:"description"`
	Source             string          `json:"source"`              // coze, n8n
	TemplateName       string          `json:"template_name"`       // workflow, streamflow
	HTTPMethod         string          `json:"http_method"`         // GET, POST, PUT
	BaseURL            string          `json:"base_url"`
	BearerToken        string          `json:"bearer_token"`
	ExternalWorkflowID string          `json:"external_workflow_id"`
	Parameters         json.RawMessage `json:"parameters"`          // JSON object
	Headers            json.RawMessage `json:"headers"`             // JSON object
	ProjectID          string          `json:"project_id"`
	CreatorDID         string          `json:"creator_did"`
	IsShared           bool            `json:"is_shared"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// CreateWorkflowRequest represents the request to create a workflow
type CreateWorkflowRequest struct {
	WorkflowName       string          `json:"workflow_name"`
	Description        string          `json:"description"`
	Source             string          `json:"source"`
	TemplateName       string          `json:"template_name"`
	HTTPMethod         string          `json:"http_method"`
	BaseURL            string          `json:"base_url"`
	BearerToken        string          `json:"bearer_token"`
	ExternalWorkflowID string          `json:"external_workflow_id"`
	Parameters         json.RawMessage `json:"parameters"`
	Headers            json.RawMessage `json:"headers"`
	ProjectID          string          `json:"project_id"`
}

// UpdateWorkflowRequest represents the request to update a workflow
type UpdateWorkflowRequest struct {
	WorkflowName       *string          `json:"workflow_name,omitempty"`
	Description        *string          `json:"description,omitempty"`
	Source             *string          `json:"source,omitempty"`
	TemplateName       *string          `json:"template_name,omitempty"`
	HTTPMethod         *string          `json:"http_method,omitempty"`
	BaseURL            *string          `json:"base_url,omitempty"`
	BearerToken        *string          `json:"bearer_token,omitempty"`
	ExternalWorkflowID *string          `json:"external_workflow_id,omitempty"`
	Parameters         *json.RawMessage `json:"parameters,omitempty"`
	Headers            *json.RawMessage `json:"headers,omitempty"`
}

// ExecuteWorkflowRequest represents the request to execute a workflow
type ExecuteWorkflowRequest struct {
	Parameters json.RawMessage `json:"parameters"`
	Headers    json.RawMessage `json:"headers"`
}

// ExecuteWorkflowResponse represents the response from executing a workflow
type ExecuteWorkflowResponse struct {
	Request  ExecuteWorkflowRequestInfo `json:"request"`
	Response ExecuteWorkflowResponseInfo `json:"response"`
}

// ExecuteWorkflowRequestInfo represents the HTTP request information
type ExecuteWorkflowRequestInfo struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

// ExecuteWorkflowResponseInfo represents the HTTP response information
type ExecuteWorkflowResponseInfo struct {
	Status     int                    `json:"status"`
	StatusText string                 `json:"status_text"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
}

// ShareWorkflowRequest represents the request to share/unshare a workflow
type ShareWorkflowRequest struct {
	IsShared bool `json:"is_shared"`
}

// HideWorkflowRequest represents the request to hide/unhide a workflow
type HideWorkflowRequest struct {
	IsHidden bool `json:"is_hidden"`
}

// SearchWorkflowRequest represents the request to search workflows
type SearchWorkflowRequest struct {
	Query     string `json:"query"`
	ProjectID string `json:"project_id"`
	TopK      int    `json:"top_k"`
	Threshold float64 `json:"threshold"`
}

// SearchWorkflowResult represents a search result
type SearchWorkflowResult struct {
	WorkflowID   string  `json:"workflow_id"`
	WorkflowName string  `json:"workflow_name"`
	Description  string  `json:"description"`
	Source       string  `json:"source"`
	TemplateName string  `json:"template_name"`
	Similarity   float64 `json:"similarity,omitempty"`
	Relevance    float64 `json:"relevance,omitempty"`
}

// SearchWorkflowResponse represents the search response
type SearchWorkflowResponse struct {
	Results      []SearchWorkflowResult `json:"results"`
	SearchMethod string                 `json:"search_method"` // "vector" or "fulltext"
}
