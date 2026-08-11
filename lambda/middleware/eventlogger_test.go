package middleware

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/ellogroup/ello-golang-clock/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

type mockSlogHandler struct {
	mock.Mock
}

func (m *mockSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return m.Called(ctx, level).Bool(0)
}

func (m *mockSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	return m.Called(ctx, r).Error(0)
}

func (m *mockSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m.Called(attrs).Get(0).(slog.Handler)
}

func (m *mockSlogHandler) WithGroup(name string) slog.Handler {
	return m.Called(name).Get(0).(slog.Handler)
}

func matchRecord(msg string, level slog.Level, attrs []slog.Attr) func(slog.Record) bool {
	return func(r slog.Record) bool {
		if r.Message != msg || r.Level != level {
			return false
		}
		var got []slog.Attr
		r.Attrs(func(a slog.Attr) bool {
			got = append(got, a)
			return true
		})
		return assert.ObjectsAreEqual(got, attrs)
	}
}

func Test_eventLoggerNoResponse_Wrap(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC)

	type args[E any] struct {
		ctx   context.Context
		event E
	}
	type mockOpts struct {
		handler func(m *mockSlogHandler)
	}
	type testCase[E any] struct {
		name       string
		options    *eventLoggerOptions
		args       args[E]
		mockOpts   mockOpts
		handlerErr error
		wantEvent  E
		wantErr    assert.ErrorAssertionFunc
	}
	tests := []testCase[string]{
		{
			name:    "logs event started and completed, handler returns nil, returns nil",
			options: &defaultEventLoggerOptions,
			args: args[string]{
				ctx:   context.Background(),
				event: "test",
			},
			mockOpts: mockOpts{handler: func(m *mockSlogHandler) {
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventStartedMsg, slog.LevelInfo, []slog.Attr{slog.Any("event", "test")},
				))).Return(nil)
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventCompletedMsg, slog.LevelInfo, []slog.Attr{slog.Duration("duration", time.Duration(0))},
				))).Return(nil)
			}},
			handlerErr: nil,
			wantEvent:  "test",
			wantErr:    assert.NoError,
		},
		{
			name:    "logs event started and completed, handler returns error, returns error",
			options: &defaultEventLoggerOptions,
			args: args[string]{
				ctx:   context.Background(),
				event: "test",
			},
			mockOpts: mockOpts{handler: func(m *mockSlogHandler) {
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventStartedMsg, slog.LevelInfo, []slog.Attr{slog.Any("event", "test")},
				))).Return(nil)
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventCompletedMsg, slog.LevelInfo, []slog.Attr{slog.Duration("duration", time.Duration(0))},
				))).Return(nil)
			}},
			handlerErr: errors.New("error"),
			wantEvent:  "test",
			wantErr:    assert.Error,
		},
		{
			name: "custom messages and levels are used",
			options: &eventLoggerOptions{
				eventStartedMsg:     "custom started",
				eventCompletedMsg:   "custom completed",
				eventStartedLevel:   slog.LevelDebug,
				eventCompletedLevel: slog.LevelWarn,
			},
			args: args[string]{
				ctx:   context.Background(),
				event: "test",
			},
			mockOpts: mockOpts{handler: func(m *mockSlogHandler) {
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					"custom started", slog.LevelDebug, []slog.Attr{slog.Any("event", "test")},
				))).Return(nil)
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					"custom completed", slog.LevelWarn, []slog.Attr{slog.Duration("duration", time.Duration(0))},
				))).Return(nil)
			}},
			handlerErr: nil,
			wantEvent:  "test",
			wantErr:    assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mHandler := new(mockSlogHandler)
			mHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)
			if tt.mockOpts.handler != nil {
				tt.mockOpts.handler(mHandler)
			}

			sut := &eventLoggerNoResponse[string]{
				clock:  clock.NewFixed(now),
				logger: slog.New(mHandler),
				opts:   tt.options,
			}
			fn := sut.Wrap(func(_ context.Context, event string) error {
				assert.Equalf(t, tt.wantEvent, event, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event)
				return tt.handlerErr
			})
			gotErr := fn(tt.args.ctx, tt.args.event)

			if !tt.wantErr(t, gotErr, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event) {
				return
			}

			mHandler.AssertExpectations(t)
		})
	}
}

