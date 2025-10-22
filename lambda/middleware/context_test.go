package middleware

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/ellogroup/ello-golang-ctx/logctx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContext_Wrap(t *testing.T) {
	type args[E any] struct {
		ctx   context.Context
		event E
	}
	type testCase[E any] struct {
		name       string
		c          contextNoResponse[E]
		args       args[E]
		handlerErr error
		wantCtx    *logctx.LogCtx
		wantEvent  E
		wantErr    assert.ErrorAssertionFunc
	}
	tests := []testCase[string]{
		{
			name: "lambda context, request id added to context, handler returns nil, returns nil",
			c:    contextNoResponse[string]{},
			args: args[string]{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
					AwsRequestID: "lambda-request-id-123",
				}),
				event: "test",
			},
			handlerErr: nil,
			wantCtx: &logctx.LogCtx{
				logctx.String("request_id", "lambda-request-id-123"),
				logctx.String("lambda_request_id", "lambda-request-id-123"),
			},
			wantEvent: "test",
			wantErr:   assert.NoError,
		},
		{
			name: "lambda context, request id added to context, handler returns error, returns error",
			c:    contextNoResponse[string]{},
			args: args[string]{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
					AwsRequestID: "lambda-request-id-123",
				}),
				event: "test",
			},
			handlerErr: errors.New("error"),
			wantCtx: &logctx.LogCtx{
				logctx.String("request_id", "lambda-request-id-123"),
				logctx.String("lambda_request_id", "lambda-request-id-123"),
			},
			wantEvent: "test",
			wantErr:   assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := tt.c.Wrap(func(ctx context.Context, event string) error {
				assert.Equalf(t, tt.wantCtx, logctx.Get(ctx), "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event)
				assert.Equalf(t, tt.wantEvent, event, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event)
				return tt.handlerErr
			})
			gotErr := fn(tt.args.ctx, tt.args.event)

			if !tt.wantErr(t, gotErr, "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event) {
				return
			}
		})
	}
}

