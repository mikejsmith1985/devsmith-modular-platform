# DevLog System

This directory contains development logs to track progress, decisions, and AI assistant activity.

## Log Types

### 1. Daily DevLog (`YYYY-MM-DD.md`)

**Purpose:** Human-written session summaries with context, decisions, and lessons learned.

**When to Create:** After each significant development session.

**Format:**
```markdown
# DevLog - Month Day, Year

## Session N: Brief Title
**Time:** Morning/Afternoon/Evening
**Participants:** Mike, Claude, Copilot, etc.
**Status:** ✅ Complete / 🚧 In Progress / ❌ Blocked

### Summary
High-level overview of what happened

### Problems Discovered
Issues encountered during the session

### Solutions Implemented
How problems were resolved

### Decisions Made
Key architectural/workflow decisions

### Action Items
What needs to happen next

### Lessons Learned
Insights for future sessions
```

**Example:** `2025-10-19.md` (667 lines covering Issues #001-#002)

---

### 2. Copilot Activity Log (`copilot-activity.md`)

**Purpose:** Automated append-only log of all AI assistant actions (Copilot, Claude, etc.)

**When Updated:** Automatically after every git commit (via post-commit hook)

**Format:**
```markdown
## YYYY-MM-DD HH:MM - Action Description
**Branch:** branch-name
**Files Changed:** N files (+X, -Y lines)
- `path/to/file1.go`
- `path/to/file2.md`

**Action:** What was done

**Commit:** `abc1234`

**Commit Message:**
```
commit message here
```

**Details:** (if provided)
```
additional context
```
---
```

**Features:**
- ✅ Automatic logging via git hook
- ✅ Captures commit context (branch, files, stats)
- ✅ Includes full commit message
- ✅ Append-only (never deletes history)
- ✅ Can be manually supplemented with `.claude/hooks/copilot-logger.sh`

---

## How Copilot Activity Logging Works

### Automatic Logging (Recommended)

Every time you make a git commit, the `post-commit` hook automatically appends an entry to `copilot-activity.md`:

1. Developer (or Copilot) makes changes
2. Developer commits: `git commit -m "feat(review): add preview mode"`
3. **Post-commit hook runs automatically**
4. Entry appended to `copilot-activity.md`
5. Developer sees: `✅ Activity logged to .docs/devlog/copilot-activity.md`

**No manual action required!**

---

### Manual Logging (Optional)

If you want to log activity without making a commit, or add extra context:

```bash
# Basic usage
.claude/hooks/copilot-logger.sh "Action description"

# With details
.claude/hooks/copilot-logger.sh "Created Issue #011" "Analytics Service specification with trend detection"
```

---

## Viewing Activity

### View Recent Activity
```bash
# Last 50 lines
tail -50 .docs/devlog/copilot-activity.md

# Last N entries (each entry ~15-20 lines)
tail -100 .docs/devlog/copilot-activity.md
```

### View Today's Activity
```bash
# Filter by today's date
grep -A 15 "$(date +%Y-%m-%d)" .docs/devlog/copilot-activity.md
```

### Search for Specific Actions
```bash
# Find all Docker-related changes
grep -A 10 "Docker" .docs/devlog/copilot-activity.md

# Find all issue creation
grep -A 10 "Issue #" .docs/devlog/copilot-activity.md
```

---

## Best Practices

### For Daily DevLogs
1. **Create one per session** (or per day if multiple short sessions)
2. **Include context:** What problem are you solving?
3. **Document decisions:** Why did you choose this approach?
4. **List action items:** What's next?
5. **Record lessons learned:** What would you do differently?

### For Copilot Activity Log
1. **Let the hook do its job** (automatic is better)
2. **Write good commit messages** (they become activity log entries)
3. **Use conventional commits** (`feat:`, `fix:`, `docs:`, etc.)
4. **Don't manually edit** copilot-activity.md (append-only)

---

## Troubleshooting

### Hook Not Running
```bash
# Check if hook exists and is executable
ls -la .git/hooks/post-commit

# Make executable
chmod +x .git/hooks/post-commit
```

### Activity Log Not Created
```bash
# Create the activity log manually
touch .docs/devlog/copilot-activity.md
```

---

## Integration with Recovery System

The activity log complements (but doesn't replace) the crash recovery system:

| Feature | Activity Log | Recovery Logs |
|---------|--------------|---------------|
| **Purpose** | Track progress | Recover from crashes |
| **When** | Every commit | During Claude sessions |
| **Retention** | Permanent | 7 days |
| **Format** | Markdown | Markdown |
| **Automated** | Yes (git hook) | Yes (session hook) |
| **Location** | `.docs/devlog/` | `.claude/recovery-logs/` |

---

## Automated Workflow Integration

### GitHub Actions: Auto-Sync and Next Issue Branch

When a PR is merged to `development`, the **Auto-Sync Workflow** automatically:

1. ✅ **Handles activity log merges** - Commits any pending `copilot-activity.md` changes
2. ✅ **Finds next issue** - Looks for the next sequential issue file (e.g., 004 → 005)
3. ✅ **Creates feature branch** - Auto-creates `feature/NNN-description` branch
4. ✅ **Posts PR comment** - Shows what's next with instructions to start working

**Workflow file:** `.github/workflows/auto-sync-next-issue.yml`

**Example:**
- Merge PR for Issue #004 (`feature/004-review-service-preview-mode`)
- Workflow automatically creates `feature/005-review-skim-mode`
- Comment posted on merged PR with next steps
- Developer just needs to: `git pull && git checkout feature/005-review-skim-mode`

**Benefits:**
- 🚀 **Zero manual work** - No need to run sync scripts
- 🔄 **Consistent workflow** - Every PR merge prepares for next issue
- 📝 **Activity log handled** - Automatic merge conflict resolution
- 👀 **Visible progress** - PR comments show what's next

**When it doesn't run:**
- PR is closed without merging
- Next issue file doesn't exist (e.g., no 006-*.md after 005)
- Non-feature branches (doesn't apply to direct `development` commits)

**Fallback: Manual Script**

If you need to start an issue manually (skip sequence, parallel work, etc.):

```bash
# Still available for edge cases
./scripts/sync-and-start-issue.sh 007 review-detailed-mode
```

---

**Created:** 2025-10-19
**Last Updated:** 2025-10-20
