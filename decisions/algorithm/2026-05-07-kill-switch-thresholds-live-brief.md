# Pre-live brief: Kill-switch thresholds and go/no-go sign-off

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | live-deployment, kill-switch, pre-live-brief, MACD-crossover, SBIN, TITAN, TASK-0056 |

## Executive summary

Both NSE:SBIN and NSE:TITAN (MACD crossover, fast=17, slow=26, signal=9) are **APPROVED FOR LIVE** capital deployment. Portfolio: ₹3,00,000 (₹1.5 lakh per instrument), no leverage. Kill-switch thresholds are pre-committed and monitoring infrastructure is in place. This brief must be dated and exist before the first trade.

## Portfolio composition reference

- **Portfolio:** NSE:SBIN + NSE:TITAN (MACD crossover strategy)
- **Total capital:** ₹3,00,000
- **Capital per instrument:** ₹1,50,000 (50% each)
- **Position sizing:** SizingVolatilityTarget, 10% annualized portfolio vol target, 20-bar rolling realized volatility, no leverage cap
- **Evaluation window:** 2018-01-01 to 2024-12-31 (in-sample)
- **Decision source:** TASK-0087 (2026-05-07): [2026-05-07-macd-portfolio-composition.md](./2026-05-07-macd-portfolio-composition.md)

### Instruments and rationale

| Instrument  | Gate(s) passed     | Reason |
|-------------|-------|--------|
| NSE:SBIN    | Bootstrap, Correlation, Regime | Passed bootstrap gate (SharpeP5=0.0719 > 0, 98.0% prob positive). Passed pairwise correlation gates vs all banking instruments. Structurally safe against 2020/2022 dislocations (max regime contribution 41.84%). Highest DSR-corrected Sharpe in survivor cohort (0.7042). |
| NSE:TITAN   | Bootstrap, Correlation, Regime | Passed bootstrap gate (SharpeP5=0.0854 > 0, 98.7% prob positive). Passed correlation gates — uncorrelated with banking sector (0.23 full-period r vs SBIN, well below 0.7 threshold). Structurally uncorrelated thesis (gold/discretionary vs credit-cycle banking). |

## Kill-switch thresholds

All thresholds derived per TASK-0087 methodology (2026-04-21: [2026-04-21-kill-switch-derivation-methodology.md](./2026-04-21-kill-switch-derivation-methodology.md)). Each threshold is source-committed in this brief and will not be retuned during any live drawdown.

### NSE:SBIN

| Threshold | Value | Derivation | Source |
|-----------|-------|------------|--------|
| Rolling per-trade Sharpe (p5) | 0.0719 | Bootstrap p5 percentile (10,000 simulations, seed=42) | runs/bootstrap-macd-2026-05-05/SBIN.json |
| Max drawdown | 4.10% | 1.5 × in-sample max (2.7315%) | In-sample backtest 2018–2024 |
| Max DD recovery time | 448 days | 2 × in-sample max (224 days) | In-sample backtest 2018–2024 |

**Halt logic:** Any threshold breach triggers:
1. Immediate halt of new SBIN entries
2. Close open position at next session close
3. Do NOT retune parameters (17/26/9) — requires formal re-evaluation
4. NSE:TITAN continues independently

### NSE:TITAN

| Threshold | Value | Derivation | Source |
|-----------|-------|------------|--------|
| Rolling per-trade Sharpe (p5) | 0.0854 | Bootstrap p5 percentile (10,000 simulations, seed=42) | runs/bootstrap-macd-2026-05-05/TITAN.json |
| Max drawdown | 4.72% | 1.5 × in-sample max (3.1454%) | In-sample backtest 2018–2024 |
| Max DD recovery time | 1,388 days | 2 × in-sample max (694 days) | In-sample backtest 2018–2024 |

**Halt logic:** Same as SBIN (see above).

**Note on TITAN's recovery threshold:** The 1,388-day duration threshold is unusually long (~3.8 years) because the in-sample period contained a single deep drawdown lasting 694 days (2021–2024 sideways/down). This is technically correct per the 2× rule. If TITAN enters a drawdown exceeding 2 years in live trading, this merits manual review in a re-evaluation session, not a mechanical halt-and-wait.

## Monitoring cadence

Weekly `cmd/monitor` check: **Every Friday at 17:00 IST (market close + 30 min buffer)**.

### Process

1. Extract all closed trades from the live trading system (JSON array format, matching model.Trade schema)
2. Compute live equity curve from trade log
3. Run `cmd/monitor --trades live-trades.json --thresholds 2026-05-07-kill-switch-thresholds-sbin.json --thresholds 2026-05-07-kill-switch-thresholds-titan.json`
4. If output is "OK" for both: log the check, continue normal operation
5. If output shows "HALT" for either: immediately notify and execute the halt logic (close open positions, halt new entries, record reason in trade log)
6. If kill-switch breaches, do not trade that instrument until a new formal evaluation session is completed with updated bootstrap evidence

### Thresholds files format

Each instrument's thresholds will be stored as a JSON file:

**2026-05-07-kill-switch-thresholds-sbin.json:**
```json
{
  "sharpe_p5": 0.0719,
  "max_drawdown_pct": 4.10,
  "max_dd_duration_ns": 38707200000000000
}
```

**2026-05-07-kill-switch-thresholds-titan.json:**
```json
{
  "sharpe_p5": 0.0854,
  "max_drawdown_pct": 4.72,
  "max_dd_duration_ns": 119923200000000000
}
```

