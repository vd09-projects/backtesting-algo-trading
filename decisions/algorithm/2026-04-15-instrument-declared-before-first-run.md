# Target instrument must be declared in writing before the first backtest run

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | backtest, instrument, anti-cherry-picking, research-hygiene, methodology, TASK-0028 |

## Context

Before running the first live backtest on NSE data, we needed a rule governing how the test
instrument is chosen. Without a rule, the natural tendency — even unintentionally — is to run
on several candidates, observe which produces the best result, and present that as the finding.
This is selection bias: the reported result is the maximum of a sample, not the expected value
of the strategy on an honestly-chosen instrument.

The rule needed to be simple enough to actually follow and clear enough that the discipline is
auditable.

## Options considered

### Option A: No rule — choose freely, even after seeing results
- **Pros**: Maximum flexibility
- **Cons**: Completely invalidates the test. If you run on 5 instruments and report the 2 that
  passed Sharpe ≥ 0.5, you haven't found edge — you've selected for it. The proliferation gate
  becomes meaningless.

### Option B: Document after the run, before publishing results
- **Pros**: Still provides a paper trail
- **Cons**: Still allows post-hoc selection. "I was always planning to use this one" is easy to
  say after the fact.

### Option C: Declare in the task acceptance criteria before any run
- **Pros**: Commits the instrument name in a version-controlled file before results exist.
  Auditable: the task block shows the name was present before the run completed.
- **Cons**: Requires discipline. Technical enforcement is not possible — the rule relies on the
  practitioner following the process.

## Decision

The instrument must be named in TASK-0028's first acceptance criterion before the first
`go run cmd/backtest` is executed. Once written, the name cannot be changed after results are
seen. The acceptance criterion reads: **"INSTRUMENT: _______"** — the blank is filled in before
running, not after.

Selection criteria for a valid instrument (Marcus): liquid Nifty 50 constituent, trading
continuously since before 2018 with no structural break in the price series, a name you would
actually consider trading with real capital. The last criterion is the important one — it prevents
choosing a historically-clean instrument that you would never actually trade.

If you run on multiple instruments to check robustness, each instrument must be declared before
its run. The primary gate check (Sharpe ≥ 0.5 for the proliferation gate) uses the first-declared
instrument only.

## Consequences

- The discipline is enforced by honor system in the task block, not by code.
- Results from any run that did not follow the declaration protocol are not valid for the
  proliferation gate decision — they can be used for exploration but cannot determine whether
  TASK-0019 or TASK-0020 proceed.
- If a future multi-instrument expansion happens, the declaration discipline should extend to
  requiring the full instrument universe to be named before any run in that universe.

## Related decisions

- [Strategy proliferation gate — Sharpe ≥ 0.5 vs buy-and-hold before variation strategies](./2026-04-10-strategy-proliferation-gate.md) — the gate this discipline protects
- [Baseline backtest period set to 2018–2024](./2026-04-15-baseline-backtest-period-2018-2024.md) — the period used in the declared run

## Revisit trigger

If expanding to multi-instrument backtesting. The declaration protocol applies to the universe,
not just a single name — revisit how the commitment is recorded when more than one instrument is
in scope.
