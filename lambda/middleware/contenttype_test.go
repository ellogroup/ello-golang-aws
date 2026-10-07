package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestContentType_Wrap(t *testing.T) {
	defaultRejected := events.APIGatewayProxyResponse{
		StatusCode: http.StatusUnsupportedMediaType,
		Body:       `{"code":"unsupported_media_type","message":"The request's Content-Type is not supported."}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
	custom := events.APIGatewayProxyResponse{StatusCode: http.StatusUnsupportedMediaType, Body: "custom"}
	ok := events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: "ok"}

	tests := []struct {
		name       string
		allowed    string
		opts       []ContentTypeOption
		event      events.APIGatewayProxyRequest
		handlerErr error
		want       events.APIGatewayProxyResponse
		wantCalled bool
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:       "allowed content type, calls handler",
			allowed:    "application/json",
			event:      events.APIGatewayProxyRequest{Headers: map[string]string{"Content-Type": "application/json"}},
			want:       ok,
			wantCalled: true,
			wantErr:    assert.NoError,
		},
		{
			name:       "lower-case header name and parameters, calls handler",
			allowed:    "application/json",
			event:      events.APIGatewayProxyRequest{Headers: map[string]string{"content-type": "Application/JSON; charset=utf-8"}},
			want:       ok,
			wantCalled: true,
			wantErr:    assert.NoError,
		},
		{
			name:       "only in multi-value headers, calls handler",
			allowed:    "application/json",
			event:      events.APIGatewayProxyRequest{MultiValueHeaders: map[string][]string{"Content-Type": {"application/json"}}},
			want:       ok,
			wantCalled: true,
			wantErr:    assert.NoError,
		},
		{
			name:       "one of several allowed, calls handler",
			allowed:    "application/json",
			opts:       []ContentTypeOption{WithContentTypeAlsoAllowed("text/plain")},
			event:      events.APIGatewayProxyRequest{Headers: map[string]string{"Content-Type": "text/plain"}},
			want:       ok,
			wantCalled: true,
			wantErr:    assert.NoError,
		},
		{
			name:       "allowed content type, handler error returned",
			allowed:    "application/json",
			event:      events.APIGatewayProxyRequest{Headers: map[string]string{"Content-Type": "application/json"}},
			handlerErr: errors.New("boom"),
			want:       ok,
			wantCalled: true,
			wantErr:    assert.Error,
		},
		{
			name:    "other content type, default 415 without calling handler",
			allowed: "application/json",
			event:   events.APIGatewayProxyRequest{Headers: map[string]string{"Content-Type": "text/plain"}},
			want:    defaultRejected,
			wantErr: assert.NoError,
		},
		{
			name:    "other content type with a custom response, custom response returned",
			allowed: "application/json",
			opts:    []ContentTypeOption{WithContentTypeRejectedResponse(custom)},
			event:   events.APIGatewayProxyRequest{Headers: map[string]string{"Content-Type": "text/plain"}},
			want:    custom,
			wantErr: assert.NoError,
		},
		{
			name:    "no content type, rejected",
			allowed: "application/json",
			event:   events.APIGatewayProxyRequest{},
			want:    defaultRejected,
			wantErr: assert.NoError,
		},
		{
			name:    "malformed content type, rejected",
			allowed: "application/json",
			event:   events.APIGatewayProxyRequest{Headers: map[string]string{"Content-Type": "application/json; ="}},
			want:    defaultRejected,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
				called = true
				return ok, tt.handlerErr
			}

			got, err := NewContentType(tt.allowed, tt.opts...).Wrap(next)(context.Background(), tt.event)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantCalled, called)
		})
	}
}
