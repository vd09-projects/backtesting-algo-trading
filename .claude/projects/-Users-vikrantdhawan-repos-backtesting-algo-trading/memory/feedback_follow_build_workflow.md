---
name: Follow build.md sub-agent workflow — do not implement inline
description: Claude must follow the build.md orchestration model exactly — spawn sub-agents for Steps 2-5 and write SESSION STATE; never implement directly in the main context
type: feedback
---

Always follow the build.md workflow when a task is picked from the backlog. Do NOT write code inline in the main context.

**Why:** The orchestration model in workflows/INDEX.md exists to keep the main context lean and create structured handoffs. Implementing directly accumulates 40-60K tokens in the main context and skips the planning/review gates. The user explicitly called this out after TASK-0041.

**How to apply:**

1. **Step 2** — Spawn a sub-agent using the template at `workflows/agents/decision-lookup.md`. Fill `{{task_id}}`, `{{task_title}}`, `{{task_context}}` from BACKLOG. Parse returned JSON into SESSION STATE.

2. **Step 3** — Spawn Marcus sub-agent (template: `workflows/agents/marcus-precheck.md`) only if task touches fill model, sizing, metrics, kill-switch, test plan, or evaluation methodology. Skip with log if a prior decision covers it.

3. **Step 4** — Spawn Priya planning sub-agent (template: `workflows/agents/priya-plan.md`). Fill slots from SESSION STATE. Parse returned JSON (`approach`, `files_to_create`, `files_to_modify`). Update SESSION STATE. Log the plan. Do NOT proceed to Step 5 without this JSON.

4. **Step 5** — Spawn Priya build sub-agent (template: `workflows/agents/priya-build.md`). Fill slots from the Step 4 verdict. Sub-agent owns the full build+lint+test loop. Main context only sees the returned JSON summary.

5. **SESSION STATE** — Write `workflows/.session-state.json` after every step. This enables resume if the session is interrupted.

**The instinct to "just write the code" when the task is clear is the exact failure mode.** Reading build.md and then coding inline violates the workflow. The templates exist; use them.
