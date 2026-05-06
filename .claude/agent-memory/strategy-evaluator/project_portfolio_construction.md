---
name: Portfolio construction state — MACD on SBIN + TITAN
description: Current portfolio construction state: MACD crossover on SBIN+TITAN expected portfolio, sizing rule, kill-switch thresholds, pending gates
type: project
---

As of 2026-05-06, Marcus gave a GO verdict on portfolio construction for the 4 MACD crossover bootstrap survivors (SBIN, BAJFINANCE, TITAN, ICICIBANK).

**Expected final portfolio:** NSE:SBIN + NSE:TITAN (subject to TASK-0085 actual correlation computation).

**Why SBIN over ICICIBANK/BAJFINANCE:** Banking cluster (SBIN, ICICIBANK, BAJFINANCE) structurally correlated — rate-sensitive, Nifty Bank constituents. SBIN has highest DSR-corrected Sharpe (0.7042). TITAN is jewelry/consumer discretionary — structurally uncorrelated with banking names.

**How to apply:** If a new strategy is being evaluated and it survives to portfolio stage, check correlation against MACD-SBIN and MACD-TITAN first. Trend-following on Nifty Bank constituents will almost certainly fail the correlation gate against MACD-SBIN.

**Capital allocation (pre-regime gate adjustment):**
- NSE:SBIN: ₹1,50,000 notional
- NSE:TITAN: ₹1,50,000 notional
- SizingVolatilityTarget: fraction = 0.10/(instrumentVol × sqrt(252)), capped at 1.0, no leverage

**Kill-switch thresholds (pre-committed 2026-05-06):**
- SBIN: Sharpe p5=0.0719, MaxDD=4.10%, MaxDDDuration=448 days
- TITAN: Sharpe p5=0.0854, MaxDD=4.72%, MaxDDDuration=1,388 days
- Source: `decisions/algorithm/2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md`

**Pending before portfolio is final:**
- TASK-0085: Run pairwise Pearson r on all 6 MACD survivor pairs — validate banking cluster assumption with actual numbers
- TASK-0086: Regime gate computation — per-regime Sharpe contributions for surviving instruments
- TASK-0087: Final portfolio composition decision file (blocked by TASK-0085 and TASK-0086)

**BAJFINANCE status:** Borderline (estimated full-period r vs SBIN: 0.55–0.72). If TASK-0085 shows SBIN/BAJFINANCE r < 0.7 full-period AND r < 0.6 both stress periods, BAJFINANCE enters the portfolio as a third instrument at ₹1,00,000 (and SBIN/TITAN rebase to ₹1,00,000 each).
