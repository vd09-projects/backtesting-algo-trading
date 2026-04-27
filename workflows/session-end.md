# Session End — Closing Ritual

Runs automatically at the end of every workflow. Operates on SESSION STATE — does NOT scan
the main conversation history. This keeps the harvest efficient regardless of session length.

## When to trigger

- Any workflow reaches its final step and routes here
- User says "wrap up," "let's commit," "I'm done," "session end"
- Natural stopping point after substantive work

---

## Execution

### Step 1 — Run session-end sub-agent

Run: `git diff --stat HEAD` in the orchestrator.

Read `workflows/agents/session-end.md`. Fill slots:
- `{{session_state_json}}` — serialize the current SESSION STATE to JSON
- `{{git_diff_stat}}` — output of the git command above

Spawn sub-agent via Agent() tool. The sub-agent:
- Invokes /task-manager in harvest mode against the execution_log + decision_marks_pending
- Invokes /decision-journal in harvest mode against decision_marks_pending
- Generates a commit message from the diff stat + task context

Parse returned JSON.

Log:
```
[AUTO] Session end — Task harvest: N tasks created.
[AUTO] Session end — Decision harvest: N decisions written.
```

### Step 2 — Mark session state complete

Session files under `workflows/sessions/` are NOT deleted — they serve as a lightweight run log.
No action needed here; the orchestrator only reads the file matching the current task ID, so
completed session files are ignored automatically and add no token cost to future sessions.

### Step 3 — Session Summary

Produce the Session Summary from SESSION STATE + sub-agent verdict:

```
═══ Session Summary — YYYY-MM-DD ═══

Task:        TASK-NNNN — <title>   (or "no task — <session type>" for review/evaluate)
Status:      done | in-progress | blocked
Quality:     PASS / FAIL / skipped

Execution log:
  (copy all entries from SESSION STATE execution_log)

Decisions written:
  (list from session-end verdict.decisions_written)

Tasks created:
  (list from session-end verdict.tasks_created)

Next up:     TASK-XXXX — <title>   (top of BACKLOG Up Next, or "check backlog")

Suggested commit:
  (from session-end verdict.suggested_commit)

Flagged (review if needed):
  (any [FLAGGED] or [WARN] entries from the execution log)
═════════════════════════════════════
```

Omit any section that is empty. If no code changes were made, omit the commit section.

---

## Notes

- If the session produced no decision_marks_pending and no execution_log entries suggesting
  new tasks, both harvests will come back empty — that is fine, log them and move on.
- If a Hard STOP fired during the session, the Summary includes the stop condition and the
  current SESSION STATE so the next session can resume from the right step.
- The Summary is the handoff artifact. Combined with the session file in `workflows/sessions/`
  (if a Hard STOP is active), it contains everything needed to restart work in a future session.
