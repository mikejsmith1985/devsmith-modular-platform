#!/bin/bash
#
# sync-and-start-issue.sh - Sync development and start new issue branch
#
# Usage:
#   ./scripts/sync-and-start-issue.sh 005 "review-skim-mode"
#   ./scripts/sync-and-start-issue.sh 006 "review-scan-mode"
#

set -euo pipefail

ISSUE_NUM="${1:-}"
BRANCH_DESC="${2:-}"

if [[ -z "$ISSUE_NUM" || -z "$BRANCH_DESC" ]]; then
    echo "Usage: $0 <issue-number> <branch-description>"
    echo ""
    echo "Examples:"
    echo "  $0 005 review-skim-mode"
    echo "  $0 011 analytics-service-foundation"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔄 Syncing Development & Starting Issue #${ISSUE_NUM}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Step 1: Check if on development
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "development" ]]; then
    echo "⚠️  You're on branch '$CURRENT_BRANCH', not 'development'"
    echo "   Switching to development..."
    git checkout development
fi

# Step 2: Stash any uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
    echo "📦 Stashing uncommitted changes..."
    git stash push -m "Auto-stash before sync (Issue #${ISSUE_NUM})"
    STASHED=true
else
    STASHED=false
fi

# Step 3: Pull latest changes
echo "⬇️  Pulling latest development..."
git pull origin development

# Step 4: Apply stashed changes if any
if [[ "$STASHED" == "true" ]]; then
    echo "📤 Applying stashed changes..."
    if git stash pop; then
        echo "✅ Stashed changes applied successfully"

        # If copilot-activity.md was modified, commit it
        if git status --porcelain | grep -q "copilot-activity.md"; then
            echo "📝 Committing merged activity log..."
            git add .docs/devlog/copilot-activity.md
            git commit -m "chore: merge activity log after sync"
        fi
    else
        echo "⚠️  Merge conflict in stashed changes!"
        echo "   Resolve conflicts manually, then run:"
        echo "   git checkout -b feature/${ISSUE_NUM}-${BRANCH_DESC}"
        exit 1
    fi
fi

# Step 5: Create feature branch
BRANCH_NAME="feature/${ISSUE_NUM}-${BRANCH_DESC}"
echo ""
echo "🌿 Creating feature branch: $BRANCH_NAME"
git checkout -b "$BRANCH_NAME"

# Step 6: Verify issue file exists
ISSUE_FILE=".docs/issues/${ISSUE_NUM}-*.md"
if ! ls $ISSUE_FILE 1> /dev/null 2>&1; then
    echo ""
    echo "⚠️  Warning: Issue file not found for #${ISSUE_NUM}"
    echo "   Expected: .docs/issues/${ISSUE_NUM}-*.md"
    echo ""
    echo "   Continue anyway? Branch created: $BRANCH_NAME"
else
    FOUND_FILE=$(ls $ISSUE_FILE 2>/dev/null | head -1)
    echo ""
    echo "📋 Issue file found: $FOUND_FILE"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Ready to start Issue #${ISSUE_NUM}!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Current branch: $BRANCH_NAME"
echo "Base: development (up to date)"
echo ""
if ls $ISSUE_FILE 1> /dev/null 2>&1; then
    echo "Next steps:"
    echo "  1. Read issue spec: cat $FOUND_FILE"
    echo "  2. Implement the feature"
    echo "  3. Commit and push"
    echo "  4. PR will be auto-created by GitHub Actions!"
else
    echo "⚠️  Create issue spec first or check issue number"
fi
echo ""