func TestContextWithResponse_Wrap(t *testing.T) {
	type args[E any] struct {
		ctx   context.Context
		event E
	}
	type testCase[E any, R any] struct {
		name        string
		c           contextWithResponse[E, R]
		args        args[E]
		handlerResp R
		handlerErr  error
		wantCtx     *logctx.LogCtx
		wantEvent   E
		wantResp    R
		wantErr     assert.ErrorAssertionFunc
	}
	tests := []testCase[string, any]{
		{
			name: "lambda context, request id added to context, handler returns string, returns string",
			c:    contextWithResponse[string, any]{},
			args: args[string]{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
					AwsRequestID: "lambda-request-id-123",
				}),
				event: "test-event",
			},
			handlerResp: "test-response",
			handlerErr:  nil,
			wantCtx: &logctx.LogCtx{
				logctx.String("request_id", "lambda-request-id-123"),
				logctx.String("lambda_request_id", "lambda-request-id-123"),
			},
			wantEvent: "test-event",
			wantResp:  "test-response",
			wantErr:   assert.NoError,
		},
		{
			name: "lambda context, request id added to context, handler returns apigw response, returns transformed response",
			c:    contextWithResponse[string, any]{},
			args: args[string]{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
					AwsRequestID: "lambda-request-id-123",
				}),
				event: "test-event",
			},
			handlerResp: events.APIGatewayProxyResponse{
				Headers: map[string]string{},
			},
			handlerErr: nil,
			wantCtx: &logctx.LogCtx{
				logctx.String("request_id", "lambda-request-id-123"),
				logctx.String("lambda_request_id", "lambda-request-id-123"),
			},
			wantEvent: "test-event",
			wantResp: events.APIGatewayProxyResponse{
				Headers: map[string]string{"x-request-id": "lambda-request-id-123"},
			},
			wantErr: assert.NoError,
		},
		{
			name: "lambda context, request id added to context, handler returns error, returns error",
			c:    contextWithResponse[string, any]{},
			args: args[string]{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
					AwsRequestID: "lambda-request-id-123",
				}),
				event: "test-event",
			},
			handlerErr: errors.New("error"),
			wantCtx: &logctx.LogCtx{
				logctx.String("request_id", "lambda-request-id-123"),
				logctx.String("lambda_request_id", "lambda-request-id-123"),
			},
			wantEvent: "test-event",
			wantErr:   assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := tt.c.Wrap(func(ctx context.Context, event string) (any, error) {
				assert.Equalf(t, tt.wantCtx, logctx.Get(ctx), "Wrap(<func>)(%v, %v)", tt.args.ctx, tt.args.event)
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

func Test_contextFromEvent(t *testing.T) {
	type args[E any] struct {
		ctx   context.Context
		event E
	}
	type testCase[E any] struct {
		name      string
		args      args[E]
		wantReqID string
		wantCtx   *logctx.LogCtx
	}
	tests := []testCase[any]{
		{
			name: "string event. empty context, handler returns empty request id and context",
			args: args[any]{
				ctx:   context.Background(),
				event: "test-event",
			},
			wantReqID: "",
			wantCtx: &logctx.LogCtx{
				logctx.String("request_id", ""),
				logctx.String("lambda_request_id", ""),
			},
		},
		{
			name: "string event. lambda context, request id added to context, handler returns request id and context",
			args: args[any]{
				ctx:   lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{AwsRequestID: "lambda-request-id-123"}),
				event: "test-event",
			},
			wantReqID: "lambda-request-id-123",
			wantCtx: &logctx.LogCtx{
				logctx.String("request_id", "lambda-request-id-123"),
				logctx.String("lambda_request_id", "lambda-request-id-123"),
			},
		},
		{
			name: "apigw event. lambda context, request id and details added to context, handler returns request id and context",
			args: args[any]{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{AwsRequestID: "lambda-request-id-123"}),
				event: events.APIGatewayProxyRequest{
					RequestContext: events.APIGatewayProxyRequestContext{
						RequestID:  "amzn-request-id-123",
						HTTPMethod: "POST",
						DomainName: "example.com",
						Path:       "/test/path",
					},
				},
			},
			wantReqID: "amzn-request-id-123",
			wantCtx: &logctx.LogCtx{
				logctx.String("request_id", "amzn-request-id-123"),
				logctx.String("lambda_request_id", "lambda-request-id-123"),
				logctx.String("amzn_request_id", "amzn-request-id-123"),
				logctx.String("request_method", "POST"),
				logctx.String("request_domain", "example.com"),
				logctx.String("request_path", "/test/path"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReqID, gotCtx := contextFromEvent(tt.args.ctx, tt.args.event)
			assert.Equalf(t, tt.wantReqID, gotReqID, "contextFromEvent(%v, %v)", tt.args.ctx, tt.args.event)
			assert.Equalf(t, tt.wantCtx, logctx.Get(gotCtx), "contextFromEvent(%v, %v)", tt.args.ctx, tt.args.event)
		})
	}
}

func Test_transformResponse(t *testing.T) {
	type args[R any] struct {
		response  R
		requestId string
	}
	type testCase[R any] struct {
		name string
		args args[R]
		want R
	}
	tests := []testCase[any]{
		{
			name: "string response, returns string unmodified",
			args: args[any]{
				response:  "test-response",
				requestId: "test-request-id",
			},
			want: "test-response",
		},
		{
			name: "apigw response, returns apigw response with request id header",
			args: args[any]{
				response:  events.APIGatewayProxyResponse{Headers: map[string]string{}},
				requestId: "test-request-id",
			},
			want: events.APIGatewayProxyResponse{Headers: map[string]string{"x-request-id": "test-request-id"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, transformResponse(tt.args.response, tt.args.requestId), "transformResponse(%v, %v)", tt.args.response, tt.args.requestId)
		})
	}
}
