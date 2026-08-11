# Ello Go AWS common packages

Common packages for integration with AWS SDK and events

## API Gateway

Common packages for integration with API Gateway 

### Response

Helpers for generating API Gateway V1 responses with a status code and body

```go
// Return a plan text response
return response.New(http.StatusOK, "plain text")

// Return a JSON response
type respBody struct {
    Message string `json:"message"`
}
return response.NewJson(http.StatusOK, respBody{Message: "json response"})

// Return an error response
return response.NewError(http.StatusBadRequest, "error message")
```

## Lambda

Helpers to start a Lambda container with middleware. The middleware will be applied in the order they are found within 
the slice.

```go
// Start a lambda that does not return a response

// handler implements interface lambda.Handler[E any]
// middlewares is a slice of interface middleware.NoResponse[E any]
lambda.Start(handler, middlewares)

// Or...
lambda.Start(handler, middlewares, func() {
    // Callback(s) to run before lambda container is shut down
})

// Start a lambda that does return a response

// handler implements interface lambda.HandlerWithResponse[E, R any]
// middlewares is a slice of interface middleware.WithResponse[E, R any]
lambda.StartWithResponse(handler, middlewares)

// Or...
lambda.StartWithResponse(handler, middlewares, func() {
    // Callback(s) to run before lambda container is shut down
})
```

## Middleware

Middleware allows interaction with incoming events and outgoing responses.

The interfaces `middleware.NoResponse[E any]` and `middleware.WithResponse[E, R any]` can be implemented to add custom 
middleware.

### Context

The context middleware adds additional information to the context of each request using the 
github.com/ellogroup/ello-golang-ctx/logctx package. This includes at the very least a request id.

For API Gateway v1 requests the context also includes the method, domain and path of the request. The response is also 
updated to include the request id within the header `x-request-id`.

### Event Logger

The event logger middleware logs the event start and end using a `*slog.Logger`. The start log record contains the
event and the end log record contains the duration of the request.

For API Gateway v1 requests the end log record also contains the status code of the response.

By default, sensitive HTTP headers (`Authorization`, `Cookie`, `X-Api-Key`) and the request `Body` are automatically
redacted in the event start log record for the following event types: `APIGatewayProxyRequest`,
`APIGatewayV2HTTPRequest`, `ALBTargetGroupRequest`, `LambdaFunctionURLRequest`, and
`APIGatewayWebsocketProxyRequest`. The `Cookies` field is also redacted for event types that carry it as a
dedicated slice. Request/response bodies routinely carry customer PII, so redaction is on unless a route is known
not to need it.

The log messages, levels, and event sanitization can be customised using functional options:

```go
logger := slog.Default()

middleware.NewEventLogger[E](logger)

middleware.NewEventLogger[E](logger,
    middleware.WithEventLoggerEventStartedMsg("Request started"),
    middleware.WithEventLoggerEventCompletedMsg("Request complete"),
    middleware.WithEventLoggerEventStartedLevel(slog.LevelInfo),
    middleware.WithEventLoggerEventCompletedLevel(slog.LevelInfo),
)

// A route that is genuinely public and bodyless-safe to log in full can preserve the body.
// Construct the Redactor once, outside the sanitizer - options are applied at construction,
// not re-processed on every event.
redactor := middleware.NewRedactor(middleware.WithBodyNotRedacted())
middleware.NewEventLogger[events.APIGatewayProxyRequest](logger,
    middleware.WithEventLoggerSanitizer(func(e events.APIGatewayProxyRequest) any {
        return redactor.Redact(e)
    }),
)

// Custom sanitizer — replaces the default header/body redaction entirely.
// Call middleware.RedactHTTPEvent inside your sanitizer to apply built-in redaction as well.
middleware.NewEventLogger[events.APIGatewayProxyRequest](logger,
    middleware.WithEventLoggerSanitizer(func(e events.APIGatewayProxyRequest) any {
        redacted := middleware.RedactHTTPEvent(e)
        // additional custom logic...
        return redacted
    }),
)
```

### Common

There are a selection of common middleware creators for different AWS events.

```go
logger := slog.Default()

middleswares := middleware.CommonS3(logger)

middleswares := middleware.CommonSNS(logger)

middleswares := middleware.CommonSQS(logger)

middleswares := middleware.CommonAPIGatewayV1(logger)
```

## Development

```shell
make build              # go build ./... — compiles the whole module
make format             # gofmt + go fix + goimports -local + go mod tidy
make static-tests       # golangci-lint + gosec + govulncheck
make unit-tests         # go test -v -cover ./...
make build-format-test  # all of the above
```

This repo includes the shared Ello AI-agent tooling (`AGENTS.md`, `CLAUDE.md`,
`.ai-context/`). `.ai-context` is a git submodule, auto-initialised by
`make build`/`make static-tests`/`make unit-tests` (skipped in CI). To pull the
latest shared standards/conventions/skills, run `make sync-ai-context`. To
seed `.agents/memory/` (progress, decisions, notes, tech debt) for a fresh
session, run `make init-memory`.
