# Git Hooks

This directory contains Git hooks used by the DevSmith Modular Platform.

## Installation

To install these hooks, copy them to `.git/hooks/`:

```bash
cp .claude/hooks/post-commit .git/hooks/post-commit
chmod +x .git/hooks/post-commit
```

## Available Hooks

### `post-commit`

Automatically logs commit activity to `.docs/devlog/copilot-activity.md`.

**Behavior:**
- **On `development` or `main` branches:** Logs every commit to the devlog
- **On feature branches:** Skips logging (keeps branches clean)
- Activity gets logged when PR merges to `development`

**Why this approach?**
- Prevents perpetual "modified file" in feature branches
- Keeps feature branches clean and focused
- Standardized commit messages already provide tracking
- Activity log is centralized metadata, not feature code

**Format:**
Each log entry includes:
- Timestamp
- Branch name
- Commit hash
- Commit message
- Files changed
- Stats

## Devlog Strategy

### Previous Approach (Removed)
- ❌ Devlog updated on every commit in every branch
- ❌ Always left `.docs/devlog/copilot-activity.md` modified
- ❌ Created noise in feature branches

### Current Approach (Implemented)
- ✅ Devlog only updated on `development`/`main` branches
- ✅ Feature branches stay clean
- ✅ Activity tracking happens at PR merge time
- ✅ No perpetual modified files

## Troubleshooting

**Hook not running?**
```bash
# Check if hook is executable
ls -l .git/hooks/post-commit

# Make executable if needed
chmod +x .git/hooks/post-commit
```

**Hook running but not logging?**
- Check you're on `development` or `main` branch: `git branch --show-current`
- Check `.docs/devlog/copilot-activity.md` exists
- Check for errors in hook output (shown after commit)

---

**Last Updated:** 2025-10-20
**Devlog Strategy:** Option C with Enhancement (feature branches clean)
