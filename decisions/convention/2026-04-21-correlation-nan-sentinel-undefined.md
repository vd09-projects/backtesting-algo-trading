# `math.NaN()` as sentinel for undefined correlation

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-21       |
| Status   | experimental     |
| Category | convention       |
| Tags     | NaN, sentinel, pearson, correlation, TASK-0027 |

## Context

TASK-0027 required a return value for `pearson` and `stressPearson` when correlation is undefined: constant series (zero variance), fewer than 2 data points, or no equity points in a stress window. Two sentinel options exist: `0.0` or `math.NaN()`.

## Decision

`math.NaN()`. Zero is a valid Pearson correlation value. Using 0 as a sentinel conflates "uncorrelated" with "undefined," which would cause a constant-series strategy (never trades) to appear uncorrelated rather than incalculable. NaN propagates through float arithmetic and is unambiguously detectable via `math.IsNaN`.

All callers guard with `math.IsNaN` before interpreting: `TooCorrelated` is only set when `!math.IsNaN(v) && v > threshold`. A NaN stress period does not suppress the full-period flag, and a NaN full-period does not suppress a stress-period flag.

## Consequences

- Output layer prints `n/a` for NaN values (`formatCorr` in `internal/output`).
- `TooCorrelated` is never triggered by a NaN value — only real computed correlations can set the flag.
- A strategy pair where one strategy never trades produces NaN for all three fields and `TooCorrelated: false`. This is correct: you cannot measure correlation for a strategy with no activity.

## Related decisions

- [Warmup first-change heuristic](../architecture/2026-04-21-correlation-warmup-first-change-heuristic.md) — how constant-series inputs are produced (strategy never trades → all-flat curve → all-zero returns → NaN).
