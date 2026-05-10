# backtesting-algo-trading

## Local ignored-file tracking

Ignored files are not committed. Use `make` targets to track what changed locally. Tracked paths are discovered dynamically from `.gitignore` — no updates needed when new ignore entries are added.

After every `git commit`, a snapshot of all ignored files is saved to `runs/temp/snapshots/<hash>.txt` automatically via the post-commit hook.

### Examples

```bash
# See ignored files added/modified since last commit (verify task output)
make ignored-new

# See all ignored files that existed at a specific commit
make ignored-at COMMIT=624a926

# See what ignored files were added or removed between two commits
make ignored-diff FROM=3a3012b TO=624a926

# List recent commits that have snapshots
make ignored-recent

# Reset the "new since" baseline to right now
make reset-ref
```

### Setup (one-time)

The post-commit hook is in `.git/hooks/post-commit` (not committed — set up per clone):

```bash
cat > .git/hooks/post-commit << 'EOF'
#!/bin/sh
mkdir -p runs/temp/snapshots
HASH=$(git log -1 --pretty=format:"%h")
MSG=$(git log -1 --pretty=format:"%s")
touch runs/temp/last-commit-ref
echo "$HASH $MSG" > runs/temp/last-commit-msg
git status --ignored --short 2>/dev/null | awk '/^!! / {print $2}' | while read path; do
  find "$path" -not -path "runs/temp/*" -type f 2>/dev/null
done | sort > "runs/temp/snapshots/${HASH}.txt"
echo "[$HASH] $MSG" >> runs/temp/snapshots/index.log
EOF
chmod +x .git/hooks/post-commit
```
