# BenchmarkReport is a separate struct, not an extension of Report

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-13       |
| Status   | accepted         |
| Category | convention       |
| Tags     | benchmark, BenchmarkReport, Report, analytics, struct-design, convention, TASK-0018 |

## Context

Adding buy-and-hold benchmark comparison (TASK-0018) required a data structure to hold benchmark metrics. The existing `analytics.Report` struct already holds strategy performance metrics. The question was whether to extend `Report` with benchmark fields or define a separate `BenchmarkReport` type.

## Options considered

### Option A: Add benchmark fields to `analytics.Report`
- **Pros**: Single struct to pass around; no changes to `output.Write` signature.
- **Cons**: `Report` fields like `TradeCount`, `WinCount`, `LossCount`, `WinRate`, and `TotalPnL` are meaningless for buy-and-hold (which has no trades). These fields would be zero-valued in the benchmark, silently misrepresenting the report's intent. Callers couldn't distinguish "strategy with 0 trades" from "benchmark report".

### Option B: Separate `BenchmarkReport` struct
- **Pros**: Each struct carries only semantically valid fields. `BenchmarkReport` has `TotalReturn`, `AnnualizedReturn`, `SharpeRatio`, `MaxDrawdown` — all meaningful for a passive position. `Report` retains its trade-centric fields. No ambiguity.
- **Cons**: Two structs to maintain; output layer needs to handle both.

## Decision

Option B — separate `BenchmarkReport` struct. The benchmark is not a strategy run. It has no trades, no win rate, no P&L in the strategy sense. Sharing a struct would create silent semantic corruption: `BenchmarkReport.TradeCount = 0` is meaningless, yet a caller iterating over a slice of `Report` values would treat it as "a strategy with no trades". Separation enforces the distinction at the type level.

`BenchmarkReport` is placed in `internal/analytics` alongside `Report`, making it a peer — both are analytics outputs, just for different subjects.

## Consequences

- `output.Config` gains a `Benchmark *analytics.BenchmarkReport` optional pointer; when nil, output is unchanged (backwards-compatible).
- Future metrics unique to the benchmark (e.g., index-relative metrics, dividend inclusion) can be added to `BenchmarkReport` without touching `Report`.
- If a future multi-strategy report is added, it will similarly get its own struct rather than extending either of these.

## Related decisions

- [io.Writer field in Config for stdout testability](../convention/2026-04-09-io-writer-in-config-for-stdout-testability.md) — establishes the Config-injection pattern; `Benchmark *BenchmarkReport` follows the same pattern (optional field, nil = feature off).
