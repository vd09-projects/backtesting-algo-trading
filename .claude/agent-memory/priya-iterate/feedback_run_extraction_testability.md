---
name: run() extraction for cmd testability
description: Extract main's wiring into run(args, stdout, stderr) to enable unit testing of flag/validation paths without subprocess or credentials
type: feedback
---

Extract `main()` wiring into `run(args []string, stdout, stderr io.Writer) error` when cmd coverage drops below threshold. Tests can then cover flag-parse failures, unknown strategy, and invalid commission without spawning a subprocess or needing live credentials.

Key design points:
- Use `flag.NewFlagSet("name", flag.ContinueOnError)` with `fs.SetOutput(stderr)` — not the global `flag` package
- Return errors instead of calling `cmdutil.Fatalf`; `main()` handles os.Exit
- For paths that need os.Exit with a specific code (report flags), return a typed `exitCodeError`; `main()` uses `errors.As` to detect and translate
- Tests only cover paths up to the provider/network boundary; paths requiring live credentials are integration-only and stay uncovered

**Why:** Coverage threshold (70%) cannot be met for cmd packages unless the flag-parse/validation layer is testable in-process. Subprocess tests would require live credentials or complex setup.

**How to apply:** Apply to any cmd/ entrypoint whose coverage is below 70%. The pattern is: extract run(), test the validation/flag paths, leave the live-provider path uncovered and note it in the coverage report.
