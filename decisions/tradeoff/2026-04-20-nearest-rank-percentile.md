# Nearest-rank percentile for bootstrap CIs (floor index)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-20       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | montecarlo, bootstrap, percentile, nearest-rank, linear-interpolation, p5, p50, p95, TASK-0024 |

## Context

`Bootstrap()` reports confidence intervals as (p5, p50, p95) of the sorted simulation distribution. The `percentile()` helper must select a value from the sorted slice. Two practical methods exist: nearest-rank (floor index) and linear interpolation between adjacent ranks.

## Options considered

### Option A: Linear interpolation (numpy default)
`result = sorted[lo] + frac * (sorted[hi] - sorted[lo])` where lo = floor(p*(n-1)), hi = lo+1.

- **Pros**: Slightly smoother estimates; consistent with numpy and scipy defaults.
- **Cons**: Produces non-observed values (interpolated between two simulation outcomes). Adds code complexity. At N=10,000, the difference from nearest-rank is undetectable — the output is already rounded to 4 decimal places.

### Option B: Nearest-rank, floor index (chosen)
`index = floor(p * n)`, return `sorted[index]`.

- **Pros**: Returns an observed simulation value — intuitive and interpretable ("the actual p5 Sharpe was X in one of the simulations"). Simple, deterministic, no floating-point interpolation. At N=10,000, p5=index 500, p95=index 9500 — fully stable estimates.
- **Cons**: On very small N (e.g., N=20 for unit tests), estimates are coarser. Tests cover this; it's a known limitation of small N, not a bug in the algorithm.

### Option C: Nearest-rank, round (alternative nearest-rank)
`index = round(p * n)`.

- **Pros**: Minimizes expected error vs. the true quantile on small N.
- **Cons**: Introduces non-determinism at the boundary (0.5 rounds differently in different languages and Go versions). Floor is deterministic everywhere.

## Decision

Floor-index nearest-rank: `index := int(math.Floor(p * float64(len(sorted))))`, clamped to `[0, len-1]`. Returns an observed simulation value. Simple, deterministic, correct for N ≥ 100. The kill-switch threshold (`p5 Sharpe`) is used to set a concrete floor for live monitoring — it must be an observed value, not an interpolation artifact.

## Consequences

- At N=10,000 (default): p5, p50, p95 are stable to the noise level of the output format (4dp).
- At very small N (testing with N=20): expect coarse quantile estimates. This is expected and tests use loose bounds.
- TASK-0026 (kill-switch) will set its threshold from `BootstrapResult.SharpeP5`. If the interpolation method is ever changed, the absolute value of `SharpeP5` shifts slightly — update TASK-0026 accordingly and re-run the bootstrap for any strategies that have been evaluated.

## Related decisions

- [Bootstrap Sharpe non-annualized per-trade](../algorithm/2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md) — the distribution being percentile-summarized.
- [Sharpe uses sample variance (n-1)](../algorithm/2026-04-10-sharpe-sample-variance.md) — sample vs. population convention applied consistently here (sample variance in Sharpe computation within each bootstrap simulation).
