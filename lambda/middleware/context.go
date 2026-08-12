package middleware

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"github.com/ellogroup/ello-golang-ctx/v2/logctx"
)

type contextNoResponse[E any] struct{}

// NewContext returns an implementation of NoResponse for the context middleware.
//
// The context middleware adds additional information to the context of each request using the
// github.com/ellogroup/ello-golang-ctx/logctx package. This includes at the very least a request id.
func NewContext[E any]() NoResponse[E] {
	return &contextNoResponse[E]{}
}

func (contextNoResponse[E]) Wrap(next func(context.Context, E) error) func(context.Context, E) error {
	return func(ctx context.Context, event E) error {
		// Get context from event
		_, ctx = contextFromEvent(ctx, event)

		// return response
		return next(ctx, event)
	}
}

type contextWithResponse[E, R any] struct{}

// NewContextWithResponse returns an implementation of WithResponse for the context middleware.
//
// The context middleware adds additional information to the context of each request using the
// github.com/ellogroup/ello-golang-ctx/logctx package. This includes at the very least a request id.
//
// For API Gateway v1 requests the context also includes the method, domain and path of the request. The response is also
// updated to include the request id within the header `x-request-id`.
func NewContextWithResponse[E, R any]() WithResponse[E, R] {
	return &contextWithResponse[E, R]{}
}

func (contextWithResponse[E, R]) Wrap(next func(context.Context, E) (R, error)) func(context.Context, E) (R, error) {
	return func(ctx context.Context, event E) (R, error) {
		// Get context from event
		requestID, ctx := contextFromEvent(ctx, event)

		// Get response
		response, err := next(ctx, event)

		// Transform response
		response = transformResponse(response, requestID)

		// Return response
		return response, err
	}
}

func contextFromEvent[E any](ctx context.Context, event E) (string, context.Context) {
	// Extract request ids
	requestID, lambdaRequestID := "", ""
	if lambdaCtx, ok := lambdacontext.FromContext(ctx); ok {
		requestID, lambdaRequestID = lambdaCtx.AwsRequestID, lambdaCtx.AwsRequestID
	}

	// Event specific context
	var additionalCtx []logctx.Field

	if apigwV1Event, ok := any(event).(events.APIGatewayProxyRequest); ok {
		// APIGatewayProxyRequest (API Gateway V1)
		amznRequestID := ""
		if id := apigwV1Event.RequestContext.RequestID; id != "" {
			requestID, amznRequestID = id, id
		}
		additionalCtx = append(additionalCtx,
			logctx.String("amzn_request_id", amznRequestID),
			logctx.String("request_method", apigwV1Event.RequestContext.HTTPMethod),
			logctx.String("request_domain", apigwV1Event.RequestContext.DomainName),
			logctx.String("request_path", apigwV1Event.RequestContext.Path),
		)
	}

	// Set context
	ctx = logctx.Add(
		ctx,
		logctx.String("request_id", requestID),
		logctx.String("lambda_request_id", lambdaRequestID),
	)
	if len(additionalCtx) > 0 {
		// Add additional ctx
		ctx = logctx.Add(
			ctx,
			additionalCtx...,
		)
	}

	return requestID, ctx
}

func transformResponse[R any](response R, requestID string) R {
	if apigwV1Response, ok := any(response).(events.APIGatewayProxyResponse); ok {
		// APIGatewayProxyResponse (API Gateway V1)
		if apigwV1Response.Headers != nil {
			// Add request id to response headers
			apigwV1Response.Headers["x-request-id"] = requestID
		}
		if transformed, ok := any(apigwV1Response).(R); ok {
			response = transformed
		}
	}
	return response
}
