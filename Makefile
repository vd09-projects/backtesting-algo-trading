REF_FILE    := runs/temp/last-commit-ref
MSG_FILE    := runs/temp/last-commit-msg
SNAPSHOTS   := runs/temp/snapshots
COMMIT      ?= $(shell git log -1 --pretty=format:"%h")

.PHONY: ignored-new ignored-at ignored-recent ignored-diff reset-ref

## Show ignored files added/modified since last git commit
ignored-new: $(REF_FILE)
	@echo "=== Ignored files new/modified since last commit ==="
	@if [ -f $(MSG_FILE) ]; then echo "Last commit: $$(cat $(MSG_FILE))"; echo ""; fi
	@find .cache/ .quality-gate/ results/ runs/ -newer $(REF_FILE) \
		-not -path "runs/temp/*" -type f 2>/dev/null

## Show ignored files snapshot at a specific commit  (usage: make ignored-at COMMIT=abc123)
ignored-at:
	@if [ ! -f "$(SNAPSHOTS)/$(COMMIT).txt" ]; then \
		echo "No snapshot for commit $(COMMIT). Snapshots only recorded from post-commit hook going forward."; \
		exit 1; \
	fi
	@echo "=== Ignored files present at commit $(COMMIT) ==="
	@cat "$(SNAPSHOTS)/$(COMMIT).txt"

## Show ignored files added between two commits  (usage: make ignored-diff FROM=abc123 TO=def456)
ignored-diff:
	@if [ -z "$(FROM)" ] || [ -z "$(TO)" ]; then \
		echo "Usage: make ignored-diff FROM=<commit> TO=<commit>"; exit 1; fi
	@echo "=== Files in $(TO) not in $(FROM) (new) ==="
	@comm -13 \
		<(sort "$(SNAPSHOTS)/$(FROM).txt" 2>/dev/null) \
		<(sort "$(SNAPSHOTS)/$(TO).txt" 2>/dev/null)
	@echo ""
	@echo "=== Files in $(FROM) not in $(TO) (removed) ==="
	@comm -23 \
		<(sort "$(SNAPSHOTS)/$(FROM).txt" 2>/dev/null) \
		<(sort "$(SNAPSHOTS)/$(TO).txt" 2>/dev/null)

## Show commit history with snapshot availability
ignored-recent:
	@echo "=== Recent commits with ignored-file snapshots ==="
	@if [ -f "$(SNAPSHOTS)/index.log" ]; then tail -10 $(SNAPSHOTS)/index.log; \
	else echo "No snapshots yet — commit something first."; fi

## Reset the 'new since' baseline to now
reset-ref:
	mkdir -p runs/temp
	touch $(REF_FILE)
	@echo "Ref reset to now."

$(REF_FILE):
	mkdir -p runs/temp/snapshots
	touch $(REF_FILE)
