package middleware

import (
	"context"
	"mime"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/ellogroup/ello-golang-aws/v2/apigw/response"
)

type contentTypeOptions struct {
	alsoAllowed []string
	rejected    *events.APIGatewayProxyResponse
}

// ContentTypeOption configures NewContentType.
type ContentTypeOption func(*contentTypeOptions)

// WithContentTypeAlsoAllowed allows further media types besides the one passed to NewContentType.
func WithContentTypeAlsoAllowed(mediaTypes ...string) ContentTypeOption {
	return func(o *contentTypeOptions) {
		o.alsoAllowed = append(o.alsoAllowed, mediaTypes...)
	}
}

// WithContentTypeRejectedResponse replaces the default 415 (response.ErrorCodeUnsupportedMedia) returned for a
// request whose Content-Type isn't allowed, e.g. to keep an application's own error format.
func WithContentTypeRejectedResponse(rejected events.APIGatewayProxyResponse) ContentTypeOption {
	return func(o *contentTypeOptions) {
		o.rejected = &rejected
	}
}

type contentType struct {
	allowed  []string
	rejected events.APIGatewayProxyResponse
}

// NewContentType returns API Gateway v1 middleware that only lets a request through when its Content-Type header's
// media type is allowed (e.g. "application/json"). Any other request, including one with no Content-Type, gets a 415
// built from response.ErrorCodeUnsupportedMedia, unless WithContentTypeRejectedResponse sets another response.
//
// Media types are compared case-insensitively and parameters are ignored, so "application/json; charset=utf-8"
// matches "application/json". The Content-Type header name is matched case-insensitively.
func NewContentType(allowed string, opts ...ContentTypeOption) WithResponse[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse] {
	o := contentTypeOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	m := &contentType{}
	for _, t := range append([]string{allowed}, o.alsoAllowed...) {
		m.allowed = append(m.allowed, strings.ToLower(strings.TrimSpace(t)))
	}
	if o.rejected != nil {
		m.rejected = *o.rejected
	} else {
		m.rejected = response.NewErrorCode(response.ErrorCodeUnsupportedMedia)
	}
	return m
}

func (m contentType) Wrap(next func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if !m.isAllowed(headerValue(event, "Content-Type")) {
			return m.rejected, nil
		}
		return next(ctx, event)
	}
}

func (m contentType) isAllowed(value string) bool {
	if value == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return false
	}
	for _, a := range m.allowed {
		if mediaType == a {
			return true
		}
	}
	return false
}

// headerValue returns the first value of the named header, matching its name case-insensitively. API Gateway passes
// headers with the casing the client sent.
func headerValue(event events.APIGatewayProxyRequest, name string) string {
	for k, v := range event.Headers {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	for k, v := range event.MultiValueHeaders {
		if strings.EqualFold(k, name) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}
