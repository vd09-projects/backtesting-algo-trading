# params column as JSON object in param-search-results.csv

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | CSV, params, JSON, serialization, cmd/param-search, TASK-0077 |

## Context

`writeResultsCSV` in `cmd/param-search` serializes each `VariantResult` to a CSV row. The params field is a `map[string]float64` whose key count varies by grid dimensionality — a 1-axis grid has 1 param key, a 3-axis grid has 3. The CSV schema had to accommodate variable-length params.

## Options considered

### Option A: One column per axis (dynamic headers)
Generate headers based on the grid's axis names: `fast-period, slow-period, dsr_sharpe, raw_sharpe, trade_count, insufficient_data`. Downstream tools get axis values as first-class columns.
- **Pros**: Native column access in pandas/Excel without JSON parsing. Sort and filter on individual params directly.
- **Cons**: CSV headers change per grid spec. Downstream tools (dashboards, reporting scripts) would need to know the axis names at read time or schema-detect dynamically. Two CSVs from different grid specs can't be concatenated without schema alignment. Requires the writer to emit axis names in a consistent order.

### Option B: Single JSON params column — chosen
Serialize the entire params map as a compact JSON object in one column: `{"fast-period":17,"slow-period":26}`.
- **Pros**: Fixed headers across all grids regardless of axis count. CSVs from different grid specs have the same schema and can be stacked. pandas handles it with `df['params'].apply(json.loads)` → expand to columns if needed. jq handles it natively. The schema is self-describing — the params object carries its own key names.
- **Cons**: Requires one extra parse step downstream to access individual param values. Not directly sortable in Excel without a formula.

## Decision

Option B. The fixed-header stability outweighs the parse-step cost. Research workflows (pandas, jq, shell) handle JSON column parsing trivially. The schema self-describing nature means a CSV from any param-search run is interpretable without knowing the grid spec used to produce it.

## Consequences

- `json.Marshal(res.Params)` in `writeResultsCSV` — the map is serialized in non-deterministic key order (Go map iteration). For reproducibility, callers who need deterministic param column order should sort the keys before comparison. This is a downstream concern and acceptable — DSR ranking is the primary sort key, not param value.
- If a future version of the tool produces one-column-per-axis CSV, that is a new output format (different flag or output path), not a change to this convention.

## Revisit trigger

If the primary consumer of param-search-results.csv is a tool that requires native column access and cannot handle JSON parsing (e.g., a BI tool with no scripting layer), revisit and add a `--wide-csv` flag that emits one column per axis alongside the JSON params column.
