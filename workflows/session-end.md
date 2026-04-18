# Session End — Closing Ritual

Every substantive session ends here. The two harvests capture work that would otherwise be
lost when the conversation ends.

## When to trigger

- Any other workflow says "→ session-end.md"
- Natural stopping point — the user has finished their work for this session
- The user says "let's wrap up," "commit," "I'm done"

## Flow

### Step 1 — Task harvest

Invoke `/task-manager` in harvest mode — it scans the conversation for implicit tasks: things
mentioned as "we should also," "TODO," "later we'll need to," "we're hardcoding this for now,"
edge cases discussed but not implemented, tests mentioned but not written.

The task-manager presents discovered tasks. User confirms which to create, modifies, or rejects.

Also prompt for status updates on in-progress tasks: "You had TASK-0042 in progress — any
update? Still going, done, or blocked?"

### Step 2 — Decision harvest

Invoke `/decision-journal` in harvest mode — it scans for inline decision marks
(`**Decision (version) — category: status**` blocks) from Marcus or Priya.

The journal presents discovered marks. User confirms which to write as decision files.

If no marks are found, the journal says so. That's fine — not every session produces decisions.

### Step 3 — Commit

Generate a commit message from the session's changes. The user can ask for this or do it
themselves.

## Notes

- If the user skips Step 1 or Step 2, the work is lost. Nudge once if marks or implicit tasks
  are present. If the user declines after the nudge, respect it.
- Not every session needs both harvests. A pure backlog grooming session has no decisions to
  harvest. A quick bug fix might have no implicit tasks. Run both; accept empty results.
- If the session produced no code changes (pure strategy evaluation, pure planning), skip
  the commit step.
