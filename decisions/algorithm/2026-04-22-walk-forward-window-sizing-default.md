# Walk-forward window sizing defaults (2yr IS / 1yr OOS)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | walk-forward, window-size, rolling, IS-OOS, regime-coverage, TASK-0022 |

## Context

WalkForwardConfig exposes InSampleWindow, OutOfSampleWindow, and StepSize as configurable parameters. The question was what the recommended defaults should be for daily-bar NSE equity strategies over the 2018–2024 baseline period, given that both SMA crossover and RSI mean-reversion produce approximately 10–20 trades per year.

## Options considered

### Option A: 2yr IS / 1yr OOS, 1yr step (selected)
- **Pros**: 4 folds over [2018-01-01, 2025-01-01); each OOS fold ~250 trading days; 2024 naturally held out; fold coverage includes pre-COVID, COVID crash/recovery, 2022 sideways
- **Cons**: 10–20 OOS trades per fold — too few to give per-fold Sharpe statistical reliability

### Option B: 1yr IS / 1yr OOS, 1yr step
- **Pros**: More folds; more regime variation visible
- **Cons**: 1yr IS window may not contain enough trades to characterize strategy behavior; IS window too short for low-frequency strategies

### Option C: 3yr IS / 1yr OOS, 1yr step
- **Pros**: More IS data per fold
- **Cons**: Only 3 folds over 2018–2024; reduces regime diversity

### Option D: Expanding windows
- **Pros**: Matches how a live system would retrain (if it trained at all)
- **Cons**: No fitting occurs — the expanding-window argument doesn't apply to stateless strategies. Early folds appear in nearly every subsequent fold, diluting regime variation.

## Decision

Fixed rolling windows: 2-year IS / 1-year OOS / 1-year step. This produces 4 folds over [2018-01-01, 2025-01-01): IS 2018–2020/OOS 2020, IS 2019–2021/OOS 2021, IS 2020–2022/OOS 2022, IS 2021–2023/OOS 2023. The 2024 calendar year falls outside as a natural consequence of the window arithmetic and is held out as a final test — not enforced by the framework, emergent from it. Fold-level Sharpe with 10–20 OOS trades is treated as indicative only; the aggregate across folds and the fold-level sign distribution carry the weight.

## Consequences

The WalkForwardConfig default values should reflect these parameters. CLI documentation should explain that per-fold Sharpe is indicative for low-frequency strategies.

## Revisit trigger

If a future strategy produces >100 trades/year on daily bars, shorter windows become viable and the 2yr/1yr defaults may be unnecessarily conservative.
