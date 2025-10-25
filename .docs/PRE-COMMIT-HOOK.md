# Pre-Commit Hook Guide

## What It Does

The `.git/hooks/pre-commit` hook automatically validates your code **before each commit**, ensuring quality and catching errors early.

**Automatic checks:**
- ✅ Code formatting (`go fmt`)
- ✅ Static analysis (`go vet`)
- ✅ Linting (`golangci-lint`)
- ✅ Tests (`go test`)
- ✅ Build validation (services compile)

**If validation fails, your commit is blocked.** This protects the codebase from broken code.

---

## Understanding the Output

### Human-Readable Dashboard

When you commit, you'll see an intelligent dashboard:

```
🔍 Pre-commit validation (standard mode)...

📁 Staged files: 13

CHECK RESULTS:
  ✓ fmt                  passed
  ✓ vet                  passed
  ✗ lint                 failed
  ✗ tests                failed

HIGH PRIORITY (Blocking): 4 issue(s)
  • [test_mock_panic] aggregator_service_test.go:125 - missing mock expectation for FindAllServices
    → Add Mock.On("FindAllServices").Return(...) - see .docs/copilot-instructions.md §5.1

  • [build_typecheck] mock_log_reader.go:36 - Error return value is not checked
    → Fix type error - this blocks tests from running

LOW PRIORITY (Can defer): 21 issue(s)
  • [style] Missing godoc comments on exported types
  • [lint] Struct field alignment optimizations
  ... and 16 more

FIX ORDER:
  1. Fix build errors
  2. Fix test failures
  3. Fix style

QUICK FIXES:
  • Auto-fix simple issues: .git/hooks/pre-commit --fix
  • Format code:           go fmt ./...
  • Run tests:             go test ./...

================================================
✗ Pre-commit validation FAILED
================================================
```

### Priority Levels

**HIGH PRIORITY (Blocking)**
These **must be fixed** before you can commit:
- `[test_mock_panic]` - Missing mock setup
- `[build_typecheck]` - Type errors, unused variables
- `[test_assertion]` - Test failures
- `[build_undefined]` - Missing functions/imports

**LOW PRIORITY (Can defer)**
Style and optimization suggestions:
- `[style]` - Missing comments, naming conventions
- `[code_quality]` - Complexity, parameter combining
- `[lint]` - Struct alignment, minor improvements

**Fix HIGH PRIORITY first!** LOW PRIORITY won't block commits once critical issues are resolved.

---

## Common Use Cases

### 1. See All Issues (Full List)

The dashboard shows only 5 low-priority issues. To see all:

```bash
.git/hooks/pre-commit --json | jq '.issues[] | "\(.type): \(.file):\(.line) - \(.message)"'
```

### 2. Auto-Fix Simple Issues

```bash
.git/hooks/pre-commit --fix
```

This automatically fixes:
- Code formatting
- Import organization
- Some style issues

### 3. Check Specific Tool Only

```bash
.git/hooks/pre-commit --check-only golangci-lint
.git/hooks/pre-commit --check-only tests
```

### 4. Quick Mode (Fast Feedback)

```bash
.git/hooks/pre-commit --quick
```

Runs only formatting + critical errors (skips full lint/test suite).

### 5. Thorough Mode (Before PR)

```bash
.git/hooks/pre-commit --thorough
```

Includes race detection and comprehensive checks.

---

## JSON Output for Tools/AI

For programmatic access or AI agents:

```bash
.git/hooks/pre-commit --json
```

Returns structured JSON:

```json
{
  "status": "failed",
  "duration": 3,
  "summary": {
    "total": 25,
    "errors": 4,
    "warnings": 21,
    "autoFixable": 11
  },
  "grouped": {
    "high": [
      {
        "type": "test_mock_panic",
        "severity": "error",
        "file": "internal/analytics/services/aggregator_service_test.go",
        "line": 125,
        "message": "Test 'TestAggregatorService_AnalyzeAggregations' - missing mock expectation for FindAllServices",
        "suggestion": "Add Mock.On(\"FindAllServices\").Return(...)",
        "autoFixable": false,
        "fixCommand": "",
        "context": "..."
      }
    ],
    "low": [...]
  }
}
```

**Fields:**
- `file`, `line` - Where the issue is
- `message` - What's wrong
- `suggestion` - How to fix it
- `autoFixable` - Can `--fix` resolve it?
- `context` - Code snippet around the issue

---

## Bypass (Emergency Only)

**⚠️ NOT RECOMMENDED** - Only use if absolutely necessary:

```bash
git commit --no-verify
```

This skips all validation and can introduce broken code.

**When it's acceptable:**
- Emergency hotfix (must still pass CI/CD)
- Work-in-progress commits on personal branches
- Blocked by hook bug (report it!)

**Never use on:**
- Pull requests to `main` or `development`
- Shared feature branches
- Final commits before PR

---

## Troubleshooting

### "Pre-commit validation PASSED but tests are failing locally"

The hook runs `go test -short ./...`. Your local tests might:
- Use different flags
- Test more packages
- Have different environment setup

Run: `go test ./...` (without `-short`) to see all test output.

### "Hook says lint passed but I see warnings"

The `.golangci.yml` config excludes test files from some linters (by design). Check:

```bash
golangci-lint run ./...
```

### "Commit blocked with no clear error message"

Run the hook manually for verbose output:

```bash
.git/hooks/pre-commit --json | jq .
```

### "Hook not running at all"

Check if it's executable:

```bash
chmod +x .git/hooks/pre-commit
```

---

## How It Works

**Location:** `.git/hooks/pre-commit` (not version controlled)

**Workflow:**
1. You run `git commit`
2. Hook intercepts before creating commit
3. Runs validation checks in parallel
4. Parses output into intelligent dashboard
5. Blocks commit if HIGH PRIORITY issues found
6. Commit proceeds if all checks pass

**Performance:**
- Standard mode: ~2-3 seconds
- Quick mode: ~1 second
- Thorough mode: ~5-10 seconds

---

## For AI Agents / Copilot

See `.github/copilot-instructions.md` Step 2.6 for:
- How to write code that passes validation first time
- Common pre-commit failure patterns
- Mock setup guidelines (§5.1)
- TDD workflow integration

---

## Configuration

Hook behavior is controlled by:
- **`.golangci.yml`** - Linter rules and exclusions
- **`go.mod`** - Go version and dependencies
- **Hook flags** - `--quick`, `--thorough`, `--fix`, etc.

No configuration needed for normal use - it works out of the box!

---

## See Also

- **[ARCHITECTURE.md](../ARCHITECTURE.md)** - Coding standards enforced by hook
- **[DevsmithTDD.md](../DevsmithTDD.md)** - TDD workflow and testing patterns
- **[.github/copilot-instructions.md](../.github/copilot-instructions.md)** - AI agent guidelines
- **[.docs/WORKFLOW-GUIDE.md](./WORKFLOW-GUIDE.md)** - Development workflow

---

**Remember:** The hook is your friend! It saves you from CI/CD failures, code review delays, and production bugs. Trust the dashboard and fix issues in priority order. 🎯
