# NSE charge rates as unexported constants in commission.go, not as OrderConfig fields

- **Date:** 2026-04-25
- **Status:** experimental
- **Category:** architecture
- **Tags:** commission, NSE, constants, OrderConfig, TASK-0038
- **Task:** TASK-0038

## Decision

NSE statutory charge rates (STT, exchange charges, SEBI charges, stamp duty rates, GST rate) are
defined as unexported package-level constants in `internal/engine/commission.go`, not as fields on
`OrderConfig` or any other caller-configurable struct.

## Rationale

These rates are statutory facts set by SEBI and the NSE, not user-configurable parameters. Putting
them in `OrderConfig` would imply they are variable inputs — which would be misleading and open the
door to accidental misconfiguration. Constants are also zero-overhead at runtime and immediately
readable alongside the arithmetic that uses them.

The only legitimate reason to make them configurable would be to support multiple exchanges or
regulatory regimes. That requirement does not exist in v1 — this is an NSE-only engine. When a
second exchange is added, the right approach is a per-exchange cost struct, not flag-level overrides.

## Rejected alternatives

- **Fields on `OrderConfig`** — implies rates are caller-controlled inputs; creates misconfiguration
  risk; adds struct bloat for values that never change in practice.
- **Config file / YAML** — same problem: rates appear configurable when they are not.

## Consequences

To update a charge rate (e.g. if SEBI changes its levy), edit `commission.go`. The change is a
one-line constant update with a re-run of the golden tests to verify the new expected round-trip
cost. No API surface changes.
