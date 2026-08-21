package response

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/aws/aws-lambda-go/events"
)

// ErrorCode identifies an error returned by an API - either one of ours (see the ErrorCode*
// constants below) or a custom one an application registers for itself with RegisterErrorCode.
// The empty ErrorCode is reserved as "unset" and is never a valid registered code.
type ErrorCode string

const (
	ErrorCodeValidationFailed ErrorCode = "validation_failed"
	ErrorCodeUnauthorized     ErrorCode = "unauthorized"
	ErrorCodeRateLimited      ErrorCode = "rate_limited"
	ErrorCodeInternalError    ErrorCode = "internal_error"
)

func init() {
	MustRegisterErrorCode(ErrorCodeValidationFailed, ErrorCodeDefinition{
		Status:  http.StatusBadRequest,
		Message: "One or more fields in the request body were invalid.",
	})
	MustRegisterErrorCode(ErrorCodeUnauthorized, ErrorCodeDefinition{
		Status:  http.StatusUnauthorized,
		Message: "Missing or invalid bearer token.",
	})
	MustRegisterErrorCode(ErrorCodeRateLimited, ErrorCodeDefinition{
		Status:  http.StatusTooManyRequests,
		Message: "Too many requests. Retry after the period in the Retry-After header.",
	})
	MustRegisterErrorCode(ErrorCodeInternalError, ErrorCodeDefinition{
		Status:  http.StatusInternalServerError,
		Message: "An unexpected error occurred. Please retry.",
	})
}

// ErrorCodeDefinition bundles the HTTP status, message, and default field-level details NewErrorCode
// uses to build a response for a given ErrorCode - see RegisterErrorCode.
type ErrorCodeDefinition struct {
	Status  int
	Message string
	Fields  []ErrorField
}

var (
	errorCodeRegistryMu sync.RWMutex
	errorCodeRegistry   = map[ErrorCode]ErrorCodeDefinition{}
)

// RegisterErrorCode registers def as the definition NewErrorCode uses for code. Call it once at
// application startup (e.g. from an init function), before request traffic begins - not on every
// request.
//
// Registering the same code with an identical def more than once is a no-op, not an error, so
// startup code that might run more than once in a process (test setup, repeated init) stays safe.
// Registering code with a def that differs from what's already registered - whether one of ours or
// one an application registered earlier - returns an error: that is a genuine naming collision
// between two unrelated definitions and must be rejected rather than silently letting the later
// registration win. Registering the empty ErrorCode also returns an error.
func RegisterErrorCode(code ErrorCode, def ErrorCodeDefinition) error {
	if code == "" {
		return fmt.Errorf("response: cannot register the empty ErrorCode")
	}

	errorCodeRegistryMu.Lock()
	defer errorCodeRegistryMu.Unlock()

	if existing, ok := errorCodeRegistry[code]; ok {
		if fieldsEqual(existing.Fields, def.Fields) && existing.Status == def.Status && existing.Message == def.Message {
			return nil
		}
		return fmt.Errorf("response: ErrorCode %q is already registered with a different definition", code)
	}
	errorCodeRegistry[code] = def
	return nil
}

// MustRegisterErrorCode is RegisterErrorCode but panics instead of returning an error. Intended for
// startup-time registration (e.g. from an init function, which can't return an error) where a
// registration conflict is a programmer error that should fail fast.
func MustRegisterErrorCode(code ErrorCode, def ErrorCodeDefinition) {
	if err := RegisterErrorCode(code, def); err != nil {
		panic(err)
	}
}

func fieldsEqual(a, b []ErrorField) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// String returns the wire code (e.g. "validation_failed") for c.
func (c ErrorCode) String() string {
	return string(c)
}

// ErrorCodeOption overrides part of the definition NewErrorCode builds a response from.
type ErrorCodeOption func(*ErrorCodeDefinition)

// WithErrorStatus overrides the HTTP status NewErrorCode would otherwise use for code.
func WithErrorStatus(status int) ErrorCodeOption {
	return func(d *ErrorCodeDefinition) { d.Status = status }
}

// WithErrorMessage overrides the message NewErrorCode would otherwise use for code. Needed
// whenever the default message can't carry details only the caller has, such as an id or field
// name a custom, application-registered ErrorCode's default message is necessarily generic about.
func WithErrorMessage(message string) ErrorCodeOption {
	return func(d *ErrorCodeDefinition) { d.Message = message }
}

// WithErrorFields attaches field-level validation details to the response NewErrorCode builds.
func WithErrorFields(fields ...ErrorField) ErrorCodeOption {
	return func(d *ErrorCodeDefinition) { d.Fields = fields }
}

// NewErrorCode creates a new error response for API Gateway using code's registered HTTP status
// and message (see RegisterErrorCode), so every caller reporting the same error produces the same
// response. Use the With* options to override any of them.
//
// NewErrorCode panics if code has no registered definition - register it first with
// RegisterErrorCode/MustRegisterErrorCode.
func NewErrorCode(code ErrorCode, opts ...ErrorCodeOption) events.APIGatewayProxyResponse {
	errorCodeRegistryMu.RLock()
	def, ok := errorCodeRegistry[code]
	errorCodeRegistryMu.RUnlock()
	if !ok {
		panic(fmt.Sprintf("response: no definition registered for ErrorCode %q - register it first with RegisterErrorCode", code))
	}
	for _, opt := range opts {
		opt(&def)
	}
	return NewError(def.Status, code, def.Message, def.Fields...)
}

// Field-level codes used across our APIs. Use these instead of inline string literals so every
// service produces the exact same values.
const (
	FieldErrorCodeRequired      = "required"
	FieldErrorCodeNotUnique     = "not_unique"
	FieldErrorCodeInvalidFormat = "invalid_format"
)

// errorCodeConstraint is satisfied by any string or integer type, named or not, so a caller can
// pass an ErrorCode - or a plain string/int - directly as an Error's code.
type errorCodeConstraint interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Error[T errorCodeConstraint] struct {
	Code    T            `json:"code"`
	Message string       `json:"message"`
	Fields  []ErrorField `json:"fields,omitempty"`
}

// NewError creates a new error response for API Gateway. code is distinct from the HTTP status and
// may be a string or any integer type (or ErrorCode). Prefer NewErrorCode for a registered
// ErrorCode - use NewError directly only for errors outside that set.
func NewError[T errorCodeConstraint](status int, code T, msg string, fields ...ErrorField) events.APIGatewayProxyResponse {
	return NewJSON(status, Error[T]{
		Code:    code,
		Message: msg,
		Fields:  fields,
	})
}

type ErrorField struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewErrorField(field, code, message string) ErrorField {
	return ErrorField{
		Code:    code,
		Field:   field,
		Message: message,
	}
}
