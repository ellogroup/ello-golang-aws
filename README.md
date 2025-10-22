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

### Metrics

The metrics middleware outputs the request start and end. The request start output contains the event and the request 
end output contains the duration of the request.

For API Gateway v1 requests the output also contains the status code of the response.

Outputters need to implement the `metrics.Outputter` interface from the github.com/ellogroup/ello-golang-metrics/metrics
package.

### Common

There are a selection of common middleware creators for different AWS events.

```go
// outputter implements metrics.Outputter

middleswares := middleware.CommonS3(outputter)

middleswares := middleware.CommonSNS(outputter)

middleswares := middleware.CommonSQS(outputter)

middleswares := middleware.CommonAPIGatewayV1(outputter)
```
