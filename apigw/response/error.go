package response

import "github.com/aws/aws-lambda-go/events"

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewError creates a new error response for API Gateway
func NewError(status int, msg string) events.APIGatewayProxyResponse {
	return NewJson(status, responseError{
		Code:    status,
		Message: msg,
	})
}
