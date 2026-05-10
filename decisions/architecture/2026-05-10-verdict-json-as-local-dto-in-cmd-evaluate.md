# verdict.json as local DTO in cmd/evaluate — not internal/output

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | serialization-dto, cmd-layer, JSON, verdict, pipeline-output, cmd/evaluate, TASK-0073 |

## Context

`cmd/evaluate` writes a `verdict.json` file summarizing the pipeline run: result (pass/killed), gate stats, survivor records, and kill records. A question arose about where to define the serialization types: should they go in `internal/output` (where `cmd/backtest`'s JSON result types live) or stay local to `cmd/evaluate`?

## Options considered

### Option A: Extend internal/output with verdict types
- **Pros**: Centralized serialization types; potential reuse by future consumers.
- **Cons**: `internal/output` is designed for per-backtest result serialization. A pipeline-level verdict aggregating gate results across three stages is a fundamentally different concern. Adding pipeline types to a per-backtest package blurs its responsibility.

### Option B: Local DTO in cmd/evaluate (chosen)
- **Pros**: `verdictJSON`, `gateStatsJSON`, `survivorRecord`, `killRecord` are defined only in `cmd/evaluate/main.go`. The DTO schema is owned by the cmd that writes it. Clean separation: `internal/output` handles per-backtest serialization; `cmd/evaluate` handles pipeline orchestration output.
- **Cons**: One more struct definition to maintain.

## Decision

Local DTO in `cmd/evaluate/main.go`. The `verdictJSON` struct and its supporting types (`gateStatsJSON`, `survivorRecord`, `killRecord`) are defined only in `cmd/evaluate`. This follows the `thresholdsFile` DTO precedent in `cmd/monitor` (2026-05-07): cmd-layer concerns stay in the cmd layer, `internal/` types stay pure.

## Consequences

- `internal/output` remains a pure per-backtest serialization package. No pipeline-level types pollute it.
- `verdictJSON` schema is owned by `cmd/evaluate`. Changes to the verdict schema require only editing `cmd/evaluate/main.go`.
- If a second cmd binary needs to read verdict.json (e.g., `cmd/param-search` post-evaluation), consider moving the DTO to `internal/cmdutil` — the same threshold used for `thresholdsFile` (one consumer = local; two consumers = extract).

## Related decisions

- [thresholdsFile DTO in cmd/monitor](2026-05-07-thresholdsfile-dto-in-cmd-monitor.md) — originating precedent: cmd-layer DTOs stay in cmd/
- [RunConfig placed in internal/output as serialization DTO](../architecture/2026-05-01-runconfig-in-internal-output.md) — contrasting case: per-backtest serialization lives in internal/output
