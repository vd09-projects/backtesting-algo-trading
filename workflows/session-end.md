# Session End — Closing Ritual

Runs automatically at the end of every workflow. No confirmation steps between harvests.
Produces the Session Summary as the final output.

## When to trigger

- Any workflow reaches its final step and routes here
- User says "wrap up," "let's commit," "I'm done," "session end"
- Natural stopping point after substantive work

---

## Execution

### Step 1 — Task harvest

Invoke `/task-manager` in harvest mode. It scans the conversation for implicit tasks:
- Phrases: "we should also," "TODO," "later we'll need to," "we're hardcoding this for now"
- Edge cases discussed but not implemented
- Tests mentioned but not written
- Technical debt introduced intentionally
- Follow-up work implied by decisions made

For each discovered potential task, auto-evaluate:
- Is it already in the backlog? → skip (log as duplicate)
- Is it a direct consequence of work just done? → create it automatically as source `session`
  with priority inferred from severity (blocker finding → high, warning → medium, note → low)
- Is it ambiguous enough that creation might be wrong? → include in Summary under "Suggested
  tasks" so user can confirm at review time

Also auto-update in-progress task statuses based on what the conversation shows was completed.

Log:
```
[AUTO] Session end — Task harvest: N tasks created, M duplicates skipped.
```

### Step 2 — Decision harvest

Invoke `/decision-journal` in harvest mode. It scans for inline decision marks from Marcus
and Priya (`**Decision (YYYY-MM.N.N) — category: status**` blocks).

For each confirmed mark, auto-write the decision file and update INDEX.md. Do not ask for
confirmation on individual marks — write them all, then report what was written in the Summary.

Exception: if two marks look like the same decision from different owners (duplicate check),
surface both in the Summary under "Decisions — needs review" rather than auto-merging.

Log:
```
[AUTO] Session end — Decision harvest: N decisions written to decisions/.
```

If no marks found:
```
[AUTO] Session end — Decision harvest: no inline marks found.
```

### Step 3 — Commit message

Generate a commit message from the session's changes. Use `git diff --stat HEAD` to see what
changed. Format:

```
<imperative verb> <what changed> (<task ID if applicable>)

- <bullet: specific change 1>
- <bullet: specific change 2>
...

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
```

Include the commit message in the Session Summary. The user runs the git command themselves.

If no code changes were made (pure strategy evaluation, planning, or decision-only session),
skip the commit step and note it in the Summary.

### Step 4 — Session Summary

Produce the Session Summary block. This is the final output of the session.

```
═══ Session Summary — YYYY-MM-DD ═══

Task:        TASK-NNNN — <title>   (or "no task — <session type>" for review/evaluate)
Status:      done | in-progress | blocked
Quality:     PASS / FAIL / skipped

Execution log:
  [AUTO]     Step 1 — Task: ...
  [AUTO]     Step 2 — Prior decisions: ...
  [DECISION] Marcus [algorithm]: ...
  [AUTO]     Step 4 — Plan: ...
  [AUTO]     Step 5 — Build: ...
  [WARN]     ...
  [CLOSED]   TASK-NNNN done. ...

Decisions written:
  decisions/algorithm/YYYY-MM-DD-<slug>.md
  decisions/convention/YYYY-MM-DD-<slug>.md

Tasks created:
  TASK-XXXX (high):   <title>
  TASK-YYYY (medium): <title>

Next up:     TASK-XXXX — <title>

Suggested commit:
  <commit message block>

Flagged (review if needed):
  <[FLAGGED] and [WARN] entries that need attention>
═════════════════════════════════════
```

Omit any section that is empty. Keep it readable.

---

## Notes

- Both harvests always run. Empty results are fine — log them and move on.
- If the session produced no code changes, omit the commit section entirely.
- If a Hard STOP fired during the session, the Summary includes the STOP condition and current
  state so the next session can resume from the right point.
- The Summary is the handoff artifact. It should contain everything needed to restart work
  in a future session without re-reading the conversation.
