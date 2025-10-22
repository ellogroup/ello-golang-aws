package lambda

import (
	"context"
	"errors"
	"github.com/ellogroup/ello-golang-aws/lambda/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockHandler[E any] struct {
	mock.Mock
}

func (m *mockHandler[E]) Handle(ctx context.Context, event E) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type mockHandlerWithResponse[E, R any] struct {
	mock.Mock
}

func (m *mockHandlerWithResponse[E, R]) Handle(ctx context.Context, event E) (R, error) {
	args := m.Called(ctx, event)
	return args.Get(0).(R), args.Error(1)
}

type idWrapperMiddleware[E []string] struct {
	id string
}

func (i idWrapperMiddleware[E]) Wrap(next func(context.Context, E) error) func(context.Context, E) error {
	return func(ctx context.Context, event E) error {
		return next(ctx, append(event, i.id))
	}
}

type idWrapperMiddlewareWithResponse[E, R []string] struct {
	id string
}

func (i idWrapperMiddlewareWithResponse[E, R]) Wrap(next func(context.Context, E) (R, error)) func(context.Context, E) (R, error) {
	return func(ctx context.Context, event E) (R, error) {
		resp, err := next(ctx, append(event, i.id))
		return append(resp, i.id), err
	}
}

func Test_wrappedHandlerFn(t *testing.T) {
	errHandler := errors.New("errHandler")

	type args[E any] struct {
		handler     Handler[E]
		middlewares []middleware.NoResponse[E]
	}
	type testCase[E any] struct {
		name    string
		args    args[E]
		event   E
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[[]string]{
		{
			name: "no middleware, handler returns nil, returns nil",
			args: args[[]string]{
				handler: func() Handler[[]string] {
					h := new(mockHandler[[]string])
					h.On("Handle", mock.Anything, []string{"event"}).Return(nil)
					return h
				}(),
				middlewares: nil,
			},
			event:   []string{"event"},
			wantErr: assert.NoError,
		},
		{
			name: "no middleware, handler returns error, returns error",
			args: args[[]string]{
				handler: func() Handler[[]string] {
					h := new(mockHandler[[]string])
					h.On("Handle", mock.Anything, []string{"event"}).Return(errHandler)
					return h
				}(),
				middlewares: nil,
			},
			event:   []string{"event"},
			wantErr: assert.Error,
		},
		{
			name: "1 level middleware, event passes through middleware, handler returns nil, returns nil",
			args: args[[]string]{
				handler: func() Handler[[]string] {
					h := new(mockHandler[[]string])
					h.On("Handle", mock.Anything, []string{"event", "middleware-1"}).Return(nil)
					return h
				}(),
				middlewares: []middleware.NoResponse[[]string]{
					idWrapperMiddleware[[]string]{"middleware-1"},
				},
			},
			event:   []string{"event"},
			wantErr: assert.NoError,
		},
		{
			name: "1 level middleware, event passes through middleware, handler returns error, returns error",
			args: args[[]string]{
				handler: func() Handler[[]string] {
					h := new(mockHandler[[]string])
					h.On("Handle", mock.Anything, []string{"event", "middleware-1"}).Return(errHandler)
					return h
				}(),
				middlewares: []middleware.NoResponse[[]string]{
					idWrapperMiddleware[[]string]{"middleware-1"},
				},
			},
			event:   []string{"event"},
			wantErr: assert.Error,
		},
		{
			name: "3 level middleware, event passes through middleware in correct order, handler returns nil, returns nil",
			args: args[[]string]{
				handler: func() Handler[[]string] {
					h := new(mockHandler[[]string])
					h.On("Handle", mock.Anything, []string{"event", "middleware-1", "middleware-2", "middleware-3"}).Return(nil)
					return h
				}(),
				middlewares: []middleware.NoResponse[[]string]{
					idWrapperMiddleware[[]string]{"middleware-1"},
					idWrapperMiddleware[[]string]{"middleware-2"},
					idWrapperMiddleware[[]string]{"middleware-3"},
				},
			},
			event:   []string{"event"},
			wantErr: assert.NoError,
		},
		{
			name: "3 level middleware, event passes through middleware in correct order, handler returns error, returns error",
			args: args[[]string]{
				handler: func() Handler[[]string] {
					h := new(mockHandler[[]string])
					h.On("Handle", mock.Anything, []string{"event", "middleware-1", "middleware-2", "middleware-3"}).Return(errHandler)
					return h
				}(),
				middlewares: []middleware.NoResponse[[]string]{
					idWrapperMiddleware[[]string]{"middleware-1"},
					idWrapperMiddleware[[]string]{"middleware-2"},
					idWrapperMiddleware[[]string]{"middleware-3"},
				},
			},
			event:   []string{"event"},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := wrappedHandlerFn(tt.args.handler, tt.args.middlewares...)
			gotErr := fn(context.Background(), tt.event)
			if !tt.wantErr(t, gotErr, "wrappedHandlerFn(%v, %v)(%v, %v)", tt.args.handler, tt.args.middlewares, context.Background(), tt.event) {
				return
			}
		})
	}
}

