---
name: Task categories that trigger Marcus pre-check
description: Which task types require Marcus methodology review vs which can skip Step 3
type: feedback
---

Commission model tasks (TASK-0038, TASK-0047) do NOT trigger Marcus. STT rates, exchange charges, and GST bases are statutory facts established by SEBI/NSE — they are architecture/convention decisions, not methodology calls. Step 3 was correctly skipped for both commission tasks.

**Why:** The methodology pre-check is for fill model choices, position sizing algorithms, performance metric definitions, kill-switch logic, and test plan methodology. Statutory charge rates have no ambiguity — they are law.

**How to apply:** Before spawning Marcus for a commission/cost-model task, check whether the question is "what rate to use" (statutory fact → skip Marcus) vs "how to model slippage or fill price" (methodology → invoke Marcus). The former is always skip.

Tasks that consistently trigger Marcus: fill model choices (TASK-0046 forced-close price), sizing methodology (vol-targeting), walk-forward window sizing, bootstrap Sharpe computation.
