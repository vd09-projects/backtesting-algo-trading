# macd-crossover × NSE:BAJFINANCE — killed at correlation-gate-stress-covid

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | macd-crossover, BAJFINANCE, correlation-gate, stress-covid, rejected, TASK-0055 |

## Decision

**macd-crossover × NSE:BAJFINANCE is killed. Gate failed: correlation-gate-stress-covid.**

SBIN–BAJFINANCE COVID crash stress-period Pearson r = 0.7176, threshold < 0.6 (exclusive). The gate fails.

DSR tiebreaker applies: SBIN (DSR=0.7042) > BAJFINANCE (DSR=0.6120). SBIN is retained; BAJFINANCE is excluded.

## Gates passed before kill

| Gate                          | Metric                                               | Result |
|-------------------------------|------------------------------------------------------|--------|
| bootstrap_gate                | SharpeP5=0.0467, ProbPositiveSharpe=97.3%            | PASS   |
| correlation-gate-full-period  | SBIN–BAJFINANCE full-period r=0.3857 (threshold < 0.7) | PASS   |
| correlation-gate-stress-covid | SBIN–BAJFINANCE COVID r=0.7176 (threshold < 0.6)     | FAIL   |

## Kill evidence

- Gate: correlation-gate-stress-covid
- Pair: SBIN–BAJFINANCE
- COVID crash window: 2020-02-01 to 2020-06-30
- Observed r: 0.7176
- Threshold: < 0.6 (exclusive)

BAJFINANCE is an NBFC (non-banking financial company). During the COVID crash, credit-quality fears and moratorium uncertainty drove NBFC and PSU bank stocks together as a sector. The stress-period correlation exceeded both the full-period pattern and the gate threshold.

## What this does NOT affect

BAJFINANCE bootstrap outcomes are valid — it passed the bootstrap gate with SharpeP5=0.0467 and ProbPositiveSharpe=97.3%. This kill is a portfolio diversification decision, not a signal quality failure. If MACD crossover is re-evaluated on a universe that does not already include SBIN, BAJFINANCE remains a candidate instrument.

## Related decisions

- [Correlation gate design (2026-04-27)](./2026-04-27-correlation-gate.md) — gate thresholds applied here
- [MACD bootstrap gate results (2026-05-05)](./2026-05-05-macd-bootstrap-gate-results.md) — source of BAJFINANCE as a survivor before this gate
- [Banking cluster — SBIN, BAJFINANCE, ICICIBANK structural correlation (2026-05-06)](./2026-05-06-banking-cluster-sbin-bajfinance-icicibank.md) — Marcus's prior rationale
- [MACD correlation gate results — SBIN and TITAN survive (2026-05-06)](./2026-05-06-macd-correlation-gate-results-sbin-titan-survivors.md) — full pairwise correlation matrix
