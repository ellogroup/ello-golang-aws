# Progress

**Last Updated:** 2026-08-11

---

## Current Status

BE-675 (redact request body by default in the event-logger middleware) is
implemented and reviewed on PR #5; the AI-integration/Go-tooling retrofit
(this repo's first as a library, not an app) has been added to the same
branch and is locally verified clean (build, format, uncapped lint, gosec,
govulncheck, unit tests).

---

## Completed

- BE-675: `RedactHTTPEvent` redacts the `Body` field by default for all 5
  supported HTTP event types; added `WithBodyNotRedacted()` as a
  non-breaking opt-out.
- Makefile/`.gitignore` realigned with the newer app repos' pattern
  (mount the working tree against the shared `diningclub/golang-dev-tools`
  image directly; `format` now runs `go fix`/`goimports -local`/`go mod
  tidy`); removed the now-orphaned `Dockerfile`.
- AI-integration/Go-tooling retrofit: `.ai-context` submodule, root files
  (`CLAUDE.md`, `AGENTS.md` — written from scratch, not copied,
  `.aiignore`, `.markdownlint.yml`), `.claude/settings.json`, skill
  wrappers (`make sync-skills`), seeded memory (`make init-memory`), a new
  root `.golangci.yml` (this repo had none before), `ensure-ai-context`/
  `sync-ai-context`/`sync-skills`/`init-memory` Makefile targets.
- Fixed all 24 pre-existing lint findings the new linter config surfaced
  (uncapped run) except one deliberately deferred — see TD-001.

---

## In Progress

- None — ready for another review pass on PR #5.

---

## Next

- Push and get the retrofit changes reviewed alongside the BE-675 fix.
- Raise with whoever owns `.ai-context`: `teams/backend/AGENTS.md` has no
  "shared library repo" category — it's written entirely from the
  app-consumer's perspective, and this retrofit had to write `AGENTS.md`'s
  "This repository" section from scratch as a result. Worth adding
  library-repo guidance to the runbook/`.ai-context` so the next library
  retrofit doesn't rediscover this.
- TD-001 (`apigw/response.NewJson` naming) — backlog, needs a coordinated
  change across every consuming repo, not just this one.

---

## Blockers

- None currently.

---

## Session Log

### 2026-08-11 — BE-675 fix + first library retrofit of the AI-tooling runbook

Implemented BE-675 (default body redaction in the event-logger
middleware), then — per review feedback on PR #5 — took this repo through
the AI-integration/Go-tooling retrofit as the first shared library (not
app) to go through it. The runbook assumes an app-shaped repo throughout
(`app/`+`test/` module split, `infrastructure/` Terraform); adapted every
step to apply at the repo root instead, and skipped anything infra/
integration-test-specific. See ADR-001 for the full rationale and
adaptation list, and the Gotchas/Lessons Learned sections in `notes.md`
for the uncapped-lint-cap gotcha and the library-shape translation notes.
Full local verification (`make build-format-test`, uncapped
`golangci-lint`, `gosec`, `govulncheck`) is clean.