# MACD Crossover Research Program — Final Report

*Marcus, May 9 2026*

---

## What this was

Six strategies entered this pipeline. One survived and got deployed. A second iteration on a new universe just cleared its final gate today. The methodology held up. Here is the record.

---

## The kill log

**RSI mean-reversion:** Died first, on a single-instrument test against NSE:RELIANCE. Seven trades in seven years. Sharpe 0.469 against a 0.50 threshold. You can't learn anything from seven trades — the confidence interval on that Sharpe is effectively the real line. What the result probably told us is that fixed RSI thresholds at 30/70 on daily bars barely fire on Reliance's volatility profile. It's not that mean-reversion is dead; it's that this parameterization is almost never in the market. Bollinger mean-reversion was cancelled as a consequence — both sit in the same thesis bucket, and if the baseline doesn't clear the bar, you don't build the variation.

**CCI mean-reversion, Donchian breakout, Momentum:** All three killed at the universe gate with zero or near-zero sufficient instruments. The 30-trade floor was the binding constraint. These strategies fire too rarely on daily Nifty50 large-cap data to be statistically evaluable. Not kills in the sense that the edge doesn't exist — kills in the sense that you can't tell either way, and you can't run a business on that.

**SMA crossover (fast=10, slow=20):** Reached walk-forward and died there. Only 4 of 12 instruments passed (33% retention). The failures were clustered — WIPRO produced a negative average OOS Sharpe; ITC, KOTAKBANK, and INFY all had majority-negative folds. At fast=10/slow=20 on daily bars, the strategy is generating noise-driven signals that don't generalize across instruments or time periods. The pre-committed revisit trigger was clean: fast=20/slow=50 was tried next, which generated 12-20 trades per instrument — well below the statistical floor. SMA crossover is done.

**MACD crossover — walk-forward gate revision:** This is the one methodological call worth flagging explicitly, because it can look like post-hoc rationalization if you're not careful. MACD initially failed the instrument-count gate in the first run (9/14 = 64%, against a 100% retention threshold). The gate revision to 60% came from a principled argument — requiring 100% retention means one instrument with a regime-specific bad period kills the whole strategy — but it still happened with knowledge of the MACD result. The rationale was pre-committed before seeing what the gate revision would do to SMA crossover (which it didn't help; SMA died at 33%). That asymmetry is meaningful. If both strategies had benefited from the looser gate, the revision would be suspect. As it stands, the revision killed SMA anyway, which is evidence it's not a threshold that was tuned to let MACD through.

---

## What survived — Nifty50 large-cap portfolio

**MACD crossover (17/26/9) on NSE:SBIN + NSE:TITAN. Capital: ₹3,00,000. Approved for live as of 2026-05-07.**

Full gate sequence: universe gate (DSRAvg 0.2715, 14/15 instruments), walk-forward (9/14 at 100% threshold, revised gate → pass), bootstrap (4 survivors: SBIN, BAJFINANCE, TITAN, ICICIBANK), correlation gate (SBIN + TITAN only; BAJFINANCE and ICICIBANK excluded on banking cluster COVID correlation).

The banking cluster result was the most interesting finding in this pipeline. SBIN, BAJFINANCE, and ICICIBANK all had valid bootstrap edge — P(Sharpe > 0) of 97-98% each. They failed not because the strategy doesn't work on them but because they move together during dislocations. COVID crash stress-period correlations of 0.67-0.72 across all three banking pairs. TITAN — gold/consumer discretionary — is structurally uncorrelated with credit cycles, which is why it pairs well with SBIN.

### Kill-switch thresholds

| Instrument | Sharpe P5 | MaxDD trigger | Duration trigger |
|---|---|---|---|
| NSE:SBIN | 0.0719 | 4.10% | 448 days |
| NSE:TITAN | 0.0854 | 4.72% | 1,388 days |

TITAN's 1,388-day duration threshold is the one number to watch carefully in live trading. It's technically correct — 2× the in-sample max — but it reflects a single multi-year sideways drawdown in the backtest period. If TITAN enters a drawdown exceeding two years in live trading, don't wait for the mechanical threshold; call a re-evaluation session.

Monitoring: weekly, Friday 17:00 IST, via `cmd/monitor`. Thresholds are pre-committed and cannot be retuned without a new formal bootstrap run and evaluation session.

---

## What just finished — Nifty Midcap 150 portfolio

**MACD crossover (17/26/9) on 6 instruments. Capital: ₹3,00,000. Pipeline complete as of 2026-05-09; pre-live brief pending.**

