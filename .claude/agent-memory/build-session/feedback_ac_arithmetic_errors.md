---
name: AC arithmetic band errors — savings vs cost transposition
description: Pattern where task acceptance criteria contain cost bands that are actually savings figures
type: feedback
---

TASK-0047 had an AC band "~₹55-60" for MIS round-trip cost that was actually the CNC-vs-MIS savings figure (~₹52.50), not the MIS cost (~₹35.74). The task context said "overstate costs by ~₹60" and this savings number was mistakenly written into the AC as the target cost.

**Why it matters:** The planner sub-agent (Priya) must verify AC arithmetic before writing golden tests. Trusting the band blindly would produce incorrect golden tests.

**How to apply:** When a task has a cost band in an AC, always hand-verify it from first principles before accepting it. Check: does the band refer to the absolute cost, or is it actually a delta/savings number? For commission model tasks specifically — compute the round-trip from the rates and compare to the AC band. If they disagree by more than 20%, flag it as a planning discrepancy.

The flag was raised correctly in Step 4 and resolved without blocking. The golden tests use the arithmetically correct values with an explanatory comment in the test file.
