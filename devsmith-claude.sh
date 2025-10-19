#!/bin/bash
# DevSmith Platform - Claude Code Startup Script
# Starts Claude Code CLI in the correct directory

set -e

echo "🤖 Starting Claude Code..."
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Change to project directory
if [ ! -d "$HOME/projects/DevSmith-Modular-Platform" ]; then
    echo -e "${RED}❌ Project directory not found${NC}"
    echo "Expected: $HOME/projects/DevSmith-Modular-Platform"
    exit 1
fi

cd $HOME/projects/DevSmith-Modular-Platform

echo "📂 Working directory: $(pwd)"
echo ""

# Check git branch
CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || echo "none")
echo "🌿 Current branch: $CURRENT_BRANCH"
echo ""

echo "========================================"
echo "✅ Ready for Claude Code!"
echo "========================================"
echo ""
echo "💡 Claude Code Role (from DevSmithRoles.md):"
echo "   - Design architecture and create implementation specs"
echo "   - Review PRs using mental models"
echo "   - Root cause analysis for complex bugs"
echo "   - DO NOT implement features (that's Aider's job)"
echo ""
echo "🚀 Launching Claude Code..."
echo ""

# Launch Claude Code CLI
# The npx command you can never remember!
npx @anthropic-ai/claude-code@latest
