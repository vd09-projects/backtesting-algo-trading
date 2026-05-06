# MACD crossover portfolio composition — final portfolio, sizing, and kill-switch thresholds

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | portfolio-composition, macd-crossover, SBIN, TITAN, vol-targeting, kill-switch, TASK-0087, live-deployment |

## Context

TASK-0055 (cross-strategy correlation and portfolio construction) is complete. The correlation gate (TASK-0085, 2026-05-06) and regime gate (TASK-0086, 2026-05-07) have produced the final two-instrument portfolio: NSE:SBIN and NSE:TITAN. Both instruments passed all gates and carry confirmed kill-switch thresholds derived from bootstrap p5 Sharpe and maximum drawdown metrics. This decision file records the final portfolio composition, capital allocation, sizing rule, and live-deployment parameters — all pre-committed before any live capital is allocated.

## Final portfolio composition

**Portfolio:** NSE:SBIN + NSE:TITAN (MACD crossover, fast=17, slow=26, signal=9)

**Evaluation period:** 2018-01-01 to 2024-12-31 (in-sample)

**Total notional capital:** ₹3,00,000 (three lakh rupees)

### Instrument inclusion and exclusion rationale

| Instrument    | Status      | Reason |
|---------------|-------------|--------|
| NSE:SBIN      | **INCLUDED** | Passed bootstrap gate (SharpeP5=0.0719 > 0, P(Sharpe > 0)=98.0%). Passed all three pairwise correlation gates (SBIN–TITAN, SBIN–BAJFINANCE, SBIN–ICICIBANK). Highest DSR-corrected Sharpe in the surviving bootstrap cohort (DSR=0.7042). Passed regime gate (RegimeConcentrated=false, max regime contribution 41.84%). |
| NSE:TITAN     | **INCLUDED** | Passed bootstrap gate (SharpeP5=0.0854 > 0, P(Sharpe > 0)=98.7%). Passed all pairwise correlation gates (TITAN–SBIN, TITAN–BAJFINANCE, TITAN–ICICIBANK). Structurally uncorrelated with banking instruments (gold/consumer discretionary thesis vs credit-cycle banking thesis). Passed regime gate (RegimeConcentrated=false, max regime contribution 45.97%). |
| NSE:BAJFINANCE | **EXCLUDED** | Passed bootstrap gate but failed correlation gate: COVID crash stress-period Pearson r=0.7176 vs SBIN (threshold r < 0.6). Excluded per DSR tiebreaker: SBIN (0.7042) > BAJFINANCE (0.6120). Recorded as excluded (correlation), not killed. |
| NSE:ICICIBANK | **EXCLUDED** | Passed bootstrap gate but failed correlation gate: COVID crash stress-period Pearson r=0.6722 vs SBIN (threshold r < 0.6). Excluded per DSR tiebreaker: SBIN (0.7042) >> ICICIBANK (0.3816). Recorded as excluded (correlation), not killed. |

The banking cluster (SBIN, BAJFINANCE, ICICIBANK) exhibited high correlation during equity dislocations (COVID 2020: r > 0.67). TITAN is structurally uncorrelated, making SBIN + TITAN the optimal two-instrument portfolio per the correlation gate.

## Capital allocation

**Base notional allocation (pre-vol-targeting):**

| Instrument | Allocation    | Fraction |
|------------|---------------|---------:|
| NSE:SBIN   | ₹1,50,000     | 50%      |
| NSE:TITAN  | ₹1,50,000     | 50%      |
| **Total**  | **₹3,00,000** | **100%** |

**Regime gate outcome:** Both NSE:SBIN and NSE:TITAN have RegimeConcentrated=false. No allocation adjustment applied.

**Final base allocation:** Unchanged from above. The base allocation remains fixed at ₹1.5 lakh per instrument, subject to vol-targeting sizing (see below).

## Position sizing — SizingVolatilityTarget

Per-instrument position size is governed by `SizingVolatilityTarget`:

```
fraction = volTarget / (instrumentVol × sqrt(252))
fraction = min(fraction, 1.0)  // no leverage
position_size = base_allocation × fraction
```

### Parameters

| Parameter        | Value           | Notes |
|------------------|-----------------|-------|
| `volTarget`      | 0.10            | 10% annualized portfolio volatility target |
| `instrumentVol`  | 20-bar rolling std dev of log-returns | Recomputed each bar; adapts to realized volatility regime |
| `sqrt(252)`      | 15.8745         | Annualisation factor for daily bars |
| Leverage cap     | 1.0 (no leverage)| Position size never exceeds base allocation, even if realized vol drops below target |

### Behaviour

The vol-targeting formula scales position size inversely with realized volatility. In high-volatility regimes (e.g., COVID crash 2020, rate hikes 2022), position sizes shrink automatically. In low-volatility regimes, position sizes approach the base allocation up to the 1.0 leverage cap. This is the primary continuous risk management mechanism alongside kill-switch thresholds.

## Kill-switch thresholds — NSE:SBIN

Bootstrap source: `runs/bootstrap-macd-2026-05-05/SBIN.json`
In-sample backtest: 2018-01-01 to 2024-01-01, 39 trades
Bootstrap parameters: 10,000 simulations, seed=42

### Thresholds

