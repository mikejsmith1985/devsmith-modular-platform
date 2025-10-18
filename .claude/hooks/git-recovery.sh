#!/bin/bash
# Git-Based Recovery Hook
# Creates automatic recovery commits to preserve work in case of crashes
# Can be restored with: git cherry-pick from recovery branch

set -euo pipefail

# Configuration
RECOVERY_BRANCH_PREFIX="claude-recovery"
DATE=$(date +%Y%m%d)
RECOVERY_BRANCH="${RECOVERY_BRANCH_PREFIX}-${DATE}"

# Get current branch (fallback to 'unknown' if not in a git repo)
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Only proceed if we're in a git repository
if [ "$CURRENT_BRANCH" = "unknown" ]; then
    echo "Not in a git repository, skipping git recovery"
    exit 0
fi

# Don't create recovery commits if we're already on a recovery branch
if [[ "$CURRENT_BRANCH" == ${RECOVERY_BRANCH_PREFIX}-* ]]; then
    echo "Already on recovery branch, skipping"
    exit 0
fi

# Check if there are any changes to commit
if git diff --quiet && git diff --cached --quiet; then
    echo "No changes to commit for recovery"
    exit 0
fi

# Save current branch for return
ORIGINAL_BRANCH="$CURRENT_BRANCH"

# Create or checkout recovery branch (based on current branch)
if git show-ref --verify --quiet "refs/heads/$RECOVERY_BRANCH"; then
    # Recovery branch exists, checkout and merge current changes
    git checkout "$RECOVERY_BRANCH" 2>/dev/null
    git merge --no-edit "$ORIGINAL_BRANCH" 2>/dev/null || true
else
    # Create new recovery branch from current branch
    git checkout -b "$RECOVERY_BRANCH" 2>/dev/null
fi

# Stage all changes (including untracked files)
git add -A

# Create recovery commit with detailed message
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
TASK_CONTEXT="${CLAUDE_TASK_CONTEXT:-No task context available}"

git commit -m "Claude auto-recovery checkpoint

Timestamp: $TIMESTAMP
Original Branch: $ORIGINAL_BRANCH
Working Directory: $(pwd)
Task Context: $TASK_CONTEXT

This is an automatic recovery commit created by Claude Code's
git-recovery hook. In case of a crash, you can restore this work:

  git checkout $ORIGINAL_BRANCH
  git cherry-pick $RECOVERY_BRANCH

Recovery branch will be cleaned up after 7 days.
" --no-verify 2>/dev/null || {
    echo "No changes to commit"
}

# Return to original branch
git checkout "$ORIGINAL_BRANCH" 2>/dev/null

# Cleanup: Delete recovery branches older than 7 days
git for-each-ref --format="%(refname:short) %(committerdate:unix)" refs/heads/${RECOVERY_BRANCH_PREFIX}-* | \
while read branch timestamp; do
    current_time=$(date +%s)
    age=$((current_time - timestamp))
    # 7 days = 604800 seconds
    if [ $age -gt 604800 ]; then
        echo "Deleting old recovery branch: $branch ($(($age / 86400)) days old)"
        git branch -D "$branch" 2>/dev/null || true
    fi
done

echo "Recovery commit created on branch: $RECOVERY_BRANCH"
