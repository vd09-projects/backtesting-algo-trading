# Analytics

**Package:** `internal/analytics/`
**Files:** `analytics.go`, `benchmark.go`, `regime.go`, `correlation.go`, `killswitch.go`, `dsr.go`

Analytics is a pure computation layer — it takes completed trade logs and equity
curves as inputs and returns metrics. It never modifies inputs, has no side effects,
and does not call the engine.

---

## Main Report

```go
analytics.Compute(trades []model.Trade, curve []model.EquityPoint, tf model.Timeframe) Report
```

### Report fields

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Trade-level metrics  (zeroed when TradeMetricsInsufficient = true)     │
├─────────────────────────────────────────────────────────────────────────┤
│  TotalPnL       float64   Sum of all RealizedPnL across closed trades   │
│  WinRate        float64   % of trades with RealizedPnL > 0  (0–100)    │
│  WinCount       int                                                      │
│  LossCount      int       Includes break-even trades (PnL <= 0)         │
│  TradeCount     int                                                      │
│  ProfitFactor   float64   GrossProfit / |GrossLoss|; 0 if no losses     │
│  AvgWin         float64   Average P&L of winning trades                 │
│  AvgLoss        float64   Average absolute P&L of losing trades         │
├─────────────────────────────────────────────────────────────────────────┤
│  Curve-level metrics  (zeroed when CurveMetricsInsufficient = true)     │
├─────────────────────────────────────────────────────────────────────────┤
│  SharpeRatio    float64   Annualized Sharpe from per-bar equity returns │
│  SortinoRatio   float64   Annualized Sortino (downside dev only)        │
│  CalmarRatio    float64   Annualized return / MaxDrawdown               │
│  TailRatio      float64   p95 return / |p5 return|                      │
│  MaxDrawdown    float64   Peak-to-trough %, 0–100                       │
│  MaxDrawdownDuration time.Duration  Wall time of max drawdown event     │
├─────────────────────────────────────────────────────────────────────────┤
│  Flags                                                                   │
├─────────────────────────────────────────────────────────────────────────┤
│  TradeMetricsInsufficient bool   true when TradeCount < 30              │
│  CurveMetricsInsufficient bool   true when len(curve) < 252             │
└─────────────────────────────────────────────────────────────────────────┘
```

### Why the guards exist

Below 30 trades, the Central Limit Theorem hasn't kicked in — win rate and
profit factor are unreliable. Below 252 bars (one NSE trading year), annualized
ratios are not interpretable. Rather than silently reporting misleading numbers,
the affected fields are zeroed and flagged.

---

## Sharpe Ratio

```
Annualized Sharpe = (mean per-bar return / std per-bar return) × √(annualizationFactor)

Annualization factors (NSE):
  daily:  252 bars/year
  15min:  252 × 25 = 6,300 bars/year  (NSE session = 375 min ÷ 15)
  5min:   252 × 75 = 18,900 bars/year
  1min:   252 × 375 = 94,500 bars/year
  weekly: 52 bars/year
```

Uses **sample variance** (n-1 denominator) consistent with the Monte Carlo bootstrap.
Zero variance → returns 0.

A Sharpe > 2.5 on daily bars for a non-HFT strategy is a red flag —
likely overfitting or insufficient sample size.

---

## Sortino Ratio

```
Downside deviation = √(Σ(min(r, 0)²) / n)    [population denominator over ALL bars]
Annualized Sortino = (mean per-bar return / downsideDev) × √(annFactor)
```

Target return = 0. Population denominator (not just negative bars) is the
Rollinger-Hoffman convention — most common in practice.

---

## Calmar Ratio

```
Annualized return = mean per-bar return × annFactor
Calmar = annualized return / (MaxDrawdown / 100)
```

Returns 0 when MaxDrawdown is zero (no losing period observed).

---

## Tail Ratio

```
p5  = 5th percentile of per-bar returns
p95 = 95th percentile of per-bar returns
TailRatio = p95 / |p5|
```

A value < 1 means the left tail is heavier than the right — the strategy
is implicitly short volatility (small steady gains, occasional large losses).

---

## Max Drawdown

**Depth (%):**
```
For each bar:
  if equity > peak: peak = equity
  dd = (peak - equity) / peak × 100
MaxDrawdown = max(dd) across all bars
```

**Duration:**
- Finds the bar that begins the worst drawdown event
- Scans forward to the first bar where equity recovers to that peak
- If equity never recovers, duration runs to the last bar

Both depth and duration are computed from the **same equity curve**, so they
always describe the same drawdown event.

---

## Benchmark (Buy-and-Hold)

```go
analytics.ComputeBenchmark(candles []model.Candle, initialCash float64) BenchmarkReport
```

```
Entry: candles[0].Open
Exit:  candles[last].Close
No transaction costs.

BenchmarkReport:
  TotalReturn      float64  // (exitPrice - entryPrice) / entryPrice × 100
  AnnualizedReturn float64  // CAGR: (1 + totalReturn/100)^(1/years) - 1, in %
  SharpeRatio      float64  // from per-bar close returns
  MaxDrawdown      float64  // peak-to-trough on close prices
