# Workflow gate: PreToolUse hook blocks production Go writes when session-state absent

- **Date:** 2026-04-26
- **Status:** accepted
- **Category:** infrastructure
- **Tags:** workflow, hook, pre-tool-use, session-state, build.md, TASK-0041
- **Task:** TASK-0041

## Decision

Check 3 was added to `.claude/hooks/pre-tool-use-edit.sh`. It blocks Edit and Write tool calls
targeting `*.go` files under `strategies/`, `internal/`, `pkg/`, or `cmd/` unless
`workflows/.session-state.json` exists. Non-Go files and the session-state file itself are not
blocked.

## Rationale

The `build.md` workflow requires Steps 1-4 (decision-lookup, Marcus pre-check, Priya planning,
session-state creation) before any implementation begins. Without a hard gate, the sub-agent
spawning requirement is advisory only — it was skipped in TASK-0041 when Priya's plan was
implemented inline rather than via a separate sub-agent call. The hook makes the workflow
requirement enforceable: code cannot be written until the session-state file, which is created at
Step 4 of build.md, is present.

The gate is per-session (`.session-state.json` must be created each session) and not per-commit.
This preserves the ability to do non-code tasks (docs, task management, decisions) without
triggering the workflow.

## Scope

Blocked: Edit/Write to `strategies/**/*.go`, `internal/**/*.go`, `pkg/**/*.go`, `cmd/**/*.go`
Not blocked: Test files, YAML, markdown, JSON, shell scripts, any file outside those dirs.
Not blocked: Creating `workflows/.session-state.json` itself (bootstrap exception).

## Pipe-tested cases (2026-04-26)

1. Production Go file, no session-state → exit 2 (BLOCKED)
2. Non-Go file in `strategies/` → exit 0 (allowed)
3. Production Go file, session-state present → exit 0 (allowed)

## Rejected alternatives

- **Warn only, no block** — insufficient; the TASK-0041 violation showed that a warning alone is
  ignored under time pressure. Hard block is the only enforcement that works.
- **Block on all files, not just .go** — too broad; task management and docs should not require
  a full workflow setup.
- **Commit hook instead** — commit hooks run after the code is written. The goal is to prevent
  the write from happening at all, not to catch it after the fact.