func Test_eventLoggerWithResponse_Wrap(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC)

	type args[E any] struct {
		ctx   context.Context
		event E
	}
	type mockOpts struct {
		handler func(m *mockSlogHandler)
	}
	type testCase[E any, R any] struct {
		name        string
		options     *eventLoggerOptions
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
			name:    "logs event started and completed, handler returns response, returns response",
			options: &defaultEventLoggerOptions,
			args: args[string]{
				ctx:   context.Background(),
				event: "test-event",
			},
			mockOpts: mockOpts{handler: func(m *mockSlogHandler) {
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventStartedMsg, slog.LevelInfo, []slog.Attr{slog.Any("event", "test-event")},
				))).Return(nil)
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventCompletedMsg, slog.LevelInfo, []slog.Attr{slog.Duration("duration", time.Duration(0))},
				))).Return(nil)
			}},
			handlerResp: "test-response",
			handlerErr:  nil,
			wantEvent:   "test-event",
			wantResp:    "test-response",
			wantErr:     assert.NoError,
		},
		{
			name:    "logs event started and completed, handler returns error, returns error",
			options: &defaultEventLoggerOptions,
			args: args[string]{
				ctx:   context.Background(),
				event: "test-event",
			},
			mockOpts: mockOpts{handler: func(m *mockSlogHandler) {
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventStartedMsg, slog.LevelInfo, []slog.Attr{slog.Any("event", "test-event")},
				))).Return(nil)
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventCompletedMsg, slog.LevelInfo, []slog.Attr{slog.Duration("duration", time.Duration(0))},
				))).Return(nil)
			}},
			handlerErr: errors.New("error"),
			wantEvent:  "test-event",
			wantErr:    assert.Error,
		},
		{
			name:    "api gateway v1 response, logs status code in completed record",
			options: &defaultEventLoggerOptions,
			args: args[string]{
				ctx:   context.Background(),
				event: "test-event",
			},
			mockOpts: mockOpts{handler: func(m *mockSlogHandler) {
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventStartedMsg, slog.LevelInfo, []slog.Attr{slog.Any("event", "test-event")},
				))).Return(nil)
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					defaultEventCompletedMsg, slog.LevelInfo, []slog.Attr{
						slog.Duration("duration", time.Duration(0)),
						slog.Int("status_code", http.StatusOK),
					},
				))).Return(nil)
			}},
			handlerResp: events.APIGatewayProxyResponse{StatusCode: http.StatusOK},
			handlerErr:  nil,
			wantEvent:   "test-event",
			wantResp:    events.APIGatewayProxyResponse{StatusCode: http.StatusOK},
			wantErr:     assert.NoError,
		},
		{
			name: "custom messages and levels are used",
			options: &eventLoggerOptions{
				eventStartedMsg:     "custom started",
				eventCompletedMsg:   "custom completed",
				eventStartedLevel:   slog.LevelDebug,
				eventCompletedLevel: slog.LevelWarn,
			},
			args: args[string]{
				ctx:   context.Background(),
				event: "test-event",
			},
			mockOpts: mockOpts{handler: func(m *mockSlogHandler) {
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					"custom started", slog.LevelDebug, []slog.Attr{slog.Any("event", "test-event")},
				))).Return(nil)
				m.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
					"custom completed", slog.LevelWarn, []slog.Attr{slog.Duration("duration", time.Duration(0))},
				))).Return(nil)
			}},
			handlerResp: "test-response",
			handlerErr:  nil,
			wantEvent:   "test-event",
			wantResp:    "test-response",
			wantErr:     assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mHandler := new(mockSlogHandler)
			mHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)
			if tt.mockOpts.handler != nil {
				tt.mockOpts.handler(mHandler)
			}

			sut := &eventLoggerWithResponse[string, any]{
				clock:  clock.NewFixed(now),
				logger: slog.New(mHandler),
				opts:   tt.options,
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

			mHandler.AssertExpectations(t)
		})
	}
}

func Test_eventLoggerNoResponse_Wrap_redactsHTTPEventByDefault(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC)
	event := events.APIGatewayProxyRequest{
		Path:    "/test",
		Headers: map[string]string{"Authorization": "Bearer secret", "Content-Type": "application/json"},
		Body:    `{"email":"jane@example.com"}`,
	}
	wantLogged := events.APIGatewayProxyRequest{
		Path:    "/test",
		Headers: map[string]string{"Authorization": redactedValue, "Content-Type": "application/json"},
		Body:    redactedValue,
	}

	mHandler := new(mockSlogHandler)
	mHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
		defaultEventStartedMsg, slog.LevelInfo, []slog.Attr{slog.Any("event", wantLogged)},
	))).Return(nil)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
		defaultEventCompletedMsg, slog.LevelInfo, []slog.Attr{slog.Duration("duration", time.Duration(0))},
	))).Return(nil)

	sut := &eventLoggerNoResponse[events.APIGatewayProxyRequest]{
		clock:  clock.NewFixed(now),
		logger: slog.New(mHandler),
		opts:   &defaultEventLoggerOptions,
	}
	fn := sut.Wrap(func(_ context.Context, e events.APIGatewayProxyRequest) error {
		assert.Equal(t, event, e, "original event must not be modified")
		return nil
	})
	_ = fn(context.Background(), event)

	mHandler.AssertExpectations(t)
}

