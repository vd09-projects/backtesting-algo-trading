# WalkForwardConfig.To is the exclusive upper bound

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | convention       |
| Tags     | walkforward, time, boundary, half-open-interval, API, TASK-0022 |

## Context

`WalkForwardConfig.To` needed a clear semantic. The rest of the codebase (`engine.Config.To`, `provider.FetchCandles`) uses half-open intervals `[from, to)` — the To value is exclusive. The question was whether WalkForwardConfig should follow the same convention.

## Decision

`WalkForwardConfig.To` is the exclusive upper bound, consistent with `engine.Config.To` and `provider.FetchCandles([from, to))`. A fold is included only when its OOS end does not exceed `cfg.To`. Callers use `2025-01-01` to mean "through end of 2024," not `2024-12-31`. This is documented in the struct's field comment.

## Consequences

Callers must be aware that `To = time.Date(2025, 1, 1, ...)` means "up to but not including 2025-01-01." This is consistent with the rest of the engine API and reduces surprise when working across the codebase.
