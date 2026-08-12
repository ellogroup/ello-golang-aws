#!/usr/bin/env bash
#
# init-memory.sh — seed .agents/memory/ from the latest documentation
# templates in the ai-context submodule.
#
# Idempotent: existing memory files are never overwritten, so this is
# safe to re-run after new template files are added upstream.
#
# Normally invoked via `make init-memory`. Run once after creating a
# repo from the template.

set -euo pipefail

SRC_DIR=".ai-context/skills/documentation/assets"
DST_DIR=".agents/memory"
FILES=(progress decisions notes techdebt)

mkdir -p "$DST_DIR"

for f in "${FILES[@]}"; do
    src="$SRC_DIR/$f.md"
    dst="$DST_DIR/$f.md"
    if [[ -f "$dst" ]]; then
        echo "Skipped $dst (already exists)"
    elif [[ ! -f "$src" ]]; then
        echo "Missing $src — is the ai-context submodule initialised?" >&2
        exit 1
    else
        cp "$src" "$dst"
        echo "Seeded $dst from $src"
    fi
done
