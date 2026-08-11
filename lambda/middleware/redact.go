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
	bodyNotRedacted bool
}

// RedactOption configures a Redactor.
type RedactOption func(*redactOptions)

// WithBodyNotRedacted leaves the event's Body field untouched instead of redacting it. Use this
// only for routes that are genuinely public and bodyless-safe to log in full - request/response
// bodies routinely carry customer PII, so the default is to redact them.
func WithBodyNotRedacted() RedactOption {
	return func(o *redactOptions) {
		o.bodyNotRedacted = true
	}
}

// Redactor redacts sensitive headers and (by default) the request body from known HTTP Lambda
// event types. Construct one with NewRedactor when you need non-default options - options are
// applied once at construction, not re-processed on every call to Sanitize. Implements Sanitizer,
// so a *Redactor can be passed directly to WithEventLoggerSanitizer.
type Redactor struct {
	opts redactOptions
}

// NewRedactor constructs a Redactor with the given options applied.
func NewRedactor(opts ...RedactOption) *Redactor {
	o := redactOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	return &Redactor{opts: o}
}

// Sanitize returns a sanitized copy of known HTTP Lambda event types with sensitive headers
// (Authorization, Cookie, X-Api-Key) and (unless configured otherwise) the request Body replaced
// with [REDACTED]. Non-HTTP event types are returned unchanged.
func (r *Redactor) Sanitize(event any) any {
	switch e := event.(type) {
	case events.APIGatewayProxyRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		e.Body = redactBody(e.Body, r.opts)
		return e
	case events.APIGatewayV2HTTPRequest:
		e.Headers = redactHeaders(e.Headers)
		if len(e.Cookies) > 0 {
			e.Cookies = []string{redactedValue}
		}
		e.Body = redactBody(e.Body, r.opts)
		return e
	case events.ALBTargetGroupRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		e.Body = redactBody(e.Body, r.opts)
		return e
	case events.LambdaFunctionURLRequest:
		e.Headers = redactHeaders(e.Headers)
		if len(e.Cookies) > 0 {
			e.Cookies = []string{redactedValue}
		}
		e.Body = redactBody(e.Body, r.opts)
		return e
	case events.APIGatewayWebsocketProxyRequest:
		e.Headers = redactHeaders(e.Headers)
		e.MultiValueHeaders = redactMultiValueHeaders(e.MultiValueHeaders)
		e.Body = redactBody(e.Body, r.opts)
		return e
	}
	return event
}

// defaultRedactor applies default options only (redact headers and body). Shared by
// RedactHTTPEvent so the common case doesn't allocate a new Redactor per call.
var defaultRedactor = NewRedactor()

// RedactHTTPEvent returns a sanitized copy of known HTTP Lambda event types with sensitive
// headers (Authorization, Cookie, X-Api-Key) and the request Body replaced with [REDACTED].
// Non-HTTP event types are returned unchanged. It can be called inside a
// WithEventLoggerSanitizer function to compose built-in redaction with custom logic.
//
// This always applies default options. For non-default behaviour (e.g. WithBodyNotRedacted),
// construct a Redactor with NewRedactor instead - options are applied once at construction,
// not on every event.
func RedactHTTPEvent(event any) any {
	return defaultRedactor.Sanitize(event)
}

func redactBody(body string, o redactOptions) string {
	if body == "" || o.bodyNotRedacted {
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
