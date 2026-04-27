# Workflow Index

## Orchestration model

Every non-trivial workflow step runs inside a sub-agent with a fresh context window.
The orchestrator's (main session's) context stays lean — it sees only structured JSON summaries,
never the raw output of skills, file reads, or build loops.

### Why this matters

After 5–6 turns, every API call ships the entire conversation history. Without sub-agents,
a full build session accumulates 40,000–60,000 tokens in the main context. With sub-agents,
the main context stays under 4,000 tokens regardless of how many steps run.

### How it works

1. Orchestrator reads the matching workflow file (build.md, evaluate.md, etc.)
2. Each step spawns a sub-agent via the Agent() tool using a template from `workflows/agents/`
3. The orchestrator reads the template, fills in `{{slots}}` from SESSION STATE, calls Agent()
4. Sub-agent does the heavy work: reads files, invokes skills, iterates
5. Sub-agent returns ONLY a structured JSON block — see `workflows/handoffs/schema.md`
6. Orchestrator updates SESSION STATE with the verdict
7. Orchestrator writes SESSION STATE to `workflows/sessions/{today}-{TASK-ID}.json` (checkpoint)

### SESSION STATE

Maintained inline in the orchestrator's response. Written to a uniquely named file under
`workflows/sessions/` after every step. Schema defined in `workflows/handoffs/schema.md`.

**File naming:** `workflows/sessions/{YYYY-MM-DD}-{TASK-ID}.json`
Example: `workflows/sessions/2026-04-27-TASK-0043.json`

One file per task run. Each new orchestrator start creates (or overwrites) its own file —
no accumulation of state across sessions, no large file to read on resume.

### Resume protocol

At session start, before routing to a workflow:

1. List files in `workflows/sessions/`. Find the most recent file whose task ID matches
   the task being started (or the most recently modified file if task is unknown).
2. If a matching file exists and `hard_stop_active` is null: resume from `step_completed + 1`
   Log: `[AUTO] Resuming from checkpoint: TASK-NNNN, step N completed. Continuing from step N+1.`
3. If `hard_stop_active` is set: present the stop condition to the user, wait for resolution
4. If no matching file exists: normal session start — create a new file named
   `workflows/sessions/{today}-{TASK-ID}.json`

Session state files are NOT deleted at session end — they serve as a lightweight run log.
The orchestrator only reads the file matching the current task ID, so old files are ignored
automatically and never add token cost.

### Sub-agent discipline

Every Agent() call must be self-contained. Never write "based on the conversation above" —
the sub-agent has no conversation history. Every piece of context must be written out explicitly.

### Continuous execution — no mid-workflow output

**Do not produce any text output between steps.** Run all steps to completion, then output
the Session Summary once. The only exception is a Hard STOP — which pauses and presents
the blocker to the user. Every other condition (flags, warnings, lint failures) is handled
silently inside the workflow. If you find yourself about to write a response before all steps
are done, stop — log the step result to SESSION STATE and immediately spawn the next agent.

### Spawn-once rule

Each sub-agent is spawned exactly once per step. If the returned JSON is structurally valid
(all required fields present, no parse error), accept it and move on. Do not re-spawn to
"improve" the result. If the JSON is malformed: re-spawn once with the same prompt plus
"Your previous response was not valid JSON. Return only the JSON block, no other text."
If it fails a second time: treat it as a Hard STOP (unresolvable blocker).

### Flag evaluation

Sub-agents return a `flag` field (null or string). The orchestrator evaluates flags
autonomously using the same Hard STOP taxonomy below. Only the three Hard STOP conditions
cause a pause. All other flags are resolved by applying prior decisions or making the call.

---

## Execution model

Workflows execute autonomously from trigger to Session Summary. Do not stop to ask permission
between steps. The user's role is to review the Summary and tune — not to approve each step.

**Execute → Log → Summarize → Tune.**

Every step produces an Execution Log entry. The session ends with a Session Summary. The user
reads the Summary and either says "looks good" or adjusts something. That adjustment is the
only mid-session interaction that should happen, and only when the user initiates it.

---

## STOP POINT taxonomy

### Hard STOP — pause and wait for user input

Only three conditions qualify:

1. **Requirements gap** — the task's intent cannot be inferred from the task block, the
   codebase, or prior decisions. A concrete piece of information only the user has is missing.
   State exactly what is missing and what the two or three most plausible interpretations are.

2. **Genuinely new methodology call** — Marcus needs to make a sizing rule, kill-switch line,
   feature verdict, or test-plan call, AND there is no prior `algorithm`-category decision that
   covers the question. Check `decisions/algorithm/` first. If a prior decision applies, use it.

3. **Unresolvable blocker** — the quality gate has an architectural finding that Priya cannot
   resolve in two iteration rounds, OR a build is fundamentally blocked and the constraint is
   external (missing data, broken dependency, incompatible interface).

When a Hard STOP fires: state the blocker, the options, and which option you'd take if forced.
Then wait. Do not continue the workflow.

### Soft STOP — log it, do not pause

- Priya makes a non-trivial tradeoff or convention decision → log `[DECISION]`, harvest at end
- A task was larger than the original scope; it was split automatically → log `[SPLIT]`
- Marcus's answer departs from a prior decision → log `[FLAGGED]`, note the tension in Summary
- Quality gate found warnings (not blockers) → log `[WARN]`, auto-create follow-up tasks
- Bootstrap or metric CI is wide / inconclusive → log `[FLAGGED]`, create a task