func Test_eventLoggerWithResponse_Wrap_redactsHTTPEventByDefault(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC)
	event := events.APIGatewayProxyRequest{
		Path:    "/test",
		Headers: map[string]string{"Authorization": "Bearer secret", "Content-Type": "application/json"},
		Body:    `{"email":"jane@example.com"}`,
	}
	wantLogged := events.APIGatewayProxyRequest{
		Path:    "/test",
		Headers: map[string]string{"Authorization": redactedValue, "Content-Type": "application/json"},
		Body:    redactedValue,
	}

	mHandler := new(mockSlogHandler)
	mHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
		defaultEventStartedMsg, slog.LevelInfo, []slog.Attr{slog.Any("event", wantLogged)},
	))).Return(nil)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
		defaultEventCompletedMsg, slog.LevelInfo, []slog.Attr{slog.Duration("duration", time.Duration(0))},
	))).Return(nil)

	sut := &eventLoggerWithResponse[events.APIGatewayProxyRequest, any]{
		clock:  clock.NewFixed(now),
		logger: slog.New(mHandler),
		opts:   &defaultEventLoggerOptions,
	}
	fn := sut.Wrap(func(_ context.Context, e events.APIGatewayProxyRequest) (any, error) {
		assert.Equal(t, event, e, "original event must not be modified")
		return nil, nil
	})
	_, _ = fn(context.Background(), event)

	mHandler.AssertExpectations(t)
}

func Test_eventLoggerNoResponse_Wrap_customSanitizer(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC)
	event := events.APIGatewayProxyRequest{Path: "/test"}

	mHandler := new(mockSlogHandler)
	mHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
		defaultEventStartedMsg, slog.LevelInfo, []slog.Attr{slog.Any("event", "sanitized")},
	))).Return(nil)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(matchRecord(
		defaultEventCompletedMsg, slog.LevelInfo, []slog.Attr{slog.Duration("duration", time.Duration(0))},
	))).Return(nil)

	opts := defaultEventLoggerOptions
	WithEventLoggerSanitizer(func(e events.APIGatewayProxyRequest) any {
		return "sanitized"
	})(&opts)

	sut := &eventLoggerNoResponse[events.APIGatewayProxyRequest]{
		clock:  clock.NewFixed(now),
		logger: slog.New(mHandler),
		opts:   &opts,
	}
	fn := sut.Wrap(func(_ context.Context, _ events.APIGatewayProxyRequest) error { return nil })
	_ = fn(context.Background(), event)

	mHandler.AssertExpectations(t)
}

func TestNewEventLogger_defaultsNotMutatedByOptions(t *testing.T) {
	// Create a logger with a custom option — previously this mutated defaultEventLoggerOptions,
	// causing subsequently created loggers to inherit the custom values.
	_ = NewEventLogger[string](slog.New(new(mockSlogHandler)), WithEventLoggerEventStartedMsg("custom"))

	mHandler := new(mockSlogHandler)
	mHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(func(r slog.Record) bool {
		return r.Message == defaultEventStartedMsg && r.Level == defaultEventStartedLevel
	})).Return(nil)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(func(r slog.Record) bool {
		return r.Message == defaultEventCompletedMsg && r.Level == defaultEventCompletedLevel
	})).Return(nil)

	sut := NewEventLogger[string](slog.New(mHandler))
	fn := sut.Wrap(func(_ context.Context, _ string) error { return nil })
	_ = fn(context.Background(), "test")

	mHandler.AssertExpectations(t)
}

func TestNewEventLoggerWithResponse_defaultsNotMutatedByOptions(t *testing.T) {
	// Create a logger with a custom option — previously this mutated defaultEventLoggerOptions,
	// causing subsequently created loggers to inherit the custom values.
	_ = NewEventLoggerWithResponse[string, string](slog.New(new(mockSlogHandler)), WithEventLoggerEventStartedMsg("custom"))

	mHandler := new(mockSlogHandler)
	mHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(func(r slog.Record) bool {
		return r.Message == defaultEventStartedMsg && r.Level == defaultEventStartedLevel
	})).Return(nil)
	mHandler.On("Handle", mock.Anything, mock.MatchedBy(func(r slog.Record) bool {
		return r.Message == defaultEventCompletedMsg && r.Level == defaultEventCompletedLevel
	})).Return(nil)

	sut := NewEventLoggerWithResponse[string, string](slog.New(mHandler))
	fn := sut.Wrap(func(_ context.Context, _ string) (string, error) { return "", nil })
	_, _ = fn(context.Background(), "test")

	mHandler.AssertExpectations(t)
}