func Test_wrappedHandlerWithResponseFn(t *testing.T) {
	errHandler := errors.New("errHandler")

	type args[E any, R any] struct {
		handler     HandlerWithResponse[[]string, []string]
		middlewares []middleware.WithResponse[[]string, []string]
	}
	type testCase[E any, R any] struct {
		name         string
		args         args[E, R]
		event        E
		wantResponse R
		wantErr      assert.ErrorAssertionFunc
	}
	tests := []testCase[[]string, []string]{
		{
			name: "no middleware, handler returns nil, returns nil",
			args: args[[]string, []string]{
				handler: func() HandlerWithResponse[[]string, []string] {
					h := new(mockHandlerWithResponse[[]string, []string])
					h.On("Handle", mock.Anything, []string{"event"}).Return([]string{"response"}, nil)
					return h
				}(),
				middlewares: nil,
			},
			event:        []string{"event"},
			wantResponse: []string{"response"},
			wantErr:      assert.NoError,
		},
		{
			name: "no middleware, handler returns error, returns error",
			args: args[[]string, []string]{
				handler: func() HandlerWithResponse[[]string, []string] {
					h := new(mockHandlerWithResponse[[]string, []string])
					h.On("Handle", mock.Anything, []string{"event"}).Return([]string{"response"}, errHandler)
					return h
				}(),
				middlewares: nil,
			},
			event:        []string{"event"},
			wantResponse: []string{"response"},
			wantErr:      assert.Error,
		},
		{
			name: "1 level middleware, event passes through middleware, handler returns nil, returns nil",
			args: args[[]string, []string]{
				handler: func() HandlerWithResponse[[]string, []string] {
					h := new(mockHandlerWithResponse[[]string, []string])
					h.On("Handle", mock.Anything, []string{"event", "middleware-1"}).Return([]string{"response"}, nil)
					return h
				}(),
				middlewares: []middleware.WithResponse[[]string, []string]{
					idWrapperMiddlewareWithResponse[[]string, []string]{"middleware-1"},
				},
			},
			event:        []string{"event"},
			wantResponse: []string{"response", "middleware-1"},
			wantErr:      assert.NoError,
		},
		{
			name: "1 level middleware, event passes through middleware, handler returns error, returns error",
			args: args[[]string, []string]{
				handler: func() HandlerWithResponse[[]string, []string] {
					h := new(mockHandlerWithResponse[[]string, []string])
					h.On("Handle", mock.Anything, []string{"event", "middleware-1"}).Return([]string{"response"}, errHandler)
					return h
				}(),
				middlewares: []middleware.WithResponse[[]string, []string]{
					idWrapperMiddlewareWithResponse[[]string, []string]{"middleware-1"},
				},
			},
			event:        []string{"event"},
			wantResponse: []string{"response", "middleware-1"},
			wantErr:      assert.Error,
		},
		{
			name: "3 level middleware, event passes through middleware in correct order, handler returns nil, returns nil",
			args: args[[]string, []string]{
				handler: func() HandlerWithResponse[[]string, []string] {
					h := new(mockHandlerWithResponse[[]string, []string])
					h.On("Handle", mock.Anything, []string{"event", "middleware-1", "middleware-2", "middleware-3"}).Return([]string{"response"}, nil)
					return h
				}(),
				middlewares: []middleware.WithResponse[[]string, []string]{
					idWrapperMiddlewareWithResponse[[]string, []string]{"middleware-1"},
					idWrapperMiddlewareWithResponse[[]string, []string]{"middleware-2"},
					idWrapperMiddlewareWithResponse[[]string, []string]{"middleware-3"},
				},
			},
			event:        []string{"event"},
			wantResponse: []string{"response", "middleware-3", "middleware-2", "middleware-1"},
			wantErr:      assert.NoError,
		},
		{
			name: "3 level middleware, event passes through middleware in correct order, handler returns error, returns error",
			args: args[[]string, []string]{
				handler: func() HandlerWithResponse[[]string, []string] {
					h := new(mockHandlerWithResponse[[]string, []string])
					h.On("Handle", mock.Anything, []string{"event", "middleware-1", "middleware-2", "middleware-3"}).Return([]string{"response"}, errHandler)
					return h
				}(),
				middlewares: []middleware.WithResponse[[]string, []string]{
					idWrapperMiddlewareWithResponse[[]string, []string]{"middleware-1"},
					idWrapperMiddlewareWithResponse[[]string, []string]{"middleware-2"},
					idWrapperMiddlewareWithResponse[[]string, []string]{"middleware-3"},
				},
			},
			event:        []string{"event"},
			wantResponse: []string{"response", "middleware-3", "middleware-2", "middleware-1"},
			wantErr:      assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := wrappedHandlerWithResponseFn(tt.args.handler, tt.args.middlewares...)
			gotResp, gotErr := fn(context.Background(), tt.event)
			if !tt.wantErr(t, gotErr, "wrappedHandlerWithResponseFn(%v, %v)(%v, %v)", tt.args.handler, tt.args.middlewares, context.Background(), tt.event) {
				return
			}
			assert.Equalf(t, tt.wantResponse, gotResp, "wrappedHandlerWithResponseFn(%v, %v)(%v, %v)", tt.args.handler, tt.args.middlewares, context.Background(), tt.event)
		})
	}
}
