# Agent Template: Evaluation-Run Session End

## Purpose
Harvests tasks and decisions from SESSION STATE after an evaluation-run session.
Operates on structured state — does NOT scan the main conversation history.

## Slots to fill
- `{{session_state_json}}` — full SESSION STATE JSON (serialized)
- `{{git_diff_stat}}` — output of `git diff --stat HEAD`

## How to use
Run `git diff --stat HEAD` in the orchestrator, fill the slots, pass as Agent() prompt.

---

## Prompt template

```
You are a step-agent. Run the session-end harvest from SESSION STATE and return structured
results. Work only from the data provided — do NOT scan the main conversation history.

SESSION STATE:
{{session_state_json}}

GIT DIFF STAT:
{{git_diff_stat}}

STEPS:

1. TASK HARVEST — invoke /task-manager in harvest mode. Pass it:
   - The execution_log entries from SESSION STATE
   - The decision_marks_pending list (these may imply follow-up tasks)
   - Look for: edge cases noted but not implemented, tests mentioned but not written,
     deferred work, tech debt introduced, follow-up implied by decisions
   - Auto-create tasks that are clear consequences of the session
   - Skip anything already in tasks/BACKLOG.md (check first)

2. DECISION HARVEST — invoke /decision-journal in harvest mode. Pass it:
   - The decision_marks_pending list from SESSION STATE
   - Write each mark as a decision file in decisions/algorithm/; update decisions/INDEX.md
   - Do not ask for confirmation — write them all, report what was written
   - Skip any mark whose content is already covered by a file written in Step 6a of the
     evaluation-run session (check decisions/INDEX.md for entries created today with an
     overlapping slug before writing)

3. COMMIT MESSAGE — generate from git diff stat and SESSION STATE context.
   Format: imperative verb + what changed + (TASK-ID if applicable), body as bullets.
   If no code changes (eval/decision-only session): set to "no commit".

Return ONLY this JSON (no other text):
{
  "step": "session_end",
  "verdict": {
    "tasks_created": [{"id": "TASK-XXXX", "title": "...", "priority": "high|medium|low"}],
    "decisions_written": ["decisions/algorithm/YYYY-MM-DD-slug.md"],
    "suggested_commit": "..."
  },
  "decision_marks": [],
  "flag": null
}
```
