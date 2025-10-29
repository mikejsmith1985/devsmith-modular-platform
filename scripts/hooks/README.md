# Git Hooks - DevSmith Platform

This directory contains Git hooks that enforce code quality standards while keeping local development fast.

## Architecture

**Key principle:** Validation happens at **push time**, NOT at commit time.

- ✅ **Local commits are FAST** — no validation delays (encourages frequent commits)
- ✅ **Push validation is COMPREHENSIVE** — full suite runs before reaching GitHub
- ✅ **Developer experience first** — colorized dashboard output for humans
- ✅ **CI/CD friendly** — JSON output for automated pipelines

## Available Hooks

### `pre-push` (NEW - Primary Hook)

**When:** Runs when you execute `git push` (before changes reach GitHub)

**Validations:**
- Code formatting (gofmt)
- Static analysis (go vet)
- Linting (golangci-lint)
- Build verification (all 4 services)
- Test execution (full or short modes)
- Security scanning (govulncheck)
- Test coverage (55% target, non-blocking warning)

**Output Modes:**
```bash
# Human-readable dashboard (default)
./scripts/hooks/pre-push
./scripts/hooks/pre-push --standard

# Quick mode (format + build only)
./scripts/hooks/pre-push --quick

# Thorough mode (full tests, no short flag)
./scripts/hooks/pre-push --thorough

# Machine-readable JSON (for agents/CI)
./scripts/hooks/pre-push --json
```

### `pre-commit` (LEGACY - DISABLED)

Now disabled to keep local commits fast. The validation that was here has been moved to `pre-push`.

### `post-commit` (Utility)

**When:** Runs after successful local commits

**Function:** Logs commit activity to `.docs/devlog/copilot-activity.md` (main/development branches only)

## Installation

From repository root:

```bash
./scripts/install-hooks.sh
```

This installs:
- ✓ `pre-push` → `.git/hooks/pre-push`
- ✓ `post-commit` → `.git/hooks/post-commit`
- ✓ Configuration examples → `.git/hooks/pre-commit-local.yaml.example`

## Configuration

### Team Configuration (Committed)
`.pre-commit-config.yaml` — applies to all team members
- Coverage thresholds (40% error, 70% warning)
- Linting rules
- TDD awareness settings
- Performance budgets

### Local Overrides (Not Committed)
`.git/hooks/pre-commit-local.yaml` — per-developer customization

To use local config:
```bash
cp .git/hooks/pre-commit-local.yaml.example .git/hooks/pre-commit-local.yaml
# Edit .git/hooks/pre-commit-local.yaml as needed
```

## Example Output

### Human Output (Terminal Dashboard)

```
════════════════════════════════════════════════════════════════
  📊 PRE-PUSH VALIDATION DASHBOARD
════════════════════════════════════════════════════════════════

▶ Code Formatting
  ✓ Code formatting OK

▶ Static Analysis (go vet)
  ✓ No issues detected by go vet

▶ Linting (golangci-lint)
  ✓ No linting issues

▶ Build Verification
  ✓ Service: portal ✓
  ✓ Service: review ✓
  ✓ Service: logs ✓
  ✓ Service: analytics ✓

▶ Running Tests
  ✓ All tests passed

▶ Security Scanning (govulncheck)
  ✓ No known vulnerabilities

▶ Test Coverage
  ✓ Coverage: 62.5% ✓

SUMMARY:
  ✓ format
  ✓ vet
  ✓ lint
  ✓ build
  ✓ tests
  ✓ security
  ✓ coverage

✅ ALL CHECKS PASSED - Ready to push!
════════════════════════════════════════════════════════════════
```

### JSON Output (For Agents/CI)

```bash
./scripts/hooks/pre-push --json
```

```json
{
  "validation": "pre-push",
  "status": "passed",
  "timestamp": "2025-10-29T14:32:15-07:00",
  "duration_seconds": 45,
  "mode": "standard",
  "checks": {
    "format": "passed",
    "vet": "passed",
    "lint": "passed",
    "build": "passed",
    "tests": "passed",
    "security": "passed",
    "coverage": "passed"
  },
  "issues": []
}
```

## Workflow

### Normal Development
```bash
# Work on feature
git add .
git commit -m "feat(review): add template fixes"  # ⚡ FAST - no validation

# Make more commits
git add .
git commit -m "test(review): add template tests"  # ⚡ FAST - no validation

# When ready to push - validation kicks in
git push  # 🔍 Full validation runs here
```

### If Push Validation Fails
```bash
$ git push
✗ format                # Format failed
✗ build                 # Build failed

# Fix the issues
go fmt ./...
# ...fix build errors...

# Try again
git push  # ✅ Should pass now
```

### Manual Testing
```bash
# Test without pushing (useful during development)
./scripts/hooks/pre-push --quick       # Just format + build
./scripts/hooks/pre-push --standard    # Full suite (default)
./scripts/hooks/pre-push --thorough    # Full + long-running tests

# Machine-readable output (for CI integration)
./scripts/hooks/pre-push --json | jq '.status'
```

## Special Cases

### Bypass Pre-Push (Emergency Only)
```bash
git push --no-verify  # ⚠️ Use rarely - quality gates exist for a reason
```

### Running Pre-Push Locally (Before Committing)
Useful to validate before committing:
```bash
# Check without hooks
./scripts/hooks/pre-push --quick

# Fix issues
go fmt ./...
golangci-lint run ./...

# Then commit
git commit -m "fix: format and lint issues"
```

## Cost Considerations

**Per the DevSmith ultra-budget model ($60/month Copilot Pro):**

- ✅ **Local pre-commit hooks:** FREE (runs locally)
- ⚠️ **Pre-push validation:** Minimal cost (only runs on push, not every commit)
- 🚫 **GitHub Actions/CI:** Only if explicitly needed (not default)

## Troubleshooting

### Hook Not Running on Push
```bash
# Verify hook is installed
ls -la .git/hooks/pre-push

# Verify it's executable
chmod +x .git/hooks/pre-push

# Run manually to debug
./.git/hooks/pre-push --json
```

### golangci-lint Not Found
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### govulncheck Not Found
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Tests Pass Locally But Fail on Push
May indicate:
- Environment differences (database, services)
- Timing-sensitive tests
- Missing test fixtures

Solutions:
```bash
# Run full tests locally
go test ./...

# Run with race detector
go test -race ./...

# Check coverage
go test -cover ./...
```

## Documentation

- **Full pre-commit guide:** `.docs/PRE-COMMIT-ENHANCEMENTS.md`
- **TDD workflow:** `.docs/copilot-instructions.md §4`
- **Architecture standards:** `ARCHITECTURE.md`

## Quick Reference

| Action | Command | Speed | When |
|--------|---------|-------|------|
| Commit | `git commit -m "..."` | ⚡ ~1s | Any time |
| Push (validates) | `git push` | 🔍 ~45s | Before GitHub |
| Manual validation | `./scripts/hooks/pre-push` | 🔍 ~45s | Before committing |
| Quick check | `./scripts/hooks/pre-push --quick` | ⚡ ~10s | Format + build only |
| JSON for CI | `./scripts/hooks/pre-push --json` | 🔍 ~45s | Automated pipelines |
