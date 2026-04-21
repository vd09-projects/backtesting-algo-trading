# Kill-switch parameters — RSI Mean-Reversion (NSE:RELIANCE, 2018–2024)

| Field    | Value        |
|----------|--------------|
| Date     | 2026-04-21   |
| Status   | accepted     |
| Category | algorithm    |
| Tags     | kill-switch, rsi-mean-reversion, NSE:RELIANCE, TASK-0026 |

## Status note

**RSI Mean-Reversion failed the proliferation gate (Sharpe 0.469 < 0.5 threshold, only 7 trades). This strategy will not be deployed live.** These parameters are recorded for completeness and as a reference for the kill-switch methodology. They must not be used as approval for live deployment.

## In-sample results (runs/rsi-mean-rev-2018-2024.json)

| Metric | Value |
|--------|-------|
| MaxDrawdown | 17.36% |
| MaxDrawdownDuration | 32,140,800,000,000,000 ns ≈ 372 days ≈ 1.02 years |
| TradeCount | 7 |
| Annualized Sharpe | 0.469 |
| TradeMetricsInsufficient | true (7 < 30) |

## Kill-switch thresholds

| Threshold | Formula | Value |
|-----------|---------|-------|
| Max drawdown | 1.5 × 17.36% | **26.03%** |
| Max drawdown duration | 2 × 372 days | **744 days ≈ 2.04 years** |
| Per-trade Sharpe p5 | bootstrap p5 | **PENDING — see below** |

## Bootstrap p5 Sharpe

The per-trade Sharpe kill-switch threshold requires a Monte Carlo bootstrap run. The existing run output (`runs/rsi-mean-rev-2018-2024.json`) predates bootstrap support (TASK-0024). To compute this value:

```bash
go run ./cmd/backtest \
  --instrument "NSE:RELIANCE" \
  --from 2018-01-01 \
  --to 2025-01-01 \
  --timeframe daily \
  --strategy rsi-mean-reversion \
  --rsi-period 14 \
  --oversold 30 \
  --overbought 70 \
  --bootstrap \
  --bootstrap-seed 42 \
  --bootstrap-n 10000
```

**Note:** with only 7 trades, the bootstrap distribution will be extremely wide. A 7-trade bootstrap is statistically meaningless — the p5 Sharpe will have enormous confidence intervals and cannot serve as a reliable kill-switch threshold. This is a critical reason not to deploy this strategy: not only does it fail the proliferation gate, but the kill-switch threshold itself is unverifiable at this sample size.

## How to apply thresholds

```go
thresholds := analytics.DeriveKillSwitchThresholds(
    bootstrapResult.SharpeP5,   // from montecarlo.Bootstrap output
    insampleReport,              // from analytics.Compute on the 2018-2024 run
)
// thresholds.MaxDrawdownPct ≈ 26.03
// thresholds.MaxDDDuration  ≈ 744 * 24 * time.Hour

alert := analytics.CheckKillSwitch(windowTrades, liveCurve, thresholds)
if alert.SharpeBreached || alert.DrawdownBreached || alert.DurationBreached {
    // halt and re-evaluate from scratch
}
```
