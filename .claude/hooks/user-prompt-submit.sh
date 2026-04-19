#!/usr/bin/env bash
# Fires on UserPromptSubmit (before Claude processes each user turn).
# Detects session wrap-up language in the prompt and injects harvest reminders
# into Claude's context before it responds.
set -euo pipefail

PROJECT="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT"

INPUT=$(cat)
PROMPT=$(echo "$INPUT" | python3 -c \
  "import json,sys; d=json.load(sys.stdin); print(d.get('prompt','').lower())" \
  2>/dev/null || echo "")

# Detect wrap-up / session-ending intent
if echo "$PROMPT" | grep -qE \
  "done for (today|now)|let.?s (stop|end|wrap|close)|wrap(ping)? up|that.?s (all|it) for|commit (this|and stop|everything)|end(ing)? (the )?session|sign off|call it (a day|done)|stop here|ending (now|session)"; then

  echo "SESSION WRAP-UP DETECTED — fire both harvests before ending:"
  echo "  /task-manager     — harvest implicit tasks from this session"
  echo "  /decision-journal — harvest inline decision marks"
  echo ""

  DIRTY=$(git diff --name-only 2>/dev/null)
  STAGED=$(git diff --cached --name-only 2>/dev/null)
  if [ -n "$DIRTY" ] || [ -n "$STAGED" ]; then
    echo "UNCOMMITTED CHANGES — commit or note before leaving:"
    [ -n "$STAGED" ] && echo "$STAGED" | sed 's/^/  [staged] /'
    [ -n "$DIRTY"  ] && echo "$DIRTY"  | sed 's/^/  [unstaged] /'
  fi
fi

exit 0
