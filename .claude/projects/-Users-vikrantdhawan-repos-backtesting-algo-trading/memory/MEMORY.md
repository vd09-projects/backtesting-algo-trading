# Memory Index

- [Run quality checks inline before Ready for review](feedback_run_checks_before_ready.md) — Priya runs golangci-lint + go test -race herself at end of build; don't wait for separate /go-quality-review
- [Follow build.md sub-agent workflow — never implement inline](feedback_follow_build_workflow.md) — Must spawn sub-agents for Steps 2-5 and write SESSION STATE; coding directly in main context violates the orchestration model
