# Code Review Session

User wants existing code reviewed. Runs the gate, fixes blockers autonomously, tracks warnings
as tasks. Ends with a Session Summary. (~5% of sessions.)

## Trigger

User says "review this code," "run a quality check on X," "pre-merge review," "what's wrong
with this package," or names a package or file to audit.

---

## Execution

### Step 1 — Determine scope and level

Infer the review level from the user's intent:
- "quick check" / "just lint" → **quick**
- "review this" / "PR review" / no qualifier → **standard**
- "deep review" / "this package is critical" / "engine internals" → **deep**
- "pre-merge" / "ready to ship" → **pre-merge**

If the user named a specific package, check what it does — engine/accounting/metrics packages
always get at least **standard** regardless of what the user said.

Log:
```
[AUTO] Step 1 — Scope: <package or files>. Level: <quick|standard|deep|pre-merge>.
```

### Step 2 — Run the review (sub-agent)

Build a minimal SESSION STATE for this code-review session:
- `workflow`: "code-review", `task_id`: null, `step_completed`: 1
- `verdicts.decision_lookup`: null (no decision lookup needed for code review)

Read `workflows/agents/priya-build.md` as a reference for structure, but for code-review
sessions spawn a simpler sub-agent:

Agent prompt (fill scope/level before spawning):
```
You are a step-agent. Your job: run /go-quality-review at {{level}} level on {{scope}},
then invoke /algo-trading-lead-dev to fix any blockers (max 2 rounds).

SCOPE: {{scope}}
LEVEL: {{level}}

STEPS:
1. Run /go-quality-review at the specified level
2. If blockers: invoke /algo-trading-lead-dev with the specific findings, ask Priya to fix
   She may push back with a tradeoff justification — that is accepted; mark it as a
   **Decision (topic) — tradeoff: accepted** and continue
3. Re-run /go-quality-review. If new blockers after round 2: set flag.
4. If gate clean: break.

Return ONLY this JSON:
{
  "step": "code_review",
  "verdict": {
    "quality_gate": "PASS | FAIL",
    "blockers_found": N,
    "warnings_found": M,
    "suggestions_found": K,
    "warnings": ["..."],
    "suggestions": ["..."],
    "files_modified": [...]
  },
  "decision_marks": [],
  "flag": null
}
```

Parse returned JSON. If `flag` is non-null after round 2: Hard STOP.

Log:
```
[AUTO] Step 2 — Quality gate: <PASS|FAIL>. Blockers: N, Warnings: M, Suggestions: K.
```

### Step 3 — Handle findings

**Blockers:** already fixed by the sub-agent in Step 2. If the gate still failed: Hard STOP.

**Warnings (should fix, creates future problems):**

Do not fix warnings in this session unless the user explicitly requests it. Instead, auto-create
a task for each distinct warning category:

| Warning | Task priority |
|---|---|
| Missing test coverage on public functions | medium |
| High cyclomatic complexity | medium |
| Tight coupling between packages | medium |
| Missing godoc on exported types | low |
| Naming or formatting issues | low |

Log each created task:
```
[WARN] <finding>. TASK-XXXX created (priority).
```

**Suggestions (nice to have):**

Log them in the Summary under "Suggestions (not tracked)." Do not create tasks for suggestions
automatically. The user can choose to act on them.

### Step 4 — Write quality-gate sentinel

If the review ends with no blocker-level findings, the quality gate passes. Write the sentinel:

```bash
mkdir -p .quality-gate && date -u +"%Y-%m-%dT%H:%M:%SZ" > .quality-gate/last-pass
```

Log:
```
[AUTO] Step 4 — Quality gate sentinel written: .quality-gate/last-pass
```

If the review still has blockers (Hard STOP case), do not write the sentinel.

### Step 5 — Session end

Go to `session-end.md`.

The decision harvest may produce results if Priya marked any tradeoff decisions while
iterating on findings. If the review was clean, the harvest comes back empty — that is fine.
