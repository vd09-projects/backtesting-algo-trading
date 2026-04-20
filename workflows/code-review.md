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

### Step 2 — Run the review

Auto-invoke `/go-quality-review` at the determined level. It produces a findings report:
blockers, warnings, suggestions.

Log:
```
[AUTO] Step 2 — Quality gate: <PASS|FAIL>. Blockers: N, Warnings: M, Suggestions: K.
```

### Step 3 — Handle findings

**Blockers (must fix before merge):**

Auto-invoke `/algo-trading-lead-dev` in iterate mode. Pass the specific blocker findings.
Priya addresses them. She may:
- Fix them straightforwardly
- Push back with reasoning — if she does, she marks a `tradeoff` decision explaining the
  intentional override (e.g., "function is long because splitting it harms readability in
  the hot loop"). Log the decision; the override is accepted.

After Priya finishes, auto-re-run `/go-quality-review` at the same level. If new blockers
surface from the fixes, iterate again. Max 2 rounds. If blockers persist after round 2: Hard STOP.

If no blockers remain after iteration: continue to warnings.

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
