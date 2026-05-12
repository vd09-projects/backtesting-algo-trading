# PriceExit TP exact-boundary tests use 20% threshold, not 10%

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-11       |
| Status   | experimental     |
| Category | convention       |
| Tags     | pkg/strategy, PriceExit, wrapper, target-profit, float64, boundary-test, test-design, TASK-0098 |

## Context

`TestPriceExit_TargetProfitFires` and `TestPriceExit_OnlyTPEnabled` verify that the TP guard fires at the exact threshold — i.e., when `bar.Close == entryPrice * (1 + targetProfitPct)`. The intent is to pin the `>=` operator (not `>`) in the implementation.

The natural test parameters would be entry price 100 and TP = 10%, giving a threshold of 110.0. However, in IEEE 754 float64:

```
100.0 * (1 + 0.10) = 110.00000000000001421085 (not 110.0 exactly)
```

Setting `bar.Close = 110.0` and asserting `>= 110.00000000000001` fails — the test would not verify the boundary at all; it would verify that the TP does *not* fire (which is the wrong test).

## Options considered

### Option A: Use 10% TP with 110.0 Close (rejected)

The "obvious" choice. Rejected because the float64 arithmetic makes it silently wrong: the test passes (Close=110.0 doesn't fire TP with a threshold of 110.00000000000001) even if the implementation uses `>=` correctly, but the test message says "TP fires at exact boundary" — which is false.

### Option B: Use a float64-exact boundary (chosen)

Find a TP percentage where `entryPrice * (1 + pct)` is exactly representable in float64. Verified:

```
100.0 * (1 + 0.20) = 120.0 exactly
100.0 * (1 + 0.05) = 105.0 exactly
100.0 * (1 - 0.05) = 95.0 exactly  (SL test, already works)
```

Use 20% for the TP boundary test: `Close = 120.0`, `targetProfitPct = 0.20`. The boundary assertion `120.0 >= 100.0 * 1.20` is true, confirming the `>=` operator.

### Option C: Use a non-boundary value (rejected)

Test with Close clearly above the threshold (e.g. 115.0 for 10% TP). This would confirm TP fires when price exceeds the threshold, but not that it fires *at* the threshold (i.e., it wouldn't catch if the implementation uses `>` instead of `>=`). This was the original 5-scenario plan's coverage level; a separate boundary test was added precisely to catch `>` vs `>=` bugs.

## Decision

Use 20% TP with Close=120.0 for exact-boundary tests. The test comment documents this choice:

```go
// 20% is used because 100*1.20 is exactly representable in float64; 100*1.10 is not.
```

This is a test-design constraint only. The production implementation uses `>=` as specified by the AC, and the threshold arithmetic works correctly for any caller-supplied value including 10%. The 20% choice only affects how the boundary test is constructed — not how the guard behaves.

## Consequences

- `TestPriceExit_TargetProfitFires` and `TestPriceExit_OnlyTPEnabled` use `targetProfitPct = 0.20` and `Close = 120.0` rather than the "natural" 10%/110.0.
- Future maintainers who see these test values should not change them to 10% without understanding the float64 representability issue. The test comment explains this, but this decision file provides the full reasoning.
- The SL boundary test is unaffected: `100 * (1 - 0.05) = 95.0` is exactly representable, so the SL test correctly uses 5%/95.0.

## Experiments

Verified in Go:

```
100.0 * (1 + 0.10) = 110.00000000000001421085  →  110.0 >= threshold is FALSE
100.0 * (1 + 0.20) = 120.00000000000000000000  →  120.0 >= threshold is TRUE
100.0 * (1 - 0.05) = 95.00000000000000000000   →  95.0  <= threshold is TRUE
```

## Revisit trigger

If the test framework is extended with an `approxEqual` helper that asserts equality within a machine-epsilon tolerance, the boundary tests could be rewritten using 10% with a tolerance check on the threshold comparison. At that point, this decision can be superseded.
