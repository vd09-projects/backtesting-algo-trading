# MACD guard condition: n <= slow+signal-1 blocks talib zero-fill initialization boundary

- **Date:** 2026-04-26
- **Status:** accepted
- **Category:** convention
- **Tags:** macd, talib, initialization, guard, crossover, TASK-0041
- **Task:** TASK-0041

## Decision

The MACD strategy guards crossover detection with `if n <= slow+signal-1 { return SignalHold }`.
This condition is required because `talib.Macd` zero-fills uninitialized positions before the
lookback is satisfied, and a naive crossover scan from index 1 would find a spurious 0-to-nonzero
transition at the initialization boundary.

## Rationale

`talib.Macd(closes, fast, slow, signal)` fills the first `slow+signal-2` output positions with 0.
The first real signal-line value appears at index `slow+signal-2` (requiring `n = slow+signal-1`
input bars). With default parameters (12, 26, 9), that is bar 34 (0-indexed bar 33).

Without the guard, the strategy sees a MACD line that crosses from 0 to the first real value and
fires a spurious buy or sell signal on bar `slow+signal-1`. This is not a real crossover — it is
the talib library filling uninitialized memory with zeros.

The guard `n <= slow+signal-1` skips all bars up to and including bar `slow+signal-1`, meaning
crossover detection begins at bar `slow+signal` — the first bar where both the current and previous
MACD and signal values are computed from real data.

## Test consequence

Test helpers `goldenCrossBar` and `goldenDeadCrossBar` construct synthetic candle series starting
the crossover pattern at index `slow+signal-1` to stay in the valid computation region. Any test
constructing a crossover at an earlier index will produce a false positive or false negative and
should be treated as a test bug, not a strategy bug.

## Rejected alternatives

- **No guard, check for zero values explicitly** — fragile: a real MACD line can legitimately cross
  through zero, so filtering `macd == 0` would suppress real signals.
- **Lookback() = slow+signal** — Lookback already returns `slow+signal-1` (the minimum bars for the
  engine to start calling Next). Increasing it to `slow+signal` would skip one more bar unnecessarily.
  The guard is the correct place to enforce the initialization boundary inside Next.
- **Use talib.MacdLookback** — correct in theory, but talib.MacdLookback is not exposed in the
  `go-talib` bindings. The formula `slow+signal-2` is the standard TA-Lib lookback derivation.
