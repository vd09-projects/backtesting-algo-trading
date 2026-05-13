# cmd/param-search

Parameter grid search with DSR correction. Finds optimal parameters for a strategy by running an N-dimensional grid search on a training window only, ranking results by Deflated Sharpe Ratio (DSR).

## Usage

```bash
go run ./cmd/param-search \
    --strategy      macd-crossover \
    --param-grid    param-grids/macd-grid.json \
    --universe      universes/nifty50-large-cap.yaml \
    --train-from    2018-01-01 \
    --train-to      2022-12-31 \
    --out-dir       results/param-search/ \
    --top-n         10
```

## Flags

| Flag | Required | Default | Description |
|---|---|---|---|
| `--strategy` | yes | — | Strategy name (see `go run ./cmd/param-search --help`) |
| `--param-grid` | yes | — | Path to param-grid JSON file (schema below) |
| `--universe` | yes | — | Path to universe YAML file |
| `--train-from` | yes | — | Training window start, YYYY-MM-DD inclusive |
| `--train-to` | yes | — | Training window end, YYYY-MM-DD exclusive |
| `--out-dir` | yes | — | Output directory (must exist) |
| `--top-n` | no | 10 | Max variants to write to CSV (0 = all) |
| `--timeframe` | no | daily | Candle timeframe: 1min, 5min, 15min, daily, weekly |
| `--commission` | no | zerodha | Commission model |
| `--cash` | no | 100000 | Starting cash per engine run |
| `--position-size` | no | 0.10 | Fraction of cash deployed per trade |
| `--slippage` | no | 0.0005 | Slippage as decimal fraction |

**There are no `--oos-from` or `--oos-to` flags.** This is architectural enforcement: parameter search must run on the training window only. OOS evaluation is the caller's responsibility via `cmd/evaluate`.

## Param-grid JSON schema

```json
{
  "axes": [
    {
      "name": "fast-period",
      "min": 5,
      "max": 30,
      "step": 5
    },
    {
      "name": "slow-period",
      "min": 20,
      "max": 60,
      "step": 10
    }
  ]
}
```

### Schema fields

| Field | Type | Description |
|---|---|---|
| `axes` | array | List of parameter axes (required, min 1) |
| `axes[].name` | string | Parameter name as registered in `internal/cmdutil/strategies.go` |
| `axes[].min` | float | Axis minimum value (inclusive) |
| `axes[].max` | float | Axis maximum value (inclusive) |
| `axes[].step` | float | Step size (must be > 0) |

The grid is the Cartesian product of all axes. A 2-axis grid with 6 fast-period values and 5 slow-period values = 30 variants total. The total variant count (GridSize) is the `nTrials` used in DSR correction.

**Axis names** must match strategy parameter names as registered in `internal/cmdutil/strategies.go`. Unrecognized names are silently ignored by strategy constructors — a typo will silently use the strategy default for that parameter.

### Common parameter names by strategy

| Strategy | Parameter names |
|---|---|
| `macd-crossover` | `fast-period`, `slow-period`, `macd-signal-period` |
| `sma-crossover` | `fast-period`, `slow-period` |
| `rsi-mean-reversion` | `rsi-period`, `oversold`, `overbought` |
| `bollinger-mean-reversion` | `bb-period`, `bb-num-std-dev` |
| `donchian-breakout` | `donchian-period` |
| `cci-mean-reversion` | `cci-period`, `cci-entry-threshold`, `cci-exit-threshold` |
| `momentum` | `lookback`, `threshold` |

## Output

`--out-dir/param-search-results.csv`:

```
params,dsr_sharpe,raw_sharpe,trade_count,insufficient_data
{"fast-period":17,"slow-period":26},0.089234,0.182410,847,false
{"fast-period":12,"slow-period":26},0.071892,0.154320,923,false
...
```

| Column | Description |
|---|---|
| `params` | JSON object of parameter values for this variant |
| `dsr_sharpe` | Mean DSR-corrected Sharpe across sufficient instruments |
| `raw_sharpe` | Mean raw Sharpe across sufficient instruments |
| `trade_count` | Total trades across all sufficient instruments |
| `insufficient_data` | `true` when all instruments had insufficient data for this variant |

Rows are sorted descending by `dsr_sharpe`. Variants where `insufficient_data=true` are sorted last.

## DSR methodology

DSR = Deflated Sharpe Ratio (Bailey & López de Prado, 2014). It corrects the observed Sharpe for the expected maximum Sharpe arising from testing multiple parameter combinations.

- `nTrials` = total grid size (number of variants tested)
- Per-variant DSR = `mean(DSR(instrumentSharpe, nTrials, instrumentTradeCount))` across sufficient instruments
- Instruments with fewer than `analytics.MinTradesForMetrics=30` trades or fewer than `analytics.MinCurvePointsForMetrics=252` equity curve points are excluded from each variant's aggregate
- This methodology matches `internal/universesweep.ApplyUniverseGate` exactly

## Workflow

This tool produces optimal **in-sample** parameters. To evaluate the selected parameters out-of-sample:

1. Run `cmd/param-search` on the training window to find top parameters
2. Read `param-search-results.csv`, pick the top variant by `dsr_sharpe`
3. Run `cmd/evaluate` with the winning parameters and a separate OOS date range:

```bash
go run ./cmd/evaluate \
    --strategy   macd-crossover \
    --universe   universes/nifty50-large-cap.yaml \
    --from       2023-01-01 \
    --to         2025-01-01 \
    --params     fast-period=17 \
    --params     slow-period=26 \
    --out-dir    results/oos-eval/
```
