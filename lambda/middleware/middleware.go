package middleware

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
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
func Common[E any](logger *slog.Logger) []NoResponse[E] {
	return []NoResponse[E]{
		NewContext[E](),
		NewEventLogger[E](logger),
	}
}

// CommonWithResponse [E, R any] returns a slice of common middleware for handlers of event type E that return a
// response type R.
func CommonWithResponse[E, R any](logger *slog.Logger) []WithResponse[E, R] {
	return []WithResponse[E, R]{
		NewContextWithResponse[E, R](),
		NewEventLoggerWithResponse[E, R](logger),
	}
}

// Common middleware for common AWS Events

// S3 is a slice of NoResponse middleware for handlers of events.S3Event.
type S3 []NoResponse[events.S3Event]

// CommonS3 returns a slice of common middleware for handlers of events.S3Event
func CommonS3(logger *slog.Logger) S3 {
	return Common[events.S3Event](logger)
}

// SNS is a slice of NoResponse middleware for handlers of events.SNSEvent.
type SNS []NoResponse[events.SNSEvent]

// CommonSNS returns a slice of common middleware for handlers of events.SNSEvent
func CommonSNS(logger *slog.Logger) SNS {
	return Common[events.SNSEvent](logger)
}

// SQS is a slice of NoResponse middleware for handlers of events.SQSEvent.
type SQS []NoResponse[events.SQSEvent]

// CommonSQS returns a slice of common middleware for handlers of events.SQSEvent
func CommonSQS(logger *slog.Logger) SQS {
	return Common[events.SQSEvent](logger)
}

// APIGatewayV1 is a slice of WithResponse middleware for handlers of events.APIGatewayProxyRequest
// that return events.APIGatewayProxyResponse.
type APIGatewayV1 []WithResponse[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse]

// CommonAPIGatewayV1 returns a slice of common middleware for handlers of events.APIGatewayProxyRequest that return
// events.APIGatewayProxyResponse
func CommonAPIGatewayV1(logger *slog.Logger) APIGatewayV1 {
	return CommonWithResponse[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse](logger)
}
