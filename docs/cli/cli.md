# CLI Entrypoints

All binaries live under `cmd/`. They are wired together from the same underlying
packages — no logic lives in `cmd/` itself beyond flag parsing and wiring.

---

## cmd/backtest — Single Instrument Backtest

The main entrypoint. Runs one strategy on one instrument and prints full analytics.

### Usage

```bash
go run ./cmd/backtest \
    --instrument "NSE:RELIANCE" \
    --from 2018-01-01 \
    --to   2025-01-01 \
    --timeframe daily \
    --cash 100000 \
    --strategy sma-crossover \
    --fast-period 10 \
    --slow-period 50 \
    --out runs/sma-result.json \
    --output-curve runs/sma-curve.csv
```

### All flags

| Flag | Default | Description |
|---|---|---|
| `--instrument` | `NSE:NIFTY 50` | Instrument identifier (e.g. `NSE:RELIANCE`, `NSE:INFY`) |
| `--from` | required | Start date `YYYY-MM-DD` (inclusive) |
| `--to` | required | End date `YYYY-MM-DD` (exclusive) |
| `--timeframe` | `daily` | `1min`, `5min`, `15min`, `daily`, `weekly` |
| `--cash` | `100000` | Starting capital in ₹ |
| `--strategy` | `stub` | `stub`, `sma-crossover`, `rsi-mean-reversion` |
| `--fast-period` | `10` | SMA crossover: fast SMA window |
| `--slow-period` | `50` | SMA crossover: slow SMA window |
| `--rsi-period` | `14` | RSI mean reversion: RSI window |
| `--oversold` | `30` | RSI mean reversion: buy when RSI below this |
| `--overbought` | `70` | RSI mean reversion: sell when RSI above this |
| `--sizing-model` | `fixed` | `fixed` or `vol-target` |
| `--vol-target` | `0.10` | Annualized vol target when `--sizing-model=vol-target` |
| `--position-size` | `0.10` | Fraction of cash per trade when `--sizing-model=fixed` |
| `--out` | (none) | Path to write JSON results (omit to skip) |
| `--output-curve` | (none) | Path to write equity curve CSV (enables regime splits) |
| `--proliferation-gate-threshold` | `0` | Sharpe threshold for PASS/FAIL gate (0 = disabled) |
| `--bootstrap` | `false` | Run Monte Carlo bootstrap |
| `--bootstrap-seed` | `42` | RNG seed for bootstrap |
| `--bootstrap-n` | `0` | Simulation count (0 = default 10,000) |

### What it outputs

```
Running strategy "sma-crossover" on NSE:RELIANCE  2018-01-01 → 2025-01-01

═══════════════════════════════════════════════════════════════
  Performance Report
═══════════════════════════════════════════════════════════════
  Total P&L              ₹  42,318.50
  Trade Count                      48
  Win Rate                      58.3%
  Profit Factor                  1.82
  Avg Win / Avg Loss        ₹1840 / ₹980
  Sharpe Ratio                   0.72
  Sortino Ratio                  1.04
  Calmar Ratio                   0.31
  Tail Ratio                     1.12
  Max Drawdown                  18.4%
  Max DD Duration             412 days

  Benchmark (Buy & Hold)
  Total Return                  92.3%
  Annualized Return             9.8%
  Sharpe Ratio                  0.61
  Max Drawdown                 24.1%

  Regime Splits (when --output-curve is set)
  Pre-COVID (2018-2019)     Sharpe: 0.41  MaxDD: 12.1%
  COVID + Recovery (2020-21) Sharpe: 1.22  MaxDD: 31.2%
  Grind (2022-2024)         Sharpe: 0.28  MaxDD: 11.4%

  Bootstrap (when --bootstrap is set)
  Sharpe p5/p50/p95:  0.31 / 0.68 / 1.12
  Drawdown p5/p50/p95: 8.2% / 14.1% / 24.8%
  P(Sharpe > 0):      94.2%
═══════════════════════════════════════════════════════════════
```

---

## cmd/sweep — 1D Parameter Sweep

Sweeps one parameter of a strategy and ranks results by Sharpe.

### Usage

```bash
go run ./cmd/sweep \
    --instrument "NSE:NIFTY 50" \
    --from 2018-01-01 \
    --to   2025-01-01 \
    --timeframe daily \
    --strategy rsi-mean-reversion \
    --sweep-param rsi-period \
    --min 7 --max 28 --step 1
```

### All flags

| Flag | Default | Description |
|---|---|---|
| `--instrument` | `NSE:NIFTY 50` | Instrument to sweep |
| `--from` | required | Start date |
| `--to` | required | End date |
| `--timeframe` | `daily` | Bar frequency |
| `--cash` | `100000` | Starting capital |
| `--strategy` | required | `sma-crossover` or `rsi-mean-reversion` |
| `--sweep-param` | required | Parameter to vary (see table below) |
| `--min` | required | Sweep range minimum |
| `--max` | required | Sweep range maximum |
| `--step` | required | Step size (must be > 0) |
| `--fast-period` | `10` | Fixed fast period (when sweeping slow-period) |
| `--slow-period` | `50` | Fixed slow period (when sweeping fast-period) |
| `--rsi-period` | `14` | Fixed RSI period (when sweeping oversold) |
| `--oversold` | `30` | Fixed oversold (when sweeping rsi-period) |
| `--overbought` | `70` | Fixed overbought (when sweeping rsi-period) |

