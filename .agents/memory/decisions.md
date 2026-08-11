# Architecture Decision Records

A log of significant technical decisions made during development of this
application. Each entry explains what was decided, why, and what alternatives
were considered.

Entries are append-only — once recorded, an ADR is never deleted. If a
decision is reversed, a new ADR is added with status Superseded and a
reference to the original.

---

## Template

### ADR-[NNN] — [Short descriptive title]

**Date:** [DATE]
**Status:** [Proposed | Accepted | Superseded by ADR-NNN]
**Author:** [Developer name or "AI-assisted — reviewed by [name]"]

**Context:**
[What situation or requirement led to this decision? What constraints existed?
What problem were we solving?]

**Decision:**
[What was decided? Be specific — avoid vague statements like "we chose the
better approach".]

**Rationale:**
[Why was this the right decision given the context? What made it preferable to
the alternatives?]

**Alternatives Considered:**

| Alternative | Reason rejected |
|---|---|
| [Option] | [Why it was not chosen] |

**Consequences:**
[What are the trade-offs? What does this decision make easier or harder in
future? Are there follow-on decisions that will need to be made?]

**Related:**
- [Link to relevant ticket, PR, or other ADR if applicable]

---

### ADR-001 — Retrofit the AI-integration/Go-tooling runbook onto this repo, adapted for a library shape

**Date:** 2026-08-11
**Status:** Accepted
**Author:** AI-assisted — reviewed by Symeon Manis

**Context:**
This is the first shared Go *library* (not a deployable app) to go through
the "Retrofitting the Go App Template (AI Integration + Go Tooling)"
runbook, raised during review of PR #5 (BE-675). The runbook assumes an
app-shaped repo throughout: a Go module rooted in `app/`, a separate
`test/` integration-test module, `infrastructure/` Terraform, and
`format-app`/`format-test`-style Makefile targets. This repo has none of
that — it's a single flat Go module at the repo root, no infra, no
integration tests, tested via native GitHub Actions rather than a
Makefile/LocalStack loop.

**Decision:**
Applied every runbook step at the repo root instead of per-module, and
skipped anything that only makes sense for a deployable app:
- `.ai-context` submodule, `CLAUDE.md`/`.aiignore`/`.markdownlint.yml`
  copied as documented.
- `AGENTS.md`'s "This repository" section written from scratch, not
  copied — describes this as a library other backend repos import, not
  an app, and explicitly notes that `.ai-context/teams/backend/AGENTS.md`
  is written from the app-consumer's perspective, not this repo's.
- `.aiignore` copied minus the app-shaped entries (`local/localstack/.volume/`,
  `app/bin/`) that can't exist here.
- Added a single root-level `.golangci.yml` (this repo previously had no
  linter config file at all — CI ran `golangci-lint-action` with pure
  defaults). Kept `sloglint` (real usage — the event-logger middleware is
  built on `log/slog`); dropped `spancheck` (zero OTel/span usage
  anywhere in this repo), per the runbook's own instruction to check
  actual dependencies before copying a linter set blindly.
- Makefile: added `ensure-ai-context`/`sync-ai-context`/`sync-skills`/
  `init-memory` targets and wired `build: ensure-ai-context`, matching
  the runbook. The "verify the Docker mount actually persists changes"
  and "drop `-it`" gotchas the runbook calls out were already satisfied
  by an earlier, unrelated Makefile realignment on this same branch.
- Fixed all 24 pre-existing lint findings the new config surfaced (see
  `notes.md`), except one (`apigw/response.NewJson` naming) deliberately
  deferred — see TD-001 in `techdebt.md`.

**Rationale:**
Retrofitting is more valuable done for real (with the actual gaps and
adaptations documented) than skipped because the runbook doesn't
literally fit this repo's shape. Every adaptation is small and mechanical
once you know the repo has no `app/`/`test/` split — the risk of doing
this retrofit was mostly in silently copying app-shaped assumptions
(Terraform paths, integration-test references) that don't apply here.

**Alternatives Considered:**

| Alternative | Reason rejected |
|---|---|
| Skip the retrofit, since the runbook doesn't cover library repos | Leaves this repo permanently without AI-agent context, memory, or a real linter config, and duplicates the "is this worth doing for a library" question for the next library repo instead of answering it once |
| Copy the runbook's app-shaped files/targets verbatim, unadapted | Would add dead references (`app/`, `test/`, `infrastructure/`) that mislead a future session into thinking this repo has structure it doesn't |

**Consequences:**
The shared `.ai-context/teams/backend/AGENTS.md` still has no dedicated
"shared library repo" category — it was written entirely from the
app-consumer's perspective. Worth raising with whoever owns `.ai-context`
so the next library retrofit doesn't have to rediscover this from
scratch. `apigw/response.NewJson`'s naming fix remains deferred (TD-001) —
picking it up requires a coordinated update across every consuming repo,
not just this one.

**Related:**
- BE-675 — PR #5 review discussion, where this retrofit was first raised
- TD-001 in `techdebt.md` — the deferred `NewJson`→`NewJSON` rename

---
