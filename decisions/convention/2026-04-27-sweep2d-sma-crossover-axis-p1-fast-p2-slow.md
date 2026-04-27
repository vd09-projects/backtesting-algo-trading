# sweep2d sma-crossover axis mapping: p1=fast-period, p2=slow-period

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | convention       |
| Tags     | sweep2d, sma-crossover, axis-mapping, cmd/sweep2d, TASK-0044 |

## Context

`cmd/sweep2d` exposes two generic sweep axes (`--p1-*` and `--p2-*`) that map to strategy parameters. When building `factoryRegistry2D` for `sma-crossover`, the axis assignment had to be fixed: which axis is fast-period, which is slow-period. The flags `--p1-name` and `--p2-name` are cosmetic — they label the CSV output but do not control which strategy dimension each axis sweeps.

## Options considered

### Option A: p1=fast-period, p2=slow-period (chosen)

- **Pros**: The natural reading of a 2D Sharpe heatmap is fast period on rows (the faster-varying axis), slow period on columns. Matches the "fast × slow grid" description in the task AC. Consistent with how `cmd/sweep` names its axes (fast-period and slow-period).
- **Cons**: No flexibility — users who want slow on rows and fast on columns must transpose in Python.

### Option B: Configurable via `--p1-param / --p2-param` flag

- **Pros**: User controls which parameter each axis sweeps.
- **Cons**: Premature generalization at 2 strategies. Adds a dispatch layer that doesn't exist in the 1D sweep. Revisit if a strategy has more than 2 meaningful sweep dimensions.

## Decision

p1→fast-period, p2→slow-period is fixed by convention in `factoryRegistry2D`. This is the natural parameter interaction surface for a crossover strategy. No flag allows axis swapping in v1 — the heatmap is always fast on rows, slow on columns. The `--p1-name` / `--p2-name` flags are cosmetic only (CSV column/row labels), not axis selectors.

## Consequences

- `--p1-name` and `--p2-name` in `cmd/sweep2d` are labels, not axis selectors. Users who try to swap axes by changing these flags will relabel the CSV but not change which parameter is swept on which axis.
- When the remaining 4 strategies are added to `factoryRegistry2D` (TASK-0061), each will need its own axis convention documented here or in TASK-0061's notes.

## Revisit trigger

If a strategy has more than 2 meaningful parameter dimensions, or if a use case arises for swapping axis orientation from the CLI, add a `--p1-param / --p2-param` flag and remove the fixed convention.
