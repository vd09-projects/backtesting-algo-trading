# Marginal ADV flagged inline rather than a separate list

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention        |
| Tags     | [universe-yaml, ADV, marginal-flag, yaml-comments, convention, TASK-0072] |

## Context

Marcus's ADV threshold call introduced two tiers: Rs 50 crore (hard minimum) and Rs 50–100 crore (marginal range, flag for scrutiny). Five instruments in the midcap universe fall in the marginal range: BHEL, SCHAEFFLER, EXIDEIND, NATIONALUM, SAIL. A convention was needed for how to mark these within the universe file.

## Options considered

### Option A: Inline comment on the instrument line (chosen)
- **Pros**: Flag is colocated with the instrument; visible in file review and in diffs; no new file or section needed.
- **Cons**: Comment text is informal; no machine-queryable structure.

### Option B: Separate `marginal:` section listing flagged instruments
- **Pros**: Machine-queryable (though no tool reads it yet).
- **Cons**: Splits the "this instrument exists" and "this instrument is marginal" information; reviewer must cross-reference two sections; schema change from the existing flat-list convention.

### Option C: Separate `marginal-instruments.yaml` file
- **Pros**: Fully decoupled; clean.
- **Cons**: Three files needed to understand the full picture (universe + sector groups + marginal flags); unnecessary for a flag that will likely be resolved by first sweep results.

## Decision

Marginal-ADV instruments are flagged with an end-of-line YAML comment on the same line as the instrument entry:

```yaml
  - NSE:BHEL        # marginal ADV (50-100 crore range) — PSU, liquidity can compress
```

The flag is informal — it exists to prompt scrutiny after sweep results, not to trigger automated behavior. If sweep results consistently show `insufficient_data=true` for these instruments, they are replaced and the flag becomes moot.

## Consequences

- Marginal-ADV flags are visible at a glance and colocated with the instrument entry.
- No `ParseUniverseFile` changes needed — comments are stripped by the YAML decoder.
- If automated ADV gating is added in the future (e.g., a pre-sweep check against a live ADV source), the comment convention should be replaced by a structured field.

## Related decisions

- [Sector grouping in universe YAML via inline comments](./2026-05-07-sector-grouping-in-universe-yaml.md) — companion convention for sector grouping
- [ADV floor for midcap daily-bar universe](../algorithm/2026-05-07-adv-floor-midcap-daily-bar-universe.md) — the methodology decision that requires this flag

## Revisit trigger

If a second universe file introduces marginal-ADV instruments, confirm the convention holds. If any tooling needs to programmatically identify marginal instruments, migrate to a structured YAML field.
