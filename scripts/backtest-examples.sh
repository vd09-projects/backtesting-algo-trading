#!/usr/bin/env bash
#
# backtest-examples.sh — Reference commands for running backtests.
#
# Prerequisites:
#   - .env file in repo root with KITE_API_KEY and KITE_API_SECRET
#   - Valid Kite Connect session (first run triggers login; token cached at
#     ~/.config/backtest/token.json and reused until 6 AM IST next day)
#   - Candles cached under .cache/zerodha/ after first fetch (no re-download)
#
# Holdout rule: 2015-2022 is training data. 2023-present is holdout.
# Do NOT run on holdout data until a strategy is being considered for live capital.
#
# Usage: copy and paste individual commands — do not run this file as a script.
# Update this file each time a new strategy is wired into cmd/backtest.
#
# ─────────────────────────────────────────────────────────────────────────────

# ── Smoke test (stub strategy — no signals, confirms pipeline wiring) ─────────

go run ./cmd/backtest \
  --instrument "NSE:NIFTY 50" \
  --from 2020-01-01 \
  --to   2020-12-31 \
  --strategy stub \
  --cash 1000000

# ── SMA Crossover ─────────────────────────────────────────────────────────────
# Trend-following. Buy on fast SMA crossing above slow SMA, sell on cross below.
# Defaults: fast=10, slow=50.

go run ./cmd/backtest \
  --instrument "NSE:NIFTY 50" \
  --from 2015-01-01 \
  --to   2022-12-31 \
  --strategy sma-crossover \
  --fast-period 10 \
  --slow-period 50 \
  --cash 1000000 \
  --out results/sma-crossover-nifty50-2015-2022.json

# ── RSI Mean-Reversion ────────────────────────────────────────────────────────
# Mean-reversion. Buy when RSI < oversold (30), sell when RSI > overbought (70).
# Defaults: period=14, oversold=30, overbought=70.
# Note: long-only, no stop-loss. Watch for dead-weight positions in 2020 crash
# and 2022 bear — RSI may stay below 30 for weeks without recovering above 70.

go run ./cmd/backtest \
  --instrument "NSE:NIFTY 50" \
  --from 2015-01-01 \
  --to   2022-12-31 \
  --strategy rsi-mean-reversion \
  --rsi-period  14 \
  --oversold    30 \
  --overbought  70 \
  --cash 1000000 \
  --out results/rsi-meanrev-nifty50-2015-2022.json

# ── Side-by-side comparison run ───────────────────────────────────────────────
# Run both strategies back-to-back on the same instrument and period.
# Compare Sharpe, max drawdown, trade count in the output JSONs.
# Neither result is meaningful without the buy-and-hold benchmark (TASK-0018).

go run ./cmd/backtest \
  --instrument "NSE:NIFTY 50" \
  --from 2015-01-01 --to 2022-12-31 \
  --strategy sma-crossover \
  --cash 1000000 \
  --out results/sma-crossover-nifty50-2015-2022.json

go run ./cmd/backtest \
  --instrument "NSE:NIFTY 50" \
  --from 2015-01-01 --to 2022-12-31 \
  --strategy rsi-mean-reversion \
  --cash 1000000 \
  --out results/rsi-meanrev-nifty50-2015-2022.json

# ── Alternative instruments ───────────────────────────────────────────────────
# NIFTY 50 index. INFY, RELIANCE, HDFCBANK for single-stock runs.

go run ./cmd/backtest \
  --instrument "NSE:INFY" \
  --from 2015-01-01 \
  --to   2022-12-31 \
  --strategy rsi-mean-reversion \
  --cash 1000000 \
  --out results/rsi-meanrev-infy-2015-2022.json

# ── Provider smoke test (confirms Kite Connect data fetch + cache) ────────────

go run ./cmd/providertest

# ── Parameter sensitivity (once TASK-0023 is implemented) ────────────────────
# Uncomment and update when --sweep flag is available.
#
# go run ./cmd/backtest \
#   --instrument "NSE:NIFTY 50" \
#   --from 2015-01-01 --to 2022-12-31 \
#   --strategy rsi-mean-reversion \
#   --sweep rsi-period=10:20:1 \
#   --cash 1000000

# ─────────────────────────────────────────────────────────────────────────────
# Results are written to results/ (gitignored). JSON fields:
#   TotalPnL, WinRate, MaxDrawdown, TradeCount, SharpeRatio
#   (ProfitFactor, SortinoRatio, CalmarRatio, TailRatio added in TASK-0016)
# ─────────────────────────────────────────────────────────────────────────────
