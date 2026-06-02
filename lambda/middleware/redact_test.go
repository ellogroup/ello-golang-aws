package middleware

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedactHTTPEvent(t *testing.T) {
	type testCase struct {
		name  string
		event any
		want  any
	}
	tests := []testCase{
		{
			name: "APIGatewayProxyRequest, sensitive headers redacted, non-sensitive preserved",
			event: events.APIGatewayProxyRequest{
				Path: "/test",
				Headers: map[string]string{
					"Authorization": "Bearer secret",
					"Cookie":        "session=abc",
					"X-Api-Key":     "key123",
					"Content-Type":  "application/json",
				},
				MultiValueHeaders: map[string][]string{
					"Authorization": {"Bearer secret"},
					"Cookie":        {"session=abc", "other=xyz"},
					"Accept":        {"application/json"},
				},
			},
			want: events.APIGatewayProxyRequest{
				Path: "/test",
				Headers: map[string]string{
					"Authorization": redactedValue,
					"Cookie":        redactedValue,
					"X-Api-Key":     redactedValue,
					"Content-Type":  "application/json",
				},
				MultiValueHeaders: map[string][]string{
					"Authorization": {redactedValue},
					"Cookie":        {redactedValue},
					"Accept":        {"application/json"},
				},
			},
		},
		{
			name: "APIGatewayProxyRequest, case-insensitive header matching",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "Bearer secret",
					"COOKIE":        "session=abc",
					"x-api-key":     "key123",
				},
			},
			want: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": redactedValue,
					"COOKIE":        redactedValue,
					"x-api-key":     redactedValue,
				},
			},
		},
		{
			name:  "APIGatewayProxyRequest, nil headers returned as nil",
			event: events.APIGatewayProxyRequest{Path: "/test"},
			want:  events.APIGatewayProxyRequest{Path: "/test"},
		},
		{
			name: "APIGatewayV2HTTPRequest, sensitive headers and cookies redacted",
			event: events.APIGatewayV2HTTPRequest{
				RawPath: "/test",
				Headers: map[string]string{
					"Authorization": "Bearer secret",
					"Content-Type":  "application/json",
				},
				Cookies: []string{"session=abc", "other=xyz"},
			},
			want: events.APIGatewayV2HTTPRequest{
				RawPath: "/test",
				Headers: map[string]string{
					"Authorization": redactedValue,
					"Content-Type":  "application/json",
				},
				Cookies: []string{redactedValue},
			},
		},
		{
			name: "APIGatewayV2HTTPRequest, empty cookies not modified",
			event: events.APIGatewayV2HTTPRequest{
				Headers: map[string]string{"Authorization": "Bearer secret"},
			},
			want: events.APIGatewayV2HTTPRequest{
				Headers: map[string]string{"Authorization": redactedValue},
			},
		},
		{
			name: "ALBTargetGroupRequest, sensitive headers redacted",
			event: events.ALBTargetGroupRequest{
				Path: "/test",
				Headers: map[string]string{
					"Authorization": "Bearer secret",
					"Content-Type":  "application/json",
				},
				MultiValueHeaders: map[string][]string{
					"Cookie": {"session=abc"},
					"Accept": {"application/json"},
				},
			},
			want: events.ALBTargetGroupRequest{
				Path: "/test",
				Headers: map[string]string{
					"Authorization": redactedValue,
					"Content-Type":  "application/json",
				},
				MultiValueHeaders: map[string][]string{
					"Cookie": {redactedValue},
					"Accept": {"application/json"},
				},
			},
		},
		{
			name: "LambdaFunctionURLRequest, sensitive headers and cookies redacted",
			event: events.LambdaFunctionURLRequest{
				RawPath: "/test",
				Headers: map[string]string{
					"Authorization": "Bearer secret",
					"Content-Type":  "application/json",
				},
				Cookies: []string{"session=abc"},
			},
			want: events.LambdaFunctionURLRequest{
				RawPath: "/test",
				Headers: map[string]string{
					"Authorization": redactedValue,
					"Content-Type":  "application/json",
				},
				Cookies: []string{redactedValue},
			},
		},
		{
			name: "APIGatewayWebsocketProxyRequest, sensitive headers redacted",
			event: events.APIGatewayWebsocketProxyRequest{
				Headers: map[string]string{
					"Authorization": "Bearer secret",
					"Content-Type":  "application/json",
				},
				MultiValueHeaders: map[string][]string{
					"Cookie": {"session=abc"},
				},
			},
			want: events.APIGatewayWebsocketProxyRequest{
				Headers: map[string]string{
					"Authorization": redactedValue,
					"Content-Type":  "application/json",
				},
				MultiValueHeaders: map[string][]string{
					"Cookie": {redactedValue},
				},
			},
		},
		{
			name:  "non-HTTP event type passed through unchanged",
			event: events.SQSEvent{Records: []events.SQSMessage{{MessageId: "123"}}},
			want:  events.SQSEvent{Records: []events.SQSMessage{{MessageId: "123"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, RedactHTTPEvent(tt.event), "RedactHTTPEvent(%v)", tt.event)
		})
	}
}

func TestRedactHTTPEvent_doesNotMutateOriginal(t *testing.T) {
	original := events.APIGatewayProxyRequest{
		Headers: map[string]string{"Authorization": "Bearer secret"},
	}

	RedactHTTPEvent(original)

	assert.Equal(t, "Bearer secret", original.Headers["Authorization"])
}
