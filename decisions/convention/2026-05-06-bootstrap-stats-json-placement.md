# Bootstrap stats nested under "bootstrap" key in output JSON

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-06       |
| Status   | experimental     |
| Category | convention       |
| Tags     | JSON, bootstrap, omitempty, serialization, internal/output, BootstrapStats, TASK-0082 |

## Context

`cmd/backtest --bootstrap` computes bootstrap distribution statistics (SharpeP5, SharpeP50, SharpeP95, ProbPositiveSharpe, WorstDrawdownP95) via `internal/montecarlo`. These were printed to stdout only; the `--out` JSON file received no bootstrap fields. The evaluation-run pipeline (TASK-0069) had to parse stdout to capture these values — fragile, breaks if the stdout format changes, and prevents machine-readable audit trails.

TASK-0082 extended the JSON output to include bootstrap stats. The question was where to place them: promoted to the top level of `jsonResult` as individual fields, or under a named nested key.

## Options considered

### Option A: Top-level fields with `omitempty`

Add `BootstrapSharpeP5 float64 \`json:"bootstrap_sharpe_p5,omitempty"\`` and six sibling fields directly to `jsonResult`.

- **Pros**: Flat JSON, consistent with how `RunConfig` fields are promoted.
- **Cons**: Fatal: `omitempty` on `float64` suppresses the value when it equals `0.0`. `SharpeP5 == 0.0` is a valid and informative bootstrap outcome (the kill-switch threshold is exactly zero). Callers could not distinguish "bootstrap not run" from "bootstrap ran and p5 Sharpe is zero". Required adding a sentinel bool field `BootstrapRan bool` to disambiguate — adding complexity and a second source of truth.

### Option B: Named nested key with pointer-to-struct (chosen)

Add a `BootstrapStats` struct and `*BootstrapStats \`json:"bootstrap,omitempty"\`` to `jsonResult`.

- **Pros**: The pointer being nil means the key is absent entirely — unambiguous "bootstrap not run". When non-nil, all fields within the struct are serialized without `omitempty`, so a zero SharpeP5 is preserved. Single source of truth: struct present means bootstrap ran. Backward compatible: existing consumers see no new keys when bootstrap is not used.
- **Cons**: Bootstrap stats are nested under `"bootstrap": {...}` rather than at the JSON root. This is a minor ergonomic difference for consumers; downstream tooling reads `result.bootstrap.sharpe_p5` instead of `result.bootstrap_sharpe_p5`.

## Decision

Bootstrap stats are placed under a named `"bootstrap"` nested key in the output JSON using a `*BootstrapStats` pointer field with `omitempty` in `jsonResult`. The float64 zero-value ambiguity of Option A makes top-level promotion with `omitempty` semantically broken for this use case. A pointer-to-struct eliminates the ambiguity: the key is either fully present (bootstrap ran) or fully absent (bootstrap not run), with no sentinel field needed.

`BootstrapStats` is defined in `internal/output` as a serialization DTO, consistent with the `RunConfig` placement decision (2026-05-01): serialization concerns stay in `internal/output`, not `pkg/model`.

The 7 fields are: `sharpe_p5`, `sharpe_p50`, `sharpe_p95`, `prob_positive_sharpe`, `worst_drawdown_p95`, `n`, `seed`. The `n` and `seed` fields are required for reproducibility — without seed, the JSON result cannot be independently re-verified.

## Consequences

Consumers reading bootstrap stats from JSON must look under the `"bootstrap"` sub-key rather than at the top level. The evaluation-run pipeline agent (TASK-0084 tracks the update) currently parses stdout; once updated to read JSON, it will use `result.bootstrap.sharpe_p5`.

Non-bootstrap JSON output is unchanged: no `"bootstrap"` key appears when `--bootstrap` was not passed.

## Related decisions

- [jsonResult struct embedding for top-level JSON merge](2026-05-01-jsonresult-struct-embedding-top-level-json.md) — `BootstrapStats` uses a named field in `jsonResult` rather than embedding, precisely because embedding would require `omitempty` on individual float64 fields.
- [RunConfig placed in internal/output as serialization DTO](../architecture/2026-05-01-runconfig-in-internal-output.md) — same principle: `BootstrapStats` belongs in `internal/output` as a serialization concern, not promoted to `pkg/model`.

## Revisit trigger

If a consumer needs to query bootstrap stats at the JSON top level (e.g., `jq .sharpe_p5` without knowing the nested path), reconsider Option A with an explicit `BootstrapRan bool` sentinel field.
