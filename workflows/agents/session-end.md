# Agent Template: Session End

## Purpose
Harvests tasks and decisions from SESSION STATE. Generates a commit message.
Operates on the structured state — does NOT scan the main conversation history.
This is the key efficiency win: instead of scanning a 40K-token conversation,
the harvest runs on ~400 tokens of structured state.

## Slots to fill
- `{{session_state_json}}` — the full SESSION STATE JSON object (serialized)
- `{{git_diff_stat}}` — output of `git diff --stat HEAD`

## How to use
Run `git diff --stat HEAD` in the main session, then fill the slots and pass as Agent() prompt.

---

## Prompt template

```
You are a step-agent. Your job: run the session-end harvest from SESSION STATE and
return structured results. You do not have access to the main conversation — work only
from the data provided below.

SESSION STATE:
{{session_state_json}}

GIT DIFF STAT:
{{git_diff_stat}}

STEPS:

1. TASK HARVEST — invoke /task-manager in harvest mode. Pass it:
   - The execution_log entries from SESSION STATE
   - The decision_marks_pending list (these may imply follow-up tasks)
   - Look for: "we should also", TODO, follow-up work implied by decisions, edge cases noted
     but not implemented, tests mentioned but not written
   - Auto-create tasks that are clear consequences of the session
   - Skip anything already in the backlog (check tasks/BACKLOG.md)

2. DECISION HARVEST — invoke /decision-journal in harvest mode. Pass it:
   - The decision_marks_pending list from SESSION STATE
   - Write each mark as a decision file; update decisions/INDEX.md
   - Do not ask for confirmation — write them all, report what was written

3. COMMIT MESSAGE — generate from git diff stat and SESSION STATE context.
   Format exactly:
   ```
   <imperative verb> <what changed> (<task_id if applicable>)

   - <bullet: specific change 1>
   - <bullet: specific change 2>

   Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
   ```
   If no code changes (strategy evaluation, planning, or decision-only session): set to "no commit".

Return ONLY this JSON (no other text):
{
  "step": "session_end",
  "verdict": {
    "tasks_created": [{"id": "TASK-XXXX", "title": "...", "priority": "high|medium|low"}],
    "decisions_written": ["decisions/algorithm/2026-04-22-...md"],
    "suggested_commit": "..."
  },
  "decision_marks": [],
  "flag": null
}
```
