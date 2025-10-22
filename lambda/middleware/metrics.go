package middleware

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/ellogroup/ello-golang-clock/clock"
	"github.com/ellogroup/ello-golang-metrics/metrics"
)

const (
	requestStartedMsg  = "Request started"
	requestCompleteMsg = "Request complete"
)

type metricsNoResponse[E any] struct {
	clock     clock.Clock
	outputter metrics.Outputter
}

// NewMetrics returns an implementation of NoResponse for the metrics middleware.
//
// The metrics middleware outputs the request start and end. The request start output contains the event and the request
// end output contains the duration of the request.
//
// Outputters need to implement the `metrics.Outputter` interface from the github.com/ellogroup/ello-golang-metrics/metrics
// package.
func NewMetrics[E any](outputter metrics.Outputter) NoResponse[E] {
	return &metricsNoResponse[E]{
		clock:     clock.NewSystem(),
		outputter: outputter,
	}
}

func (m metricsNoResponse[E]) Wrap(next func(context.Context, E) error) func(context.Context, E) error {
	return func(ctx context.Context, event E) error {
		// Log when the request starts
		start := m.clock.Now()
		m.outputter.Output(ctx, requestStartedMsg, []metrics.Field{{Name: "event", Val: event}})

		err := next(ctx, event)

		// Log when the request completes
		m.outputter.Output(ctx, requestCompleteMsg, []metrics.Field{{Name: "duration", Val: m.clock.Since(start)}})

		// Return response
		return err
	}
}

type metricsWithResponse[E, R any] struct {
	clock     clock.Clock
	outputter metrics.Outputter
}

// NewMetricsWithResponse returns an implementation of WithResponse for the metrics middleware.
//
// The metrics middleware outputs the request start and end. The request start output contains the event and the request
// end output contains the duration of the request.
//
// For API Gateway v1 requests the output also contains the status code of the response.
//
// Outputters need to implement the `metrics.Outputter` interface from the github.com/ellogroup/ello-golang-metrics/metrics
// package.
func NewMetricsWithResponse[E, R any](outputter metrics.Outputter) WithResponse[E, R] {
	return &metricsWithResponse[E, R]{
		outputter: outputter,
		clock:     clock.NewSystem(),
	}
}

func (m metricsWithResponse[E, R]) Wrap(next func(context.Context, E) (R, error)) func(context.Context, E) (R, error) {
	return func(ctx context.Context, event E) (R, error) {
		// Log when the request starts
		start := m.clock.Now()
		m.outputter.Output(ctx, requestStartedMsg, []metrics.Field{{Name: "event", Val: event}})

		response, err := next(ctx, event)

		// Log when the request completes
		f := []metrics.Field{
			{Name: "duration", Val: m.clock.Since(start)},
		}

		if apigwV1Response, ok := any(response).(events.APIGatewayProxyResponse); ok {
			// APIGatewayProxyResponse (API Gateway V1)
			f = append(f, metrics.Field{Name: "status_code", Val: apigwV1Response.StatusCode})
		}

		// Log when the request completes
		m.outputter.Output(ctx, requestCompleteMsg, f)

		// Return response
		return response, err
	}
}
