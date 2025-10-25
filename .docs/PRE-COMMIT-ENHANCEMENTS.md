# Pre-Commit Hook Enhancements v2.1

## Summary

The pre-commit hook has been significantly enhanced with production-ready quality gates while maintaining TDD workflow compatibility and respecting the 90-second performance budget.

---

## 📦 Installation

### Quick Setup

```bash
# Clone the repository and install hooks
./scripts/install-hooks.sh
```

This will:
- ✅ Install the pre-commit hook to `.git/hooks/`
- ✅ Copy local config example to `.git/hooks/`
- ✅ Make the hook executable

### Manual Setup

```bash
# Copy hook manually
cp scripts/hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Copy config example (optional)
cp scripts/hooks/pre-commit-local.yaml.example .git/hooks/pre-commit-local.yaml.example
```

### Dependencies

**Required** (will fail gracefully if missing):
- `go` (1.23+)
- `gofmt`
- `golangci-lint` - [Install guide](https://golangci-lint.run/usage/install/)

**Optional** (enhanced features):
- `govulncheck` - Security scanning: `go install golang.org/x/vuln/cmd/govulncheck@latest`
- `yq` - Config parsing (uses defaults if missing)

---

## 🎯 What's New

### 1. **Test Coverage Requirements** ✅
- **Error Threshold**: Blocks commits below 40% coverage
- **Warning Threshold**: Warns below 70% coverage
- **TDD-Aware**: Automatically skipped during RED phase
- **Cached**: Results cached for 5 minutes for performance

**Example Output**:
```bash
📊 Checking test coverage...
  ⚠️  Coverage 45% < 70% (recommended)
```

### 2. **Security Vulnerability Scanning** 🔒
- **Tool**: `govulncheck` (official Go vulnerability database)
- **Mode**: Runs in standard mode (every commit)
- **Cached**: Results cached for 24 hours
- **Offline Support**: Gracefully skips if network unavailable
- **Performance**: ~10-30s first run, <1s cached

**Installation**:
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

**Example Output**:
```bash
🔒 Checking for security vulnerabilities...
  ✓ No known vulnerabilities (cached)
```

### 3. **Enhanced Linting Rules** 📋
**New Blocking Linters** (errors that prevent commit):
- `typecheck` - Type errors
- `errcheck` - Unchecked errors
- `staticcheck` - Static analysis bugs

**New Warning Linters** (non-blocking):
- `gosec` - Security issues
- `unused` - Unused code (skipped in TDD RED)
- `ineffassign` - Ineffectual assignments
- `govet` - Suspicious constructs
- `gocritic` - Code quality
- `gocyclo` - Cyclomatic complexity
- `dupl` - Duplicate code
- `goconst` - Repeated constants

### 4. **Early Import Cycle Detection** 🔄
- Runs `go list` before build to detect cycles early
- Fails fast with clear error message
- Prevents wasted build time

**Example Output**:
```bash
🔄 Checking for import cycles...
  ✗ Import cycle detected
  → Break cycle by moving shared types to common package
```

### 5. **Conditional Race Detection** 🏁
- **Modes**: `always`, `conditional`, `never`
- **Default**: `conditional` (only runs if goroutines detected)
- **Smart Detection**: Scans for `go func`, `select {`, goroutines
- **Performance**: Only pays cost when needed

**Example Output**:
```bash
  → Detected concurrent code (goroutines/channels)
🏁 Running race detector on concurrent code...
  ✓ No race conditions detected
```

### 6. **TDD Workflow Awareness** 🔴
- **Auto-Detection**: Identifies RED phase by checking for:
  - `undefined:` errors
  - `declared and not used` errors
  - `imported and not used` errors
- **Behavior**: Runs all checks but doesn't block in RED phase
- **Skipped Checks**: Coverage and unused code (meaningless in RED)

**Example Output**:
```bash
🔴 TDD RED phase detected - checks will run but won't block
```

### 7. **Configuration System** ⚙️
**Two-Level Config**:
1. **Team Config** (`.pre-commit-config.yaml` in repo root)
   - Committed to repository
   - Applies to all developers
   - Team-wide standards

2. **Local Override** (`.git/hooks/pre-commit-local.yaml`)
   - Not committed (local only)
   - Overrides team settings
   - Individual developer preferences

**Copy Example**:
```bash
cp .git/hooks/pre-commit-local.yaml.example .git/hooks/pre-commit-local.yaml
# Edit to customize your settings
```

---

## 📊 Performance Budget

| Mode | Target Time | Actual Time | Checks Included |
|------|-------------|-------------|-----------------|
| **Quick** | <15s | ~10-15s | fmt, build |
| **Standard** | <60s | ~45-75s | fmt, vet, lint, tests, coverage, security, cycles, race (conditional) |
| **Thorough** | <90s | ~70-90s | All + race (always) |

---

## 🔧 Configuration Examples

### Team Config (`.pre-commit-config.yaml`)

```yaml
# Coverage thresholds
coverage:
  enabled: true
  error_threshold: 40    # Block below 40%
  warning_threshold: 70  # Warn below 70%
  blocking: true

# Security scanning
security:
  enabled: true
  govulncheck:
    enabled: true
    mode: "standard"
    cache_duration: 86400  # 24 hours
    allow_offline: true

# Race detection
race_detection:
  enabled: true
  mode: "conditional"    # Only if goroutines detected
  blocking: true

# TDD awareness
tdd:
  enabled: true
  detect_red_phase: true
  red_phase_behavior: "warn"  # Run but don't block
```

### Local Override Examples

**Beginner Developer** (more lenient):
```yaml
coverage:
  error_threshold: 20
  warning_threshold: 50
  blocking: false
```

**Senior Developer** (stricter):
```yaml
coverage:
  error_threshold: 60
  warning_threshold: 80
race_detection:
  mode: "always"
```

**Offline Mode**:
```yaml
security:
  enabled: false
coverage:
  enabled: false
```

**Fast Iteration**:
```yaml
default_mode: "quick"
coverage:
  enabled: false
race_detection:
  enabled: false
```

---

## 🚀 Usage

### Standard Workflow
```bash
# Normal commit (runs standard mode)
git add .
git commit -m "feat: add new feature"

# With auto-fix
git commit -m "feat: add new feature" --no-verify
.git/hooks/pre-commit --fix
git add .
git commit -m "feat: add new feature"

# Test specific mode
.git/hooks/pre-commit --quick
.git/hooks/pre-commit --standard
.git/hooks/pre-commit --thorough
```

### TDD Workflow
```bash
# RED Phase - write failing test
git add internal/service/user_test.go
git commit -m "test: add user validation test (RED)"
# ✅ Commits even with undefined references

# GREEN Phase - implement feature
git add internal/service/user.go
git commit -m "feat: implement user validation (GREEN)"
# ✅ All checks run normally

# REFACTOR Phase
git add internal/service/user.go
git commit -m "refactor: improve user validation (REFACTOR)"
# ✅ All quality checks enforced
```

### Viewing Results
```bash
# JSON output for parsing
.git/hooks/pre-commit --json | jq '.summary'

# View specific check
.git/hooks/pre-commit --check-only coverage

# Explain a test failure
.git/hooks/pre-commit --explain "TestUserValidation"

# Get fix suggestions
.git/hooks/pre-commit --suggest-fix "user_test.go:45"
```

---

## 📈 What Gets Checked

### Quick Mode
- ✅ Code formatting (gofmt)
- ✅ Critical build errors
- ⏭️ Everything else skipped

### Standard Mode (Default)
- ✅ Code formatting
- ✅ go vet analysis
- ✅ golangci-lint (blocking + warning linters)
- ✅ All tests (short mode)
- ✅ Import cycle detection
- ✅ Test coverage (40% block / 70% warn)
- ✅ Security vulnerabilities (govulncheck)
- ✅ Race conditions (if goroutines detected)

### Thorough Mode
- ✅ Everything from Standard
- ✅ Race detection (always, not conditional)
- ✅ Full test suite

---

## ⚠️ TDD Phase Behavior

### RED Phase (Test-First)
**Detected When**:
- `undefined:` errors present
- `declared and not used` warnings
- Tests failing with expected failures

**What Happens**:
- 🟢 Format checks: **RUN + BLOCK**
- 🟢 Import cycles: **RUN + BLOCK**
- 🟡 Build errors: **RUN + WARN** (expected in RED)
- 🟡 Test failures: **RUN + WARN** (expected in RED)
- 🟡 Coverage: **SKIPPED** (meaningless in RED)
- 🟡 Unused code: **SKIPPED** (expected in RED)
- 🟡 Security: **RUN + WARN**
- 🟡 Race detection: **RUN + WARN**

### GREEN/REFACTOR Phase
**All checks enforced normally**

---

## 🔍 Troubleshooting

### "govulncheck not found"
```bash
# Install it
go install golang.org/x/vuln/cmd/govulncheck@latest

# Verify
which govulncheck
```

### "yq command not found" (config loading)
```bash
# Install yq (optional - configs work without it, just uses defaults)
brew install yq  # macOS
sudo apt install yq  # Linux

# Or use defaults (works fine without yq)
```

### "Coverage check too slow"
```yaml
# In .git/hooks/pre-commit-local.yaml
coverage:
  enabled: false  # Disable locally, CI will still check
```

### "Race detection times out"
```yaml
# Increase timeout or disable
race_detection:
  timeout: 60  # Increase from default 30s
  # OR
  enabled: false
```

### "Too many warnings"
```yaml
# Adjust linting strictness
linting:
  warning_linters: []  # Disable warnings, keep only blocking
```

---

## 📊 Metrics & Validation

### Coverage Tracking
```bash
# View current coverage
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Security Scan Details
```bash
# Full vulnerability report
govulncheck ./...

# Check specific module
govulncheck -mode=source ./internal/service
```

### Race Condition Details
```bash
# Full race detection
go test -race ./...

# Specific package
go test -race ./internal/service
```

---

## 🎓 Best Practices

### 1. **Commit Frequently**
- Small commits complete faster
- Easier to identify issues
- Better git history

### 2. **Use TDD Workflow**
- Write test first (RED)
- Implement feature (GREEN)
- Improve code (REFACTOR)
- Hook respects each phase

### 3. **Fix Warnings Regularly**
- Don't let warnings accumulate
- Address security issues promptly
- Keep coverage above 70%

### 4. **Configure for Your Workflow**
- Use local overrides for personal preferences
- Don't commit local config
- Respect team standards in shared config

### 5. **Monitor Performance**
- Check hook duration in output
- Adjust timeouts if needed
- Use quick mode for rapid iteration

---

## 📚 Related Documentation

- [TDD Workflow Guide](.docs/copilot-instructions.md)
- [Mock Testing Patterns](.docs/copilot-instructions.md#51-mock-patterns)
- [Docker Validation](.docs/DOCKER-VALIDATION.md)
- [golangci-lint Config](.golangci.yml)

---

## 🚦 CI/CD Integration

The pre-commit hook is designed to catch issues locally, but CI should run the same checks:

```yaml
# .github/workflows/ci.yml
- name: Pre-commit checks
  run: |
    .git/hooks/pre-commit --thorough --json > validation.json

- name: Upload results
  uses: actions/upload-artifact@v3
  with:
    name: validation-results
    path: validation.json
```

---

## 📝 Version History

### v2.1 (Current)
- ✅ Coverage requirements (40% block / 70% warn)
- ✅ Security vulnerability scanning (govulncheck)
- ✅ Enhanced linting (gosec, unused, ineffassign)
- ✅ Early import cycle detection
- ✅ Conditional race detection
- ✅ TDD-aware checking (run but don't block in RED)
- ✅ Two-level configuration system
- ✅ Performance optimizations (caching, parallel execution)

### v2.0 (Previous)
- JSON output support
- Auto-fix capabilities
- Multiple modes (quick/standard/thorough)
- Agent integration
- Smart error parsing

---

**Last Updated**: 2025-10-23
**Hook Version**: 2.1
**Performance Budget**: 90 seconds
**Compatibility**: Go 1.23+
