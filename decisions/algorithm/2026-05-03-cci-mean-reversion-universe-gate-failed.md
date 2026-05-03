# CCI mean-reversion fails universe gate — DSR-corrected average Sharpe negative

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | cci-mean-reversion, universe-gate, DSR, thin-sample, TASK-0052, kill |

## Context

TASK-0052 ran `cmd/universe-sweep` for cci-mean-reversion (CCI period=20, entry on CCI < -100, exit on CCI cross above 0, long-only) across all 15 Nifty50 large-cap instruments, 2018-01-01 to 2024-01-01, with `--commission zerodha_full`. The universe gate requires: DSR-corrected average Sharpe > 0 AND >= 40% of sufficient instruments show positive Sharpe.

A sufficient instrument is one where `insufficient_data=false` (trade_count >= 30 per gate design). CCI at period=20 produced 31–38 trades per sufficient instrument — a thin but passing sample.

The signal frequency audit (2026-05-02) had pre-cleared CCI: avg 35.3 trades/instrument, 0 COVID clustering violations. The signal audit used a different trade count (the audit run showed higher counts such as NSE:RELIANCE=42) because the audit ran to 2024-01-01 with different fill timing than the universe sweep. The universe sweep counts (31–38) are the authoritative figures for gate purposes.

## Options considered

N/A — gate application, not a design decision. The gate criteria are fixed (2026-04-25).

## Decision

**CCI mean-reversion is killed.** Gate applied as specified.

**Numeric evidence:**

- Sufficient instruments (trade_count >= 30, insufficient_data=false): 12 of 15
  - NSE:ITC (36 trades, raw Sharpe=+0.669)
  - NSE:HDFCBANK (38 trades, raw Sharpe=+0.443)
  - NSE:ICICIBANK (32 trades, raw Sharpe=+0.420)
  - NSE:RELIANCE (33 trades, raw Sharpe=+0.291)
  - NSE:TITAN (32 trades, raw Sharpe=+0.282)
  - NSE:TCS (31 trades, raw Sharpe=+0.251)
  - NSE:AXISBANK (34 trades, raw Sharpe=+0.207)
  - NSE:HINDUNILVR (30 trades, raw Sharpe=+0.094)
  - NSE:BAJFINANCE (31 trades, raw Sharpe=+0.060)
  - NSE:KOTAKBANK (32 trades, raw Sharpe=-0.016)
  - NSE:SBIN (32 trades, raw Sharpe=-0.134)
  - NSE:WIPRO (33 trades, raw Sharpe=-0.173)
- Excluded (insufficient_data=true, trade_count < 30): NSE:LT (28), NSE:INFY (28), NSE:MARUTI (28)
- SufficientInstrumentCount: 12 — passes the >= 10 floor
- PositiveSharpe instruments: 9/12 = 75.0% — passes the >= 40% condition
- Raw average Sharpe (12 sufficient instruments): +0.1996
- DSR-corrected average Sharpe (nTrials=12, ~31–36 trades/instrument): **-0.0960** — **fails DSRAvg > 0 condition**

The DSR correction penalises the multiple-comparison problem inherent in testing across 12 instruments. With a thin sample (~31–36 trades/instrument), the standard error on each per-instrument Sharpe is high, and testing 12 instruments raises nTrials sufficiently that the correction pushes the already-modest raw average negative.

Three instruments stand out individually: NSE:ITC (DSR-corrected Sharpe ≈ +0.388), NSE:HDFCBANK (DSR-corrected Sharpe ≈ +0.169), and NSE:ICICIBANK (positive but close to boundary). These cleared DSR correction on their own, but this does not permit re-testing CCI in isolation on those instruments. Instrument-specific edge requires a separate test structure per gate design — the universe sweep evaluates the strategy's breadth of edge, not cherry-picked pairs.

**Gates passed before kill:**

| Gate condition | Result |
|----------------|--------|
| SufficientInstrumentCount >= 10 | PASS (12 of 15) |
| PassFraction >= 0.40 (raw positive Sharpe) | PASS (9/12 = 0.750) |
| DSR-corrected average Sharpe > 0 | **FAIL (-0.0960)** |

## Consequences

- CCI mean-reversion does not advance to walk-forward (TASK-0053).
- TASK-0053 proceeds with the two survivors from the original six-strategy evaluation: macd-crossover (14 eligible instruments) and sma-crossover (12 eligible instruments). The CCI kill does not alter the TASK-0053 survivor list.
- The pre-committed tiebreaker (prefer Bollinger over CCI at correlation gate if both survive) is now moot — both Bollinger and CCI are killed, for different reasons. Bollinger failed at the sufficient-instrument count stage; CCI failed the DSR correction despite a higher pass fraction and raw average.
- The signal audit (2026-05-02) correctly identified CCI as viable on frequency grounds. The universe sweep gate is a harder test — frequency is necessary but not sufficient for passing the DSR-corrected Sharpe requirement.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate applied here
- [CCI mean-reversion signal audit: PROCEED](./2026-05-02-cci-mean-reversion-signal-audit-proceed.md) — pre-pipeline frequency audit that cleared CCI to enter the pipeline
- [TASK-0051 routing decision](./2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity.md) — pipeline routing context

## Revisit trigger

If the project revisits CCI with a materially different period or entry/exit logic that produces 50+ trades/instrument (reducing DSR correction sensitivity), this kill should be reconsidered. At ~31–36 trades, the standard error on the per-instrument Sharpe is large enough that even a genuine edge can be erased by the correction. The instrument-specific signals (ITC, HDFCBANK) are noted but do not constitute grounds for cherry-picked re-evaluation under the current gate design.