### No STOP — fully automatic

- Routing between skills (Priya ↔ Marcus ↔ quality gate)
- Applying prior decisions from `decisions/`
- Fixing lint, formatting, race conditions (`golangci-lint --fix`, re-run once)
- Creating tasks from discoveries or reviewer findings
- Session-end harvests (tasks + decisions)
- Archiving completed tasks
- Generating commit messages

---

## Prior Decision Lookup protocol

Before routing to Marcus or Priya for any non-trivial call, check the decision journal:

- **Methodology / algorithm question** → read `decisions/algorithm/` index entries. If a prior
  decision covers the question, apply it and log `[AUTO] Applied prior decision: <title>`.
  Only invoke Marcus if the question is genuinely new.

- **Architecture / convention question** → read `decisions/architecture/` and
  `decisions/convention/` index entries. Same logic — apply existing decisions, do not re-litigate.

- **Quality gate check** → read `.quality-gate/last-pass`. If current for the files changed,
  skip the gate and log `[AUTO] Quality gate: already current, skipped`.

If prior decisions conflict with each other, surface the conflict in the Summary as `[FLAGGED]`
and apply the more recent one.

---

## Auto-routing rules

| Situation | Action |
|---|---|
| Task picked from backlog | Auto-invoke Priya to plan |
| Task touches fill model, sizing, metrics, kill-switch, or test plan | Check journal first; if no prior decision → auto-invoke Marcus before Priya plans |
| Priya says she is flagging for Marcus | Auto-invoke Marcus; return to Priya with his answer |
| Build complete | Auto-run quality gate (standard for `internal/` changes; quick otherwise) |
| Lint or format failures | Auto-fix (`golangci-lint --fix`), re-run once |
| Quality gate clean | Auto-verify acceptance criteria; auto-close if all met |
| Acceptance criteria gap | Log `[FLAGGED]`, create a follow-up task, close this task for what it achieved |
| Methodology pivot mid-build ("this isn't a code bug") | Auto-invoke Marcus; return to Priya with his ruling |
| Session ends | Auto-harvest tasks, auto-harvest decisions, auto-generate commit message |

---

## Execution Log format

Every autonomous action produces a one-line log entry. Accumulate these through the session
and include them verbatim in the Session Summary.

```
[AUTO]     <Step N — what happened and the result>
[DECISION] <Owner> [<category>]: <one-line summary of the decision>
[SPLIT]    <TASK-NNNN split into TASK-XXXX and TASK-YYYY — reason>
[FLAGGED]  <What was flagged and why — follow-up task ID if created>
[WARN]     <Quality gate warning — severity, finding, follow-up task ID>
[CLOSED]   <TASK-NNNN done. All criteria met. Archived.>
[STOP]     <Hard STOP fired — which condition, what is needed>
```

---

## Session Summary format

Every workflow ends by producing this block. This is the user's primary interaction point.

```
═══ Session Summary — YYYY-MM-DD ═══

Task:        TASK-NNNN — <title>
Status:      done | in-progress | blocked
Quality:     PASS / FAIL / skipped

Execution log:
  [AUTO]     ...
  [DECISION] ...
  [CLOSED]   ...

Decisions made:
  <Owner> [<category>]:  <one-line>

Tasks created:
  TASK-XXXX (<priority>): <title>

Next up:     TASK-XXXX — <title>

Flagged (not blocking):
  <Anything in [FLAGGED] or [WARN] that the user should know>
═════════════════════════════════════
```

If nothing was flagged, omit that section. Keep it clean.

---

## Session-type triggers

Identify the session type from the user's opening message. Load only the matching workflow.
If no trigger matches, respond directly — not every interaction needs a workflow.

| User intent | Workflow |
|---|---|
| "What's next" / picks a task / mentions backlog | Check task type: feature/refactor → `build.md`, bug → `bugfix.md` |
| Has backtest results, run outputs, or metrics to review | `review.md` |
| New strategy idea, edge thesis, "is this worth building?" | `evaluate.md` |
| "Review this code" / "run quality check" / "pre-merge review" | `code-review.md` |
| "What have we decided about X?" / "why did we choose Y?" | Answer directly using the decision journal in query mode. If the answer implies action, route to the relevant workflow. |
| "Show backlog" / "reprioritize" / "any stale tasks?" | Answer directly using task-manager in review mode. No workflow file needed. |
| Quick question (syntax, definitions, "what's half-Kelly?") | Answer directly. No workflow. |

---

## Rules that do not change

- **Any workflow can be exited early.** If the user says "stop here" or "skip the rest," honor it.
  Run session-end.md for whatever work was done.
- **Marcus owns edge, sizing, methodology.** Priya owns code, structure, infra. Neither overrides
  the other in their domain. This routing is non-negotiable.
- **Context carrying.** At every skill transition, pass: task ID, key decisions made so far,
  what the next skill needs to know. Never rely on the next skill to reconstruct context.
- **TDD is not optional.** Priya writes tests before implementation. The quality gate verifies
  this. No exceptions unless the user explicitly says so.
