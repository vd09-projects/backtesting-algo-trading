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

**How to apply:** When planning a new cmd/ binary, check cmd/sweep/main.go as the reference implementation. The flags struct and factoryRegistry pattern are non-negotiable for testability.
