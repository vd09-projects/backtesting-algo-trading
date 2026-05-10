# backtesting-algo-trading

## Local ignored-file tracking

Ignored files are not committed. Use `make` targets to track what changed locally. Tracked paths are discovered dynamically from `.gitignore` — no updates needed when new ignore entries are added.

No snapshots stored. Uses git commit timestamps + `find -newer` — works retroactively on any commit.

### Examples

```bash
# Ignored files added/modified since last commit (verify task output)
make ignored-new

# Ignored files modified during a specific commit (since its parent up to that commit)
make ignored-since COMMIT=624a926

# Reset the "new since" baseline to right now
make reset-ref
```

### Setup (one-time)

The post-commit hook is in `.git/hooks/post-commit` (not committed — set up per clone):

```bash
cat > .git/hooks/post-commit << 'EOF'
#!/bin/sh
mkdir -p runs/temp
touch runs/temp/last-commit-ref
git log -1 --pretty=format:"%h %s" > runs/temp/last-commit-msg
EOF
chmod +x .git/hooks/post-commit
```
