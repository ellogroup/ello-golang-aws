package response

import (
	"fmt"
	"sync"
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

func TestNewError_ErrorCodeArgument(t *testing.T) {
	t.Run("an ErrorCode value can be passed directly, without converting to string", func(t *testing.T) {
		got := NewError(401, ErrorCodeUnauthorized, "bad request")
		want := events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       `{"code":"unauthorized","message":"bad request"}`,
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
	t.Run("uses the registered status and message", func(t *testing.T) {
		got := NewErrorCode(ErrorCodeUnauthorized)
		want := events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       `{"code":"unauthorized","message":"Missing or invalid bearer token."}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		assert.Equal(t, want, got)
	})

	t.Run("WithErrorMessage overrides the default message", func(t *testing.T) {
		got := NewErrorCode(ErrorCodeValidationFailed, WithErrorMessage("One or more query parameters were invalid."))
		want := events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"code":"validation_failed","message":"One or more query parameters were invalid."}`,
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
			WithErrorFields(NewErrorField("email", FieldErrorCodeInvalidFormat, "must be a valid email")),
		)
		want := events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"code":"validation_failed","message":"One or more fields in the request body were invalid.","fields":[{"code":"invalid_format","field":"email","message":"must be a valid email"}]}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		assert.Equal(t, want, got)
	})

	t.Run("options apply independently of each other", func(t *testing.T) {
		got1 := NewErrorCode(ErrorCodeValidationFailed)
		got2 := NewErrorCode(ErrorCodeValidationFailed, WithErrorMessage("custom message"))
		assert.NotEqual(t, got1.Body, got2.Body, "the first call's definition must not be mutated by the second call's options")
	})

	t.Run("unregistered ErrorCode panics", func(t *testing.T) {
		assert.Panics(t, func() {
			NewErrorCode(ErrorCode("definitely_not_registered"))
		})
	})
}

func TestBuiltinErrorCodeDefinitions_registerEveryErrorCode(t *testing.T) {
	builtins := []ErrorCode{
		ErrorCodeValidationFailed,
		ErrorCodeUnauthorized,
		ErrorCodeRateLimited,
		ErrorCodeInternalError,
	}

	errorCodeRegistryMu.RLock()
	defer errorCodeRegistryMu.RUnlock()
	for _, code := range builtins {
		_, ok := errorCodeRegistry[code]
		assert.True(t, ok, "%s has no registered definition", code)
	}
}

func TestNewErrorCode_builtins(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want events.APIGatewayProxyResponse
	}{
		{
			code: ErrorCodeValidationFailed,
			want: events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body:       `{"code":"validation_failed","message":"One or more fields in the request body were invalid."}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
		{
			code: ErrorCodeUnauthorized,
			want: events.APIGatewayProxyResponse{
				StatusCode: 401,
				Body:       `{"code":"unauthorized","message":"Missing or invalid bearer token."}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
		{
			code: ErrorCodeRateLimited,
			want: events.APIGatewayProxyResponse{
				StatusCode: 429,
				Body:       `{"code":"rate_limited","message":"Too many requests. Retry after the period in the Retry-After header."}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
		{
			code: ErrorCodeInternalError,
			want: events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       `{"code":"internal_error","message":"An unexpected error occurred. Please retry."}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			assert.Equal(t, tt.want, NewErrorCode(tt.code))
		})
	}
}

func TestErrorCodeRegistry_concurrentAccess(t *testing.T) {
	const goroutines = 50

	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Go(func() {
			code := ErrorCode(fmt.Sprintf("test_concurrent_code_%d", i))
			def := ErrorCodeDefinition{Status: 400, Message: "concurrent"}
			assert.NoError(t, RegisterErrorCode(code, def))
			assert.NoError(t, RegisterErrorCode(code, def)) // idempotent re-registration, racing lookups below
			_ = NewErrorCode(code)
			_ = NewErrorCode(ErrorCodeUnauthorized) // concurrent reads of a shared, already-registered code
		})
	}
	wg.Wait()
}

