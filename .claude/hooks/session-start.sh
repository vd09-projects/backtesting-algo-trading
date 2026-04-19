#!/usr/bin/env bash
# Fires on SessionStart (startup + resume). Injects orientation context into Claude's
# first turn: in-progress tasks, uncommitted changes, quality-gate status, mandatory rules.
set -euo pipefail

PROJECT="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT"

# Consume stdin (Claude Code pipes JSON input; we don't need it here)
cat > /dev/null

echo "=== SESSION START ==="
echo ""

# --- In-progress tasks ---
IN_PROGRESS=$(grep -B2 "Status:.*in-progress" tasks/BACKLOG.md 2>/dev/null \
  | grep "TASK-" | sed 's/.*\[/[/;s/\].*/]/' | tr '\n' ' ')
if [ -n "$IN_PROGRESS" ]; then
  echo "IN PROGRESS: $IN_PROGRESS"
else
  echo "IN PROGRESS: none"
fi

# --- Uncommitted changes ---
DIRTY=$(git diff --name-only 2>/dev/null | head -8)
STAGED=$(git diff --cached --name-only 2>/dev/null | head -4)
if [ -n "$DIRTY" ] || [ -n "$STAGED" ]; then
  echo "UNCOMMITTED CHANGES:"
  [ -n "$STAGED" ] && echo "$STAGED" | sed 's/^/  [staged] /'
  [ -n "$DIRTY"  ] && echo "$DIRTY"  | sed 's/^/  [unstaged] /'
fi

# --- Quality gate sentinel status ---
SENTINEL="$PROJECT/.quality-gate/last-pass"
if [ -f "$SENTINEL" ]; then
  PENDING=$(find internal/ pkg/ -name "*.go" ! -name "*_test.go" \
    -newer "$SENTINEL" 2>/dev/null | head -3)
  if [ -n "$PENDING" ]; then
    echo ""
    echo "QUALITY GATE PENDING: production code changed since last /go-quality-review:"
    echo "$PENDING" | sed 's/^/  /'
  fi
else
  HAS_CODE=$(find internal/ pkg/ -name "*.go" ! -name "*_test.go" 2>/dev/null | head -1)
  if [ -n "$HAS_CODE" ]; then
    echo ""
    echo "QUALITY GATE: no sentinel found — run /go-quality-review to establish baseline"
  fi
fi

# --- Mandatory rules reminder ---
echo ""
echo "MANDATORY GATES (CLAUDE.md — non-optional):"
echo "  1. TDD          — write the failing test BEFORE implementation"
echo "  2. quality-gate — /go-quality-review before marking done if internal/ or pkg/ touched"
echo "  3. workflows    — read workflows/INDEX.md after every skill terminal state"
echo "  4. decisions    — /decision-journal for any design choice made"
echo "  5. session-end  — /task-manager + /decision-journal harvests both fire"
echo ""
echo "Next: read workflows/INDEX.md to determine which workflow applies to this session."