```

CAGR uses actual elapsed calendar time (not trading days) to stay comparable
across strategies run over different date ranges.

---

## Regime Splits

```go
analytics.ComputeRegimeSplits(
    curve []model.EquityPoint,
    regimes []Regime,
    tf model.Timeframe,
) []RegimeReport
```

Pre-defined NSE regimes (2018–2024):

| Regime | Window | Market character |
|---|---|---|
| Pre-COVID | 2018-01-01 to 2020-01-31 | Mixed trending and sideways, moderate vol |
| COVID Crash + Recovery | 2020-02-01 to 2021-06-30 | Sharp crash, V-shaped recovery, high vol |
| Post-recovery | 2021-07-01 to 2024-12-31 | Grinding uptrend, 2022 rate-hike bear, lower vol |

Each `RegimeReport` contains Sharpe and MaxDrawdown for the slice of the equity
curve that falls within that period.

### Regime gate

A strategy passes the regime gate if no single regime accounts for ≥ 70% of the
total Sharpe concentration. Concentration is measured as:

```
contribution[i] = abs(S[i]) / sum(abs(S[j]) for all j)
```

where `S[i]` is the per-trade Sharpe for regime `i` (can be negative). Using
absolute values avoids sign-cancellation when regimes have mixed signs.

The gate **flags but does not kill** — a regime-concentrated strategy receives
**half-weight** in portfolio construction rather than being excluded. If a strategy
has zero trades in any regime window, it is treated as concentrated by default.

```
RegimeConcentrated = false  →  full weight eligible
RegimeConcentrated = true   →  at most 50% of computed vol-target weight
```

---

## Correlation

```go
analytics.ComputeCorrelation(a, b NamedCurve) PairCorrelation
analytics.ComputeMatrix(curves []NamedCurve) CorrelationMatrix
```

```
PairCorrelation:
  FullPeriod     float64  // Pearson r, full backtest period (2018-01-01 to 2024-12-31)
  Crash2020      float64  // Pearson r, COVID crash: 2020-02-01 to 2020-06-30
  Correction2022 float64  // Pearson r, rate-hike bear: 2022-01-01 to 2022-12-31
  TooCorrelated  bool     // true if FullPeriod >= 0.7 OR either stress period >= 0.6
```

**Series**: daily log-returns of the equity curve (`ln(equity[t] / equity[t-1])`).
Flat segments (days with no open position, equity unchanged) contribute a return
of zero — they are real trading days, not missing data.

**Warmup trimming:** Leading bars where equity equals the initial value (flat —
no position taken yet) are trimmed before computing returns. A strategy that
never trades produces a constant series → NaN correlation (not zero, which
would be misleading).

NaN is the sentinel for undefined correlation (constant series, empty window).
`TooCorrelated` is only set when the value is not NaN and exceeds the threshold.

### Correlation gate thresholds

| Window | Pass condition |
|---|---|
| Full period | Pearson r < 0.7 |
| Either stress period | Pearson r < 0.6 |

Both conditions must hold. Failing either triggers the tiebreaker: keep the
strategy with higher DSR-corrected Sharpe. If DSR-Sharpe is within 5%, prefer
the strategy from a different edge bucket (trend-following vs. mean-reversion).

---

## Kill Switch

```go
// Step 1: derive thresholds from in-sample results + bootstrap
thresholds := analytics.DeriveKillSwitchThresholds(
    bootstrapResult.SharpeP5,
    inSampleReport,
)
// thresholds.SharpeP5       = bootstrap p5 Sharpe (from montecarlo)
// thresholds.MaxDrawdownPct = 1.5 × in-sample max drawdown
// thresholds.MaxDDDuration  = 2 × in-sample max drawdown duration

// Step 2: check live metrics against thresholds
alert := analytics.CheckKillSwitch(windowTrades, liveCurve, thresholds)
```

```
KillSwitchAlert:
  SharpeBreached    bool     // rolling per-trade Sharpe < SharpeP5
  DrawdownBreached  bool     // current drawdown > MaxDrawdownPct
  DurationBreached  bool     // time in drawdown > MaxDDDuration
  RollingPerTradeSharpe float64
  CurrentDrawdownPct    float64
  CurrentDDDuration     time.Duration
```

The per-trade Sharpe in `CheckKillSwitch` uses the **identical formula** as the
Monte Carlo bootstrap — `mean(ReturnOnNotional) / std(ReturnOnNotional)`, sample
variance, no annualization. This is intentional: the kill-switch comparison must
be apples-to-apples with the threshold it was derived from.

---

## Deflated Sharpe Ratio (DSR)

```go
analytics.DSR(observedSharpe, nTrials, nObservations float64) float64
```

Corrects the observed Sharpe ratio for the **expected maximum Sharpe** that arises
purely from testing multiple independent strategies.

```
E[max SR] = (1−γ)·Φ⁻¹(1 − 1/nTrials) + γ·Φ⁻¹(1 − 1/(nTrials·e))
            where γ = Euler-Mascheroni constant ≈ 0.5772

SE = 1 / √(nObservations − 1)

DSR = observedSharpe − E[max SR] × SE
```

A positive DSR means the strategy's Sharpe exceeds what multiple testing alone
would predict. Used in the 2D sweep to automatically penalize the peak Sharpe
for having searched a large grid.
