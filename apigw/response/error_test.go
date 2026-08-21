package response

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestNewError_StringCode(t *testing.T) {
	tests := []struct {
		name   string
		status int
		code   string
		msg    string
		fields []ErrorField
		want   events.APIGatewayProxyResponse
	}{
		{
			name:   "no fields, returns json response without fields key",
			status: 401,
			code:   "unauthorized",
			msg:    "bad request",
			want: events.APIGatewayProxyResponse{
				StatusCode: 401,
				Body:       `{"code":"unauthorized","message":"bad request"}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
		{
			name:   "with fields, returns json response with fields key",
			status: 400,
			code:   "validation_failed",
			msg:    "validation failed",
			fields: []ErrorField{
				NewErrorField("email", FieldErrorCodeInvalidFormat, "must be a valid email"),
				NewErrorField("age", "out_of_range", "must be between 0 and 120"),
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body:       `{"code":"validation_failed","message":"validation failed","fields":[{"code":"invalid_format","field":"email","message":"must be a valid email"},{"code":"out_of_range","field":"age","message":"must be between 0 and 120"}]}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
		{
			name:   "empty message, still serialises message key",
			status: 500,
			code:   "internal_error",
			msg:    "",
			want: events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       `{"code":"internal_error","message":""}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewError(tt.status, tt.code, tt.msg, tt.fields...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewError_IntCode(t *testing.T) {
	t.Run("int code differs from status", func(t *testing.T) {
		got := NewError(400, 1001, "domain error", NewErrorField("id", FieldErrorCodeRequired, "id is required"))
		want := events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"code":1001,"message":"domain error","fields":[{"code":"required","field":"id","message":"id is required"}]}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		assert.Equal(t, want, got)
	})
}

func TestNewErrorField(t *testing.T) {
	got := NewErrorField("email", FieldErrorCodeInvalidFormat, "must be a valid email")
	assert.Equal(t, ErrorField{
		Code:    "invalid_format",
		Field:   "email",
		Message: "must be a valid email",
	}, got)
}

func TestErrorCode_String(t *testing.T) {
	assert.Equal(t, "validation_failed", ErrorCodeValidationFailed.String())
	assert.Equal(t, "internal_error", ErrorCodeInternalError.String())
}

func TestNewErrorCode(t *testing.T) {
	t.Run("uses the registered status, code and message", func(t *testing.T) {
		got := NewErrorCode(ErrorCodeUnauthorized)
		want := events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       `{"code":"unauthorized","message":"Missing or invalid bearer token."}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		assert.Equal(t, want, got)
	})

	t.Run("WithErrorMessage overrides the default message", func(t *testing.T) {
		got := NewErrorCode(ErrorCodeCustomerNotFound, WithErrorMessage("No customer with id `abc123` was found."))
		want := events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"code":"customer_not_found","message":"No customer with id ` + "`abc123`" + ` was found."}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		assert.Equal(t, want, got)
	})

	t.Run("WithErrorStatus overrides the default status", func(t *testing.T) {
		got := NewErrorCode(ErrorCodeValidationFailed, WithErrorStatus(422))
		want := events.APIGatewayProxyResponse{
			StatusCode: 422,
			Body:       `{"code":"validation_failed","message":"One or more fields in the request body were invalid."}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		assert.Equal(t, want, got)
	})

	t.Run("WithErrorFields attaches field-level validation details", func(t *testing.T) {
		got := NewErrorCode(ErrorCodeValidationFailed,
			WithErrorMessage("One or more query parameters were invalid."),
			WithErrorFields(NewErrorField("nearLocation", FieldErrorCodeInvalidFormat, "Must be either a place name or a `lat,lng` coordinate pair.")),
		)
		want := events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"code":"validation_failed","message":"One or more query parameters were invalid.","fields":[{"code":"invalid_format","field":"nearLocation","message":"Must be either a place name or a ` + "`lat,lng`" + ` coordinate pair."}]}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		assert.Equal(t, want, got)
	})

	t.Run("options apply independently of each other", func(t *testing.T) {
		got1 := NewErrorCode(ErrorCodeOfferNotFound)
		got2 := NewErrorCode(ErrorCodeOfferNotFound, WithErrorMessage("custom message"))
		assert.NotEqual(t, got1.Body, got2.Body, "the first call's definition must not be mutated by the second call's options")
	})

	t.Run("unregistered ErrorCode panics", func(t *testing.T) {
		assert.Panics(t, func() {
			NewErrorCode(errorCodeCount)
		})
	})
}

// TestErrorCodeDefinitions_registerEveryErrorCode guards against a new ErrorCode constant being
// added without a corresponding errorCodeDefinitions entry, which NewErrorCode would only catch
// at call time via its panic.
func TestErrorCodeDefinitions_registerEveryErrorCode(t *testing.T) {
	for code := ErrorCode(0); code < errorCodeCount; code++ {
		_, ok := errorCodeDefinitions[code]
		assert.True(t, ok, "ErrorCode(%d) has no registered definition", code)
	}
	assert.Len(t, errorCodeDefinitions, int(errorCodeCount), "errorCodeDefinitions has an entry not backed by an ErrorCode constant")
}
