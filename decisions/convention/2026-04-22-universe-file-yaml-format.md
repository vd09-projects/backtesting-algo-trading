# Universe file uses YAML with top-level `instruments:` key

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | convention       |
| Tags     | YAML, universe-file, file-format, universesweep, TASK-0035 |

## Context

Defining the file format for universe files (lists of instruments to sweep). The file must be human-editable and machine-parseable. `gopkg.in/yaml.v3` is already in `go.mod`.

## Options considered

### Option A: Plain text, one instrument per line
- **Pros**: Maximally simple; no parser dependency.
- **Cons**: No extensibility. Adding per-instrument metadata (exchange override, lot size, comment) later requires breaking the format or a second file. Also no self-documenting key.

### Option B: YAML bare sequence (`- NSE:RELIANCE` at root)
- **Pros**: Slightly terser.
- **Cons**: Fragile to extension — adding a top-level `name:` or `description:` field would break the parser.

### Option C: YAML with top-level `instruments:` key (chosen)
- **Pros**: Extensible. Adding `name:`, `description:`, `asset_class:` fields later does not break existing files or the parser. The `instruments:` key mirrors the terminology used throughout the codebase.
- **Cons**: One extra line.

## Decision

Universe files use `gopkg.in/yaml.v3` with the schema `type universeFile struct { Instruments []string }`. The file format is:

```yaml
instruments:
  - NSE:RELIANCE
  - NSE:INFY
```

## Consequences

`ParseUniverseFile` decodes `universeFile`, validates non-empty, deduplicates (order-preserving). The named key is the stable contract; new top-level fields can be added to `universeFile` as optional fields without changing the instruments list parsing.
