package response

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"golang.org/x/exp/constraints"
)

// ErrorCode identifies a known error returned by our APIs. Each ErrorCode has an associated HTTP
// status, wire code, and default message - see NewErrorCode.
type ErrorCode int

const (
	ErrorCodeValidationFailed ErrorCode = iota
	ErrorCodeCustomerAlreadyExists
	ErrorCodeCustomerNotFound
	ErrorCodeOfferNotFound
	ErrorCodeOfferNotRedeemable
	ErrorCodeRedemptionNotFound
	ErrorCodeUnauthorized
	ErrorCodeRateLimited
	ErrorCodeInternalError

	// errorCodeCount is one past the last ErrorCode constant - used to check every constant above
	// has a registered definition (see error_test.go) and as a guaranteed-unregistered value.
	errorCodeCount
)

type errorCodeDefinition struct {
	status  int
	code    string
	message string
	fields  []ErrorField
}

var errorCodeDefinitions = map[ErrorCode]errorCodeDefinition{
	ErrorCodeValidationFailed: {
		status:  http.StatusBadRequest,
		code:    "validation_failed",
		message: "One or more fields in the request body were invalid.",
	},
	ErrorCodeCustomerAlreadyExists: {
		status:  http.StatusConflict,
		code:    "customer_already_exists",
		message: "A customer with this externalCustomerRef already exists.",
	},
	ErrorCodeCustomerNotFound: {
		status:  http.StatusNotFound,
		code:    "customer_not_found",
		message: "The customer was not found.",
	},
	ErrorCodeOfferNotFound: {
		status:  http.StatusNotFound,
		code:    "offer_not_found",
		message: "The offer was not found.",
	},
	ErrorCodeOfferNotRedeemable: {
		status:  http.StatusConflict,
		code:    "offer_not_redeemable",
		message: "This offer is no longer redeemable.",
	},
	ErrorCodeRedemptionNotFound: {
		status:  http.StatusNotFound,
		code:    "redemption_not_found",
		message: "The redemption was not found.",
	},
	ErrorCodeUnauthorized: {
		status:  http.StatusUnauthorized,
		code:    "unauthorized",
		message: "Missing or invalid bearer token.",
	},
	ErrorCodeRateLimited: {
		status:  http.StatusTooManyRequests,
		code:    "rate_limited",
		message: "Too many requests. Retry after the period in the Retry-After header.",
	},
	ErrorCodeInternalError: {
		status:  http.StatusInternalServerError,
		code:    "internal_error",
		message: "An unexpected error occurred. Please retry.",
	},
}

// String returns the wire code (e.g. "validation_failed") for c.
func (c ErrorCode) String() string {
	return errorCodeDefinitions[c].code
}

// ErrorCodeOption overrides part of the definition NewErrorCode builds a response from.
type ErrorCodeOption func(*errorCodeDefinition)

// WithErrorStatus overrides the HTTP status NewErrorCode would otherwise use for code.
func WithErrorStatus(status int) ErrorCodeOption {
	return func(d *errorCodeDefinition) { d.status = status }
}

// WithErrorMessage overrides the message NewErrorCode would otherwise use for code. Needed
// whenever the default message can't carry details only the caller has, such as an id or field
// name (e.g. customer_not_found's default is generic; a handler that knows the id should override
// it with the specific message).
func WithErrorMessage(message string) ErrorCodeOption {
	return func(d *errorCodeDefinition) { d.message = message }
}

// WithErrorFields attaches field-level validation details to the response NewErrorCode builds.
func WithErrorFields(fields ...ErrorField) ErrorCodeOption {
	return func(d *errorCodeDefinition) { d.fields = fields }
}

// NewErrorCode creates a new error response for API Gateway using code's predefined HTTP status,
// wire code, and message, so every caller reporting the same error produces the same response.
// Use the With* options to override any of them.
//
// NewErrorCode panics if code has no registered definition. Every ErrorCode constant in this
// package is registered, so this can only happen if a new ErrorCode is added here without one.
func NewErrorCode(code ErrorCode, opts ...ErrorCodeOption) events.APIGatewayProxyResponse {
	def, ok := errorCodeDefinitions[code]
	if !ok {
		panic(fmt.Sprintf("response: no definition registered for ErrorCode(%d)", code))
	}
	for _, opt := range opts {
		opt(&def)
	}
	return NewError(def.status, def.code, def.message, def.fields...)
}

// Field-level codes used across our APIs. Use these instead of inline string literals so every
// service produces the exact same values.
const (
	FieldErrorCodeRequired       = "required"
	FieldErrorCodeNotUnique      = "not_unique"
	FieldErrorCodeInvalidFormat  = "invalid_format"
	FieldErrorCodeOfferWithdrawn = "offer_withdrawn"
)

type Error[T string | constraints.Integer] struct {
	Code    T            `json:"code"`
	Message string       `json:"message"`
	Fields  []ErrorField `json:"fields,omitempty"`
}

// NewError creates a new error response for API Gateway. code is distinct from the HTTP status and
// may be a string or any integer type. Prefer NewErrorCode for one of our known ErrorCode values -
// use NewError directly only for errors outside that set.
func NewError[T string | constraints.Integer](status int, code T, msg string, fields ...ErrorField) events.APIGatewayProxyResponse {
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