func TestRegisterErrorCode(t *testing.T) {
	t.Run("registering a new code succeeds and is usable via NewErrorCode", func(t *testing.T) {
		code := ErrorCode("test_widget_jammed")
		err := RegisterErrorCode(code, ErrorCodeDefinition{
			Status:  409,
			Message: "The widget is jammed.",
		})
		assert.NoError(t, err)

		got := NewErrorCode(code)
		want := events.APIGatewayProxyResponse{
			StatusCode: 409,
			Body:       `{"code":"test_widget_jammed","message":"The widget is jammed."}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		assert.Equal(t, want, got)
	})

	t.Run("registering the same code with an identical definition is a no-op", func(t *testing.T) {
		code := ErrorCode("test_idempotent_code")
		def := ErrorCodeDefinition{Status: 400, Message: "idempotent"}

		assert.NoError(t, RegisterErrorCode(code, def))
		assert.NoError(t, RegisterErrorCode(code, def))
	})

	t.Run("nil and empty Fields are treated as equal when checking for idempotent re-registration", func(t *testing.T) {
		code := ErrorCode("test_nil_vs_empty_fields")

		assert.NoError(t, RegisterErrorCode(code, ErrorCodeDefinition{Status: 400, Message: "msg", Fields: nil}))
		assert.NoError(t, RegisterErrorCode(code, ErrorCodeDefinition{Status: 400, Message: "msg", Fields: []ErrorField{}}))
	})

	t.Run("identical non-empty Fields are treated as equal when checking for idempotent re-registration", func(t *testing.T) {
		code := ErrorCode("test_identical_fields")
		def := ErrorCodeDefinition{Status: 400, Message: "msg", Fields: []ErrorField{NewErrorField("email", FieldErrorCodeInvalidFormat, "must be a valid email")}}

		assert.NoError(t, RegisterErrorCode(code, def))
		assert.NoError(t, RegisterErrorCode(code, def))
	})

	t.Run("differing non-empty Fields are treated as a conflict", func(t *testing.T) {
		code := ErrorCode("test_differing_fields")

		assert.NoError(t, RegisterErrorCode(code, ErrorCodeDefinition{Status: 400, Message: "msg", Fields: []ErrorField{NewErrorField("email", FieldErrorCodeInvalidFormat, "must be a valid email")}}))
		err := RegisterErrorCode(code, ErrorCodeDefinition{Status: 400, Message: "msg", Fields: []ErrorField{NewErrorField("email", FieldErrorCodeRequired, "email is required")}})
		assert.Error(t, err)
	})

	t.Run("registering the same code with a different definition returns an error", func(t *testing.T) {
		code := ErrorCode("test_conflicting_code")

		assert.NoError(t, RegisterErrorCode(code, ErrorCodeDefinition{Status: 400, Message: "first"}))
		err := RegisterErrorCode(code, ErrorCodeDefinition{Status: 400, Message: "second"})
		assert.Error(t, err)
	})

	t.Run("registering a built-in code with a different definition returns an error", func(t *testing.T) {
		err := RegisterErrorCode(ErrorCodeUnauthorized, ErrorCodeDefinition{Status: 418, Message: "different"})
		assert.Error(t, err)
	})

	t.Run("registering the empty ErrorCode returns an error", func(t *testing.T) {
		err := RegisterErrorCode(ErrorCode(""), ErrorCodeDefinition{Status: 400, Message: "msg"})
		assert.Error(t, err)
	})
}

func TestMustRegisterErrorCode(t *testing.T) {
	t.Run("does not panic on a new or identical registration", func(t *testing.T) {
		code := ErrorCode("test_must_register_ok")
		def := ErrorCodeDefinition{Status: 400, Message: "msg"}
		assert.NotPanics(t, func() {
			MustRegisterErrorCode(code, def)
			MustRegisterErrorCode(code, def)
		})
	})

	t.Run("panics on a conflicting registration", func(t *testing.T) {
		code := ErrorCode("test_must_register_conflict")
		MustRegisterErrorCode(code, ErrorCodeDefinition{Status: 400, Message: "first"})
		assert.Panics(t, func() {
			MustRegisterErrorCode(code, ErrorCodeDefinition{Status: 400, Message: "second"})
		})
	})
}
