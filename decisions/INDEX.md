# Decision Index

<!-- 
  This file is maintained by the decision-journal skill.
  Entries are in YAML format for machine-friendly querying.
  Newest entries go at the top. Do not manually reorder.
-->

```yaml
decisions:
  - id: 2026-04-03-no-pyramiding-v1
    title: "No pyramiding — single position per instrument enforced in v1"
    date: 2026-04-03
    status: accepted
    category: tradeoff
    tags: [portfolio, position-sizing, pyramiding, v1-scope, engine]
    path: tradeoff/2026-04-03-no-pyramiding-v1.md
    summary: "Buy signals on an already-open position are silently skipped in v1. Pyramiding deferred until a concrete strategy requires scale-in behaviour."

  - id: 2026-04-02-trade-pnl-stored-not-computed
    title: "Trade.RealizedPnL stored on the struct, not computed on-demand"
    date: 2026-04-02
    status: accepted
    category: convention
    tags: [trade, pnl, analytics, engine, commission, slippage]
    path: convention/2026-04-02-trade-pnl-stored-not-computed.md
    summary: "Engine computes and stores RealizedPnL at close time so analytics never needs to know about commission models or slippage — keeping it a pure read-only reporting layer."

  - id: 2026-04-02-strategy-lookback-as-interface-method
    title: "Strategy.Lookback() as a first-class interface method"
    date: 2026-04-02
    status: accepted
    category: architecture
    tags: [strategy, interface, engine, lookback, no-lookahead]
    path: architecture/2026-04-02-strategy-lookback-as-interface-method.md
    summary: "Lookback is a required interface method so the engine enforces the no-lookahead constraint universally, rather than delegating it to individual strategies."
```
