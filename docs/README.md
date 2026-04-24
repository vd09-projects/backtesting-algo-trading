# Documentation Index

This directory contains the complete technical documentation for the backtesting-algo-trading engine.
Each subdirectory covers one major flow or subsystem.

---

## Structure

| Folder | What it covers |
|---|---|
| [architecture/](architecture/overview.md) | End-to-end data flow, package dependency rules, core invariants |
| [engine/](engine/engine.md) | Backtest event loop, portfolio, order simulation, sizing models |
| [provider/](provider/provider.md) | Data provider interface, Zerodha implementation, caching, auth |
| [model/](model/model.md) | All shared domain types (Candle, Trade, Signal, Position, etc.) |
| [strategies/](strategies/strategies.md) | Strategy interface and concrete implementations |
| [analytics/](analytics/analytics.md) | Performance metrics, benchmark, regime splits, correlation, kill switch, DSR |
| [montecarlo/](montecarlo/montecarlo.md) | Bootstrap simulation, Sharpe/drawdown distributions, kill-switch threshold derivation |
| [sweeps/](sweeps/sweeps.md) | 1D parameter sweep, 2D grid sweep, universe sweep |
| [walkforward/](walkforward/walkforward.md) | Walk-forward validation harness |
| [cli/](cli/cli.md) | All CLI entrypoints and their flags |

---

## Quick orientation

```
You want to...                          Read...
─────────────────────────────────────────────────────────────
Understand the full system at a glance  architecture/overview.md
Run your first backtest                 cli/cli.md
Add a new strategy                      strategies/strategies.md
Understand what metrics mean            analytics/analytics.md
Understand how bootstrap works          montecarlo/montecarlo.md
Tune parameters systematically         sweeps/sweeps.md
Test cross-instrument robustness       sweeps/sweeps.md (universe sweep)
Validate against overfitting           walkforward/walkforward.md
Understand how market data gets in     provider/provider.md
```
