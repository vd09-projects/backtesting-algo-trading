---
name: Session file naming and location convention
description: Where build session state files live and how they are named
type: project
---

Session files: `workflows/sessions/{YYYY-MM-DD}-{TASK-ID}.json` — one file per task run. The orchestrator checks for an existing file at session start to enable resume. Completed session files are not deleted; they serve as a lightweight run log. The orchestrator only reads the file matching the current task ID, so old files add no token cost.

**Why:** Enables resume from step_completed + 1 if a session is interrupted. Also provides a lightweight audit trail of which tasks were run on which dates.

**How to apply:** At session start, always check workflows/sessions/ for a file matching {today}-{TASK-ID}.json before initializing fresh state. If found and hard_stop_active is null, resume from step_completed + 1.

As of 2026-04-27, completed session files in the directory:
- 2026-04-27-TASK-0039.json (TASK-0039, completed)
- 2026-04-27-TASK-0043.json (TASK-0043, completed)
- 2026-04-27-TASK-0047.json (TASK-0047, completed)
