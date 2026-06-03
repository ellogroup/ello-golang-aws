package middleware

import (
	"github.com/aws/aws-lambda-go/events"
	"strings"
)

const redactedValue = "[REDACTED]"

var sensitiveHeaders = map[string]struct{}{
	"authorization": {},
	"cookie":        {},
	"x-api-key":     {},
}

// RedactHTTPEvent returns a sanitized copy of known HTTP Lambda event types with sensitive headers
// (Authorization, Cookie, X-Api-Key) replaced with [REDACTED]. Non-HTTP event types are returned unchanged.
// It can be called inside a WithEventLoggerSanitizer function to compose built-in redaction with custom logic.
func RedactHTTPEvent(event any) any {
	switch e := event.(type) {
	case events.APIGatewayProxyRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		return e
	case events.APIGatewayV2HTTPRequest:
		e.Headers = redactHeaders(e.Headers)
		if len(e.Cookies) > 0 {
			e.Cookies = []string{redactedValue}
		}
		return e
	case events.ALBTargetGroupRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		return e
	case events.LambdaFunctionURLRequest:
		e.Headers = redactHeaders(e.Headers)
		if len(e.Cookies) > 0 {
			e.Cookies = []string{redactedValue}
		}
		return e
	case events.APIGatewayWebsocketProxyRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		return e
	}
	return event
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
