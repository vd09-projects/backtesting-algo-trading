# macd-crossover × NSE:ICICIBANK — killed at correlation-gate-stress-covid

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | macd-crossover, ICICIBANK, correlation-gate, stress-covid, rejected, TASK-0055 |

## Decision

**macd-crossover × NSE:ICICIBANK is killed. Gate failed: correlation-gate-stress-covid.**

SBIN–ICICIBANK COVID crash stress-period Pearson r = 0.6722, threshold < 0.6 (exclusive). The gate fails.

DSR tiebreaker applies: SBIN (DSR=0.7042) > ICICIBANK (DSR=0.3816). SBIN is retained; ICICIBANK is excluded.

## Gates passed before kill

| Gate                          | Metric                                                   | Result |
|-------------------------------|----------------------------------------------------------|--------|
| bootstrap_gate                | SharpeP5=0.0229, ProbPositiveSharpe=96.2%                | PASS   |
| correlation-gate-full-period  | SBIN–ICICIBANK full-period r=0.4243 (threshold < 0.7)    | PASS   |
| correlation-gate-stress-covid | SBIN–ICICIBANK COVID r=0.6722 (threshold < 0.6)          | FAIL   |

## Kill evidence

- Gate: correlation-gate-stress-covid
- Pair: SBIN–ICICIBANK
- COVID crash window: 2020-02-01 to 2020-06-30
- Observed r: 0.6722
- Threshold: < 0.6 (exclusive)

ICICIBANK is a large private sector bank. During the COVID crash (Feb–Jun 2020), SBIN fell approximately 45% and ICICIBANK fell approximately 42% from January highs — near-identical crash behaviour driven by rate sensitivity and credit cycle exposure. The stress-period correlation (0.6722) exceeded the gate threshold despite a moderate full-period r (0.4243), confirming that the co-movement is regime-conditional and concentrated in dislocations.

ICICIBANK also has the lowest DSR-corrected Sharpe among the four bootstrap survivors (DSR=0.3816), making the retention decision decisive in favour of SBIN.

## What this does NOT affect

ICICIBANK bootstrap outcomes are valid — it passed the bootstrap gate with SharpeP5=0.0229 and ProbPositiveSharpe=96.2%. This kill is a portfolio diversification decision, not a signal quality failure. If MACD crossover is re-evaluated on a universe that does not already include SBIN, ICICIBANK remains a candidate instrument.

## Related decisions

- [Correlation gate design (2026-04-27)](./2026-04-27-correlation-gate.md) — gate thresholds applied here
- [MACD bootstrap gate results (2026-05-05)](./2026-05-05-macd-bootstrap-gate-results.md) — source of ICICIBANK as a survivor before this gate
- [Banking cluster — SBIN, BAJFINANCE, ICICIBANK structural correlation (2026-05-06)](./2026-05-06-banking-cluster-sbin-bajfinance-icicibank.md) — Marcus's prior rationale
- [MACD correlation gate results — SBIN and TITAN survive (2026-05-06)](./2026-05-06-macd-correlation-gate-results-sbin-titan-survivors.md) — full pairwise correlation matrix
