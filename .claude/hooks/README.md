# Claude Code Recovery Hooks

This directory contains crash recovery mechanisms to prevent work loss when Claude Code experiences V8 engine crashes.

## Overview

These hooks provide a multi-layered recovery system:

1. **Session Logging**: Tracks all actions to markdown logs
2. **Git-Based Recovery**: Auto-commits to recovery branches
3. **Recovery Helper**: Interactive tool to restore work after crashes

## Files

### `session-logger.sh`
Logs all significant Claude Code actions to daily markdown files in `.claude/recovery-logs/`.

**What it logs**:
- Timestamp of each action
- Action type (Edit, Write, Bash, etc.)
- Description and details
- Current branch and working directory

**Log retention**: 7 days

### `git-recovery.sh`
Automatically commits all changes to a daily recovery branch (`claude-recovery-YYYYMMDD`).

**How it works**:
- Detects changes in the working directory
- Creates/updates recovery branch for today
- Commits with detailed context message
- Returns to original branch

**Branch retention**: 7 days

### `recovery-helper.sh`
Interactive script to recover work after a crash.

**Commands**:
```bash
# Check recovery status
.claude/hooks/recovery-helper.sh status

# View recent session logs
.claude/hooks/recovery-helper.sh list

# Restore work from recovery branch
.claude/hooks/recovery-helper.sh restore
```

## Usage

### Manual Invocation (During Claude Session)

You can call these hooks manually during a Claude Code session:

```bash
# Log current action
.claude/hooks/session-logger.sh "Edit" "Updated ARCHITECTURE.md Section 5" "Details here"

# Create recovery commit
.claude/hooks/git-recovery.sh
```

### After a Crash

1. **Check recovery status**:
   ```bash
   .claude/hooks/recovery-helper.sh status
   ```
   This shows:
   - Available recovery branches
   - Recent session logs
   - Todo list status

2. **Review session logs** to see what was in progress:
   ```bash
   .claude/hooks/recovery-helper.sh list
   ```
   Select a log file to view its contents.

3. **Restore work** from recovery branch:
   ```bash
   .claude/hooks/recovery-helper.sh restore
   ```
   This will:
   - Show available recovery branches
   - Let you select which one to restore
   - Cherry-pick commits to your current branch

4. **Check todo list** to see what tasks were pending:
   ```bash
   cat .claude/todos.json
   ```

### Automatic Invocation (Future Enhancement)

These hooks can be configured to run automatically using Claude Code's hook system (when supported):

**`.claude/settings.local.json`**:
```json
{
  "hooks": {
    "post-tool-use": ".claude/hooks/session-logger.sh",
    "post-edit": ".claude/hooks/git-recovery.sh"
  }
}
```

## Recovery Workflow Example

**Scenario**: Claude Code crashes while updating ARCHITECTURE.md

1. **Restart Claude Code**

2. **Check what was lost**:
   ```bash
   .claude/hooks/recovery-helper.sh status
   ```

   Output:
   ```
   ================================
   Recovery Status Check
   ================================

   Current Branch: feature/initial-setup

   Available Recovery Branches:
   ✓ claude-recovery-20251018
      Last commit: 2 minutes ago - Claude auto-recovery checkpoint

   Recent Session Logs:
   ✓ session-20251018.md (143 lines)

   Todo List Status:
   ✓ Todo list found (.claude/todos.json)
      In Progress: 1 | Pending: 2 | Completed: 3
   ```

3. **Review session log**:
   ```bash
   .claude/hooks/recovery-helper.sh list
   ```

   Select `session-20251018.md` to see:
   ```markdown
   ## 18:05:32 - Edit
   **Description**: Updating ARCHITECTURE.md Section 5 for OpenHands
   **Branch**: feature/initial-setup
   ```

4. **Restore work**:
   ```bash
   .claude/hooks/recovery-helper.sh restore
   ```

   Select recovery branch and confirm:
   ```
   Commits to be applied:
   a1b2c3d Claude auto-recovery checkpoint

   Proceed with recovery? (y/n) y
   ✓ Recovery successful!
   ```

5. **Resume work**: Check `.claude/todos.json` to see what task was in progress

## Environment Variables

The hooks support these environment variables for context:

- `CLAUDE_TASK_CONTEXT`: Description of current task (used in recovery commits)
- `CLAUDE_TOOL_NAME`: Name of the tool being used
- `CLAUDE_TOOL_DESCRIPTION`: Description of the tool action

These can be set by Claude Code if supported, or manually:

```bash
export CLAUDE_TASK_CONTEXT="Updating ARCHITECTURE.md with OpenHands hybrid workflow"
.claude/hooks/git-recovery.sh
```

## Integration with Todo List

The `.claude/todos.json` file acts as a recovery log itself:

```json
[
  {
    "content": "Update ARCHITECTURE.md Section 5",
    "activeForm": "Updating ARCHITECTURE.md Section 5",
    "status": "in_progress"
  },
  {
    "content": "Update ARCHITECTURE.md Section 14",
    "status": "pending"
  }
]
```

After a crash, this shows exactly what was being worked on.

## Cleanup

Both hooks automatically clean up old data:

- **Session logs**: Deleted after 7 days
- **Recovery branches**: Deleted after 7 days

Manual cleanup:
```bash
# Delete all recovery branches
git branch | grep 'claude-recovery-' | xargs git branch -D

# Delete all session logs
rm -rf .claude/recovery-logs/
```

## Troubleshooting

### "Not in a git repository"
The git-recovery hook only works in git repositories. This is expected behavior for non-git projects.

### Recovery branch conflicts
If cherry-picking fails with conflicts:
```bash
# Resolve conflicts manually, then:
git cherry-pick --continue

# Or abort:
git cherry-pick --abort
```

### No changes to commit
If git-recovery reports "No changes to commit", it means there are no uncommitted changes to save. This is normal.

## Future Enhancements

1. **Claude Code native integration**: Hooks triggered automatically on tool use
2. **Conversation export**: Save full conversation history (requires Claude Code API)
3. **Auto-resume**: Detect crash and offer to resume last session
4. **Cloud backup**: Sync recovery data to cloud storage

## Credits

Created as part of the DevSmith Modular Platform V8 crash mitigation strategy.

**Related Documents**:
- `ARCHITECTURE.md` - Section 5 (Development Tools)
- `DevSmithRoles.md` - Agent roles and responsibilities
