---
name: cmd/ entrypoint pattern
description: Structural pattern for all cmd/ CLI binaries in this project — flags struct, factory registry, validate helper, cmdutil.BuildProvider
type: project
---

All cmd/ entrypoints in this project follow a consistent structure established across cmd/sweep, cmd/universe-sweep, and cmd/sweep2d:

1. **flags struct** (e.g. `flags2D`) — groups parsed flag values for passing to the validate helper; keeps the validator testable without constructing a `flag.FlagSet`. Tests construct the struct literal and mutate one field per case.

2. **parseAndValidateFlags** helper — validates required flags, step constraints, date parse, and timeframe parse. Returns typed parsed values. Always returns an error (not calls `cmdutil.Fatalf`) so it is testable.

3. **factoryRegistry / strategyRegistry** — dispatches on `--strategy` string to return a strategy factory or instance. Extracted as a named function (not inline in main) to keep main() under cyclop limit of 15.

4. **cmdutil.BuildProvider(ctx)** — all binaries delegate provider construction here; never copy-paste the buildProvider logic.

5. **main()** stays under cyclomatic complexity 15 (golangci-lint cyclop limit). Extract helpers aggressively.

**Why:** Established pattern from TASK-0023 (cmd/sweep), TASK-0035 (cmd/universe-sweep), TASK-0044 (cmd/sweep2d). The flags struct replaces individual parameters when parameter count exceeds ~7.

**Batch CLI variant (cmd/fetch-history, TASK-0070):** Batch/CI tools that accept a direct access token (not OAuth) use a `providerFactory func(fetchFlags)(provider.DataProvider,error)` parameter in `run()` instead of `cmdutil.BuildProvider`. The factory receives parsed flags so main() passes `buildProductionProvider` directly. This eliminates global state and keeps dry-run paths from constructing a provider. Reference: cmd/fetch-history/main.go.

**Pipeline orchestrator variant (cmd/evaluate, TASK-0073):** Multi-stage pipeline CLIs that call multiple internal packages in sequence use `providerFactory func(context.Context)(provider.DataProvider,error)` in `run()` (context-based, not flags-based). Cyclop limit requires splitting run() into parseEvalFlags + buildPipeline + runPipeline + stage helpers. Stage outputs to dated subdirectory. All pure helpers (gate functions, write helpers, flag parsing) testable; provider-dependent stages are integration-only. Reference: cmd/evaluate/main.go.

**How to apply:** When planning a new cmd/ binary, check cmd/sweep/main.go as the reference implementation for interactive tools, cmd/fetch-history/main.go for batch/CI tools, cmd/evaluate/main.go for pipeline orchestrators. The flags struct and factoryRegistry pattern are non-negotiable for testability.
