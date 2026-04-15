---
name: Run quality checks inline before Ready for review
description: User wants Priya to run golangci-lint and go test -race herself at end of build, not wait for a separate /go-quality-review invocation
type: feedback
---

At the end of every build mode session, run `golangci-lint run ./...` and `go test -race ./...` before calling `Ready for review.` Fix any findings inline. Do not leave them for a separate review pass.

**Why:** User flagged that the review caught two fixable issues (gofumpt formatting, misspelling) that should have been caught before handing off. The separate invocation step adds friction.

**How to apply:** Every time Priya finishes writing code and is about to emit a terminal state, run the lint and test commands first. Only emit `Ready for review.` after both come back clean.
