#!/usr/bin/env bash
# Fires on PostToolUse when the Skill tool completes.
# Maps which skill just finished to the expected next build-workflow step,
# injecting a concrete routing suggestion into Claude's context before the
# next turn. Cannot block (PostToolUse is observational only).
#
# NOTE: this hook relies on the matcher "Skill" firing for the Skill tool.
# If the tool_name does not match, the hook exits silently — no harm done.
set -euo pipefail

PROJECT="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT"

INPUT=$(cat)

TOOL_NAME=$(echo "$INPUT" | python3 -c \
  "import json,sys; d=json.load(sys.stdin); print(d.get('tool_name',''))" \
  2>/dev/null || echo "")

# Only act if this is actually the Skill tool
if [ "$TOOL_NAME" != "Skill" ]; then
  exit 0
fi

SKILL=$(echo "$INPUT" | python3 -c \
  "import json,sys; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('skill',''))" \
  2>/dev/null || echo "")

# Map completed skill → next expected build-workflow step
case "$SKILL" in
  task-manager)
    echo ""
    echo "WORKFLOW: task-manager completed."
    echo "  If a task was just surfaced → next: /algo-trading-lead-dev for planning (build.md Step 2)"
    echo "  If criteria were verified and task closed → next: session-end (build.md Step 6)"
    ;;
  algo-trading-lead-dev)
    echo ""
    echo "WORKFLOW: algo-trading-lead-dev completed."
    echo "  If plan was delivered → confirm with user, then continue in build mode (build.md Step 3)"
    echo "  If build is done ('Ready for review') → next: /go-quality-review standard level (build.md Step 4)"
    ;;
  algo-trading-veteran)
    echo ""
    echo "WORKFLOW: algo-trading-veteran completed."
    echo "  Return to /algo-trading-lead-dev with Marcus's answer to unblock planning or build (build.md Step 2/3)"
    ;;
  go-quality-review)
    echo ""
    echo "WORKFLOW: go-quality-review completed."
    echo "  If clean (no blockers) → next: /task-manager to verify criteria + mark done (build.md Step 5)"
    echo "  If blockers found → return to /algo-trading-lead-dev to fix, then re-run reviewer"
    ;;
  decision-journal)
    echo ""
    echo "WORKFLOW: decision-journal completed."
    echo "  If decisions were recorded → check if any imply new tasks (/task-manager if so)"
    echo "  If this was the session-end harvest → proceed to commit (session-end Step 3)"
    ;;
  *)
    # Unknown skill — remind to check INDEX.md
    if [ -n "$SKILL" ]; then
      echo ""
      echo "WORKFLOW: $SKILL completed — consult workflows/INDEX.md for next step."
    fi
    ;;
esac

exit 0
