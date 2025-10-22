# Aider Setup Implementation Spec

**Created:** 2025-10-19
**Issue:** #TBD (Infrastructure setup)
**Estimated Complexity:** Low
**Target Service:** Infrastructure / Development Environment

---

## Overview

### Feature Description
Replace OpenHands with Aider as the autonomous implementation agent, configured to work with local Ollama models (no API costs).

### User Story
As Mike (Project Orchestrator), I want a working autonomous coding agent that uses local LLMs so that I can develop features overnight without API costs or configuration headaches.

### Success Criteria
- [ ] Aider installed and configured with Ollama integration
- [ ] Successfully completes first test feature (Portal User model creation)
- [ ] Documentation updated to reflect Aider workflow instead of OpenHands
- [ ] Startup scripts updated for Aider usage

---

## Context for Cognitive Load Management

### Bounded Context
**Service:** Development Infrastructure
**Domain:** AI-Assisted Development Tooling
**Related Entities:**
- Aider (autonomous coding agent)
- Ollama (local LLM provider)
- qwen2.5-coder:32b (code generation model)

**Context Boundaries:**
- ✅ **Within scope:** Aider setup, Ollama integration, model configuration, workflow documentation
- ❌ **Out of scope:** Application code changes, service architecture

---

## Implementation Details

### 1. Install Aider

```bash
# Activate Python virtual environment
source ~/ollama-venv/bin/activate

# Install aider-chat
pip install aider-chat

# Verify installation
aider --version
```

**Expected output:** `aider X.X.X`

---

### 2. Configure Ollama Model

```bash
# Pull the best local model for Go development
ollama pull qwen2.5-coder:32b

# Verify model is available
ollama list | grep qwen
```

**Why qwen2.5-coder:32b:**
- Superior Go code generation compared to DeepSeek
- Proper struct tag formatting (backticks, spacing)
- Better understanding of Go idioms
- 32B parameter size balances quality with performance

**Alternative models (if 32B is too large):**
- `qwen2.5-coder:14b` - Faster, still good quality
- `codellama:34b` - Slightly worse at Go, but solid backup

---

### 3. Create Aider Startup Script

**File:** `~/projects/DevSmith-Modular-Platform/devsmith-aider.sh`

```bash
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
```

**Make it executable:**
```bash
chmod +x ~/projects/DevSmith-Modular-Platform/devsmith-aider.sh
```

---

### 4. Test Aider with First Feature

**Create test branch:**
```bash
cd ~/projects/DevSmith-Modular-Platform
git checkout development
git pull
git checkout -b feature/001-portal-user-model
```

**Run Aider:**
```bash
./devsmith-aider.sh
```

**Test prompt for Aider:**
```
Create the file internal/portal/models/user.go with a User struct that has:
- ID (uint)
- GitHubID (int64)
- Username (string)
- Email (string)
- CreatedAt (time.Time)
- UpdatedAt (time.Time)

Include proper package declaration, imports, and JSON struct tags.
Follow Go naming conventions.
```

**Expected Aider workflow:**
1. Creates the directory structure
2. Writes the file with correct syntax
3. Shows the diff
4. Asks for confirmation

**Verify the generated code:**
```bash
# Check syntax
go fmt ./internal/portal/models/user.go

# View the file
cat ./internal/portal/models/user.go
```

**Expected output:**
```go
package models

import "time"

// User represents a user in the portal system
type User struct {
	ID        uint      `json:"id"`
	GitHubID  int64     `json:"github_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

---

### 5. Update Project Documentation

**Files to update:**

#### DevSmithRoles.md

**Change:**
```markdown
### 3. Primary Implementation Agent (OpenHands + Ollama)
- **Role:** **Autonomous code generator and implementer** - executes 70-80% of development work.
```

**To:**
```markdown
### 3. Primary Implementation Agent (Aider + Ollama)
- **Role:** **Autonomous code generator and implementer** - executes 70-80% of development work.
- **Tools:**
  - Aider CLI (autonomous agent framework with excellent local LLM support)
  - Ollama with model: `qwen2.5-coder:32b`
  - Go toolchain: `go test`, `go build`, `air` (hot reload)
  - Git for version control
- **Strengths:**
  - **Excellent local LLM integration** - no API key issues like OpenHands
  - **Local execution** - no API costs, no rate limits
  - **Git-aware** - understands repo structure and history
  - **Incremental edits** - makes surgical changes, not full rewrites
  - **Superior Go support** - qwen2.5-coder generates proper Go code
- **Limitations:**
  - Smaller context window than Claude (but sufficient for single features)
  - Interactive by default (use `--yes` flag for full autonomy)
  - May need guidance on complex architectural decisions (defers to Claude)
```

