DOCKER_RUN = docker run --rm -i --platform linux/amd64 -v $(CURDIR):/src -w /src
GO_DEV_TOOLS = diningclub/golang-dev-tools:latest

.PHONY: build-format-test
build-format-test: build format test

.PHONY: build
build:
	$(DOCKER_RUN) $(GO_DEV_TOOLS) go build ./...

.PHONY: format
format:
	$(DOCKER_RUN) $(GO_DEV_TOOLS) sh -c "gofmt -w ./ && go fix ./... && goimports -local github.com/ellogroup -w ./ && go mod tidy"

.PHONY: test
test: static-tests unit-tests

.PHONY: static-tests
static-tests:
	$(DOCKER_RUN) $(GO_DEV_TOOLS) sh -c "golangci-lint run -v && gosec ./... && govulncheck ./..."

.PHONY: unit-tests
unit-tests:
	$(DOCKER_RUN) $(GO_DEV_TOOLS) go test -v -cover ./...