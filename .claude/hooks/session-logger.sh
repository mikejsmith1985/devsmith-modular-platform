#!/bin/bash
# Session Logger Hook
# Logs all significant actions to a persistent markdown file
# This provides a recovery trail in case of Claude Code crashes

set -euo pipefail

# Configuration
LOG_DIR=".claude/recovery-logs"
DATE=$(date +%Y%m%d)
LOG_FILE="$LOG_DIR/session-$DATE.md"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Create log directory if it doesn't exist
mkdir -p "$LOG_DIR"

# Initialize log file if it doesn't exist
if [ ! -f "$LOG_FILE" ]; then
    cat > "$LOG_FILE" <<EOF
# Claude Code Session Log - $DATE
Branch: $CURRENT_BRANCH
Started: $(date +"%Y-%m-%d %H:%M:%S")

---

EOF
fi

# Function to log an action
log_action() {
    local action_type="$1"
    local description="$2"
    local details="${3:-}"

    cat >> "$LOG_FILE" <<EOF
## $(date +"%H:%M:%S") - $action_type

**Description**: $description
**Branch**: $CURRENT_BRANCH
**Working Directory**: $(pwd)

EOF

    if [ -n "$details" ]; then
        cat >> "$LOG_FILE" <<EOF
**Details**:
\`\`\`
$details
\`\`\`

EOF
    fi

    cat >> "$LOG_FILE" <<EOF
---

EOF
}

# Export the log_action function so it can be sourced
export -f log_action

# If called with arguments, log them directly
if [ $# -gt 0 ]; then
    log_action "$@"
fi

# Cleanup: Keep only last 7 days of logs
find "$LOG_DIR" -name "session-*.md" -mtime +7 -delete 2>/dev/null || true

echo "Session logged to: $LOG_FILE"
