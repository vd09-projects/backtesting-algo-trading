---
name: cmd/evaluate pattern and status
description: cmd/evaluate pipeline orchestrator (TASK-0073): testability pattern, verdict.json DTO, gate thresholds, follow-up tasks
type: project
---

cmd/evaluate is a pure cmd-layer pipeline orchestrator (universe sweep → WF → bootstrap). Done 2026-05-10.

Key pattern: `run(args, stdout, stderr, providerFactory func(context.Context)(provider.DataProvider,error))`. providerFactory signature differs from cmd/fetch-history's `providerFactory func(fetchFlags)` — cmd/evaluate uses context-based factory (simpler; no parsed flags needed by provider).

Gate thresholds encoded:
- Universe: DSRAvg > 0 AND PassFraction >= 40% (via universesweep.ApplyUniverseGate)
- WF: floor(0.60 × universe_gate_passes) AND >= 6 (applyWFGate standing order)
- Bootstrap: SharpeP5 > 0 AND ProbPos > 0.80 (applyBootstrapGate)

verdict.json DTO: local to cmd/evaluate (verdictJSON, gateStatsJSON, survivorRecord, killRecord). Not in internal/output.

Stage outputs: --out-dir/YYYY-MM-DD-{strategy}-{timeframe}/ dated subdirectory.

Coverage: 65% (integration-only gap: runWalkForward/runBootstrap/collectTrades require live provider producing positive-Sharpe trades — cannot mock reliably because DSR gate requires MinTradesForMetrics=30 trades with positive Sharpe, which needs realistic candle series).

Follow-up tasks created:
- TASK-0102: wire universesweep.Result.Trades to skip bootstrap engine re-run
- TASK-0103: makeStageDir → os.MkdirAll for retry-after-error
- TASK-0104: applyWFGate traceability to decision document

TASK-0077 (cmd/param-search) now unblocked.

**Why:** Coverage gap lesson — flat/oscillating candle mocks don't produce positive-Sharpe strategies (sma-crossover on flat series: zero trades; on symmetric oscillator: near-zero DSR-corrected Sharpe fails gate). Integration-only pipeline paths are accepted with comment.

**How to apply:** When building similar pipeline CLIs, accept that stages 2+ coverage is integration-only. Focus coverage effort on pure helpers (gate functions, write helpers, flag parsing).
