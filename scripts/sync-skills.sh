#!/usr/bin/env bash
#
# sync-skills.sh — generate Claude Code skill wrappers under
# .claude/skills/<name>/SKILL.md for every skill in .ai-context/skills/
# (universal) and .ai-context/teams/backend/skills/ (team-owned — this
# is a backend-team template repo) that declares a `command:` field in
# its frontmatter.
#
# Each wrapper is a one-line pointer to the canonical SKILL.md in
# ai-context. Prose updates upstream propagate without anything to
# regenerate here.
#
# Team-owned skill directories that only override assets (no SKILL.md
# of their own — e.g. teams/backend/skills/spec/assets/) are skipped
# naturally, since the glob below requires a SKILL.md to be present.
#
# Idempotent: existing wrappers are skipped, never overwritten or
# deleted.
#
# Per-repo skills under .agents/skills/ are not scanned — those are
# managed manually.
#
# Normally invoked via `make sync-skills`.

set -euo pipefail

SKILLS_DIR=".ai-context/skills"
BACKEND_SKILLS_DIR=".ai-context/teams/backend/skills"
TARGET_DIR=".claude/skills"

if [[ ! -d "$SKILLS_DIR" ]]; then
    echo "Error: $SKILLS_DIR not found — is the ai-context submodule initialised?" >&2
    exit 1
fi

mkdir -p "$TARGET_DIR"

created=0
skipped=0

for skill_file in "$SKILLS_DIR"/*/SKILL.md "$BACKEND_SKILLS_DIR"/*/SKILL.md; do
    [[ -f "$skill_file" ]] || continue

    # Extract the `command:` value from the YAML frontmatter only
    # (between the first two `---` markers). Strip the leading slash.
    command=$(awk '
        /^---$/ { if (++c == 2) exit; next }
        c == 1 && /^command:[[:space:]]*\// {
            gsub(/^command:[[:space:]]*\//, "")
            gsub(/[[:space:]]+$/, "")
            print
            exit
        }
    ' "$skill_file")

    # Skills with no `command:` field are auto-triggered — no wrapper.
    [[ -z "$command" ]] && continue

    wrapper_dir="$TARGET_DIR/$command"
    wrapper="$wrapper_dir/SKILL.md"

    if [[ -e "$wrapper" ]]; then
        skipped=$((skipped + 1))
        continue
    fi

    mkdir -p "$wrapper_dir"
    cat > "$wrapper" <<EOF
Read and follow \`$skill_file\`.
EOF
    echo "→ Created $wrapper (→ $skill_file)"
    created=$((created + 1))
done

echo
echo "Done. $created wrapper(s) created, $skipped already present."
