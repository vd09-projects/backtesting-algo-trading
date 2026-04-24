# Parameter Sweeps

Three sweep harnesses exist. Each answers a different question.

| Harness | Question | Package |
|---|---|---|
| 1D Sweep | "Which value of this one parameter is best, and is the best region wide or a knife-edge?" | `internal/sweep/` |
| 2D Grid Sweep | "How do these two parameters interact, and is the peak Sharpe real after correcting for the number of trials?" | `internal/sweep2d/` |
| Universe Sweep | "Does this strategy work across many instruments, or only on the one I tested it on?" | `internal/universesweep/` |

---

## 1D Parameter Sweep

**Package:** `internal/sweep/`
**CLI:** `cmd/sweep`

### What it does

Runs the same strategy with different values of one parameter over a fixed
instrument and date range. Ranks results by Sharpe ratio. Identifies the
**plateau** — the parameter range where Sharpe stays within 80% of the peak.

```
Sweep rsi-period from 7 to 28, step 1:

rsi-period=14: Sharpe=0.82  ◄── peak
rsi-period=12: Sharpe=0.79
rsi-period=16: Sharpe=0.78
rsi-period=10: Sharpe=0.74  ◄── plateau floor (80% of 0.82 = 0.656; these all qualify)
rsi-period=18: Sharpe=0.71
rsi-period=20: Sharpe=0.54  ◄── drops out of plateau
rsi-period=8:  Sharpe=0.21
rsi-period=7:  Sharpe=-0.1  ◄── negative

Plateau: MinParam=7, MaxParam=16, Count=6
```

A wide plateau (many parameter values produce similar results) is a good sign —
the strategy is not sensitive to the exact parameter choice, which suggests the
edge is real. A cliff-edge optimum (only one or two parameter values produce
acceptable Sharpe) is a red flag.

### Flow

```
sweep.Run(ctx, cfg, provider)
      │
      ├── Generate parameter steps: [Min, Min+Step, ..., Max]
      │   (integer step counting avoids floating-point drift)
      │
      ├── For each parameter value:
      │     s := cfg.StrategyFactory(value)     ← caller defines the mapping
      │     eng := engine.New(cfg.EngineConfig)
      │     eng.Run(ctx, provider, s)
      │     analytics.Compute(trades, curve, tf) → Sharpe, PnL, TradeCount, MaxDD
      │
      ├── Sort results descending by Sharpe
      │
      └── computePlateau(results) → *PlateauRange (nil if peak Sharpe ≤ 0)
```

Runs are **sequential** (one after another). The engine is stateless per run.

### Supported strategy + parameter combinations (cmd/sweep)

| Strategy | Sweep param | Fixed params |
|---|---|---|
| `sma-crossover` | `fast-period` | `--slow-period` |
| `sma-crossover` | `slow-period` | `--fast-period` |
| `rsi-mean-reversion` | `rsi-period` | `--oversold`, `--overbought` |
| `rsi-mean-reversion` | `oversold` | `--rsi-period`; overbought = 100 − oversold (symmetric) |

### Config

```go
sweep.Config{
    ParameterName:   "rsi-period",
    Min:             7,
    Max:             28,
    Step:            1,
    Timeframe:       model.TimeframeDaily,
    EngineConfig:    engine.Config{...},    // fixed across all runs
    StrategyFactory: func(v float64) (strategy.Strategy, error) {...},
}
```

---

## 2D Grid Sweep

**Package:** `internal/sweep2d/`

### What it does

Sweeps two parameters simultaneously as a grid. All `param1 × param2` combinations
are tested. Returns the full grid plus a **DSR-corrected peak Sharpe**.

```
Sweep fast-period [5..20] × slow-period [20..100]:

         slow=20  slow=40  slow=60  slow=80  slow=100
fast=5   -0.12    0.31     0.42     0.38     0.21
fast=10   0.18    0.55     0.71     0.68     0.44     ← peak at fast=10, slow=60 (0.71)
fast=15   0.09    0.41     0.58     0.52     0.33
fast=20  -0.05    0.22     0.38     0.31     0.18

VariantCount = 5 × 5 = 25
PeakSharpe = 0.71
DSRCorrectedPeakSharpe = DSR(0.71, 25, 1764)   ← penalized for 25 trials tested
```

