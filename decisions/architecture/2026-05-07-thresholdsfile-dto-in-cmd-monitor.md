# thresholdsFile DTO in cmd/monitor, not JSON tags on KillSwitchThresholds

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | serialization-dto, JSON-tags, package-boundary, pure-computation, cmd/monitor, TASK-0048 |

## Context

`cmd/monitor` needs to read a pre-committed kill-switch thresholds file from disk.
`analytics.KillSwitchThresholds` holds the three threshold values (SharpeP5, MaxDrawdownPct,
MaxDDDuration) but has no JSON tags — it is a pure computation type used by `CheckKillSwitch`
with no prior serialization concern.

Two options for JSON deserialization:

### Option A: Add JSON tags to `analytics.KillSwitchThresholds` (rejected)

```go
type KillSwitchThresholds struct {
    SharpeP5       float64       `json:"sharpe_p5"`
    MaxDrawdownPct float64       `json:"max_drawdown_pct"`
    MaxDDDuration  time.Duration `json:"max_dd_duration_ns"`
}
```

- **Pros**: One type handles both computation and serialization.
- **Cons**: Couples the computation layer to a serialization concern. `internal/analytics` is a
  pure, dependency-free computation package — adding JSON tags there signals it is also a
  serialization type, which blurs its responsibility. `time.Duration` does not marshal as
  nanoseconds by default (it marshals as a float64 number of seconds via the `json.Marshaler`
  interface), so a custom marshaler would also be needed, further complicating the analytics type.

### Option B: Local `thresholdsFile` DTO in `cmd/monitor` (chosen)

```go
// In cmd/monitor/main.go only.
type thresholdsFile struct {
    SharpeP5        float64 `json:"sharpe_p5"`
    MaxDrawdownPct  float64 `json:"max_drawdown_pct"`
    MaxDDDurationNs int64   `json:"max_dd_duration_ns"`
}
```

Deserialization reads into `thresholdsFile`, then a one-line conversion builds the analytics struct:

```go
return analytics.KillSwitchThresholds{
    SharpeP5:       tf.SharpeP5,
    MaxDrawdownPct: tf.MaxDrawdownPct,
    MaxDDDuration:  time.Duration(tf.MaxDDDurationNs),
}, nil
```

- **Pros**: `internal/analytics` stays free of serialization concerns. `MaxDDDuration` is
  serialized explicitly as int64 nanoseconds — unambiguous, no custom marshaler, round-trips
  perfectly. The DTO is local to `cmd/monitor`, so it cannot be accidentally imported by other
  packages.
- **Cons**: One extra struct type in cmd/monitor. The conversion is three lines. Neither is a
  real cost.

## Decision

Local `thresholdsFile` DTO in `cmd/monitor/main.go`. `analytics.KillSwitchThresholds` receives
no JSON tags. The DTO handles the JSON layer; the conversion from DTO to analytics struct is a
one-liner in `loadThresholds`. `MaxDDDuration` is serialized as `int64` nanoseconds
(`max_dd_duration_ns` JSON key) — standard `time.Duration` integer encoding, no custom marshaler.

## Consequences

- `analytics.KillSwitchThresholds` remains a pure computation type: no serialization dependency,
  no custom marshaler, no JSON-awareness.
- The thresholds file schema is owned by `cmd/monitor`; changes to serialization format are
  localized there.
- Users writing thresholds JSON must use the `max_dd_duration_ns` key with a nanosecond int64.
  This is documented in the `cmd/monitor` package-level doc comment.

## Related decisions

- [Kill-switch derivation methodology](../algorithm/2026-04-21-kill-switch-derivation-methodology.md) — defines the three threshold fields this DTO serializes.
- [Kill-switch API keeps analytics free of montecarlo](2026-04-21-kill-switch-analytics-to-montecarlo-boundary.md) — same philosophy: analytics stays dependency-free of external concerns.

## Revisit trigger

If a second cmd/ binary also needs to read `KillSwitchThresholds` from disk, reconsider whether
the DTO should move to a shared location (e.g., `internal/cmdutil`). One consumer justifies a
local DTO; two consumers justify extraction.
