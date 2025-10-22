package response

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		status int
		body   string
	}
	tests := []struct {
		name string
		args args
		want events.APIGatewayProxyResponse
	}{
		{
			name: "status and body set, returns response",
			args: args{123, "test-123"},
			want: events.APIGatewayProxyResponse{StatusCode: 123, Body: "test-123", Headers: map[string]string{}},
		},
		{
			name: "status and body empty, returns response",
			args: args{0, ""},
			want: events.APIGatewayProxyResponse{StatusCode: 0, Body: "", Headers: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, New(tt.args.status, tt.args.body), "New(%v, %v)", tt.args.status, tt.args.body)
		})
	}
}

func TestNewJson(t *testing.T) {
	type testBody struct {
		Num int    `json:"num"`
		Str string `json:"str"`
	}
	type args struct {
		status int
		body   any
	}
	tests := []struct {
		name string
		args args
		want events.APIGatewayProxyResponse
	}{
		{
			name: "status and body set, returns json response",
			args: args{123, testBody{456, "test-789"}},
			want: events.APIGatewayProxyResponse{StatusCode: 123, Body: `{"num":456,"str":"test-789"}`, Headers: map[string]string{
				"Content-Type": "application/json",
			}},
		},
		{
			name: "status but no body set, returns non-json response",
			args: args{status: 123},
			want: events.APIGatewayProxyResponse{StatusCode: 123, Body: "", Headers: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewJson(tt.args.status, tt.args.body), "NewJson(%v, %v)", tt.args.status, tt.args.body)
		})
	}
}
