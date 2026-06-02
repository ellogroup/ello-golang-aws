package middleware

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/ellogroup/ello-golang-clock/clock"
	"log/slog"
)

const (
	defaultEventStartedMsg     = "Request started"
	defaultEventCompletedMsg   = "Request complete"
	defaultEventStartedLevel   = slog.LevelInfo
	defaultEventCompletedLevel = slog.LevelInfo
)

type eventLoggerOptions struct {
	eventStartedMsg     string
	eventCompletedMsg   string
	eventStartedLevel   slog.Level
	eventCompletedLevel slog.Level
	sanitizer           func(any) any
}

var defaultEventLoggerOptions = eventLoggerOptions{
	eventStartedMsg:     defaultEventStartedMsg,
	eventCompletedMsg:   defaultEventCompletedMsg,
	eventStartedLevel:   defaultEventStartedLevel,
	eventCompletedLevel: defaultEventCompletedLevel,
}

type EventLoggerOption func(*eventLoggerOptions)

func WithEventLoggerEventStartedMsg(msg string) EventLoggerOption {
	return func(e *eventLoggerOptions) {
		e.eventStartedMsg = msg
	}
}
func WithEventLoggerEventCompletedMsg(msg string) EventLoggerOption {
	return func(e *eventLoggerOptions) {
		e.eventCompletedMsg = msg
	}
}
func WithEventLoggerEventStartedLevel(l slog.Level) EventLoggerOption {
	return func(e *eventLoggerOptions) {
		e.eventStartedLevel = l
	}
}
func WithEventLoggerEventCompletedLevel(l slog.Level) EventLoggerOption {
	return func(e *eventLoggerOptions) {
		e.eventCompletedLevel = l
	}
}

// WithEventLoggerSanitizer sets a custom function to transform the event before it is logged.
// When provided, it replaces the default built-in HTTP header redaction. To compose both, call
// RedactHTTPEvent inside your sanitizer function.
func WithEventLoggerSanitizer[E any](fn func(E) any) EventLoggerOption {
	return func(opts *eventLoggerOptions) {
		opts.sanitizer = func(e any) any {
			return fn(e.(E))
		}
	}
}

func sanitizeEvent(opts *eventLoggerOptions, event any) any {
	if opts.sanitizer != nil {
		return opts.sanitizer(event)
	}
	return RedactHTTPEvent(event)
}

type eventLoggerNoResponse[E any] struct {
	clock  clock.Clock
	logger *slog.Logger
	opts   *eventLoggerOptions
}

// NewEventLogger returns an implementation of NoResponse middleware.
//
// The event logger middleware logs the event start and end. The event start log record contains the event and the event
// end log record contains the duration of the event.
func NewEventLogger[E any](logger *slog.Logger, options ...EventLoggerOption) NoResponse[E] {
	opts := defaultEventLoggerOptions
	l := &eventLoggerNoResponse[E]{
		clock:  clock.NewSystem(),
		logger: logger,
		opts:   &opts,
	}
	for _, option := range options {
		option(l.opts)
	}
	return l
}

func (l eventLoggerNoResponse[E]) Wrap(next func(context.Context, E) error) func(context.Context, E) error {
	return func(ctx context.Context, event E) error {
		// Log when the event starts
		start := l.clock.Now()
		l.logger.LogAttrs(ctx, l.opts.eventStartedLevel, l.opts.eventStartedMsg, slog.Any("event", sanitizeEvent(l.opts, event)))

		err := next(ctx, event)

		// Log when the event completes
		l.logger.LogAttrs(ctx, l.opts.eventCompletedLevel, l.opts.eventCompletedMsg, slog.Duration("duration", l.clock.Since(start)))

		// Return response
		return err
	}
}

type eventLoggerWithResponse[E, R any] struct {
	clock  clock.Clock
	logger *slog.Logger
	opts   *eventLoggerOptions
}

// NewEventLoggerWithResponse returns an implementation of WithResponse middleware.
//
// The event logger middleware logs the event start and end. The event start log record contains the event and the event
// end log record contains the duration of the event.
//
// For API Gateway v1 requests the log record also contains the status code of the response.
func NewEventLoggerWithResponse[E, R any](logger *slog.Logger, options ...EventLoggerOption) WithResponse[E, R] {
	opts := defaultEventLoggerOptions
	l := &eventLoggerWithResponse[E, R]{
		clock:  clock.NewSystem(),
		logger: logger,
		opts:   &opts,
	}
	for _, option := range options {
		option(l.opts)
	}
	return l
}

func (l eventLoggerWithResponse[E, R]) Wrap(next func(context.Context, E) (R, error)) func(context.Context, E) (R, error) {
	return func(ctx context.Context, event E) (R, error) {
		// Log when the event starts
		start := l.clock.Now()
		l.logger.LogAttrs(ctx, l.opts.eventStartedLevel, l.opts.eventStartedMsg, slog.Any("event", sanitizeEvent(l.opts, event)))

		response, err := next(ctx, event)

		// Log when the event completes
		attr := []slog.Attr{
			slog.Duration("duration", l.clock.Since(start)),
		}

		if apigwV1Response, ok := any(response).(events.APIGatewayProxyResponse); ok {
			// APIGatewayProxyResponse (API Gateway V1)
			attr = append(attr, slog.Int("status_code", apigwV1Response.StatusCode))
		}

		l.logger.LogAttrs(ctx, l.opts.eventCompletedLevel, l.opts.eventCompletedMsg, attr...)

		// Return response
		return response, err
	}
}
