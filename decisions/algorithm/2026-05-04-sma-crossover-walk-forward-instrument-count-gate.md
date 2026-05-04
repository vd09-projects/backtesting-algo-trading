# SMA crossover fails walk-forward instrument-count gate — 4 of 12 eligible instruments pass

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-04       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | sma-crossover, walk-forward, instrument-count-gate, TASK-0053, kill |

## Context

TASK-0053 walk-forward validation ran sma-crossover (fast=10, slow=20) on 12 eligible instruments from the TASK-0052 universe gate handoff. The instrument-count gate requires the strategy to pass walk-forward on at least as many instruments as it passed the universe gate (12). Data: 2018-01-01 to 2024-01-01, 2yr IS / 1yr OOS / 1yr step, 4 folds, commission=zerodha_full.

At the universe gate (TASK-0052), SMA crossover passed with a thin margin: DSRAvg=0.0969, PassFraction=12/14=85.7%. The revisit trigger set at that stage noted: "if fewer than 6 instruments pass walk-forward, consider whether SMA crossover has sufficient universe breadth for portfolio inclusion."

## Decision

SMA crossover is killed at the instrument-count gate. 4 of 12 instruments pass the per-instrument walk-forward gate; 8 fail. Required: 12 passes. Actual: 4 passes (33% retention). This is below the revisit trigger threshold of 6 instruments set at the universe gate stage.

**Passing instruments** (OverfitFlag=False, NegFoldFlag=False):

| Instrument | AvgISSharpe | AvgOOSSharpe | OOSISRatio | NegFoldCount |
|---|---|---|---|---|
| NSE:LT | 0.093825 | 0.103298 | 1.1010 | 1 |
| NSE:RELIANCE | 0.204219 | 0.313945 | 1.5373 | 1 |
| NSE:TITAN | 0.136220 | 0.168151 | 1.2344 | 0 |
| NSE:AXISBANK | -0.030905 | 0.070649 | -2.2860 | 1 |

**Failing instruments:**

| Instrument | Failure reason | Key metrics |
|---|---|---|
| NSE:INFY | OverfitFlag=True, NegFoldFlag=True, majority-negative-folds triggered | OOSISRatio=0.3887, NegFoldCount=3 |
| NSE:HINDUNILVR | OverfitFlag=True, NegFoldFlag=True | OOSISRatio=0.3437, NegFoldCount=2 |
| NSE:ICICIBANK | NegFoldFlag=True | NegFoldCount=2 |
| NSE:SBIN | OverfitFlag=True, NegFoldFlag=True | OOSISRatio=0.2828, NegFoldCount=2 |
| NSE:WIPRO | OverfitFlag=True, NegFoldFlag=True | OOSISRatio=-0.1737, AvgOOSSharpe=-0.017435, NegFoldCount=2 |
| NSE:TCS | OverfitFlag=True | OOSISRatio=0.3138 |
| NSE:ITC | OverfitFlag=True, NegFoldFlag=True | AvgOOSSharpe=-0.245697, NegFoldCount=2 |
| NSE:KOTAKBANK | OverfitFlag=True, NegFoldFlag=True, majority-negative-folds triggered | AvgOOSSharpe=-0.685197, NegFoldCount=4 |

Note: NSE:INFY and NSE:KOTAKBANK also trigger the majority-negative-folds rule (NegFoldCount > FoldCount/2 = 2).

## Consequences

SMA crossover does not advance to bootstrap (TASK-0054). The thin DSRAvg at the universe gate (0.0969) proved predictive — the strategy lacks robust cross-instrument regime stability at fast=10/slow=20 parameters. The 33% instrument retention rate and the scale of the failures (5 of 8 failing instruments have both OverfitFlag and NegFoldFlag simultaneously) suggest this is not a borderline kill.

The 4 passing instruments (LT, RELIANCE, TITAN, AXISBANK) show modest but positive OOS Sharpe. However, the gate requires all 12 universe-gate-passing instruments to also pass walk-forward — a 100% retention requirement that this strategy falls far short of.

## Related decisions

- [SMA crossover passes universe gate](./2026-05-03-sma-crossover-universe-gate-passed.md) — the prior gate; revisit trigger for <6 walk-forward passes was set here
- [Walk-forward OOS/IS Sharpe threshold](./2026-04-22-walk-forward-oos-is-sharpe-threshold.md) — defines OverfitFlag and NegativeFoldFlag thresholds applied here
- [Cross-instrument universe gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate that established the instrument-count retention requirement

## Revisit trigger

If the parameter space is explored (e.g., slower SMA periods like fast=20/slow=50 that generate fewer but higher-quality signals), re-test with a fresh universe sweep before re-entering the walk-forward gate. The current parameter set (fast=10, slow=20) is too fast for daily Nifty50 bars — it generates noise-driven signals that don't generalize across instruments or regimes.
