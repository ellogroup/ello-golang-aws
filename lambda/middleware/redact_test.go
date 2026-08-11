package middleware

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestRedactHTTPEvent(t *testing.T) {
	type testCase struct {
		name  string
		event any
		want  any
	}
	tests := []testCase{
		{
			name: "APIGatewayProxyRequest, sensitive headers and body redacted, non-sensitive headers preserved",
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
				Body: `{"email":"jane@example.com"}`,
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
				Body: redactedValue,
			},
		},
		{
			name:  "APIGatewayProxyRequest, empty body left empty",
			event: events.APIGatewayProxyRequest{Path: "/test"},
			want:  events.APIGatewayProxyRequest{Path: "/test"},
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
			name: "APIGatewayV2HTTPRequest, sensitive headers, cookies, and body redacted",
			event: events.APIGatewayV2HTTPRequest{
				RawPath: "/test",
				Headers: map[string]string{
					"Authorization": "Bearer secret",
					"Content-Type":  "application/json",
				},
				Cookies: []string{"session=abc", "other=xyz"},
				Body:    `{"email":"jane@example.com"}`,
			},
			want: events.APIGatewayV2HTTPRequest{
				RawPath: "/test",
				Headers: map[string]string{
					"Authorization": redactedValue,
					"Content-Type":  "application/json",
				},
				Cookies: []string{redactedValue},
				Body:    redactedValue,
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
			name: "ALBTargetGroupRequest, sensitive headers and body redacted",
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
				Body: `{"email":"jane@example.com"}`,
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
				Body: redactedValue,
			},
		},
		{
			name: "LambdaFunctionURLRequest, sensitive headers, cookies, and body redacted",
			event: events.LambdaFunctionURLRequest{
				RawPath: "/test",
				Headers: map[string]string{
					"Authorization": "Bearer secret",
					"Content-Type":  "application/json",
				},
				Cookies: []string{"session=abc"},
				Body:    `{"email":"jane@example.com"}`,
			},
			want: events.LambdaFunctionURLRequest{
				RawPath: "/test",
				Headers: map[string]string{
					"Authorization": redactedValue,
					"Content-Type":  "application/json",
				},
				Cookies: []string{redactedValue},
				Body:    redactedValue,
			},
		},
		{
			name: "APIGatewayWebsocketProxyRequest, sensitive headers and body redacted",
			event: events.APIGatewayWebsocketProxyRequest{
				Headers: map[string]string{
					"Authorization": "Bearer secret",
					"Content-Type":  "application/json",
				},
				MultiValueHeaders: map[string][]string{
					"Cookie": {"session=abc"},
				},
				Body: `{"email":"jane@example.com"}`,
			},
			want: events.APIGatewayWebsocketProxyRequest{
				Headers: map[string]string{
					"Authorization": redactedValue,
					"Content-Type":  "application/json",
				},
				MultiValueHeaders: map[string][]string{
					"Cookie": {redactedValue},
				},
				Body: redactedValue,
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

func TestNewRedactor_Redact(t *testing.T) {
	event := events.APIGatewayProxyRequest{
		Headers: map[string]string{"Authorization": "Bearer secret"},
		Body:    `{"status":"ok"}`,
	}

	t.Run("default options match RedactHTTPEvent", func(t *testing.T) {
		assert.Equal(t, RedactHTTPEvent(event), NewRedactor().Redact(event))
	})

	t.Run("WithBodyNotRedacted preserves body", func(t *testing.T) {
		want := events.APIGatewayProxyRequest{
			Headers: map[string]string{"Authorization": redactedValue},
			Body:    `{"status":"ok"}`,
		}
		assert.Equal(t, want, NewRedactor(WithBodyNotRedacted()).Redact(event))
	})

	t.Run("options are applied once at construction, not per Redact call", func(t *testing.T) {
		redactor := NewRedactor(WithBodyNotRedacted())
		for range 3 {
			got, ok := redactor.Redact(event).(events.APIGatewayProxyRequest)
			assert.True(t, ok)
			assert.Equal(t, `{"status":"ok"}`, got.Body)
		}
	})
}
