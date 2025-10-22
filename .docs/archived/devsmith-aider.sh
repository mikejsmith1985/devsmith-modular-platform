#!/bin/bash
# DevSmith Platform - Aider Autonomous Coding Agent
# Starts Aider with local Ollama model for autonomous development

set -e

echo "🚀 Starting DevSmith Aider Agent..."
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

if [ "$CURRENT_BRANCH" == "main" ] || [ "$CURRENT_BRANCH" == "master" ]; then
    echo -e "${YELLOW}⚠️  Warning: You're on $CURRENT_BRANCH branch${NC}"
    read -p "Switch to development branch? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git checkout development
    fi
fi

echo ""
echo "🦙 Checking Ollama..."
if ! systemctl is-active --quiet ollama; then
    echo "Starting Ollama..."
    sudo systemctl start ollama
    sleep 2
fi

# Test Ollama connection
if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Ollama running"

    # Check if qwen2.5-coder model is available
    if ollama list | grep -q "qwen2.5-coder:32b"; then
        echo -e "${GREEN}✓${NC} qwen2.5-coder:32b model ready"
    else
        echo -e "${YELLOW}⚠${NC}  qwen2.5-coder:32b not found, pulling now..."
        ollama pull qwen2.5-coder:32b
    fi
else
    echo -e "${RED}❌ Ollama not responding${NC}"
    exit 1
fi

echo ""
echo "🐍 Activating Python environment..."
if [ ! -d "$HOME/ollama-venv" ]; then
    echo -e "${RED}❌ Virtual environment not found${NC}"
    echo "Run: python3 -m venv ~/ollama-venv && source ~/ollama-venv/bin/activate && pip install aider-chat"
    exit 1
fi

source $HOME/ollama-venv/bin/activate

# Verify Aider is installed
if ! command -v aider &> /dev/null; then
    echo -e "${RED}❌ Aider not found${NC}"
    echo "Run: pip install aider-chat"
    exit 1
fi

echo -e "${GREEN}✓${NC} Aider ready"

echo ""
echo "========================================"
echo "✅ Ready to code!"
echo "========================================"
echo ""
echo "🤖 Starting Aider with local Ollama model..."
echo "   Model: qwen2.5-coder:32b"
echo "   Mode: Interactive (use --yes for autonomous)"
echo "   Branch: $(git branch --show-current)"
echo ""
echo "💡 Aider Commands:"
echo "   /add <file>     - Add file to chat context"
echo "   /drop <file>    - Remove file from chat context"
echo "   /ls             - List files in context"
echo "   /commit         - Commit changes"
echo "   /diff           - Show uncommitted changes"
echo "   /undo           - Undo last change"
echo "   /help           - Show all commands"
echo ""
echo "🚀 Launching..."
echo ""

# Launch Aider
# --model: Specify Ollama model
# --no-auto-commits: Manual commit control (change to --auto-commits for autonomous)
# --edit-format: Use 'whole' for complete file rewrites (safer for structured code)
aider \
    --model ollama/qwen2.5-coder:32b \
    --no-auto-commits \
    --edit-format whole
