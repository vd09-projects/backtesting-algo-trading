# Monte Carlo Bootstrap

**Package:** `internal/montecarlo/`
**File:** `montecarlo.go`

The Monte Carlo bootstrap answers the question:
"Is this strategy's Sharpe ratio real, or is it sampling luck?"

It does this by resampling the actual per-trade returns 10,000 times and
computing the distribution of Sharpe ratios across those resamples. If the
p5 Sharpe (the worst 5% of simulations) is still positive, there's more
confidence the edge is real.

---

## How Bootstrap Works

```
Input: []model.Trade  (closed trades from a backtest run)

For each trade:
  ReturnOnNotional = RealizedPnL / (EntryPrice × Quantity)

returns = [r1, r2, r3, ..., rN]    (one per trade)

For i in 1..NSimulations (default 10,000):
  resample = N draws WITH REPLACEMENT from returns
  sharpes[i]   = sampleSharpe(resample)
  drawdowns[i] = worstDrawdown(resample)

sort(sharpes)
sort(drawdowns)

Output: BootstrapResult{
  MeanSharpe         = mean(sharpes)
  SharpeP5           = sharpes[floor(0.05 × N)]
  SharpeP50          = sharpes[floor(0.50 × N)]
  SharpeP95          = sharpes[floor(0.95 × N)]
  WorstDrawdownP5    = drawdowns[floor(0.05 × N)]
  WorstDrawdownP50   = drawdowns[floor(0.50 × N)]
  WorstDrawdownP95   = drawdowns[floor(0.95 × N)]
  ProbPositiveSharpe = count(sharpes > 0) / N
}
```

---

## Key Formulas

### Per-trade Sharpe (non-annualized)

```
mean = Σ(returns) / n
std  = √(Σ(r - mean)² / (n-1))    [sample variance, n-1 denominator]
Sharpe = mean / std
```

No annualization factor — this is a per-trade Sharpe, not a per-bar Sharpe.
The kill-switch monitoring uses the **same formula** so the threshold and the
live metric are comparable on identical scales.

### Worst Drawdown (geometric compounding)

```
equity = 1.0
peak   = 1.0
for r in returns:
  equity = equity × (1 + r)
  equity = max(equity, 0)         [floor at 0: losses can't exceed notional]
  peak   = max(peak, equity)
  dd = (peak - equity) / peak × 100
worstDrawdown = max(dd)
```

Geometric compounding is used rather than arithmetic accumulation because
it correctly models compound returns. For small per-trade returns the difference
is negligible, but the convention is established correctly for future strategies
with larger individual trade returns.

---

## Configuration

```go
montecarlo.Bootstrap(trades []model.Trade, cfg BootstrapConfig) BootstrapResult

type BootstrapConfig struct {
    NSimulations int    // default 10,000 when 0
    Seed         int64  // RNG seed — MUST be logged with results for reproducibility
}
```

The seed is critical. Two runs with the same seed produce bit-identical results.
Different seeds produce different samples. Always log the seed alongside results.

RNG: `math/rand/v2` with `rand.NewPCG(uint64(seed), 0)` — PCG64, high statistical
quality, fast, part of Go stdlib since 1.22.

---

## How Bootstrap connects to the Kill Switch

```
Backtest run
     │
     ▼
montecarlo.Bootstrap(trades, cfg)
     │
     └── BootstrapResult.SharpeP5  ──────────────────────────┐
                                                              │
analytics.DeriveKillSwitchThresholds(sharpeP5, inSampleReport)│
     │                                                        │
     │  thresholds.SharpeP5       = SharpeP5 ────────────────┘
     │  thresholds.MaxDrawdownPct = inSample.MaxDrawdown × 1.5
     │  thresholds.MaxDDDuration  = inSample.MaxDDDuration × 2
     │
     └── store thresholds (before going live)


Live monitoring (periodic):
  alert := analytics.CheckKillSwitch(recentTrades, liveCurve, thresholds)

  if alert.SharpeBreached    → halt: rolling per-trade Sharpe below p5 baseline
  if alert.DrawdownBreached  → halt: drawdown beyond 1.5× worst in-sample
  if alert.DurationBreached  → halt: stuck in drawdown beyond 2× worst in-sample recovery
```

---

## Bootstrap Gate

A strategy passes the bootstrap gate if **both** conditions hold:

```
SharpeP5 > 0                     (5th-percentile bootstrap Sharpe is positive)
ProbPositiveSharpe > 0.80        (>80% of 10,000 simulations produce positive Sharpe)
```

Failing either condition kills the strategy — the kill is recorded in `decisions/algorithm/`.
Both conditions are required. A strategy with `SharpeP5 = 0.01` but `ProbPositiveSharpe = 0.72`
fails: the distribution is barely above zero and not concentrated enough above the line.

The `SharpeP5` value from a passing bootstrap run feeds directly into
`analytics.DeriveKillSwitchThresholds` — it becomes the live monitoring floor.

---

## Interpreting Results

| Metric | Good signal | Caution |
|---|---|---|
| `SharpeP5 > 0` | Strategy has positive edge even in the worst 5% of simulations | — |
| `SharpeP5 < 0` | 5% of draw sequences produce a losing strategy | Kill-switch threshold is negative — not useful |
| `ProbPositiveSharpe > 0.80` | >80% of simulations are profitable | Below 0.70 is concerning |
| `WorstDrawdownP95` | 95th-percentile drawdown — expect to see this over a long run | Compare against your risk tolerance |
| `MeanSharpe` | Central estimate of the strategy's per-trade Sharpe | Not the same as the annualized Sharpe in the main Report |

---

## Why Not Annualized Sharpe?

The bootstrap operates on **per-trade returns**, not per-bar returns. The number
of trades per year varies (strategy, regime, instrument). Annualizing per-trade
Sharpe would require knowing the average hold time, which is variable. Using
non-annualized per-trade Sharpe avoids this and makes the kill-switch comparison
exact: the threshold was computed from per-trade returns; the live check uses
per-trade returns from the same window size.
