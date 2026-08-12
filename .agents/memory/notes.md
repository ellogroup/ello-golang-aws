# Notes

Working notes, gotchas, lessons learned, and useful references accumulated
during development of this application.

Entries are append-only within each section. The Scratch section may be
cleared between sessions if it no longer contains relevant content.

---

## Gotchas

Things that caused unexpected problems and how they were resolved. Future
sessions and developers should read this before starting work.

- **[DATE] [Short title]:** [What the problem was, what caused it, how it was
  resolved. Be specific enough that someone encountering the same issue can
  recognise it and know the fix.]

- **2026-08-11 the AI-tooling runbook's lint-finding cap hid real findings, exactly as it warns:** Enabling the new `.golangci.yml` for the first time surfaced 22 issues on a capped `golangci-lint run -v`, but re-running uncapped (`--max-same-issues=0 --max-issues-per-linter=0`) revealed 24 — 2 more `forcetypeassert` findings (`lambda/lambda_test.go:29`, `lambda/middleware/eventlogger_test.go:31`) had been hidden by the default per-linter cap. Always do the uncapped run before declaring a retrofit's lint pass clean, exactly as the runbook says — the cap hid findings on the very first repo this happened to be checked on.

---

## Lessons Learned

Patterns and approaches that proved effective or ineffective. Worth knowing
for next time.

- **[DATE] [Short title]:** [What was learned and why it matters.]

- **2026-08-11 the AI-tooling retrofit runbook translates cleanly to a library repo, once you stop looking for app/ and test/:** Every step applies at the repo root instead of per-module; the only step that needed real judgment was step 5 (linter config) — this repo had zero `.golangci.yml` before this retrofit (CI ran `golangci-lint-action` with pure defaults), so "check the app's actual dependencies before copying" meant deciding sloglint/spancheck relevance from scratch rather than adjusting an existing config. See ADR-001 for the full adaptation.

---

## Useful References

Links and resources that are specifically relevant to this application — not
general documentation, but things that were actually useful during development.

- **[Title]:** [URL or reference] — [Why it is useful and in what context.]

---

## Scratch

Temporary working notes from the current or recent session. May be cleared
once no longer relevant.

[Working notes here.]
