---
name: Multiple strategy parameter variants are allowed
description: When porting a strategy to a new timeframe, create 2-3 named variants; evaluation decides survivors
type: feedback
---

When porting a strategy to a new timeframe (e.g., daily → 5-min), create 2–3 named variants with different parameter calibrations rather than choosing one upfront.

**Why:** Parameter calibration for a new timeframe involves genuine uncertainty. Let the evaluation pipeline resolve it empirically rather than making an arbitrary a-priori choice.

**How to apply:** Naming convention `{strategy}-{timeframe}-{variant}` (e.g., `macd-5min-fast`, `macd-5min-standard`, `macd-5min-slow`). Max 3 variants. Register all in GlobalRegistry. Each counts as independent nTrials in DSR correction. If all 3 fail universe gate, strategy is killed — not retried with more variants. Do NOT force a single parameter set — this is explicitly approved by the user.
