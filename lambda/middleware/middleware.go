package middleware

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/ellogroup/ello-golang-metrics/metrics"
)

// NoResponse [E any] interface should be implemented for middleware of handlers of event type E that do not return a
// response.
type NoResponse[E any] interface {
	// Wrap returns a function that receives an event and returns an error. The next parameter is the next function to
	// be called in the chain of middleware and handlers, and should always be called unless the desired outcome is to
	// prevent the request from proceeding (i.e. a validation middleware could return an error straight away and
	// prevent an invalid request from being further processed)
	Wrap(next func(context.Context, E) error) func(context.Context, E) error
}

// WithResponse [E, R any] interface should be implemented for middleware of handlers of event type E that return a
// response type R.
type WithResponse[E, R any] interface {
	// Wrap returns a function that receives an event and returns a response/error. The next parameter is the next
	// function to be called in the chain of middleware and handlers, and should always be called unless the desired
	// outcome is to prevent the request from proceeding (i.e. a validation middleware could return an error straight
	// away and prevent an invalid request from being further processed)
	Wrap(next func(context.Context, E) (R, error)) func(context.Context, E) (R, error)
}

// Common [E any] returns a slice of common middleware for handlers of event type E that do not return a
// response.
func Common[E any](outputter metrics.Outputter) []NoResponse[E] {
	return []NoResponse[E]{
		NewContext[E](),
		NewMetrics[E](outputter),
	}
}

// CommonWithResponse [E, R any] returns a slice of common middleware for handlers of event type E that return a
// response type R.
func CommonWithResponse[E, R any](outputter metrics.Outputter) []WithResponse[E, R] {
	return []WithResponse[E, R]{
		NewContextWithResponse[E, R](),
		NewMetricsWithResponse[E, R](outputter),
	}
}

// Common middleware for common AWS Events

type S3 []NoResponse[events.S3Event]

// CommonS3 returns a slice of common middleware for handlers of events.S3Event
func CommonS3(outputter metrics.Outputter) S3 {
	return Common[events.S3Event](outputter)
}

type SNS []NoResponse[events.SNSEvent]

// CommonSNS returns a slice of common middleware for handlers of events.SNSEvent
func CommonSNS(outputter metrics.Outputter) SNS {
	return Common[events.SNSEvent](outputter)
}

type SQS []NoResponse[events.SQSEvent]

// CommonSQS returns a slice of common middleware for handlers of events.SQSEvent
func CommonSQS(outputter metrics.Outputter) SQS {
	return Common[events.SQSEvent](outputter)
}

type APIGatewayV1 []WithResponse[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse]

// CommonAPIGatewayV1 returns a slice of common middleware for handlers of events.APIGatewayProxyRequest that return
// events.APIGatewayProxyResponse
func CommonAPIGatewayV1(outputter metrics.Outputter) APIGatewayV1 {
	return CommonWithResponse[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse](outputter)
}
