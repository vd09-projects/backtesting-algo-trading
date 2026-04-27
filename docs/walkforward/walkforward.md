# Walk-Forward Validation

**Package:** `internal/walkforward/`
**File:** `walkforward.go`

Walk-forward is a regime-stability test. It checks whether a strategy's
out-of-sample (OOS) performance degrades significantly compared to its
in-sample (IS) performance across multiple time windows.

**This is not parameter optimization.** Parameters are fixed before the
walk-forward run. The harness tests whether fixed parameters that looked
good in-sample still hold OOS — a signal of overfitting to a specific
market regime.

---

## Window Structure

```
Full evaluation range: 2018-01-01 → 2024-12-31

Default sizing: InSampleWindow=2y, OutOfSampleWindow=1y, StepSize=1y

Fold 1:  IS [2018-01-01 → 2020-01-01)  OOS [2020-01-01 → 2021-01-01)
Fold 2:  IS [2019-01-01 → 2021-01-01)  OOS [2021-01-01 → 2022-01-01)
Fold 3:  IS [2020-01-01 → 2022-01-01)  OOS [2022-01-01 → 2023-01-01)
Fold 4:  IS [2021-01-01 → 2023-01-01)  OOS [2023-01-01 → 2024-01-01)
Fold 5:  IS [2022-01-01 → 2024-01-01)  OOS [2024-01-01 → 2025-01-01)

Fold 6 would need OOS end = 2026-01-01, which is after cfg.To → excluded
```

Windows step by `StepSize`. A fold is excluded when its OOS end exceeds `cfg.To`.

---

## Per-Fold Computation

```
For each fold:
  ┌─────────────────────────────────────────────────────────────┐
  │  IS run:                                                    │
  │    engine.Config{From: isStart, To: isEnd, ...}            │
  │    eng.Run(ctx, provider, strategy)                         │
  │    isTrades = eng.Portfolio().ClosedTrades()                │
  │    isSharpe = perTradeSharpe(isTrades)                      │
  │                                                             │
  │  OOS run:                                                   │
  │    engine.Config{From: oosStart, To: oosEnd, ...}          │
  │    eng.Run(ctx, provider, strategy)                         │
  │    oosTrades = eng.Portfolio().ClosedTrades()               │
  │    oosSharpe = perTradeSharpe(oosTrades)                    │
  └─────────────────────────────────────────────────────────────┘

  WindowResult{
    InSampleSharpe:    isSharpe,
    OutOfSampleSharpe: oosSharpe,
    TradeCount:        len(oosTrades),
    Degenerate:        len(oosTrades) == 0,
  }
```

IS and OOS runs within a fold are sequential (not parallel with each other).
The fold as a whole is one unit of parallel work.

---

## Parallelism

```
Folds run in parallel via errgroup:
  g, gctx := errgroup.WithContext(ctx)
  g.SetLimit(runtime.GOMAXPROCS(0))   ← cap to available CPUs

  for i := range windows:
    g.Go(func() { results[i] = runFold(...) })

  g.Wait()
```

Results are written to a pre-allocated slice at fixed indices — fold ordering
is preserved regardless of which goroutine finishes first.

---

## Degenerate Folds

A fold with zero OOS closed trades is marked `Degenerate = true`.

Degenerate folds are **excluded from all scoring** — they don't count toward
averages, they don't contribute to flag checks, and they don't increment
`DeduplicatedFoldCount`.

"No trades" ≠ overfitting. A long-only strategy may simply not generate signals
in a short OOS window. Treating that as a failure would be incorrect.

---

## Scoring

```
Over non-degenerate folds only:

AvgInSampleSharpe    = mean(WindowResult.InSampleSharpe)
AvgOutOfSampleSharpe = mean(WindowResult.OutOfSampleSharpe)
DeduplicatedFoldCount = count of non-degenerate folds

OverfitFlag     = AvgOutOfSampleSharpe < 0.5 × AvgInSampleSharpe
NegativeFoldFlag = count(OOS Sharpe < 0) >= 2
NegativeFoldCount = count(non-degenerate folds with OOS Sharpe < 0)
```

### OverfitFlag

If the average OOS Sharpe is less than half the average IS Sharpe, the strategy
is fitting the training regime rather than capturing transferable edge.

Example:
```
AvgISSharpe  = 0.80
AvgOOSSharpe = 0.30  → 0.30 < 0.5 × 0.80 = 0.40  → OverfitFlag = true
```

### NegativeFoldFlag

Two or more folds with negative OOS Sharpe suggests the strategy loses money
in a meaningful fraction of out-of-sample windows — more than bad luck.

---

## Report

```go
walkforward.Report{
    Windows               []WindowResult   // all folds, including degenerate
    AvgInSampleSharpe     float64
    AvgOutOfSampleSharpe  float64
    DeduplicatedFoldCount int
    OverfitFlag           bool
    NegativeFoldFlag      bool
    NegativeFoldCount     int
}
```

---

## Config

```go
walkforward.WalkForwardConfig{
    InSampleWindow:    2 * 365 * 24 * time.Hour,   // 2 years
    OutOfSampleWindow: 365 * 24 * time.Hour,        // 1 year
    StepSize:          365 * 24 * time.Hour,        // advance 1 year per fold
    Instrument:        "NSE:RELIANCE",
    From:              time.Date(2018, 1, 1, ...),
    To:                time.Date(2025, 1, 1, ...),
}

walkforward.EngineConfigTemplate{
    InitialCash:          100000,
    OrderConfig:          model.OrderConfig{SlippagePct: 0.0005, CommissionModel: model.CommissionZerodha},
    PositionSizeFraction: 0.10,
    SizingModel:          model.SizingFixed,
    VolatilityTarget:     0,
}
```

The `EngineConfigTemplate` is separate from `WalkForwardConfig` by design:
the harness owns `Instrument`, `From`, `To` per fold; the caller owns
cost model and sizing. This prevents accidentally setting the wrong date
range in the engine config template.

---

## Walk-Forward Gate

A strategy passes the walk-forward gate if **both** flags are false:

```
OverfitFlag     = false    (OOS Sharpe >= 50% of IS Sharpe on average)
NegativeFoldFlag = false   (fewer than 2 OOS folds with negative Sharpe)
```

Both are required — not either/or. A strategy can fail by overfitting (IS/OOS
degradation) or by being regime-unstable (repeated negative OOS folds) independently.
Failing either condition kills the strategy at this gate; it does not proceed to
bootstrap or correlation evaluation.

---

## Interpreting Results

| Result | Interpretation |
|---|---|
| OverfitFlag = false, NegativeFoldFlag = false | Strategy passes the walk-forward gate — proceed to bootstrap. |
| OverfitFlag = true | IS/OOS degradation is significant. Parameters may be curve-fit. Revisit parameter choice. |
| NegativeFoldFlag = true | Strategy loses money in 2+ time windows. Either the regime changed or the edge was never real. |
| All folds degenerate | Strategy never trades in OOS windows. Signal frequency too low for the OOS window size. |
| OOS Sharpe > IS Sharpe | Unusual but possible (lucky OOS period). Not a problem, but worth investigating. |