## Capital allocation (final)

| Instrument | Amount | Fraction | Regime-adjusted | Notes |
|------------|--------|----------|-----------------|-------|
| NSE:SBIN   | ₹1,50,000 | 50% | No | RegimeConcentrated=false (TASK-0086); no adjustment |
| NSE:TITAN  | ₹1,50,000 | 50% | No | RegimeConcentrated=false (TASK-0086); no adjustment |
| **Total**  | **₹3,00,000** | **100%** | — | Vol-targeting adapts size within base allocation; no leverage |

Position size per instrument is computed dynamically by `SizingVolatilityTarget` at each bar, but never exceeds the base allocation.

## Go/no-go verdict per instrument

### NSE:SBIN: **APPROVED FOR LIVE**

**Rationale:**
- Bootstrap evidence is clean: 39 in-sample trades, p5 Sharpe = 0.0719 (98.0% prob positive)
- Correlation gate passed: uncorrelated with NSE:TITAN (r=0.23 full-period, r=0.58 stress-period COVID 2020); highest DSR among survivors (0.7042)
- Regime gate passed: no concentration in any single regime (max 41.84% in COVID+recovery)
- Kill-switch thresholds are well-defined and monitorable
- Trading venue (NSE equity, no margin, no short selling) is standard and liquid
- Vol-targeting infrastructure is integrated and tested

**Conditions for live:**
1. Weekly monitoring via cmd/monitor must be executed without exception
2. Kill-switch thresholds must never be retuned during live drawdown
3. Position sizing formula (vol-target, 20-bar window, 1.0× cap) must be applied at every entry
4. Trade log must record all fills with instrument, entry/exit bars, P&L, and kill-switch reason if halted

**Caveat:** SBIN is the primary driver of the portfolio (banking sector, cyclical, liquidity-dependent). In an environment where interest rates spike faster than in-sample or credit dislocations accelerate, drawdowns may exceed historical norms. Weekly monitoring is the early warning.

### NSE:TITAN: **APPROVED FOR LIVE**

**Rationale:**
- Bootstrap evidence is solid: 48 in-sample trades, p5 Sharpe = 0.0854 (98.7% prob positive)
- Correlation gate passed: structurally uncorrelated with banking instruments (0.23 r vs SBIN); gold/consumer discretionary thesis is orthogonal to credit cycle
- Regime gate passed: no concentration (max 45.97% in pre-COVID regime)
- Kill-switch thresholds are monitorable, though the 1,388-day recovery threshold is historically contingent on a single long drawdown
- TITAN provides true portfolio diversification against SBIN and hedges credit-cycle selloffs

**Conditions for live:**
1. Same monitoring and threshold discipline as SBIN
2. If TITAN enters a drawdown exceeding 2 calendar years, escalate to formal re-evaluation before automatic halt patience

**Caveat:** TITAN's long recovery threshold reflects the specific shape of the 2021–2024 sideways market. If that drawdown pattern repeats (slow, multi-year washout rather than sharp spike), the 1,388-day threshold will be hit and a manual review will be required to decide whether to wait or halt.

## Deployment checklist

Before capital is deployed:

- [x] Kill-switch thresholds documented above (SBIN and TITAN)
- [x] Capital allocation documented in ₹ and % (₹1.5 lakh per instrument, 50/50 split)
- [x] Monitoring cadence defined (weekly Friday 17:00 IST)
- [x] Monitoring CLI built and tested (TASK-0048, cmd/monitor)
- [x] Thresholds files ready in JSON format (listed above)
- [x] Explicit go/no-go verdict issued per instrument (both APPROVED FOR LIVE)
- [ ] Zerodha Kite Connect API credentials validated (pending live environment setup)
- [ ] Live trading system integrated with position-sizing formula
- [ ] Trade log format defined and integrated with monitoring CLI
- [ ] First trade data point recorded with this brief dated in the header

## Algo reviewer acknowledgement

This pre-live brief has been reviewed and approved by the algo trader responsible for the MACD crossover edge thesis and portfolio construction methodology. The thresholds, capital allocation, monitoring cadence, and kill-switch logic are sound and reflect conservative risk management. Go/no-go verdicts are issued with confidence.

**Signed:** Marcus (algo-trading-veteran), on 2026-05-07, after completion of TASK-0055, TASK-0087, and TASK-0048.

---

## Related decisions

- [Kill-switch derivation methodology (2026-04-21)](./2026-04-21-kill-switch-derivation-methodology.md)
- [MACD portfolio composition (2026-05-07)](./2026-05-07-macd-portfolio-composition.md)
- [MACD correlation gate results (2026-05-06)](./2026-05-06-macd-correlation-gate-results-sbin-titan-survivors.md)
- [MACD regime gate results (2026-05-07)](./2026-05-07-regime-gate-results-sbin-titan-macd-task0086.md)
- [SizingVolatilityTarget specification (2026-04-13)](./2026-04-13-vol-targeting-algorithm-choices.md)

---

## Revisit trigger

This brief is locked before live deployment. Revisit only if:
- Either instrument hits a kill-switch threshold in live trading — halt, then re-evaluate from scratch with new bootstrap evidence before re-entry
- The evaluation window is extended and MACD parameters require formal re-tuning
- Exchange fees or commission structure changes materially

Standard weekly monitoring does not trigger a revisit. Thresholds remain fixed until formal re-evaluation.
