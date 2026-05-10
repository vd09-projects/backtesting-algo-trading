REF_FILE    := runs/temp/last-commit-ref
MSG_FILE    := runs/temp/last-commit-msg
COMMIT      ?= $(shell git log -1 --pretty=format:"%h")

.PHONY: ignored-new ignored-since ignored-recent reset-ref

## Show ignored files added/modified since last git commit
ignored-new: $(REF_FILE)
	@echo "=== Ignored files new/modified since last commit ==="
	@if [ -f $(MSG_FILE) ]; then echo "Last commit: $$(cat $(MSG_FILE))"; echo ""; fi
	@git status --ignored --short 2>/dev/null | awk '/^!! / {print $$2}' | while read path; do \
		find "$$path" -newer $(REF_FILE) -not -path "runs/temp/*" -type f 2>/dev/null; \
	done

## Show ignored files modified since a specific commit's parent up to that commit  (usage: make ignored-since COMMIT=abc123)
ignored-since:
	@THIS_T=$$(git log -1 --format="%ct" $(COMMIT) 2>/dev/null); \
	PARENT_T=$$(git log -1 --format="%ct" $(COMMIT)^ 2>/dev/null); \
	if [ -z "$$THIS_T" ]; then echo "Commit $(COMMIT) not found."; exit 1; fi; \
	THIS_REF=$$(mktemp); PARENT_REF=$$(mktemp); \
	touch -t $$(date -r $$THIS_T +"%Y%m%d%H%M.%S") "$$THIS_REF"; \
	if [ -n "$$PARENT_T" ]; then \
		touch -t $$(date -r $$PARENT_T +"%Y%m%d%H%M.%S") "$$PARENT_REF"; \
	else \
		touch -t 197001010000 "$$PARENT_REF"; \
	fi; \
	echo "=== Ignored files modified in commit $(COMMIT) ==="; \
	git status --ignored --short 2>/dev/null | awk '/^!! / {print $$2}' | while read path; do \
		find "$$path" -newer "$$PARENT_REF" ! -newer "$$THIS_REF" -not -path "runs/temp/*" -type f 2>/dev/null; \
	done; \
	rm -f "$$THIS_REF" "$$PARENT_REF"

## Reset the 'new since' baseline to now
reset-ref:
	mkdir -p runs/temp
	touch $(REF_FILE)
	@echo "Ref reset to now."

$(REF_FILE):
	mkdir -p runs/temp
	touch $(REF_FILE)
