package middleware

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/ellogroup/ello-golang-clock/clock"
	"github.com/ellogroup/ello-golang-metrics/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
	"time"
)

type mockOutputter struct {
	mock.Mock
}

func (m *mockOutputter) Output(ctx context.Context, label string, fields []metrics.Field) {
	m.Called(ctx, label, fields)
}

func Test_metricsNoResponse_Wrap(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC)

	type args[E any] struct {
		ctx   context.Context
		event E
	}
	type mockOpts struct {
		outputter func(m *mockOutputter)
	}
	type testCase[E any] struct {
		name       string
		args       args[E]
		mockOpts   mockOpts
		handlerErr error
		wantEvent  E
		wantErr    assert.ErrorAssertionFunc
	}
	tests := []testCase[string]{
		{
			name: "metrics outputted, handler returns nil, returns nil",
			args: args[string]{
				ctx:   context.Background(),
				event: "test",
			},
			mockOpts: mockOpts{outputter: func(m *mockOutputter) {
				m.On("Output", mock.Anything, requestStartedMsg, []metrics.Field{{Name: "event", Val: "test"}}).Return()
				m.On("Output", mock.Anything, requestCompleteMsg, []metrics.Field{{Name: "duration", Val: time.Duration(0)}}).Return()
			}},
			handlerErr: nil,
			wantEvent:  "test",
			wantErr:    assert.NoError,
		},
		{
			name: "metrics outputted, handler returns error, returns error",
			args: args[string]{
				ctx:   context.Background(),
				event: "test",
			},
			mockOpts: mockOpts{outputter: func(m *mockOutputter) {
				m.On("Output", mock.Anything, requestStartedMsg, []metrics.Field{{Name: "event", Val: "test"}}).Return()
				m.On("Output", mock.Anything, requestCompleteMsg, []metrics.Field{{Name: "duration", Val: time.Duration(0)}}).Return()
			}},
			handlerErr: errors.New("error"),
			wantEvent:  "test",
			wantErr:    assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mOutputter := new(mockOutputter)
			if tt.mockOpts.outputter != nil {
				tt.mockOpts.outputter(mOutputter)
			}

			sut := &metricsNoResponse[string]{
				clock:     clock.NewFixed(now),
				outputter: mOutputter,
			}
			fn := sut.Wrap(func(_ context.Context, event string) error {
				assert.Equalf(t, tt.wantEvent, event, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event)
				return tt.handlerErr
			})
			gotErr := fn(tt.args.ctx, tt.args.event)

			if !tt.wantErr(t, gotErr, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event) {
				return
			}

			mOutputter.AssertExpectations(t)
		})
	}
}

func Test_metricsWithResponse_Wrap(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC)

	type args[E any] struct {
		ctx   context.Context
		event E
	}
	type mockOpts struct {
		outputter func(m *mockOutputter)
	}
	type testCase[E any, R any] struct {
		name        string
		args        args[E]
		mockOpts    mockOpts
		handlerResp R
		handlerErr  error
		wantEvent   E
		wantResp    R
		wantErr     assert.ErrorAssertionFunc
	}
	tests := []testCase[string, any]{
		{
			name: "metrics outputted, handler returns nil, returns nil",
			args: args[string]{
				ctx:   context.Background(),
				event: "test-event",
			},
			mockOpts: mockOpts{outputter: func(m *mockOutputter) {
				m.On("Output", mock.Anything, requestStartedMsg, []metrics.Field{{Name: "event", Val: "test-event"}}).Return()
				m.On("Output", mock.Anything, requestCompleteMsg, []metrics.Field{{Name: "duration", Val: time.Duration(0)}}).Return()
			}},
			handlerResp: "test-response",
			handlerErr:  nil,
			wantEvent:   "test-event",
			wantResp:    "test-response",
			wantErr:     assert.NoError,
		},
		{
			name: "metrics outputted, handler returns error, returns error",
			args: args[string]{
				ctx:   context.Background(),
				event: "test-event",
			},
			mockOpts: mockOpts{outputter: func(m *mockOutputter) {
				m.On("Output", mock.Anything, requestStartedMsg, []metrics.Field{{Name: "event", Val: "test-event"}}).Return()
				m.On("Output", mock.Anything, requestCompleteMsg, []metrics.Field{{Name: "duration", Val: time.Duration(0)}}).Return()
			}},
			handlerErr: errors.New("error"),
			wantEvent:  "test-event",
			wantErr:    assert.Error,
		},
		{
			name: "apigw response, metrics outputted including http response, handler returns nil, returns nil",
			args: args[string]{
				ctx:   context.Background(),
				event: "test-event",
			},
			mockOpts: mockOpts{outputter: func(m *mockOutputter) {
				m.On("Output", mock.Anything, requestStartedMsg, []metrics.Field{{Name: "event", Val: "test-event"}}).Return()
				m.On("Output", mock.Anything, requestCompleteMsg, []metrics.Field{
					{Name: "duration", Val: time.Duration(0)},
					{Name: "status_code", Val: http.StatusOK},
				}).Return()
			}},
			handlerResp: events.APIGatewayProxyResponse{StatusCode: http.StatusOK},
			handlerErr:  nil,
			wantEvent:   "test-event",
			wantResp:    events.APIGatewayProxyResponse{StatusCode: http.StatusOK},
			wantErr:     assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mOutputter := new(mockOutputter)
			if tt.mockOpts.outputter != nil {
				tt.mockOpts.outputter(mOutputter)
			}

			sut := &metricsWithResponse[string, any]{
				clock:     clock.NewFixed(now),
				outputter: mOutputter,
			}
			fn := sut.Wrap(func(ctx context.Context, event string) (any, error) {
				assert.Equalf(t, tt.wantEvent, event, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event)
				return tt.handlerResp, tt.handlerErr
			})
			gotResp, gotErr := fn(tt.args.ctx, tt.args.event)

			if !tt.wantErr(t, gotErr, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event) {
				return
			}
			assert.Equalf(t, tt.wantResp, gotResp, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event)
		})
	}
}
