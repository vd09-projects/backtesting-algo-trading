#!/usr/bin/env bash
# Fires on Stop (end of every Claude turn). Warns — but does not block — when
# internal/ or pkg/ production code has changed since the last quality-gate pass.
# Blocking every turn would cause alert fatigue; the hard block lives in
# pre-tool-use-edit.sh (intercepts the BACKLOG.md "Status: done" edit).
set -euo pipefail

PROJECT="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT"

# Consume stdin (JSON with last_assistant_message — not needed for file-state checks)
cat > /dev/null

# Any unstaged/staged production code changes in internal/ or pkg/?
INTERNAL_DIRTY=$(git diff --name-only -- internal/ pkg/ 2>/dev/null \
  | grep '\.go$' | grep -v '_test\.go' || true)
INTERNAL_STAGED=$(git diff --cached --name-only -- internal/ pkg/ 2>/dev/null \
  | grep '\.go$' | grep -v '_test\.go' || true)

ALL_DIRTY="${INTERNAL_DIRTY}${INTERNAL_STAGED}"
if [ -z "$ALL_DIRTY" ]; then
  exit 0  # nothing to warn about
fi

SENTINEL="$PROJECT/.quality-gate/last-pass"
GATE_NEEDED=false

if [ ! -f "$SENTINEL" ]; then
  GATE_NEEDED=true
else
  NEWER=$(find internal/ pkg/ -name "*.go" ! -name "*_test.go" \
    -newer "$SENTINEL" 2>/dev/null | head -1)
  [ -n "$NEWER" ] && GATE_NEEDED=true
fi

if [ "$GATE_NEEDED" = "true" ]; then
  echo ""
  echo "QUALITY GATE PENDING: production code in internal/ or pkg/ changed since last /go-quality-review."
  echo "Changed files:"
  echo "$ALL_DIRTY" | sort -u | sed 's/^/  /'
  echo ""
  echo "Run /go-quality-review before marking any task done or committing."
fi

exit 0
