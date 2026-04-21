---
name: Workflow session-end must execute autonomously in one turn
description: Session-end steps (task harvest, decision harvest, commit message, session summary) must all run without stopping for user input between them
type: feedback
---

Session-end sub-steps — task harvest, decision harvest, commit message generation, session summary — must all execute in one continuous turn. Do not pause between them or surface intermediate state that implicitly invites the user to say "continue."

**Why:** The build workflow (`workflows/INDEX.md`) explicitly says "Execute autonomously from trigger to Session Summary. Do not stop to ask permission between steps." The user's role is to read the final Summary and tune — not to approve each step. When session-end required three separate user prompts ("Continue from where you left off", "continue"), it violated this contract.

**How to apply:** When the quality gate passes and acceptance criteria are verified, the next single turn must: invoke `/task-manager` (harvest mode), invoke `/decision-journal` (harvest mode), generate the commit message, and output the Session Summary block — all without yielding to the user between any of them. The only reason to stop mid-session-end is a Hard STOP condition (requirements gap, new methodology call, unresolvable blocker). "I finished the decision harvest" is not a stopping point.
