# Kill-switch derivation methodology

| Field    | Value        |
|----------|--------------|
| Date     | 2026-04-21   |
| Status   | accepted     |
| Category | algorithm    |
| Tags     | kill-switch, live-monitoring, sharpe, drawdown, bootstrap, TASK-0026 |

## Context

Before any strategy runs with real capital, three halt conditions must be defined and committed to in writing. Without them, a normal drawdown turns into parameter tweaking and re-running — the mechanism by which a system becomes overfit to its own live history.

The kill-switch has one rule when hit: halt and re-evaluate from scratch. Never retune parameters while in a drawdown.

## The three thresholds

### 1. Rolling per-trade Sharpe threshold

**Source:** p5 of the bootstrap distribution from `internal/montecarlo.Bootstrap`.

**Live computation:** rolling window of recent closed trades (typically 6 months of trade history). Per-trade Sharpe = `mean(ReturnOnNotional) / std(ReturnOnNotional)`, sample variance (n-1), **no annualization factor**.

**Trigger:** rolling per-trade Sharpe drops below the p5 threshold.

**Why p5:** p5 is not the expected level (p50); it is the level below which only 5% of bootstrap simulations landed. Breaching it means the live performance is in the left tail of what the backtest said was possible. That is not a drawdown — that is evidence the edge may not be there.

**Critical constraint:** the live computation and the bootstrap computation must use the same formula. `DeriveKillSwitchThresholds` takes `montecarlo.BootstrapResult.SharpeP5` as a raw `float64` and `CheckKillSwitch` computes the rolling Sharpe with `computePerTradeSharpe` — both use `mean(r)/std(r)` on `ReturnOnNotional`, sample variance, no annualization. This is enforced by the TASK-0024 algorithm decision.

### 2. Maximum drawdown threshold

**Formula:** `1.5 × Report.MaxDrawdown` (in-sample worst drawdown, percent 0–100).

**Live computation:** current drawdown from all-time equity peak to latest bar, via `computeCurrentDrawdownDepth`.

**Trigger:** current drawdown exceeds the threshold.

**Why 1.5×:** the in-sample worst drawdown is the most extreme thing the strategy experienced in training. 1.5× gives a half-sigma buffer. A strategy that breaches 1.5× in-sample max DD is doing something the backtest never saw — that warrants a halt, not a wait.

### 3. Maximum drawdown recovery time threshold

**Formula:** `2 × Report.MaxDrawdownDuration` (in-sample worst recovery duration).

**Live computation:** time elapsed since the most recent all-time equity high, via `computeCurrentDDDuration`. Zero when equity is at or above its all-time high.

**Trigger:** strategy has been in continuous drawdown longer than the threshold.

**Why 2×:** the in-sample worst recovery sets the expectation. 2× is the point at which the drawdown is outlasting anything the backtest demonstrated — the strategy is in a regime the in-sample period never recovered from at this speed.

## Implementation

```
internal/analytics/killswitch.go
  KillSwitchThresholds    — struct holding the three thresholds
  KillSwitchAlert         — struct holding breach flags and current metric values
  DeriveKillSwitchThresholds(sharpeP5 float64, inSample Report) KillSwitchThresholds
  CheckKillSwitch(windowTrades []model.Trade, curve []model.EquityPoint, thresholds KillSwitchThresholds) KillSwitchAlert
```

`DeriveKillSwitchThresholds` is intentionally agnostic of `internal/montecarlo` — the caller passes `BootstrapResult.SharpeP5` as a plain `float64`. This keeps `internal/analytics` free of the simulation dependency.

## Per-strategy records

Each strategy that produces bootstrap results must have a companion decision file in `decisions/algorithm/` documenting its specific threshold values. These records are the pre-commitment — they must be written before any live deployment, not derived after observing live results.

See:
- [SMA Crossover kill-switch parameters](2026-04-21-kill-switch-sma-crossover.md)
- [RSI Mean-Reversion kill-switch parameters](2026-04-21-kill-switch-rsi-mean-reversion.md)
