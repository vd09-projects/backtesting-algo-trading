# Workflow Index

How to use: when a session starts or a skill reaches a terminal state, scan the triggers below
to find the right workflow. Load only the matching workflow file. If no trigger matches, respond
directly — not every interaction needs a workflow.

## Rules

- **User decides at decision points.** Workflows name options; the user picks. Never auto-invoke
  a skill without the user's say-so. Suggest the next step, wait for confirmation.
- **If multiple triggers match,** ask the user which to focus on first.
- **Context carrying:** at each skill transition, restate the task ID (if any), the key decisions
  made so far, and what the next skill needs to know. Don't rely on the user to carry context.
- **Any workflow can be exited early.** If the user says "skip the reviewer" or "that's enough,"
  honor it. Jump to session-end.md if there's work to harvest.

---

## Session start triggers

| User intent | Workflow |
|---|---|
| "What's next" / picks a task / mentions backlog | Check task type: feature/refactor → `build.md`, bug → `bugfix.md` |
| Has backtest results, run outputs, or metrics to review | `review.md` |
| New strategy idea, edge thesis, "is this worth building?" | `evaluate.md` |
| "Review this code" / "run quality check" / "pre-merge review" | `code-review.md` |
| "What have we decided about X?" / "why did we choose Y?" | Inline: invoke `/decision-journal` in query mode. If the answer leads to action (revisit a decision, make a new one), invoke the relevant skill and use `session-end.md` to harvest. |
| "Show backlog" / "reprioritize" / "any stale tasks?" | Inline: invoke `/task-manager` in review mode. If stale decisions surface too, also invoke `/decision-journal` review. No workflow file needed. |
| Quick question (Go syntax, "what's half-Kelly?", definitions) | No workflow. Answer directly. |

## Mid-session triggers

| Situation | Action |
|---|---|
| A skill hits a terminal state | Check the active workflow for what comes next. Suggest, don't auto-invoke. |
| Priya says she needs methodology input | The active workflow should have a Marcus detour. Follow it. |
| Marcus gives a verdict that implies build work | Suggest transitioning to Priya or task-manager per the active workflow. |
| Reviewer finds blockers | Route back to Priya for iteration per the active workflow. |
| User pivots ("actually, forget this, let's do something else") | The current workflow is abandoned. Check triggers for the new intent. |
| User asks "what should happen next?" | Re-read the active workflow's current step and suggest. |

## Session end trigger

Any natural stopping point where work was done → `session-end.md`. Every other workflow ends
by pointing here.
