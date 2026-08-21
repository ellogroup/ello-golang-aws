# Agent Context — This Repository

> Read this file first, then follow the reading order below. The shared
> standards for all Ello repositories live in the `ai-context` submodule at
> `.ai-context/`. If `.ai-context/` is empty, run `make sync-ai-context`
> before continuing.

---

## Reading order

1. `.ai-context/AGENTS.md` — universal standards for all teams.
2. `.ai-context/standards/security.md` — absolute security constraints.
3. `.ai-context/teams/backend/AGENTS.md` — Go / Lambda / API Gateway conventions for backend repos. Read this with one caveat: it's written from the perspective of a deployable app (`app/` module root, `infrastructure/` Terraform, one Lambda per route) that *consumes* this repo as a dependency — this repo has no such shape (see below).
4. `.ai-context/skills/documentation/SKILL.md` — how to maintain `.agents/memory/`.
5. `.ai-context/skills/spec/SKILL.md` — spec-driven development workflow.
6. This repository's `README.md` — usage examples and package overview.
7. `.agents/memory/progress.md`, `decisions.md`, `notes.md`, `techdebt.md` — session memory. If absent, run `make init-memory` to seed them from `.ai-context/skills/documentation/assets/`.

Load other documents under `.ai-context/` on demand — for example
`.ai-context/teams/backend/conventions/go.md` before writing Go.
`.ai-context/teams/backend/conventions/api-design.md` and
`.ai-context/standards/ci-cd.md` don't apply here — this repo has no HTTP
API of its own and no Terraform-managed CI/CD pipeline (see below).

---

## This repository

This repository is **`ello-golang-aws`**, a shared Go library — not a
deployable service. It provides AWS Lambda / API Gateway middleware and
response helpers that other Ello backend repos import as a Go module
dependency (`github.com/ellogroup/ello-golang-aws/v2`), versioned via git tags.

There is no `app/` module root, no `infrastructure/` (this repo owns no cloud
resources of its own), and no `test/` integration-test module — the Go module
is rooted at the repo root, and testing is unit-tests-only (`go test ./...`),
run via native GitHub Actions (`.github/workflows/tests.yml`), not a
Makefile-driven local/LocalStack loop. Most of `.ai-context/teams/backend/`'s
app-shaped guidance (Terraform, integration tests, one Lambda per route) does
not apply here — this repo *is* one of the primitives that guidance assumes
apps import, not an app itself.

| Concern | Where to look |
|---|---|
| Local commands | `Makefile`, `README.md` |
| API Gateway response helpers | `apigw/response/` — `New`, `NewJson`, `NewError` |
| Lambda start + middleware chaining | `lambda/` — `Start`, `StartWithResponse` |
| Middleware implementations | `lambda/middleware/` — request context (`context.go`), structured event logging (`eventlogger.go`), sensitive-data redaction (`redact.go`) |
| Session memory / handoff notes | `.agents/memory/` |

This section is intentionally thin on first retrofit. As the package surface
grows — new middleware, new event types, non-obvious design decisions —
update this section so a new session can orient in 30 seconds without
rediscovering it. This is a stable overview, not a change log: don't
duplicate `.agents/memory/decisions.md` or `progress.md` here, just keep the
high-level "what does this library provide and how is it shaped" summary
current.

---

## Updating shared context

`.ai-context/` is a git submodule pinned to a specific commit. To pull the
latest shared context:

```bash
make sync-ai-context
```

Treat each update as a dependency bump — review the diff before committing
the new submodule pointer.
