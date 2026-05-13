---
name: cmd/param-search pattern and DSR aggregation convention
description: internal/paramsearch built 2026-05-13; DSR aggregation convention established; follow-up tasks for test fix and doc contract
type: project
---

cmd/param-search done 2026-05-13 (TASK-0077). New `internal/paramsearch` package handles N-dimensional grid search. Key conventions established:

**DSR aggregation:** Per-variant DSR = mean(analytics.DSR(instrumentSharpe, gridSize, instrumentTradeCount)) across sufficient instruments. Matches universesweep.ApplyUniverseGate. nTrials = total Cartesian product (gridSize), not nSufficientVariants.

**InsufficientData filtering:** Mirrors universesweep.runInstrument — TradeMetricsInsufficient || CurveMetricsInsufficient instruments excluded from per-variant aggregate. All-insufficient variant flagged InsufficientData=true, sorted last, present in CSV.

**CSV output:** params column as JSON object (single column, stable headers across any grid dimensionality). Other columns: dsr_sharpe, raw_sharpe, trade_count, insufficient_data.

**No OOS flags:** --oos-from and --oos-to deliberately absent. Architectural enforcement of Marcus standing order. TestRun_NoOOSFlag verifies this.

**Follow-up tasks:**
- TASK-0110: done 2026-05-13 — replaced inline time.Parse with cmdutil.ParseDateRange
- TASK-0111: done 2026-05-13 — TestDSRRanking_SortedDescending now uses tradingStub; fired=true guard added
- TASK-0112: StrategyFactory godoc: add read-only params contract (still open)

**Why:** BuildSearchPipeline split from RunSearchPipeline to stay under cyclop limit (16 vs 15 max). registerFlags exported for TestRun_NoOOSFlag FlagSet inspection.

**How to apply:** When adding a new cmd search/sweep tool, follow internal/paramsearch structure. DSR convention (match ApplyUniverseGate, not DSR(avgSharpe)) is the established methodology for any cross-instrument ranking tool.
