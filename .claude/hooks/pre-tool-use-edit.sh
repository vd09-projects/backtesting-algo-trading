#!/usr/bin/env bash
# Fires on PreToolUse for Edit and Write tool calls.
#
# Two checks:
#   1. BACKLOG.md being edited to mark a task done → BLOCK if quality gate not current.
#      This is the hard enforcement point for the quality gate rule.
#   2. Production code in internal/ or pkg/ being edited → TDD reminder (warn, no block).
#
# Exit 2 blocks the tool call. Exit 0 allows it.
set -euo pipefail

PROJECT="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT"

# Parse JSON input from stdin
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | python3 -c \
  "import json,sys; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('file_path',''))" \
  2>/dev/null || echo "")
NEW_STRING=$(echo "$INPUT" | python3 -c \
  "import json,sys; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('new_string',''))" \
  2>/dev/null || echo "")

# --- Check 1: Marking a task done in BACKLOG.md ---
if echo "$FILE_PATH" | grep -q "BACKLOG\.md" \
  && echo "$NEW_STRING" | grep -qiE "\*\*Status:\*\* done|Status:.*done"; then

  SENTINEL="$PROJECT/.quality-gate/last-pass"

  # Any uncommitted production code changes?
  INTERNAL_CHANGED=$(git diff --name-only -- internal/ pkg/ 2>/dev/null \
    | grep '\.go$' | grep -v '_test\.go' || true)
  INTERNAL_STAGED=$(git diff --cached --name-only -- internal/ pkg/ 2>/dev/null \
    | grep '\.go$' | grep -v '_test\.go' || true)
  ALL_CHANGED="${INTERNAL_CHANGED}${INTERNAL_STAGED}"

  if [ -n "$ALL_CHANGED" ]; then
    # Production code changed — verify quality gate is current
    if [ ! -f "$SENTINEL" ]; then
      echo "BLOCKED: Cannot mark task done."
      echo "Reason: internal/ or pkg/ production code was changed but /go-quality-review has never been run."
      echo "Fix: run /go-quality-review, then retry marking the task done."
      exit 2
    fi

    NEWER=$(find internal/ pkg/ -name "*.go" ! -name "*_test.go" \
      -newer "$SENTINEL" 2>/dev/null | head -1)
    if [ -n "$NEWER" ]; then
      echo "BLOCKED: Cannot mark task done."
      echo "Reason: production code in internal/ or pkg/ was modified after the last /go-quality-review pass."
      echo "Changed since last pass:"
      find internal/ pkg/ -name "*.go" ! -name "*_test.go" -newer "$SENTINEL" 2>/dev/null \
        | head -5 | sed 's/^/  /'
      echo "Fix: run /go-quality-review, then retry marking the task done."
      exit 2
    fi
  fi
fi

# --- Check 2: Editing production code in internal/ or pkg/ (TDD reminder) ---
if echo "$FILE_PATH" | grep -qE "^(internal|pkg)/" \
  && ! echo "$FILE_PATH" | grep -q "_test\.go"; then
  echo "TDD: have you written the failing test first? (CLAUDE.md: TDD is mandatory)"
fi

exit 0
