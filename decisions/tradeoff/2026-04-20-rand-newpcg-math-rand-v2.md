# `math/rand/v2` with `rand.NewPCG` for bootstrap PRNG

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-20       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | montecarlo, bootstrap, prng, pcg64, math-rand-v2, determinism, reproducibility, seed, TASK-0024 |

## Context

The Monte Carlo bootstrap requires a deterministic, seedable pseudo-random number generator. The `BootstrapConfig.Seed` field is an explicit `int64`, logged in the output header, so that results are reproducible across runs and comparable across parameter sweep variants. The PRNG choice directly impacts: (a) statistical quality of the bootstrap, (b) determinism guarantees, and (c) reproducibility across machines.

## Options considered

### Option A: `math/rand` (v1) with `rand.New(rand.NewSource(seed))`
The stdlib v1 random package.

- **Pros**: Available without any import path change; familiar.
- **Cons**: v1 uses an LCG-based generator (not PCG) with weaker statistical properties. The global state (`rand.Seed()`) is deprecated and shared — risky in code that will eventually parallelize across sweep folds. The v1 API is superseded as of Go 1.22.

### Option B: `crypto/rand`
Cryptographically secure random bytes.

- **Pros**: Strongest entropy.
- **Cons**: Non-deterministic — cannot be seeded. Bootstrap **must** be deterministic: same seed → same result, every time. This is required for comparing bootstrap CIs across sweep variants and for reproducing a specific result from the TASK-LOG. `crypto/rand` is ruled out entirely.

### Option C: `math/rand/v2` with `rand.NewPCG(seed, 0)` (chosen)
Go 1.22+ stdlib PRNG using the Permuted Congruential Generator (PCG64).

- **Pros**: PCG64 passes the BigCrush statistical test suite. v2 API eliminates the global state concern. Each `rand.New(rand.NewPCG(seed, 0))` instance is independent and deterministic. No new dependency — already in the Go stdlib at v2. The `Intn(n)` method produces uniform integers directly without bias.
- **Cons**: Requires Go 1.22+. The repo already targets `go1.25.0` — not a concern.

## Decision

`math/rand/v2` with `rand.NewPCG(uint64(seed), 0)`. The zero stream constant is fixed so the seed is the sole source of variation. Each call to `Bootstrap()` creates a fresh, local `*rand.Rand` — no shared state, safe for future parallelization.

The seed is surfaced in two places:
1. `BootstrapConfig.Seed` (caller-controlled, defaults to 42 in the CLI)
2. Output header: `--- Bootstrap (N sims, seed=S) ---`

Reproducibility is a first-class property of the bootstrap API.

## Consequences

- Any run with the same `Seed` and `NSimulations` produces bit-identical results, regardless of machine or Go version patch (within the same Go minor version — PCG64 is stable).
- Sweep variants using the same seed are comparable: the noise term is identical, so CI differences reflect only parameter differences.
- If bootstrap is ever parallelized (one goroutine per simulation), each goroutine must construct its own `rand.Rand` with a derived seed (e.g., `seed + goroutine_index`). The shared-state risk is pre-empted by the current per-call construction pattern.

## Related decisions

- [Bootstrap Sharpe non-annualized per-trade](../algorithm/2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md) — the algorithm this PRNG serves.
- [Sweep executes sequentially — no errgroup](../tradeoff/2026-04-15-sweep-sequential-execution-no-errgroup.md) — current concurrency posture; bootstrap parallelization is deferred.
