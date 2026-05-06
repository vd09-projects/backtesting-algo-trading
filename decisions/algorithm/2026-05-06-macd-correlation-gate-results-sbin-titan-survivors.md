# MACD correlation gate results — SBIN and TITAN survive, BAJFINANCE and ICICIBANK excluded

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-06       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | macd-crossover, correlation-gate, pairwise-pearson, portfolio-construction, SBIN, BAJFINANCE, TITAN, ICICIBANK, DSR-tiebreaker, banking-cluster, TASK-0085 |

## Context

TASK-0069 (bootstrap gate) produced 4 survivors for MACD crossover (17/26/9): SBIN, BAJFINANCE, TITAN, and ICICIBANK. Before assembling a portfolio these instruments must pass the correlation gate defined in `decisions/algorithm/2026-04-27-correlation-gate.md`. The gate requires that every pair in the portfolio has:

- Full-period Pearson r < 0.7 (2018-01-01 to 2024-12-31)
- COVID crash Pearson r < 0.6 (2020-02-01 to 2020-06-30)
- Rate-hike bear Pearson r < 0.6 (2022-01-01 to 2022-12-31)

All three conditions must hold for the pair to coexist. Failing any one triggers the DSR-corrected Sharpe tiebreaker: retain the higher-DSR instrument, record the other as "excluded (correlation)".

Daily log-return equity curves were generated via `cmd/backtest` for each surviving instrument (MACD 17/26/9, 2018-2024, `--commission zerodha_full`) and saved to `runs/equity-curves-macd-2026-05-06/`. Pairwise correlation was computed using `cmd/correlate`.

A pre-run code fix was required: `internal/analytics/correlation.go` had wrong stress window dates — crash2020Start was 2020-01-14 (should be 2020-02-01) and correction2022End was 2022-06-30 (should be 2022-12-31). Both were corrected against the authoritative decision file before running. Tests updated to match.

## Correlation results

All 6 pairs evaluated. Gate thresholds: full-period r < 0.7 (exclusive), stress-period r < 0.6 (exclusive).

| Pair                  | Full-Period | COVID 2020 | Rate-Hike 2022 | Status  |
|-----------------------|-------------|------------|----------------|---------|
| SBIN – BAJFINANCE     | 0.3857      | **0.7176** | 0.3317         | **FAIL** |
| SBIN – TITAN          | 0.1785      | 0.4220     | 0.1860         | PASS     |
| SBIN – ICICIBANK      | 0.4243      | **0.6722** | 0.3525         | **FAIL** |
| BAJFINANCE – TITAN    | 0.2657      | 0.4651     | 0.3004         | PASS     |
| BAJFINANCE – ICICIBANK| 0.3637      | **0.7197** | 0.4150         | **FAIL** |
| TITAN – ICICIBANK     | 0.1838      | 0.4440     | 0.3004         | PASS     |

Three pairs fail on the COVID crash stress window (2020-02-01 to 2020-06-30). Full-period and rate-hike bear windows pass for all pairs. The stress-period correlation is the binding constraint — all three banking names move together during equity dislocations.

## Decision

**SBIN and TITAN are retained. BAJFINANCE and ICICIBANK are excluded (correlation).**

### DSR tiebreaker application

DSR-corrected Sharpe ranking (from universe sweep): SBIN=0.7042 > BAJFINANCE=0.6120 > TITAN=0.6022 > ICICIBANK=0.3816.

- SBIN–BAJFINANCE fail: retain SBIN (0.7042 > 0.6120), exclude BAJFINANCE.
- SBIN–ICICIBANK fail: retain SBIN (0.7042 > 0.3816), exclude ICICIBANK.
- BAJFINANCE–ICICIBANK fail: BAJFINANCE already excluded — no additional decision needed.

After applying tiebreaker, SBIN and TITAN form the surviving portfolio pair. TITAN passed all three pairs it was evaluated in (SBIN–TITAN, BAJFINANCE–TITAN, TITAN–ICICIBANK) and was never the lower-DSR instrument in a failing pair.

### Why this result was expected

All three failing pairs are banking names (SBIN=PSU bank, BAJFINANCE=NBFC, ICICIBANK=private bank). During COVID crash (Feb–Jun 2020), Indian banking and NBFC stocks sold off together: credit-quality fears, moratorium uncertainty, and liquidity stress drove the sector as a unit. Marcus anticipated SBIN/ICICIBANK correlation r ≈ 0.67 and SBIN/BAJFINANCE r ≈ 0.72 in the 2020 stress period. Actual values (0.6722 and 0.7176) are consistent with this prior.

TITAN is a luxury consumer/retail business (Tata Watches, Tanishq). Its equity curve is driven by consumer discretionary spending and brand premium, not credit cycles. This structural difference explains its low correlation with all banking names in both normal and stress regimes.

## Consequences

- Portfolio for MACD crossover (17/26/9) is SBIN + TITAN.
- Capital allocation: ₹1.5 lakh notional per instrument (₹3 lakh total), subject to vol-targeting sizing — pre-committed in `decisions/algorithm/2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md`.
- BAJFINANCE and ICICIBANK are recorded as "excluded (correlation)" — not gate failures. Both had valid bootstrap outcomes; they are excluded because a better-diversifying alternative existed within the same strategy family.
- TASK-0087 (portfolio composition file) can now be written using SBIN + TITAN as the confirmed instrument pair, pending TASK-0086 (regime gate).
- The stress-window dates correction (`crash2020Start` and `correction2022End` in `internal/analytics/correlation.go`) affects all future `cmd/correlate` runs. All correlation tests pass with the corrected dates.

## Related decisions

- [Correlation gate — maximum inter-strategy correlation](./2026-04-27-correlation-gate.md) — gate definition this decision applies
- [MACD crossover bootstrap gate: 4 survivors](./2026-05-05-macd-bootstrap-gate-results.md) — source of the 4 instruments evaluated here
- [MACD portfolio sizing — SBIN + TITAN, vol-targeting](./2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md) — capital allocation pre-committed before this gate ran
- [Banking cluster — SBIN, BAJFINANCE, ICICIBANK structural correlation](./2026-05-06-banking-cluster-sbin-bajfinance-icicibank.md) — Marcus's prior rationale for expecting this outcome

## Revisit trigger

If MACD crossover is re-evaluated on a different instrument universe (e.g., Nifty Midcap 150 — TASK-0072), correlation must be re-run on that universe's bootstrap survivors. The banking cluster result above is specific to the Nifty50 large-cap universe and the 2018–2024 evaluation window.
