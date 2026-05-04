---
name: Walk-forward kill + parameter retune methodology
description: When a strategy is killed at walk-forward, re-testing with new parameters is clean only if the revisit trigger was pre-committed before WF results were analyzed
type: feedback
---

When a strategy fails the walk-forward gate and the user wants to re-test with different parameters, Marcus applies a strict test: was the new parameter set named as a future candidate BEFORE the per-instrument WF results were analyzed?

- Pre-committed revisit trigger in kill decision = CLEAN (TASK-0067: SMA fast=20/slow=50 was named in the TASK-0053 kill decision)
- Choosing parameters based on which instruments failed the WF gate = CONTAMINATED (implicit fold-structure tuning)

**Why:** Changing parameters after seeing which instruments failed WF risks fitting the new parameters to the fold structure of the failing instruments, even if framed as "going to standard defaults."

**How to apply:** At Step 3 (Marcus pre-check), always check the kill decision's revisit trigger text before ruling on whether a parameter retune is clean. If the proposed new parameters appear in the revisit trigger, they're clean. If they don't appear there, treat as a potential methodology violation and invoke Marcus.

Also: if MACD or other strategy parameters came from a single-instrument plateau sweep (not the full universe), be careful — changing those parameters post-WF failure may constitute implicit tuning against the WF results of that specific instrument. Marcus ruled: instrument-specific plateau parameter + cross-instrument WF failure ≠ grounds to change the parameter.