**Update workflow section:**
```markdown
### Standard Feature Development (80% of work)

...

3. **Autonomous Implementation** (Aider + Ollama):
   - Mike triggers Aider: `./devsmith-aider.sh`
   - In Aider chat: Paste Claude's implementation spec or reference issue #
   - For full autonomy: Restart with `--yes` flag
   - Aider works through the spec:
     - Creates feature branch from `development`
     - Writes tests first (TDD per `DevsmithTDD.md`)
     - Implements feature following Claude's spec
     - Shows diffs for review
     - Commits with conventional commit messages
   - **Duration**: 15 minutes - 2 hours depending on feature complexity
   - **Crash-proof**: Aider stores context in `.aider` directory

4. **PR Creation** (Aider + Mike):
   - Aider commits with: `/commit` command and conventional message
   - Mike pushes and creates PR manually (for now)
   - PR includes:
     - Link to issue (`Closes #[number]`)
     - Implementation summary
     - Test results
```

---

#### README.md

**Add Aider section:**
```markdown
## Development Workflow

### Using Aider for Autonomous Coding

This project uses Aider with local Ollama models for AI-assisted development.

**Quick Start:**
```bash
# 1. Start Aider
./devsmith-aider.sh

# 2. In Aider, add files to context
/add internal/portal/**/*.go

# 3. Give instructions (or paste implementation spec)
Create the GitHub OAuth handler following the spec in issue #42

# 4. Review changes and commit
/diff
/commit
```

**Aider Configuration:**
- Model: `qwen2.5-coder:32b` (via Ollama)
- Edit format: `whole` (complete file rewrites)
- Auto-commits: Disabled (manual review)

**For fully autonomous mode:**
```bash
aider --model ollama/qwen2.5-coder:32b --yes --auto-commits
```

See `DevSmithRoles.md` for full workflow details.
```

---

#### LESSONS_LEARNED.md

**Add new section:**
```markdown
## 9. AI Agent Selection

### 9.1 OpenHands API Key Issues with Local Models

**Problem:**
- OpenHands 0.58 requires API keys even for local Ollama models
- Configuration with `OPENHANDS_API_KEY=none` fails at chat initialization
- "Error authenticating with the LLM provider" despite Ollama responding
- 6+ hours spent troubleshooting configuration with no resolution

