DOCKER_RUN = docker run --rm -i --platform linux/amd64 -v $(CURDIR):/src -w /src
GO_DEV_TOOLS = diningclub/golang-dev-tools:latest

.PHONY: build-format-test
build-format-test: build format test

.PHONY: build
build: ensure-ai-context
	$(DOCKER_RUN) $(GO_DEV_TOOLS) go build ./...

# Initialise the ai-context submodule if it is missing. Skipped in CI —
# pipelines do not need the AI agent context to build or test the app.
.PHONY: ensure-ai-context
ensure-ai-context:
	@if [ ! -f .ai-context/AGENTS.md ] && [ -z "$$CI" ]; then \
		echo "Initialising ai-context submodule..."; \
		git submodule update --init --depth 1 .ai-context || true; \
	fi

# Pull the latest shared AI context. Run this when you want the latest
# standards, conventions, and skills from ellogroup/ai-context. After
# bumping the submodule, sync the Claude Code skill wrappers so any
# newly-added skills get a wrapper too.
.PHONY: sync-ai-context
sync-ai-context:
	git submodule update --remote --merge .ai-context
	$(MAKE) sync-skills
	@echo "ai-context updated. Review .ai-context/ and commit the pointer if appropriate."

# Generate one-line Claude Code skill wrappers under .claude/skills/<name>/
# for every skill in .ai-context/skills/ that declares a `command:` field.
# Idempotent — existing wrappers are skipped, never overwritten or deleted.
.PHONY: sync-skills
sync-skills: ensure-ai-context
	@./scripts/sync-skills.sh

# Seed .agents/memory/ from the latest documentation templates in the
# ai-context submodule. Wraps scripts/init-memory.sh — idempotent,
# existing files are never overwritten.
.PHONY: init-memory
init-memory: ensure-ai-context
	@./scripts/init-memory.sh

.PHONY: format
format:
	$(DOCKER_RUN) $(GO_DEV_TOOLS) sh -c "gofmt -w ./ && go fix ./... && goimports -local github.com/ellogroup -w ./ && go mod tidy"

.PHONY: test
test: static-tests unit-tests

.PHONY: static-tests
static-tests:
	$(DOCKER_RUN) $(GO_DEV_TOOLS) sh -c "golangci-lint config verify && golangci-lint run -v && gosec ./... && govulncheck ./..."

.PHONY: unit-tests
unit-tests:
	$(DOCKER_RUN) $(GO_DEV_TOOLS) go test -v -cover ./...