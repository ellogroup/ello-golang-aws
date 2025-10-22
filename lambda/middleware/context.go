package middleware

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/ellogroup/ello-golang-ctx/logctx"
)

type contextNoResponse[E any] struct{}

// NewContext returns an implementation of NoResponse for the context middleware.
//
// The context middleware adds additional information to the context of each request using the
// github.com/ellogroup/ello-golang-ctx/logctx package. This includes at the very least a request id.
func NewContext[E any]() NoResponse[E] {
	return &contextNoResponse[E]{}
}

func (c contextNoResponse[E]) Wrap(next func(context.Context, E) error) func(context.Context, E) error {
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

func (c contextWithResponse[E, R]) Wrap(next func(context.Context, E) (R, error)) func(context.Context, E) (R, error) {
	return func(ctx context.Context, event E) (R, error) {
		// Get context from event
		requestId, ctx := contextFromEvent(ctx, event)

		// Get response
		response, err := next(ctx, event)

		// Transform response
		response = transformResponse(response, requestId)

		// Return response
		return response, err
	}
}

func contextFromEvent[E any](ctx context.Context, event E) (string, context.Context) {
	// Extract request ids
	requestId, lambdaRequestId := "", ""
	if lambdaCtx, ok := lambdacontext.FromContext(ctx); ok {
		requestId, lambdaRequestId = lambdaCtx.AwsRequestID, lambdaCtx.AwsRequestID
	}

	// Event specific context
	var additionalCtx []logctx.Field

	if apigwV1Event, ok := any(event).(events.APIGatewayProxyRequest); ok {
		// APIGatewayProxyRequest (API Gateway V1)
		amznRequestId := ""
		if id := apigwV1Event.RequestContext.RequestID; id != "" {
			requestId, amznRequestId = id, id
		}
		additionalCtx = append(additionalCtx,
			logctx.String("amzn_request_id", amznRequestId),
			logctx.String("request_method", apigwV1Event.RequestContext.HTTPMethod),
			logctx.String("request_domain", apigwV1Event.RequestContext.DomainName),
			logctx.String("request_path", apigwV1Event.RequestContext.Path),
		)
	}

	// Set context
	ctx = logctx.Add(
		ctx,
		logctx.String("request_id", requestId),
		logctx.String("lambda_request_id", lambdaRequestId),
	)
	if len(additionalCtx) > 0 {
		// Add additional ctx
		ctx = logctx.Add(
			ctx,
			additionalCtx...,
		)
	}

	return requestId, ctx
}

func transformResponse[R any](response R, requestId string) R {
	if apigwV1Response, ok := any(response).(events.APIGatewayProxyResponse); ok {
		// APIGatewayProxyResponse (API Gateway V1)
		if apigwV1Response.Headers != nil {
			// Add request id to response headers
			apigwV1Response.Headers["x-request-id"] = requestId
		}
		response = any(apigwV1Response).(R)
	}
	return response
}