The DSR correction is automatic — you don't have to remember to apply it.
A negative DSR-corrected Sharpe means the peak Sharpe you found is fully
explained by random search over the grid.

### Flow

```
sweep2d.Run(ctx, cfg, provider)
      │
      ├── Generate p1Steps and p2Steps
      │
      ├── Pre-allocate grid[len(p1)][len(p2)]
      │
      ├── errgroup with GOMAXPROCS ceiling:
      │   For each (i, j):
      │     g.Go(func() {
      │       s := cfg.StrategyFactory(p1Steps[i], p2Steps[j])
      │       eng.Run(ctx, provider, s)
      │       grid[i][j] = GridCell{Sharpe, TradeCount, MaxDrawdown}
      │     })
      │
      ├── g.Wait()  ← all cells complete
      │
      ├── peakSharpe = max(all grid cells)
      └── DSRCorrectedPeakSharpe = analytics.DSR(peak, variantCount, nObs)
```

Runs in **parallel** (one goroutine per grid cell, bounded by GOMAXPROCS).
Results are written at fixed `[i][j]` indices — no mutex needed, deterministic output.

### Output

Results are exported as CSV via `sweep2d.WriteCSV()`:
```
param1_value,param2_value,sharpe,trade_count,max_drawdown
5,20,-0.12,8,12.4
5,40,0.31,14,8.2
...
```

---

## Universe Sweep

**Package:** `internal/universesweep/`
**CLI:** `cmd/universe-sweep`

### What it does

Runs a **fixed strategy** (same parameters) across a list of instruments.
Identifies which instruments the strategy works on and which it doesn't.

This is the cross-instrument robustness check. If a strategy only works on
one or two instruments out of 15, the backtest result may be instrument-specific
noise rather than a transferable edge.

### Flow

```
ParseUniverseFile("universes/nifty50-large-cap.yaml")
      │
      └── []string instruments  (deduplicated, order preserved)

universesweep.Run(ctx, cfg, provider)
      │
      ├── Pre-allocate results[len(instruments)]
      │
      ├── errgroup with GOMAXPROCS ceiling:
      │   For each i, instrument:
      │     g.Go(func() {
      │       engCfg := cfg.EngineConfig
      │       engCfg.Instrument = instrument   ← stamp instrument per run
      │       eng.Run(ctx, provider, cfg.Strategy)
      │       report = analytics.Compute(trades, curve, cfg.Timeframe)
      │       results[i] = Result{
      │           Instrument:       instrument,
      │           Sharpe:           report.SharpeRatio,
      │           TradeCount:       report.TradeCount,
      │           TotalPnL:         report.TotalPnL,
      │           MaxDrawdown:      report.MaxDrawdown,
      │           InsufficientData: report.TradeMetricsInsufficient ||
      │                             report.CurveMetricsInsufficient,
      │       }
      │     })
      │
      ├── g.Wait()
      ├── sort results descending by Sharpe
      └── WriteCSV(stdout, report)
```

### Universe file format (YAML)

```yaml
# universes/nifty50-large-cap.yaml
instruments:
  - NSE:RELIANCE
  - NSE:TCS
  - NSE:HDFCBANK
  - NSE:INFY
  - NSE:HINDUNILVR
```

Duplicates are silently removed. An empty `instruments:` list is an error.

### Output (CSV to stdout)

```
instrument,sharpe,trade_count,total_pnl,max_drawdown,insufficient_data
NSE:TCS,0.842341,48,128450.23,12.4521,false
NSE:INFY,0.731204,42,98320.10,14.2100,false
NSE:RELIANCE,0.210043,12,12400.50,22.1000,true
```

`insufficient_data=true` rows have Sharpe zeroed (not hidden — still present so
you can see the instrument failed the data gate).

### Concurrency

Runs in **parallel** (one goroutine per instrument, bounded by GOMAXPROCS).
Pre-allocated slice at fixed indices — deterministic ordering before the final
Sharpe sort regardless of goroutine scheduling.