| Metric | Value | Derivation |
|--------|-------|------------|
| **Rolling per-trade Sharpe (SharpeP5)** | 0.0719 | Bootstrap p5 percentile (from `internal/model/bootstrap.go` — lowest 5% of per-trade Sharpe across all 10,000 simulations). Breach triggers halt. |
| **Max drawdown (MaxDD)** | 4.10% | 1.5 × in-sample MaxDrawdown (2.7315% from SBIN.json). Applied to rolling equity curve. Breach triggers halt. |
| **Max DD duration (MaxDDDuration)** | 448 days | 2 × in-sample MaxDrawdownDuration (224 days from SBIN.json, converted from 19,353,600,000,000,000 ns). Counts consecutive days with equity ≤ running peak. Breach triggers halt. |

### Halt logic

When any SBIN threshold is breached:
1. Halt new SBIN entries immediately.
2. Close any open SBIN position at the next session close (or current bar close if session end is reached).
3. Record halt reason in trade log: "kill-switch — [SharpeP5|MaxDD|MaxDDDuration] threshold breached".
4. Do NOT retune MACD parameters (17/26/9) during halt. If re-entry is considered, require a new Marcus evaluation session with updated bootstrap evidence.
5. NSE:TITAN continues operating independently unless its own kill-switch triggers.

## Kill-switch thresholds — NSE:TITAN

Bootstrap source: `runs/bootstrap-macd-2026-05-05/TITAN.json`
In-sample backtest: 2018-01-01 to 2024-01-01, 48 trades
Bootstrap parameters: 10,000 simulations, seed=42

### Thresholds

| Metric | Value | Derivation |
|--------|-------|------------|
| **Rolling per-trade Sharpe (SharpeP5)** | 0.0854 | Bootstrap p5 percentile. Breach triggers halt. |
| **Max drawdown (MaxDD)** | 4.72% | 1.5 × in-sample MaxDrawdown (3.1454% from TITAN.json). Breach triggers halt. |
| **Max DD duration (MaxDDDuration)** | 1,388 days | 2 × in-sample MaxDrawdownDuration (694 days from TITAN.json, converted from 59,961,600,000,000,000 ns). Breach triggers halt. |

### Halt logic (same as SBIN)

When any TITAN threshold is breached:
1. Halt new TITAN entries immediately.
2. Close any open TITAN position at the next session close.
3. Record halt reason in trade log.
4. Do NOT retune parameters during halt.
5. NSE:SBIN continues operating independently unless its own kill-switch triggers.

### Note on TITAN's MaxDDDuration

TITAN's max DD duration threshold of 1,388 days (~3.8 years) is unusually long relative to the 6-year in-sample period. This reflects a single deep multi-year drawdown (2021–2024 sideways/down market) that lasted 694 days in the backtest. The threshold is technically correct per the 2×in-sample rule but should prompt qualitative review if TITAN enters drawdown exceeding 2 years in live trading — such extended drawdowns warrant a manual evaluation session rather than a mechanical halt wait.

## Regime gate outcome — final confirmation

**TASK-0086 (2026-05-07):** Both instruments evaluated against regime concentration gate (70% threshold on any single regime).

| Instrument | RegimeConcentrated | Max Regime Contribution | Allocation Adjustment |
|------------|--------------------|-----------------------|----------------------|
| NSE:SBIN   | false              | 41.84% (COVID+recovery)| none                 |
| NSE:TITAN  | false              | 45.97% (Pre-COVID)    | none                 |

Both pass the gate. Base allocation remains ₹1.5 lakh per instrument.

## Live deployment checklist

Before capital is allocated:

- [ ] Kill-switch threshold values are recorded in this file (above).
- [ ] A monitoring routine (`cmd/monitor`, TASK-0048) is in place to check thresholds weekly.
- [ ] Trade log format is decided (pending TASK-0048 decision on live trade log schema).
- [ ] Zerodha Kite Connect API credentials are validated and tested.
- [ ] The trading venue (NSE, standard equity, no margin/short selling) is confirmed.
- [ ] Position sizing is wired into the live execution system — vol-targeting formula is running before each order.
- [ ] Pre-live brief (TASK-0056) has been signed off by the algo reviewer.

## Related decisions

- [Kill-switch derivation methodology (2026-04-21)](./2026-04-21-kill-switch-derivation-methodology.md) — the framework these thresholds are derived from
- [Vol-targeting algorithm choices (2026-04-13)](./2026-04-13-vol-targeting-algorithm-choices.md) — SizingVolatilityTarget spec
- [MACD bootstrap gate results (2026-05-05)](./2026-05-05-macd-bootstrap-gate-results.md) — bootstrap p5 Sharpe source; 4 survivors
- [Banking cluster structural correlation (2026-05-06)](./2026-05-06-banking-cluster-sbin-bajfinance-icicibank.md) — prior expectation; why banking pairs fail
- [MACD correlation gate results (2026-05-06)](./2026-05-06-macd-correlation-gate-results-sbin-titan-survivors.md) — confirmed SBIN + TITAN survivors; BAJFINANCE and ICICIBANK excluded
- [MACD portfolio sizing — SBIN + TITAN (2026-05-06)](./2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md) — pre-committed capital allocation and kill-switch derivation
- [Regime gate results — SBIN and TITAN (2026-05-07)](./2026-05-07-regime-gate-results-sbin-titan-macd-task0086.md) — regime concentration confirmation; both pass

## Revisit trigger

This decision is locked before live deployment. Revisit and update only if:
- Live capital is halted on either instrument and subsequent evaluation produces new bootstrap evidence with different p5 thresholds.
- The evaluation window is extended (2024 data now available) and a new bootstrap run materially shifts kill-switch thresholds.
- Zerodha commission structure changes or exchange fees change, affecting backtest validity.

Standard live monitoring (weekly via `cmd/monitor`) does not trigger a revisit. Thresholds remain fixed until formal re-evaluation.
