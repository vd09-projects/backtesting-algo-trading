---
name: Instrument-count gate is 100% retention
description: Walk-forward instrument-count gate requires WF pass count >= universe gate pass count — not a percentage floor
type: project
---

The instrument-count gate as specified in TASK-0053 AC requires a strategy to pass walk-forward on AT LEAST AS MANY instruments as it passed the universe gate. This means 100% retention:

- MACD passed universe on 14 instruments → needs 14 WF passes
- SMA passed universe on 12 instruments → needs 12 WF passes

This is an extremely strict gate. No partial credit. A strategy passing 13/14 WF instruments still fails.

**Why this matters for future runs:** If the user revisits the gate criteria, the decision to require 100% retention was in the TASK-0053 acceptance criteria (not a Marcus call). To relax it, the user would need to update the gate criteria before re-running.

**Context:** Both strategies failed in 2026-05-04 run. MACD's 64% retention (9/14) is the stronger case for gate relaxation; SMA's 33% (4/12) suggests a more fundamental fit issue.
