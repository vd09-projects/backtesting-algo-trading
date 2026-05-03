# Memory Index

- [cyclop factory table pattern](feedback_cyclop_factory_table.md) — Use map[string]builderFunc dispatch table when factory switch hits cyclop limit (7+ strategy cases)
- [errcheck in validated closures](feedback_errcheck_in_closures.md) — Use explicit panic with invariant message instead of `_` in closures after prior validation; errcheck fires on `s, _ :=`
- [forvar loop variable capture Go 1.22+](feedback_forvar_go125.md) — Remove `i := i` copies in range loops; Go 1.22+ per-iteration semantics makes them unneeded and linter-flagged
- [run() extraction for cmd testability](feedback_run_extraction_testability.md) — Extract main wiring into run(args, stdout, stderr) to unit-test flag/validation paths without subprocess or live credentials
