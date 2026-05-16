---
name: Composite signals must be regime filters, not oscillator conjunctions
description: Multi-signal strategies acceptable only as regime filters; oscillator stacking (MACD AND RSI AND WMA) rejected
type: feedback
---

Composite signal strategies are acceptable only as regime filters on a base strategy, not as oscillator conjunctions.

**Why:** Oscillator conjunctions (MACD AND RSI AND WMA all agree) use highly correlated indicators — no independent information is added, trade count drops sharply, confidence intervals widen. The user explicitly asked about these but Marcus and the user aligned that filters with independent economic rationale are the right model.

**How to apply:** Any composite signal proposal must state: (1) what independent mechanism the filter adds, (2) why that mechanism should improve the base signal's edge specifically. Accepted combinations: MACD+VWAP (institutional benchmark regime), RSI+volume confirmation (exhaustion vs drift), SMA+session-timing (morning institutional participation). VWAP requires TASK-0131 to be done first.
