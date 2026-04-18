# Strategy Evaluation

User has a new strategy idea and wants Marcus's assessment before committing to building anything.
Pure methodology — no code yet. (~10% of sessions.)

## Trigger

User describes a strategy idea, asks "is this worth building," shares an edge thesis, or asks
about a new market/instrument. Usually sounds like: "I want to try momentum on BTC perps,"
"I have an idea for a mean-reversion strategy," "what do you think about this edge."

## Flow

### Step 1 — Marcus interrogates

Invoke `/algo-trading-veteran` — share the user's idea. Marcus runs his 5-minute interrogation:
edge thesis in one sentence, instrument/timeframe/capital/capacity, stage, data.

This is often multi-turn. Marcus asks, user answers, Marcus digs deeper. Let this run — the
interrogation IS the value. Don't rush to the verdict.

Marcus will typically deliver:
- Edge thesis assessment (which of the five buckets, why it might work, why it might not)
- Test plan (which data, which methodology, which stress periods)
- Sizing recommendation (if the idea has legs)
- Kill-switch line (pre-committed halt conditions)
- Verdict: go / iterate / kill

He'll mark specific calls as `algorithm` decisions inline (sizing rules, kill-switch lines,
feature verdicts — NOT general principles).

### Step 2 — Handle the verdict

**If "go" with sizing and kill-switch marked:**
The idea passed Marcus's filter. Two options:

Option A — create implementation tasks now. Invoke `/task-manager` to create tasks for building
the backtest, the engine support, the data pipeline. User confirms which tasks to create.

Option B — start planning with Priya immediately. Invoke `/algo-trading-lead-dev` to plan the
implementation. This works if the user wants to start building in the same session.

Suggest both options. Let the user pick. If they pick B, this effectively transitions into
`build.md` starting at Step 2.

**If "iterate":**
Stay with Marcus. The user refines the idea — different instrument, different timeframe,
different edge thesis. Marcus re-evaluates. This can loop several times. When Marcus finally
gives a go or kill, handle that verdict.

Sometimes the iterate loop surfaces a question Priya could answer — "is it feasible to
build a 1-minute-bar engine?" or "how hard is it to add funding rate handling?" If so, briefly
invoke `/algo-trading-lead-dev` for a feasibility check, then return to Marcus with the answer.

**If "kill":**
The idea didn't pass. Marcus explains why. The rejection itself is valuable — it prevents
wasting a month on something that won't work. Proceed to session-end to harvest the rejection
as an `algorithm: rejected` decision. This preserves the reasoning so nobody re-proposes the
same idea in 3 months without knowing it was already evaluated.

### Step 3 — Session end

Go to `session-end.md`.

Even for "kill" verdicts — the decision marks should be harvested so the journal captures
the rejection and its reasoning.
