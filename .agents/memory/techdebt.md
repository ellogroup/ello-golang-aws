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
| Medium | 0 | 0 | 1 | 1 |
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

**Status:** Resolved (option b)
**Severity:** Medium
**Category:** Code Quality
**Created:** 2026-08-11
**Resolved:** 2026-08-12
**Created by:** AI-assisted — reviewed by Symeon Manis
**Owner:** Backend team
**Linked ticket:** BE-676

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

**Resolution taken:**
Option (b) — added `NewJSON` as a new function with the real
implementation; `NewJson` is now a thin deprecated wrapper delegating to
it (`//nolint:revive` kept on the alias, `//nolint:staticcheck` on the
one test exercising it intentionally). Existing consumers keep working
unchanged and can migrate to `NewJSON` at their own pace; `NewJson` can
be removed in a future major version once consumers have moved over.

---

### TD-002 — `apigw/response.ErrorCode` set is domain-specific, not generic

**Status:** Resolved (option b, extended)
**Severity:** Medium
**Category:** Architecture
**Created:** 2026-08-21
**Resolved:** 2026-08-21
**Created by:** AI-assisted — reviewed by Dave Richards
**Owner:** Backend team
**Linked ticket:** None

**What is the debt?**
`apigw/response.ErrorCode` and its registered definitions
(`ErrorCodeCustomerNotFound`, `ErrorCodeOfferNotFound`,
`ErrorCodeRedemptionNotFound`, `ErrorCodeCustomerAlreadyExists`,
`ErrorCodeOfferNotRedeemable`) are named after entities specific to the
Ello B2B API (customers, offers, redemptions) — copied literally from
that API's OpenAPI spec (Confluence: "OpenAPI Specification", space EP,
page 1223098369). `ello-golang-aws` is a shared, generic library
imported by multiple backend repos, so baking one API's entity-specific
codes into it doesn't generalise: any other consuming service with
different domain entities either can't use `NewErrorCode` for its own
"not found" / "already exists" cases, or this library has to keep
growing a new `ErrorCode` constant + definition per entity per
consumer, which is the wrong place for that coupling.

**Why does it exist?**
Added while implementing `NewError`/`NewErrorCode` to match the Ello
B2B API's documented error shape exactly, so callers reporting the same
error produce the same status/code/message. The BA was on annual leave
at the time, so we couldn't confirm whether the OpenAPI spec's
per-entity codes (`customer_not_found`, `offer_not_found`, ...) are a
fixed external contract we must mirror exactly, or whether a more
generic scheme (e.g. `not_found` / `already_exists` parameterised by
entity) would be acceptable there too. We went with the literal,
spec-matching set as the safe default in the meantime.

**What is the risk if unaddressed?**
Every new "not found" / "already exists" case, in this API or a
different consuming service, needs a new `ErrorCode` constant and
`errorCodeDefinitions` entry added to this shared library — coupling a
generic package to specific business domains it shouldn't need to know
about, and this set will not transfer as-is to a service with different
entities.

**Resolution taken:**
Option (b), extended into a general mechanism rather than a one-off move.
`ErrorCode` changed from an `iota`-based `int` (closed, compile-time-only
set) to `string` (the wire code is now the identity itself), backed by an
open, `sync.RWMutex`-guarded registry. `RegisterErrorCode`/
`MustRegisterErrorCode` let any consuming application register its own
`ErrorCode` definitions once at startup (e.g. from an `init` function),
with `NewErrorCode` working identically for built-in and app-registered
codes — no separate "generator" struct to construct and thread through
handlers; this package's own built-ins register themselves through the
same public function. The 5 Ello-B2B-specific codes
(`ErrorCodeCustomerAlreadyExists`, `ErrorCodeCustomerNotFound`,
`ErrorCodeOfferNotFound`, `ErrorCodeOfferNotRedeemable`,
`ErrorCodeRedemptionNotFound`) and `FieldErrorCodeOfferWithdrawn` were
removed from this library entirely; only the truly generic codes remain
(`ErrorCodeValidationFailed`, `ErrorCodeUnauthorized`,
`ErrorCodeRateLimited`, `ErrorCodeInternalError`).

**Follow-up required in the consuming Ello B2B API repo:** it must define
and register its own `ErrorCode` constants (mirroring the 5 removed ones)
via `response.RegisterErrorCode`/`MustRegisterErrorCode` in its own
startup code, and update its handlers from `response.ErrorCodeCustomerNotFound`
(etc.) to its own package's equivalent, before it can upgrade to this
version of `ello-golang-aws`. Nothing has been tagged/released yet, so no
consumer is broken by this today, but this is a breaking change for
whenever that repo does upgrade.

---
