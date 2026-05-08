# Walk-forward year-to-duration uses 365×24h, not time.AddDate

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-08       |
| Status   | experimental     |
| Category | convention       |
| Tags     | walk-forward, time, duration, year, 365-days, AddDate, fold-count, determinism |

## Context

`cmd/walk-forward` accepts `--is-years`, `--oos-years`, and `--step-years` integer flags and must convert them to `time.Duration` for the `WalkForwardConfig` struct. Two options exist: `time.AddDate(n, 0, 0)` (calendar-exact) or `n * 365 * 24 * time.Hour` (fixed 365-day approximation).

The existing `internal/walkforward` test suite uses `365 * 24 * time.Hour` throughout, with inline comments documenting the leap-year arithmetic.

## Decision

Use `365 * 24 * time.Hour` per year, not `time.AddDate`. The conversion is: `year := 365 * 24 * time.Hour; window := time.Duration(n) * year`.

`time.AddDate(n, 0, 0)` gives calendar-exact year boundaries but makes fold count depend on how many leap years fall in the outer window — a 7-year window produces a different number of folds depending on which 7 years it spans. That's mildly surprising for a flag documented as "in years." The fixed-day approximation keeps fold arithmetic predictable and matches the existing test suite. The error (roughly 1 day per 4 years) is irrelevant for backtest windows measured in years.

## Consequences

Fold boundaries shift by up to ~1 day per leap year versus calendar-exact arithmetic. Immaterial for daily-bar strategy evaluation. If sub-day precision on fold boundaries ever matters, this convention should be revisited.

## Related decisions

- [Walk-forward window sizing defaults (2yr IS / 1yr OOS / 1yr step)](../algorithm/2026-04-22-walk-forward-window-sizing-default.md) — the defaults this convention applies to

## Revisit trigger

If walk-forward is extended to intraday strategies where fold boundaries at day-level precision matter.
