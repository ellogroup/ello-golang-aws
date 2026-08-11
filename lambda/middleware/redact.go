package middleware

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

const redactedValue = "[REDACTED]"

var sensitiveHeaders = map[string]struct{}{
	"authorization": {},
	"cookie":        {},
	"x-api-key":     {},
}

type redactOptions struct {
	includeBody bool
}

// RedactOption configures RedactHTTPEvent.
type RedactOption func(*redactOptions)

// WithBodyIncluded preserves the event's Body field instead of redacting it. Use this only for
// routes that are genuinely public and bodyless-safe to log in full - request/response bodies
// routinely carry customer PII, so the default is to redact them.
func WithBodyIncluded() RedactOption {
	return func(o *redactOptions) {
		o.includeBody = true
	}
}

// RedactHTTPEvent returns a sanitized copy of known HTTP Lambda event types with sensitive headers
// (Authorization, Cookie, X-Api-Key) and the request Body replaced with [REDACTED]. Non-HTTP event
// types are returned unchanged. Pass WithBodyIncluded() to preserve the body for a route that is
// known not to carry sensitive data.
// It can be called inside a WithEventLoggerSanitizer function to compose built-in redaction with custom logic.
func RedactHTTPEvent(event any, opts ...RedactOption) any {
	o := redactOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	switch e := event.(type) {
	case events.APIGatewayProxyRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		e.Body = redactBody(e.Body, o)
		return e
	case events.APIGatewayV2HTTPRequest:
		e.Headers = redactHeaders(e.Headers)
		if len(e.Cookies) > 0 {
			e.Cookies = []string{redactedValue}
		}
		e.Body = redactBody(e.Body, o)
		return e
	case events.ALBTargetGroupRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		e.Body = redactBody(e.Body, o)
		return e
	case events.LambdaFunctionURLRequest:
		e.Headers = redactHeaders(e.Headers)
		if len(e.Cookies) > 0 {
			e.Cookies = []string{redactedValue}
		}
		e.Body = redactBody(e.Body, o)
		return e
	case events.APIGatewayWebsocketProxyRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		e.Body = redactBody(e.Body, o)
		return e
	}
	return event
}

func redactBody(body string, o redactOptions) string {
	if body == "" || o.includeBody {
		return body
	}
	return redactedValue
}

func redactHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		if _, sensitive := sensitiveHeaders[strings.ToLower(k)]; sensitive {
			out[k] = redactedValue
		} else {
			out[k] = v
		}
	}
	return out
}

func redactMultiValueHeaders(headers map[string][]string) map[string][]string {
	if headers == nil {
		return nil
	}
	out := make(map[string][]string, len(headers))
	for k, v := range headers {
		if _, sensitive := sensitiveHeaders[strings.ToLower(k)]; sensitive {
			out[k] = []string{redactedValue}
		} else {
			out[k] = v
		}
	}
	return out
}
