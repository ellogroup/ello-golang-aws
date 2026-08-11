# Claude Code Configuration

> Universal agent context for this repository lives in `AGENTS.md`. The
> shared Claude Code configuration (loading order, memory behaviour) lives
> in `.ai-context/CLAUDE.md`. This file is for Claude-specific
> configuration that applies only to this repository.

@AGENTS.md
@.ai-context/CLAUDE.md

---

## Permissions

Tool-level deny rules for this repository are enforced in
`.claude/settings.json`. Do not relax these without team approval.
Developer-local overrides live in `.claude/settings.local.json` and are
gitignored.

---

## Repository-specific Claude behaviour

There is currently no Claude-specific behaviour beyond what is configured
above. If you add slash commands, hooks, or repository-local settings,
document them in this section. Anything tool-agnostic belongs in
`AGENTS.md`.
