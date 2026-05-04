# MACD crossover fails walk-forward instrument-count gate — 9 of 14 eligible instruments pass

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-04       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | macd-crossover, walk-forward, instrument-count-gate, TASK-0053, kill |

## Context

TASK-0053 walk-forward validation ran macd-crossover (fast=17, slow=26, signal=9) on 14 eligible instruments from the TASK-0052 universe gate handoff. The instrument-count gate requires the strategy to pass walk-forward on at least as many instruments as it passed the universe gate (14). Per the 2026-04-22 walk-forward threshold decision, each instrument×pair passes if OverfitFlag=false AND NegativeFoldFlag=false. Data: 2018-01-01 to 2024-01-01, 2yr IS / 1yr OOS / 1yr step, 4 folds, commission=zerodha_full.

At the universe gate (TASK-0052), MACD crossover was the strongest performer: DSRAvg=0.2715, PassFraction=14/15=93.3%. All 14 positive-Sharpe instruments were eligible for walk-forward.

## Decision

MACD crossover is killed at the instrument-count gate. 9 of 14 instruments pass the per-instrument walk-forward gate; 5 fail. Required: 14 passes. Actual: 9 passes (64% retention).

**Passing instruments** (OverfitFlag=False, NegFoldFlag=False):

| Instrument | AvgISSharpe | AvgOOSSharpe | OOSISRatio | NegFoldCount |
|---|---|---|---|---|
| NSE:SBIN | 0.341046 | 0.262855 | 0.7707 | 0 |
| NSE:BAJFINANCE | 0.258502 | 0.304601 | 1.1783 | 0 |
| NSE:TITAN | 0.332900 | 0.471607 | 1.4167 | 1 |
| NSE:LT | 0.264513 | 0.219953 | 0.8315 | 1 |
| NSE:ICICIBANK | 0.281298 | 0.232430 | 0.8263 | 0 |
| NSE:INFY | 0.224193 | 0.249079 | 1.1110 | 1 |
| NSE:AXISBANK | 0.080157 | 0.180131 | 2.2472 | 1 |
| NSE:ITC | 0.056336 | 0.214099 | 3.8004 | 1 |
| NSE:KOTAKBANK | -0.008622 | 0.061420 | -7.1235 | 1 |

**Failing instruments:**

| Instrument | Failure reason | Key metric |
|---|---|---|
| NSE:TCS | NegativeFoldFlag=True | NegFoldCount=2 |
| NSE:RELIANCE | OverfitFlag=True | OOSISRatio=0.4822 (< 0.50 threshold) |
| NSE:HINDUNILVR | OverfitFlag=True | OOSISRatio=0.3416 |
| NSE:WIPRO | OverfitFlag=True | OOSISRatio=0.4331 |
| NSE:HDFCBANK | NegativeFoldFlag=True | NegFoldCount=2 |

## Consequences

MACD crossover does not advance to bootstrap (TASK-0054). The pipeline from TASK-0052 terminates here for this strategy under the current gate design.

The 9 passing instruments show positive average OOS Sharpe ranging from 0.062 (KOTAKBANK) to 0.472 (TITAN), suggesting real regime-stable edge on a subset of the universe. The failures are clustered in two failure modes: OverfitFlag on large-cap defensives (RELIANCE, HINDUNILVR, WIPRO) and NegativeFoldFlag on higher-volatility instruments (TCS, HDFCBANK). A 64% pass rate on instruments is notable signal — but the gate design requires 100% retention relative to the universe gate count.

## Related decisions

- [MACD crossover passes universe gate](./2026-05-03-macd-crossover-universe-gate-passed.md) — the prior gate this strategy passed with DSRAvg=0.2715
- [Walk-forward OOS/IS Sharpe threshold](./2026-04-22-walk-forward-oos-is-sharpe-threshold.md) — defines OverfitFlag and NegativeFoldFlag thresholds applied here
- [Cross-instrument universe gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate that established the instrument-count retention requirement

## Revisit trigger

If the instrument-count gate threshold is revisited (e.g., relaxed from 100% retention to 60-70% retention), MACD crossover with 9/14 passing instruments should be re-evaluated — its per-instrument OOS Sharpes on passing pairs are materially positive. The overfit failures (RELIANCE, HINDUNILVR, WIPRO) may reflect regime concentration in those specific instruments rather than a fundamental strategy flaw.
