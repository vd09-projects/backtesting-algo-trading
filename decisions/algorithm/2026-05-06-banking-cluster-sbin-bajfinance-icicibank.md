# Banking cluster (SBIN, BAJFINANCE, ICICIBANK) treated as correlated for portfolio construction

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-06       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | correlation-gate, banking-cluster, instrument-selection, SBIN, ICICIBANK, BAJFINANCE, TITAN, MACD-crossover, TASK-0055 |

## Context

TASK-0055 (cross-strategy correlation and portfolio construction) requires applying the correlation gate (full-period Pearson r < 0.7 AND stress-period r < 0.6) to the 4 MACD crossover bootstrap survivors: SBIN, BAJFINANCE, TITAN, ICICIBANK. The correlation gate was designed for multi-strategy correlation, but here applies to same-strategy/different-instrument pairs. The relevant question is whether running MACD on a second banking instrument (ICICIBANK or BAJFINANCE) adds genuine diversification when the first (SBIN) is already in the portfolio.

Marcus evaluated the 6 pairwise combinations before actual correlation numbers were computed, based on structural sector analysis and historical behaviour during the COVID crash (2020) and rate-hike bear (2022).

## Decision

The banking cluster (SBIN, BAJFINANCE, ICICIBANK) is treated as structurally correlated instruments for portfolio construction purposes.

SBIN/ICICIBANK is expected to fail the correlation gate. Both are Nifty50 banking constituents — SBIN is PSU banking and ICICIBANK is private banking, but both are rate-sensitive and track Nifty Bank. In March 2020, SBIN fell approximately 45% and ICICIBANK fell approximately 42% from January highs — almost identical crash behaviour. The full-period Pearson r is expected to exceed 0.7, and the COVID stress-period r is expected to exceed 0.6.

SBIN/BAJFINANCE is assessed as borderline. BAJFINANCE is an NBFC (non-banking financial company), not a bank, but its price behaviour is correlated with banking sentiment and rate cycles. Estimated full-period r in the range 0.55–0.72. Stress-period correlation depends on the timing of the COVID-triggered credit stress hitting NBFCs.

TITAN is structurally uncorrelated with banking names. It is a gold-linked consumer discretionary company (jewelry retail). Its drawdown during COVID was driven by retail store closures and gold price volatility — a different mechanism from the credit-cycle concerns driving banking stocks. TITAN full-period r vs any of the three financial instruments is expected in the 0.45–0.60 range, passing the correlation gate.

The DSR tiebreaker (from the correlation gate decision 2026-04-27) applies when pairs fail: SBIN (DSR=0.7042) > BAJFINANCE (DSR=0.6120) > TITAN (DSR=0.6022) > ICICIBANK (DSR=0.3816). SBIN dominates ICICIBANK decisively. If SBIN/BAJFINANCE also fails, SBIN dominates BAJFINANCE.

**Expected final portfolio after running the correlation gate: NSE:SBIN + NSE:TITAN.**

## Consequences

- ICICIBANK is expected to be excluded with reason "excluded (correlation)" — not a gate failure. It was a valid strategy that happened to be correlated with SBIN, which has a higher DSR.
- BAJFINANCE's fate depends on actual Pearson r computation in TASK-0085. If borderline (r < 0.7 full-period AND r < 0.6 stress-period), it enters the portfolio as a third instrument at ~₹1 lakh notional alongside SBIN and TITAN.
- The correlation gate does not re-open the bootstrap gate. Excluded instruments are not killed — they are recorded as excluded (correlation) with the DSR tiebreaker rationale.
- If only SBIN passes from the banking cluster, the portfolio is 2-instrument (SBIN + TITAN) at ₹1.5 lakh each. If BAJFINANCE also passes, it becomes 3-instrument at ₹1 lakh each.

## Related decisions

- [Correlation gate design (2026-04-27)](./2026-04-27-correlation-gate.md) — the gate thresholds applied here; tiebreaker rule (DSR-corrected Sharpe) defined there
- [MACD bootstrap gate results (2026-05-05)](./2026-05-05-macd-bootstrap-gate-results.md) — the 4 survivors entering this gate; DSR values from TASK-0052 universe sweep cited in that decision
- [Portfolio sizing and kill-switch thresholds (2026-05-06)](./2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md) — downstream sizing decision conditioned on this outcome

## Revisit trigger

If actual pairwise Pearson r for SBIN/ICICIBANK < 0.7 full-period AND < 0.6 stress-period (both windows), admit both instruments to the portfolio and re-run the allocation math. This decision is a prior, not a finding — it must be validated against the computed correlation matrix in TASK-0085 before being treated as settled.