**Lesson:**
- **Evaluate AI tools based on actual local LLM support, not marketing claims**
- Test basic functionality (chat initialization) before committing to a tool
- Some tools are cloud-first with broken local modes
- API key validation should be provider-aware (local providers don't need keys)

**Action for This Project:**
- Use Aider instead of OpenHands
- Aider designed specifically for local LLMs
- No API key requirements for Ollama integration
- Better Go code generation with qwen2.5-coder:32b

### 9.2 Model Selection for Go Development

**Problem:**
- DeepSeek-Coder v2:16b generates malformed Go code
- Missing backticks in struct tags: `json:"id` instead of `` `json:"id"` ``
- Missing spaces: `time.Timejson:` instead of `time.Time `json:`
- Resulted in compilation errors on every attempt

**Lesson:**
- **Test model quality with language-specific code before using in production**
- Not all "code" models are equal across languages
- Larger models aren't always better (DeepSeek 16B worse than Qwen 32B for Go)

**Action for This Project:**
- Use qwen2.5-coder:32b as primary model (excellent Go support)
- Fallback to codellama:34b if needed
- Always verify generated Go code with `go fmt` and `go build`
- Never trust AI-generated code without compilation check
```

---

### 6. Create Aider Configuration File

**File:** `~/projects/DevSmith-Modular-Platform/.aider.conf.yml`

```yaml
# Aider configuration for DevSmith Platform
#
# This file configures Aider's behavior for this project

# Model configuration
model: ollama/qwen2.5-coder:32b

# Edit format - how Aider modifies files
# Options: whole, diff, udiff
# 'whole' is safest for Go code (avoids patch conflicts)
edit-format: whole

# Git integration
auto-commits: false  # Manual commit control
dirty-commits: false # Don't commit with dirty working directory

# Context management
map-tokens: 1024  # Tokens for repository map

# Show diffs before applying
show-diffs: true

# Lint and test commands (optional - runs after changes)
# lint-cmd: golangci-lint run
# test-cmd: go test ./...

# Files to always exclude from context
ignore:
  - "*.md"  # Don't include docs unless explicitly added
  - "vendor/"
  - "node_modules/"
  - ".git/"
  - "*.log"
  - "coverage.out"
```

---

## Implementation Checklist

### Phase 1: Setup Aider
- [ ] Activate Python venv: `source ~/ollama-venv/bin/activate`
- [ ] Install Aider: `pip install aider-chat`
- [ ] Verify installation: `aider --version`
- [ ] Pull Ollama model: `ollama pull qwen2.5-coder:32b`
- [ ] Verify model: `ollama list | grep qwen`

### Phase 2: Configuration
- [ ] Create startup script: `devsmith-aider.sh`
- [ ] Make executable: `chmod +x devsmith-aider.sh`
- [ ] Create Aider config: `.aider.conf.yml`
- [ ] Test startup script: `./devsmith-aider.sh`

### Phase 3: Test with First Feature
- [ ] Create test branch: `git checkout -b feature/001-portal-user-model`
- [ ] Launch Aider: `./devsmith-aider.sh`
- [ ] Give test prompt (User model creation)
- [ ] Verify generated code compiles: `go fmt ./internal/portal/models/user.go`
- [ ] Check struct tags are correct (backticks, spacing)
- [ ] Commit test: `/commit` in Aider

### Phase 4: Update Documentation
- [ ] Update DevSmithRoles.md (change OpenHands → Aider)
- [ ] Update README.md (add Aider workflow section)
- [ ] Update LESSONS_LEARNED.md (add AI agent selection lessons)
- [ ] Remove/archive OpenHands troubleshooting notes

### Phase 5: Clean Up Old Setup
- [ ] Stop OpenHands container: `docker stop openhands-app`
- [ ] Remove OpenHands container: `docker rm openhands-app`
- [ ] Archive OpenHands config: `mv ~/openhands-config ~/openhands-config.old`
- [ ] Remove Open Interpreter Go support files (no longer needed):
  - [ ] `go_language.py`
  - [ ] `devsmith-start-with-go.sh`
  - [ ] `test_go_interpreter.py`
  - [ ] `GO_SUPPORT_GUIDE.md`
  - [ ] `QUICK_START_GO.md`

### Phase 6: Commit Infrastructure Changes
- [ ] Stage all changes: `git add .`
- [ ] Commit: `git commit -m "chore(infra): replace OpenHands with Aider for local LLM support"`
- [ ] Push: `git push origin development`
- [ ] Merge to main once verified working

---

## Success Metrics

**Aider is considered successfully configured when:**

1. ✅ Startup script runs without errors
2. ✅ Aider connects to Ollama successfully
3. ✅ Test feature (User model) generates valid Go code
4. ✅ Generated code compiles: `go fmt` passes
5. ✅ Struct tags are properly formatted (backticks present, spacing correct)
6. ✅ Aider can commit changes via `/commit` command
7. ✅ Documentation updated to reflect new workflow

---

## Troubleshooting

### Issue: Aider can't connect to Ollama

**Check:**
```bash
curl http://localhost:11434/api/tags
```

**Fix:**
```bash
sudo systemctl start ollama
```

### Issue: qwen2.5-coder:32b too large for system

**Solution:** Use smaller model
```bash
ollama pull qwen2.5-coder:14b
# Update .aider.conf.yml: model: ollama/qwen2.5-coder:14b
```

### Issue: Aider generates incorrect Go code

**Check model:**
```bash
ollama list
```

**Verify you're using qwen2.5-coder, not DeepSeek:**
```bash
# In Aider
/model ollama/qwen2.5-coder:32b
```

### Issue: Aider is slow

**Options:**
1. Use smaller model (14B instead of 32B)
2. Reduce context with `/drop <file>` commands
3. Use `--edit-format diff` instead of `whole`

---

## References

- Aider documentation: https://aider.chat/docs/
- Ollama model library: https://ollama.com/library
- DevSmithRoles.md - Hybrid workflow
- LESSONS_LEARNED.md - OpenHands issues

---

**Next Steps:**
1. Execute Phase 1-3 of checklist (setup and test)
2. If test succeeds, execute Phase 4-6 (documentation and cleanup)
3. Once complete, move to Issue #002: Portal Authentication (GitHub OAuth)
