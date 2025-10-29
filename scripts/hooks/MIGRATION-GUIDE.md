# Pre-Commit Hook → Pre-Push Migration Guide

## What Changed

**Old approach (pre-commit):**
- Validation ran on **every commit** (slow, 30-90 seconds)
- Multiple conflicting hook files in `.git/hooks/`
- Config complexity with YAML parsing
- TDD-unfriendly (blocked RED phase work)

**New approach (pre-push):**
- Validation runs **only on `git push`** (fast local commits)
- Single, consolidated `pre-push` hook
- Beautiful color-coded dashboard + JSON output
- Timeout protection (prevents hanging)
- Budget-friendly (no extra CI/CD costs)

## Migration Steps

### 1. Install the New Hook
```bash
./scripts/install-hooks.sh
```

This installs:
- ✓ `pre-push` → `.git/hooks/pre-push` (NEW)
- ✓ `post-commit` → `.git/hooks/post-commit` (activity logging)

### 2. Disable Old Pre-Commit Hook (Optional)
The old pre-commit hook is still there but can be safely disabled:
```bash
# Option A: Remove it
rm .git/hooks/pre-commit

# Option B: Disable it (keep as backup)
mv .git/hooks/pre-commit .git/hooks/pre-commit.disabled
chmod -x .git/hooks/pre-commit.disabled
```

### 3. Update Your Workflow

**Before (old way):**
```bash
git add .
git commit -m "feat: new feature"  # ⏳ 30-90s wait for validation
```

**After (new way):**
```bash
git add .
git commit -m "feat: new feature"  # ⚡ Instant (no validation)
git push                           # 🔍 Full validation runs here
```

## Key Differences

| Aspect | Old (Pre-Commit) | New (Pre-Push) |
|--------|------------------|----------------|
| **Timing** | On every commit | Only on push |
| **Local commit speed** | 30-90s | ~1s ⚡ |
| **TDD-friendly** | ❌ Blocks RED phase | ✅ Allows RED phase |
| **Developer friction** | High | Low |
| **Token cost** | High (many commits) | Low (fewer validates) |
| **Output** | Complex YAML config | Simple color dashboard |
| **JSON support** | Limited | Full support |
| **Timeout protection** | None (hangs possible) | Protected (30 timeouts) |

## Validation Modes

```bash
# Default (format + vet + lint + build + tests + security + coverage)
git push

# Or test locally first:
./scripts/hooks/pre-push --standard      # Default mode
./scripts/hooks/pre-push --quick         # Format + build only (⚡ fast)
./scripts/hooks/pre-push --thorough      # Full tests (no -short flag)
./scripts/hooks/pre-push --json          # Machine-readable for CI

# Bypass (emergency only)
git push --no-verify
```

## Configuration

### Team Config (Committed)
`.pre-commit-config.yaml` — shared across team

```yaml
coverage:
  error_threshold: 40    # Block if below
  warning_threshold: 70  # Warn if below
```

### Local Config (Not Committed)
`.git/hooks/pre-commit-local.yaml` — individual overrides

```bash
cp .git/hooks/pre-commit-local.yaml.example .git/hooks/pre-commit-local.yaml
# Edit to customize for your machine
```

## Timeout Configuration

Adjustable in `scripts/hooks/pre-push` if needed:

```bash
TIMEOUT_FORMAT=10          # gofmt timeout
TIMEOUT_VET=15            # go vet timeout
TIMEOUT_LINT=30           # golangci-lint timeout
TIMEOUT_BUILD=60          # go build timeout per service
TIMEOUT_TESTS_SHORT=45    # Short test timeout
TIMEOUT_TESTS_FULL=120    # Full test timeout
TIMEOUT_SECURITY=30       # govulncheck timeout
```

## Example Output

### Human Dashboard
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

### JSON Output
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

## Troubleshooting

### Q: Will pre-push block legitimate work?

**A:** No. Pre-push is **warnings-friendly**:
- ✅ Timeout warnings are non-blocking (push proceeds)
- ✅ Coverage warnings are non-blocking (unless critical)
- ⛔ Only **hard errors block** (broken builds, failed tests)

### Q: I forgot to fix something — can I push anyway?

**A:** Yes, two options:

```bash
# Option 1: Force push (use sparingly)
git push --no-verify

# Option 2: Fix and re-push (recommended)
go fmt ./...
golangci-lint fix ./... (if available)
git add .
git commit -m "fix: lint issues"
git push
```

### Q: Tests pass locally but fail on push

**A:** Run pre-push validation locally to debug:
```bash
./scripts/hooks/pre-push --standard
./scripts/hooks/pre-push --thorough  # For full tests
go test -race ./...                  # Race detector
go test -cover ./...                 # Coverage details
```

### Q: Hook not running on push

**A:** Verify installation:
```bash
ls -la .git/hooks/pre-push
chmod +x .git/hooks/pre-push
./.git/hooks/pre-push --quick  # Test manually
```

## Benefits Summary

### For Developers
- ⚡ Fast local commits (friction-free TDD)
- 📊 Beautiful colorized feedback
- 🔍 Comprehensive validation only at push
- ⏱️ Protected from hangs (timeout guards)
- 🎯 Quick feedback on what failed

### For DevSmith Budget
- 💰 No extra CI/CD costs (local validation)
- 🤝 Tokens conserved (fewer validation runs)
- ⚙️ Reliable (timeout-protected, no hanging)

### For Code Quality
- ✅ Catches issues before GitHub
- 🔐 Security scanning (govulncheck)
- 📈 Coverage tracking (55% target)
- 🏗️ Build verification (all 4 services)

## Next Steps

1. **Install hooks:** `./scripts/install-hooks.sh`
2. **Test locally:** `./scripts/hooks/pre-push --quick`
3. **Read full guide:** `.docs/PRE-COMMIT-ENHANCEMENTS.md`
4. **Delete old config (optional):** `rm -f .git/hooks/pre-commit.disabled`

---

Questions? Check:
- `scripts/hooks/README.md` — Hook documentation
- `ARCHITECTURE.md` — Architecture standards
- `.docs/copilot-instructions.md §4` — TDD workflow

