#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}📦 Installing Git hooks...${NC}"
echo ""

# Get the root directory of the git repo
GIT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
if [ -z "$GIT_ROOT" ]; then
    echo -e "${YELLOW}⚠️  Not in a git repository${NC}"
    exit 1
fi

HOOKS_SRC="$GIT_ROOT/scripts/hooks"
HOOKS_DEST="$GIT_ROOT/.git/hooks"

# Check if source hooks exist
if [ ! -d "$HOOKS_SRC" ]; then
    echo -e "${YELLOW}⚠️  Hooks directory not found: $HOOKS_SRC${NC}"
    exit 1
fi

# Install pre-commit hook
if [ -f "$HOOKS_SRC/pre-commit" ]; then
    echo -e "${BLUE}→ Installing pre-commit hook...${NC}"
    cp "$HOOKS_SRC/pre-commit" "$HOOKS_DEST/pre-commit"
    chmod +x "$HOOKS_DEST/pre-commit"
    echo -e "${GREEN}  ✓ Pre-commit hook installed${NC}"
else
    echo -e "${YELLOW}  ⚠️  pre-commit hook not found in $HOOKS_SRC${NC}"
fi

# Install local config example
if [ -f "$HOOKS_SRC/pre-commit-local.yaml.example" ]; then
    echo -e "${BLUE}→ Installing local config example...${NC}"
    cp "$HOOKS_SRC/pre-commit-local.yaml.example" "$HOOKS_DEST/pre-commit-local.yaml.example"
    echo -e "${GREEN}  ✓ Local config example installed${NC}"
    echo -e "${BLUE}  ℹ️  Copy to pre-commit-local.yaml to customize: cp $HOOKS_DEST/pre-commit-local.yaml.example $HOOKS_DEST/pre-commit-local.yaml${NC}"
else
    echo -e "${YELLOW}  ⚠️  Local config example not found${NC}"
fi

echo ""
echo -e "${GREEN}✅ Git hooks installation complete!${NC}"
echo ""
echo -e "${BLUE}📖 Documentation:${NC}"
echo -e "   Pre-commit guide: .docs/PRE-COMMIT-ENHANCEMENTS.md"
echo ""
echo -e "${BLUE}🧪 Test the hook:${NC}"
echo -e "   .git/hooks/pre-commit --quick      # Fast validation"
echo -e "   .git/hooks/pre-commit --standard   # Full validation"
echo ""
echo -e "${BLUE}⚙️  Configuration:${NC}"
echo -e "   Team config:  .pre-commit-config.yaml (committed)"
echo -e "   Local config: .git/hooks/pre-commit-local.yaml (not committed)"
echo ""