### Supported sweep-param combinations

| Strategy | sweep-param | Notes |
|---|---|---|
| `sma-crossover` | `fast-period` | `--slow-period` is fixed |
| `sma-crossover` | `slow-period` | `--fast-period` is fixed |
| `rsi-mean-reversion` | `rsi-period` | `--oversold`, `--overbought` are fixed |
| `rsi-mean-reversion` | `oversold` | overbought = 100 − oversold (symmetric); `--rsi-period` fixed |

---

## cmd/universe-sweep — Cross-Instrument Sweep

Runs a fixed strategy across all instruments in a universe YAML file.

### Usage

```bash
go run ./cmd/universe-sweep \
    --universe universes/nifty50-large-cap.yaml \
    --strategy sma-crossover \
    --from 2020-01-01 \
    --to   2024-12-31 \
    --timeframe daily \
    --cash 100000
```

### All flags

| Flag | Default | Description |
|---|---|---|
| `--universe` | required | Path to YAML universe file |
| `--strategy` | required | `sma-crossover` or `rsi-mean-reversion` |
| `--from` | required | Start date |
| `--to` | required | End date |
| `--timeframe` | `daily` | Bar frequency |
| `--cash` | `100000` | Starting capital |
| `--position-size` | `0.10` | Fraction of cash per trade |
| `--slippage` | `0.0005` | Slippage fraction |
| `--fast-period` | `10` | SMA fast period |
| `--slow-period` | `50` | SMA slow period |
| `--rsi-period` | `14` | RSI period |
| `--oversold` | `30` | RSI oversold threshold |
| `--overbought` | `70` | RSI overbought threshold |

### Output

CSV written to stdout:
```
instrument,sharpe,trade_count,total_pnl,max_drawdown,insufficient_data
NSE:TCS,0.842341,48,128450.23,12.4521,false
NSE:INFY,0.731204,42,98320.10,14.2100,false
NSE:RELIANCE,0.000000,8,4200.00,22.1000,true
```

Sorted descending by Sharpe. `insufficient_data=true` when trade count < 30
or bar count < 252.

---

## cmd/correlate — Strategy Correlation Analysis

Loads two previously-exported equity curve CSVs and computes pairwise
Pearson correlation over the full period and two NSE stress windows.

### Usage

```bash
go run ./cmd/correlate \
    --curve-a runs/sma-crossover-curve.csv:"SMA Crossover" \
    --curve-b runs/rsi-mean-rev-curve.csv:"RSI Mean Rev"
```

### Output

```
SMA Crossover vs RSI Mean Rev
  Full period:          r = 0.23
  COVID crash (2020):   r = 0.41
  Correction (2022):    r = 0.18
  TooCorrelated:        false
```

Equity curves must be generated with `--output-curve` from `cmd/backtest`.

---

## cmd/rsi-diagnostic — Signal Frequency Debugger

Runs an RSI strategy and counts how many bars fire each signal type.
Use before committing to a backtest to verify the strategy will actually trade.

### Usage

```bash
go run ./cmd/rsi-diagnostic \
    --instrument "NSE:RELIANCE" \
    --from 2018-01-01 \
    --to   2025-01-01 \
    --rsi-period 14 \
    --oversold 30 \
    --overbought 70
```

### Output

```
Signal frequency for RSI(14) oversold=30 overbought=70
  Total bars:   1764
  Buy signals:    18  (1.0%)
  Sell signals:   16  (0.9%)
  Hold:         1730  (98.1%)
```

A very low signal rate (< 10 trades over a multi-year period) means you won't
have enough trades for reliable statistics.

---

## cmd/authtest — Credential Verification

Loads credentials, triggers the OAuth login flow if needed, and verifies
the token is working by fetching instrument metadata.

```bash
go run ./cmd/authtest
```

Run this first when setting up a new machine or after changing credentials.

---

## cmd/providertest — End-to-End Provider Smoke Test

Fetches a short candle series for a known instrument and prints the first
few candles. Verifies the full provider pipeline (auth → API call → cache write → parse).

```bash
go run ./cmd/providertest
```

---

## Common Setup

All `cmd/*` binaries:
1. Load `.env` from the working directory (if present)
2. Read `KITE_API_KEY` and `KITE_API_SECRET` from environment
3. Build the Zerodha provider (handles auth automatically)
4. Parse remaining flags specific to that binary

The `.env` file format:
```
KITE_API_KEY=your_key_here
KITE_API_SECRET=your_secret_here
```
