package lambda

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ellogroup/ello-golang-aws/lambda/middleware"
)

// Handler [E any] interface should be implemented for handlers of event type E that do not return a response.
type Handler[E any] interface {
	// Handle handles an event.
	Handle(ctx context.Context, event E) error
}

// Start initiates a lambda container for a handler and middleware of events that do not return a response.
// sigTermCallbacks are callbacks to be triggered when the lambda container is closed.
func Start[E any](handler Handler[E], middlewares []middleware.NoResponse[E], sigTermCallbacks ...func()) {
	lambda.StartWithOptions(
		wrappedHandlerFn(handler, middlewares...),
		lambda.WithEnableSIGTERM(sigTermCallbacks...),
	)
}

func wrappedHandlerFn[E any](handler Handler[E], middlewares ...middleware.NoResponse[E]) func(context.Context, E) error {
	handlerFn := handler.Handle
	for i := len(middlewares) - 1; i >= 0; i-- {
		handlerFn = middlewares[i].Wrap(handlerFn)
	}
	return handlerFn
}

// HandlerWithResponse [E, R any] interface should be implemented for handlers of event type E that return a response
// type R.
type HandlerWithResponse[E, R any] interface {
	// Handle handles an event and returns a response.
	Handle(ctx context.Context, event E) (R, error)
}

// StartWithResponse initiates a lambda container for a handler and middleware of events that return a response type R.
// sigTermCallbacks are callbacks to be triggered when the lambda container is closed.
func StartWithResponse[E, R any](handler HandlerWithResponse[E, R], middlewares []middleware.WithResponse[E, R], sigTermCallbacks ...func()) {
	lambda.StartWithOptions(
		wrappedHandlerWithResponseFn(handler, middlewares...),
		lambda.WithEnableSIGTERM(sigTermCallbacks...),
	)
}

func wrappedHandlerWithResponseFn[E, R any](handler HandlerWithResponse[E, R], middlewares ...middleware.WithResponse[E, R]) func(context.Context, E) (R, error) {
	handlerFn := handler.Handle
	for i := len(middlewares) - 1; i >= 0; i-- {
		handlerFn = middlewares[i].Wrap(handlerFn)
	}
	return handlerFn
}