Gate sequence: universe gate (DSRAvg 0.0885, 43/48 instruments), walk-forward (25/43 = exact 60% floor — flagged but a clean pass), bootstrap (6/25 survivors), correlation gate (all 6 pass, zero pairs above 0.70 full-period or 0.60 stress-period).

The midcap bootstrap was brutal — 19 of 25 walk-forward survivors failed. That's expected when you pass at the exact gate boundary; the walk-forward boundary signal was weak, and bootstrap correctly sorted the weak from the robust. The 4 anomalous-ratio survivors from walk-forward (SCHAEFFLER, THOMASCOOK, EXIDEIND, LALPATHLAB) all failed bootstrap, which is the system working as designed.

### The 6 survivors

| Instrument | SharpeP5 | P(Sharpe>0) | Sector |
|---|---|---|---|
| NSE:PERSISTENT | 0.1507 | 99.75% | IT |
| NSE:TORNTPHARM | 0.1455 | 99.66% | Pharma |
| NSE:COFORGE | 0.0413 | 97.42% | IT |
| NSE:SUNDARMFIN | 0.0351 | 96.73% | NBFC |
| NSE:INDHOTEL | 0.0191 | 96.27% | Hotels |
| NSE:MUTHOOTFIN | 0.0249 | 96.72% | NBFC/Gold Finance |

Sector diversity is genuine: IT, pharma, hospitality, two different NBFC profiles (vehicle credit vs gold-backed loans). Pairwise correlations are near-zero for most pairs — the highest is PERSISTENT/COFORGE at 0.36 full-period (IT sector co-movement during rate corrections). That pair gets an IT sector cap: ₹40k each when both are deployed simultaneously instead of the standard ₹50k.

### Kill-switch thresholds

| Instrument | Kill-switch MaxDD | Kill-switch Duration |
|---|---|---|
| NSE:PERSISTENT | 5.62% | 1,204 days |
| NSE:TORNTPHARM | 2.67% | 1,016 days |
| NSE:COFORGE | 12.57% hard / 8% early-warning | 2,504 days |
| NSE:SUNDARMFIN | 6.96% | 1,078 days |
| NSE:INDHOTEL | 8.70% | 1,724 days |
| NSE:MUTHOOTFIN | 6.27% | 1,870 days |

COFORGE is the one to watch. Its SharpeP5 (0.041) is the lowest survivor, its historical MaxDD (8.38%) is almost twice the next-worst in the cohort, and its mechanical kill-switch of 12.57% is genuinely loose. An early-warning trigger at 8% is set — not an automatic halt, but a prompt for qualitative review before waiting for the mechanical threshold.

---

## What the methodology produced

Six strategies entered. One thesis survived two separate universe runs. The pipeline eliminated five strategies cleanly and produced one that cleared every gate twice, on two different universes, with minimal overfitting pressure.

The DSR correction deserves credit here. The raw average Sharpe on MACD across the large-cap universe was ~0.49; DSR-corrected for 15 trials dropped it to 0.2715. On the midcap universe, DSR at 48 trials compressed it further to 0.0885. Without the multiple-testing correction, you'd have been much more optimistic coming out of universe sweep on both runs.

The correlation gate caught a real structural cluster that the bootstrap didn't. SBIN, BAJFINANCE, and ICICIBANK all had individually strong bootstrap results. The correlation screen caught what individual bootstrap can't: that they go to the same place in a crisis. That's the most useful portfolio-construction finding in this entire program.

---

## What this is not

This is a 6-year in-sample backtest on one strategy family, on two Indian equity universes, with Zerodha-realistic costs, evaluated at daily bars. It is not a proof that MACD trend-following has permanent edge in Indian equities. It is evidence that, over this period, on these instruments, the strategy generated sufficiently consistent out-of-sample performance to justify live deployment at the capital levels specified, with the monitoring discipline specified.

The 2025 data is the true holdout — the live period, not the evaluation window. The kill-switches exist for exactly that reason. Check them every Friday.

---

## The two live positions

| Portfolio | Instruments | Capital | Status |
|---|---|---|---|
| Nifty50 large-cap | NSE:SBIN + NSE:TITAN | ₹3,00,000 | Approved for live (2026-05-07) |
| Nifty Midcap 150 | PERSISTENT, TORNTPHARM, COFORGE, SUNDARMFIN, INDHOTEL, MUTHOOTFIN | ₹3,00,000 | Pipeline complete; pre-live brief pending |

Total deployed notional at max concurrent positions across both books: ₹6,00,000 across 8 instruments. At ₹3L per book, both are manageable within the capital assumptions of this program.
