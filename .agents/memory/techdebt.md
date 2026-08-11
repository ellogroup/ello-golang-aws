# Tech Debt Register

A structured log of known technical debt in this application. Entries are
created at the point debt is identified — not retrospectively.

Honest entries are more useful than polished ones. The reason a piece of debt
exists is as important as the debt itself.

---

## Summary

| Severity | Open | In Progress | Resolved | Accepted |
|---|---|---|---|---|
| Critical | 0 | 0 | 0 | 0 |
| High | 0 | 0 | 0 | 0 |
| Medium | 0 | 0 | 0 | 1 |
| Low | 0 | 0 | 0 | 0 |

---

## Template

### TD-[NNN] — [Short descriptive title]

**Status:** [Open | In Progress | Resolved | Accepted]
**Severity:** [Critical | High | Medium | Low]
**Category:** [Security | Performance | Architecture | Code Quality | Dependencies | Operational]
**Created:** [DATE]
**Created by:** [Developer name or "AI-assisted — reviewed by [name]"]
**Owner:** [Team or person responsible for resolution]
**Linked ticket:** [Ticket reference or None]

**What is the debt?**
[Clear, specific description of the problem. Not "this could be better" but
what exactly is wrong or missing.]

**Why does it exist?**
[Honest reason — time pressure, unclear requirements, legacy decision, known
shortcut taken deliberately, etc. This is not a blame record — it is context.]

**What is the risk if unaddressed?**
[Concrete impact. Not "it might cause problems" but what specifically breaks,
degrades, or becomes harder as a result of leaving this unresolved.]

**Proposed resolution:**
[What needs to be done to fix it. Does not need to be a full solution — a
direction is sufficient.]

**Effort estimate:** [S — hours | M — days | L — week+ | XL — significant]
**Resolution target:** [Current sprint | This quarter | Next quarter | Backlog | Accepted risk]

---

### TD-001 — `apigw/response.NewJson` violates Go naming convention (should be `NewJSON`)

**Status:** Accepted
**Severity:** Medium
**Category:** Code Quality
**Created:** 2026-08-11
**Created by:** AI-assisted — reviewed by Symeon Manis
**Owner:** Backend team
**Linked ticket:** None

**What is the debt?**
`revive`'s `var-naming` rule (enabled as part of this repo's first
AI-integration/Go-tooling retrofit) flags `apigw/response.NewJson` —
initialisms should be all-caps per Go convention, so it should be
`NewJSON`. Suppressed with `//nolint:revive` on the declaration rather
than fixed.

**Why does it exist?**
`NewJson` is a public, exported function already imported by every
consuming backend repo (`ello-membership-service-app`, `ello-vendr-app`,
and others) as `response.NewJson(...)`. Renaming it is a breaking API
change for every one of those repos, not a safe mechanical fix like the
rest of this retrofit's lint findings — it needs a coordinated,
deliberate rollout, not a drive-by rename inside an unrelated retrofit.

**What is the risk if unaddressed?**
None functionally — this is a naming-convention violation, not a bug.
The only risk is the inconsistency reads oddly next to correctly-named
exports elsewhere in the same package.

**Proposed resolution:**
Either (a) rename to `NewJSON` as a deliberate breaking change, bump
this module's version accordingly, and update every consuming repo's
import in the same coordinated effort, or (b) add `NewJSON` as a new
function doing the same thing, mark `NewJson` deprecated via its GoDoc
comment, and let consumers migrate at their own pace before removing
`NewJson` in a future major version.

**Effort estimate:** M — days (coordinated change across every consuming repo, not just this one)
**Resolution target:** Backlog

---
