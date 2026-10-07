package middleware

import (
	"context"
	"mime"
	"slices"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type contentType struct {
	allowed  []string
	rejected events.APIGatewayProxyResponse
}

// NewContentType returns API Gateway v1 middleware that only lets a request through when its Content-Type header's
// media type is one of allowed (e.g. "application/json"). Any other request, including one with no Content-Type, gets
// rejected instead, so each application keeps its own error format (typically a 415).
//
// Media types are compared case-insensitively and parameters are ignored, so "application/json; charset=utf-8"
// matches "application/json". The Content-Type header name is matched case-insensitively.
func NewContentType(rejected events.APIGatewayProxyResponse, allowed ...string) WithResponse[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse] {
	a := make([]string, 0, len(allowed))
	for _, t := range allowed {
		a = append(a, strings.ToLower(strings.TrimSpace(t)))
	}
	return &contentType{allowed: a, rejected: rejected}
}

func (m contentType) Wrap(next func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if !m.isAllowed(headerValue(event, "Content-Type")) {
			return m.rejected, nil
		}
		return next(ctx, event)
	}
}

func (m contentType) isAllowed(value string) bool {
	if value == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return false
	}
	return slices.Contains(m.allowed, mediaType)
}

// headerValue returns the first value of the named header, matching its name case-insensitively. API Gateway passes
// headers with the casing the client sent.
func headerValue(event events.APIGatewayProxyRequest, name string) string {
	for k, v := range event.Headers {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	for k, v := range event.MultiValueHeaders {
		if strings.EqualFold(k, name) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}
