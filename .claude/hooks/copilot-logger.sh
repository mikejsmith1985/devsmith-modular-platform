#!/bin/bash
#
# copilot-logger.sh - Log Copilot/AI assistant activity to copilot-activity.md
#
# Usage:
#   .claude/hooks/copilot-logger.sh "Action description" "Details (optional)"
#
# Example:
#   .claude/hooks/copilot-logger.sh "Created Issue #011" "Analytics Service specification"
#

set -euo pipefail

ACTIVITY_LOG=".docs/devlog/copilot-activity.md"

# Get parameters
ACTION_DESC="${1:-"Unnamed action"}"
DETAILS="${2:-""}"

# Get current context
TIMESTAMP=$(date "+%Y-%m-%d %H:%M")
BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
LAST_COMMIT=$(git log -1 --format='%h' 2>/dev/null || echo "none")

# Get files changed in last commit (if exists)
if [[ "$LAST_COMMIT" != "none" ]]; then
    FILES_CHANGED=$(git diff-tree --no-commit-id --name-only -r "$LAST_COMMIT" 2>/dev/null | head -10)
    NUM_FILES=$(echo "$FILES_CHANGED" | wc -l)
    STATS=$(git show --stat --format="" "$LAST_COMMIT" | tail -1)
    COMMIT_MSG=$(git log -1 --format='%s' "$LAST_COMMIT" 2>/dev/null)
else
    FILES_CHANGED="No commit yet"
    NUM_FILES=0
    STATS=""
    COMMIT_MSG=""
fi

# Create log entry
LOG_ENTRY="
## $TIMESTAMP - $ACTION_DESC
**Branch:** $BRANCH
**Files Changed:** $STATS
"

# Add file list (format as markdown list)
if [[ "$FILES_CHANGED" != "No commit yet" ]]; then
    LOG_ENTRY+="$(echo "$FILES_CHANGED" | sed 's/^/- `/' | sed 's/$/`/')"
    LOG_ENTRY+="

"
fi

# Add action description
LOG_ENTRY+="**Action:** $ACTION_DESC
"

# Add details if provided
if [[ -n "$DETAILS" ]]; then
    LOG_ENTRY+="
**Details:**
$DETAILS
"
fi

# Add commit info
if [[ "$LAST_COMMIT" != "none" ]]; then
    LOG_ENTRY+="
**Commit:** \`$LAST_COMMIT\`

**Commit Message:**
\`\`\`
$COMMIT_MSG
\`\`\`
"
fi

LOG_ENTRY+="
---
"

# Append to activity log
echo "$LOG_ENTRY" >> "$ACTIVITY_LOG"

echo "✅ Logged activity to $ACTIVITY_LOG"
