---
name: 5-min is primary strategy timeframe
description: All new strategy development targets 5-min bars; daily is secondary; 1-min pending verification
type: feedback
---

5-min bars are the primary development timeframe for new strategies. Daily-bar strategies already live (MACD portfolio) stay running but no new daily strategies are built.

**Why:** 5-min gives 75× more evaluation points per year (18,900 vs 252), statistical tests have real power. Existing daily strategies are all indicator-crossover (same edge bucket) and most failed the universe gate. NSE 5-min data confirmed 5+ years available.

**How to apply:** When evaluating or designing a new strategy, default to 5-min bars. If Marcus or the user proposes a daily-bar strategy, flag that we're focusing on 5-min and ask whether it has a 5-min adaptation. 1-min: don't commit until TASK-0129 empirical verification of Kite historical depth completes.
