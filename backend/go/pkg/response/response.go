package response

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
)

// Success creates a success response
func Success(data interface{}) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"data":    data,
	})

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}
}

// Error creates an error response
func Error(statusCode int, message string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"success": false,
		"error":   message,
	})

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}
}

// BadRequest creates a 400 error response
func BadRequest(message string) events.APIGatewayProxyResponse {
	return Error(400, message)
}

// Unauthorized creates a 401 error response
func Unauthorized(message string) events.APIGatewayProxyResponse {
	return Error(401, message)
}

// Forbidden creates a 403 error response
func Forbidden(message string) events.APIGatewayProxyResponse {
	return Error(403, message)
}

// NotFound creates a 404 error response
func NotFound(message string) events.APIGatewayProxyResponse {
	return Error(404, message)
}

// InternalError creates a 500 error response
func InternalError(message string) events.APIGatewayProxyResponse {
	return Error(500, message)
}
