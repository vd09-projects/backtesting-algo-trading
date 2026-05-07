# Sector grouping in universe YAML via inline comments

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention        |
| Tags     | [universe-yaml, yaml-comments, sector-grouping, universe-file-format, convention, TASK-0072] |

## Context

Universe YAML files (e.g., `universes/nifty50-large-cap.yaml`) contain a flat list of instrument strings under the `instruments:` key. As the midcap universe grew to 25 instruments across 9 sectors, a flat list became harder to review and audit manually. A structuring convention was needed.

## Options considered

### Option A: Sector grouping via YAML comments (chosen)
- **Pros**: Zero schema change — `ParseUniverseFile` already strips comments and returns a flat string slice. Human-readable in both the file and diffs. No new fields or parsing logic needed.
- **Cons**: Comments are not machine-queryable; tools reading the YAML cannot programmatically extract sector metadata.

### Option B: Sector as a YAML key grouping instruments under nested lists
- **Pros**: Machine-queryable; sector metadata is structured.
- **Cons**: Breaking change to the schema; `ParseUniverseFile` would need updating; all existing universe files would need migration; adds complexity for a query use case that doesn't exist yet.

### Option C: No grouping — keep flat list, add a separate sector-mapping file
- **Pros**: Maximally simple file.
- **Cons**: Splits related information into two files; harder to review; duplicate maintenance burden.

## Decision

Use YAML comment lines (`# --- Sector Name ---`) to group instrument entries by sector within the `instruments:` block. Established in `universes/nifty-midcap-liquid.yaml`. Should be adopted in future universe files of ≥ 10 instruments.

The nifty50-large-cap.yaml file has only 15 instruments with no sector grouping — no migration needed unless a new version of that file is produced.

## Consequences

- Universe files ≥ 10 instruments use sector comment groups; smaller files may omit them.
- `ParseUniverseFile` requires no changes — YAML v3 decoder ignores comments.
- If a machine-queryable sector field is needed in the future, the schema should be extended with a structured `sectors:` block; the comment convention would then be deprecated.

## Related decisions

- [Marginal ADV flagged inline rather than a separate list](./2026-05-07-marginal-adv-flagged-inline.md) — companion convention for marginal-ADV annotation

## Revisit trigger

If any tooling needs to read sector metadata programmatically (e.g., sector-diversification check in universe construction), migrate to a structured YAML schema rather than extending comment parsing.
