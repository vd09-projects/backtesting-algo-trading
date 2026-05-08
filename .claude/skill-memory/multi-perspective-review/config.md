# Project Config

## Reviewer Overrides
always_include:
  - Tech Debt Sentinel
  - Domain Logic Reviewer

always_exclude:
  - Observability & Debuggability Reviewer
  - Data Integrity & Migration Reviewer

## Project Context
domain: "Go algorithmic backtesting engine for Indian equity markets (NSE/BSE — Nifty50, BankNifty)"
primary_languages: Go
architecture: "Event-driven single-process engine. DataProvider → Engine ← Strategy → Trade Log → Analytics → Output. No DB, no HTTP, no migrations."
urgency_default: normal
debt_tolerance: normal

## Custom Triage Rules
- internal/engine/ change → always include Concurrency & State Safety Reviewer (goroutine lifecycle, hot loop)
- pkg/strategy/ change → always include API & Contract Reviewer (Strategy interface is the central contract)
- pkg/provider/ change → always include API & Contract Reviewer + Dependency & Coupling Reviewer (DataProvider isolation must be absolute)
- internal/analytics/ change → always include Domain Logic Reviewer (P&L/Sharpe/drawdown math must be exact)
- strategies/ change → always include Domain Logic Reviewer
- pkg/model/ change → always include Ripple Effect Analyst (shared domain types, breakage is silent)
- cmd/universe-sweep/ change → always include Ripple Effect Analyst (strategy registry)

## Reviewer Voice Tuning
- Naming Guardian: enforce Go idioms — PascalCase exports, consistent receiver names per type, no stutter (pkg.PkgType)
- Domain Logic Reviewer: flag any P&L or metric formula that deviates from standard finance definitions
- Tech Debt Sentinel: flag hardcoded instrument tokens, magic numbers in signal logic, untracked TODOs in hot paths
