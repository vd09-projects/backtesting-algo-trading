#!/usr/bin/env bash
# Fires on SessionEnd (when the Claude Code session exits/clears).
# Cannot block — informational only. Prints harvest reminders and
# uncommitted change summary so nothing is forgotten before closing.
set -euo pipefail

PROJECT="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT"

cat > /dev/null  # consume stdin

echo ""
echo "=== SESSION ENDING ==="
echo ""
echo "MANDATORY HARVESTS — fire both before closing:"
echo "  /task-manager     — implicit tasks from this session"
echo "  /decision-journal — inline decision marks"
echo ""

DIRTY=$(git diff --name-only 2>/dev/null)
STAGED=$(git diff --cached --name-only 2>/dev/null)
AHEAD=$(git rev-list --count origin/main..HEAD 2>/dev/null || echo "?")

if [ -n "$DIRTY" ] || [ -n "$STAGED" ]; then
  echo "UNCOMMITTED CHANGES:"
  [ -n "$STAGED" ] && echo "$STAGED" | sed 's/^/  [staged] /'
  [ -n "$DIRTY"  ] && echo "$DIRTY"  | sed 's/^/  [unstaged] /'
  echo ""
fi

if [ "$AHEAD" != "0" ] && [ "$AHEAD" != "?" ]; then
  echo "Commits ahead of origin/main: $AHEAD (not yet pushed)"
fi
