package response

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// New creates a new plain text response for API Gateway
func New(status int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    map[string]string{},
		Body:       body,
	}
}

// NewJson creates a new JSON response for API Gateway. Body will be converted into JSON using json.Marshal(). If there
// is an error marshalling the body, the response will be left blank.
//
// Deprecated: use NewJSON instead. NewJson violates Go naming convention (initialisms should
// be all-caps) - kept as a thin wrapper for existing consumers, tracked as tech debt (TD-001).
//
//nolint:revive // deprecated alias, see comment above
func NewJson(status int, body any) events.APIGatewayProxyResponse {
	return NewJSON(status, body)
}

// NewJSON creates a new JSON response for API Gateway. Body will be converted into JSON using json.Marshal(). If
// there is an error marshalling the body, the response will be left blank.
func NewJSON(status int, body any) events.APIGatewayProxyResponse {
	res := events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    map[string]string{},
	}

	if body != nil {
		if j, err := json.Marshal(body); err == nil {
			res.Body = string(j)
			res.Headers["Content-Type"] = "application/json"
		}
	}

	return res
}
