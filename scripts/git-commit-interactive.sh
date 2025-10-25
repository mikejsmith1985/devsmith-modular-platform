#!/bin/bash
# Interactive Git Commit Script
# Provides terminal-based confirmation for bypassing pre-commit hooks
# Usage: ./scripts/git-commit-interactive.sh "commit message"

set -e

if [ $# -eq 0 ]; then
    echo "Usage: ./scripts/git-commit-interactive.sh \"commit message\""
    exit 1
fi

COMMIT_MSG="$@"

# Run pre-commit validation
echo "🔍 Running pre-commit validation..."
if .git/hooks/pre-commit; then
    echo ""
    echo "✅ Pre-commit validation PASSED"
    echo "Proceeding with commit..."
    git commit -m "$COMMIT_MSG"
else
    echo ""
    echo "⚠️  Pre-commit validation FAILED"
    echo ""
    read -p "Do you want to bypass pre-commit checks? (yes/NO): " confirm
    
    if [[ "$confirm" != "yes" ]]; then
        echo "Commit cancelled."
        exit 1
    fi
    
    echo ""
    read -p "This is not recommended. Type 'BYPASS' to confirm: " bypass_confirm
    
    if [[ "$bypass_confirm" != "BYPASS" ]]; then
        echo "Commit cancelled."
        exit 1
    fi
    
    echo ""
    echo "⚠️  Bypassing pre-commit checks with your explicit permission..."
    git commit --no-verify -m "$COMMIT_MSG"
fi

echo ""
echo "✅ Commit successful!"
