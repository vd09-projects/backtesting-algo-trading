# Kill-switch parameters — SMA Crossover (NSE:RELIANCE, 2018–2024)

| Field    | Value        |
|----------|--------------|
| Date     | 2026-04-21   |
| Status   | accepted     |
| Category | algorithm    |
| Tags     | kill-switch, sma-crossover, NSE:RELIANCE, TASK-0026 |

## Status note

**SMA Crossover failed the proliferation gate (Sharpe 0.447 < 0.5 threshold). This strategy will not be deployed live.** These parameters are recorded for completeness and as a reference for the kill-switch methodology. They must not be used as approval for live deployment.

## In-sample results (runs/sma-crossover-2018-2024.json)

| Metric | Value |
|--------|-------|
| MaxDrawdown | 16.38% |
| MaxDrawdownDuration | 98,236,800,000,000,000 ns ≈ 1,137 days ≈ 3.11 years |
| TradeCount | 22 |
| Annualized Sharpe | 0.447 |
| TradeMetricsInsufficient | true (22 < 30) |

## Kill-switch thresholds

| Threshold | Formula | Value |
|-----------|---------|-------|
| Max drawdown | 1.5 × 16.38% | **24.57%** |
| Max drawdown duration | 2 × 1,137 days | **2,274 days ≈ 6.22 years** |
| Per-trade Sharpe p5 | bootstrap p5 | **PENDING — see below** |

## Bootstrap p5 Sharpe

The per-trade Sharpe kill-switch threshold requires a Monte Carlo bootstrap run. The existing run output (`runs/sma-crossover-2018-2024.json`) predates bootstrap support (TASK-0024). To compute this value:

```bash
go run ./cmd/backtest \
  --instrument "NSE:RELIANCE" \
  --from 2018-01-01 \
  --to 2025-01-01 \
  --timeframe daily \
  --strategy sma-crossover \
  --fast-period 10 \
  --slow-period 50 \
  --bootstrap \
  --bootstrap-seed 42 \
  --bootstrap-n 10000
```

**Note:** with only 22 trades, the bootstrap distribution will be wide and the p5 Sharpe will be unreliable. The Central Limit Theorem requires ~30 observations minimum. This is a further reason not to deploy this strategy — the kill-switch threshold itself cannot be established with confidence.

## How to apply thresholds

```go
thresholds := analytics.DeriveKillSwitchThresholds(
    bootstrapResult.SharpeP5,   // from montecarlo.Bootstrap output
    insampleReport,              // from analytics.Compute on the 2018-2024 run
)
// thresholds.MaxDrawdownPct ≈ 24.57
// thresholds.MaxDDDuration  ≈ 2274 * 24 * time.Hour

alert := analytics.CheckKillSwitch(windowTrades, liveCurve, thresholds)
if alert.SharpeBreached || alert.DrawdownBreached || alert.DurationBreached {
    // halt and re-evaluate from scratch
}
```
