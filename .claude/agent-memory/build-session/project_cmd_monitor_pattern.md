---
name: cmd/monitor pattern — pure file-reading CLI, no DataProvider
description: cmd/monitor is a kill-switch monitoring CLI that reads JSON files only — no Zerodha auth, no BuildProvider needed; use run() extraction pattern
type: project
---

cmd/monitor (TASK-0048, 2026-05-07) is a pure file-reading binary with no DataProvider dependency. Key patterns established:

- No BuildProvider, no KITE_API_KEY, no auth flow — pure os.ReadFile + json.Unmarshal
- run(args, stdout, stderr) extraction from cmd/walk-forward applied here too
- thresholdsFile DTO in cmd/monitor (not JSON tags on analytics.KillSwitchThresholds)
- Synthetic equity curve from sorted trades by ExitTime (no separate --curve file)
- Trade log format: JSON array of model.Trade (decisions/convention/2026-05-07-live-trade-log-json-array-format.md)
- Exit code 0=OK, 1=any breach — enables cron/shell scripting

**Why:** analytics.KillSwitchThresholds is a pure computation type; keep JSON concerns in cmd layer only.

**How to apply:** Any future cmd/ binary that only reads files (no provider) follows this pattern. No cmdutil.BuildProvider call needed.
