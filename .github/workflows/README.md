# GitHub Actions Workflows

This directory contains automated CI/CD workflows for the DevSmith Modular Platform.

## Quality Philosophy

**Pre-commit hook = Quality Gate** ✅
**CI = Lightweight Safety Net** 🛡️

The pre-commit hook (`.git/hooks/pre-commit`) is comprehensive and catches issues locally:
- Code formatting (go fmt)
- Static analysis (go vet)
- Comprehensive linting (golangci-lint)
- Full builds (all 4 services)
- Tests (go test -short)
- Misplaced code detection

CI validates what pre-commit can't:
- Docker image builds
- Things that slip through if developers use `--no-verify`

## Active Workflows

### 🏗️ `ci.yml` - Continuous Integration
**Trigger:** Push/PR to `development` or `main`
**Purpose:** Lightweight validation of deployment artifacts

**Jobs:**
1. **Build Services** - Validates all 4 services compile (`portal`, `review`, `logs`, `analytics`)
2. **Docker Build** - Validates Docker images build successfully (can't do in pre-commit)
3. **Quick Lint** - Fast golangci-lint pass (catches `--no-verify` commits)
4. **CI Summary** - Aggregates results (useful for branch protection)

**Why this approach:**
- ✅ Only fails for real problems (build errors, Docker issues)
- ✅ No database tests (avoids schema drift false failures)
- ✅ Fast (<3 minutes typical)
- ✅ Doesn't duplicate pre-commit extensively

---

### 🔗 `link-pr-to-issue-and-validate.yml` (renamed from auto-sync-next-issue.yml)
**Trigger:** PR opened/updated to `development` or `main`
**Purpose:** Enforce GitHub Issues workflow standards and link PRs to issues

**What it does:**
1. Extracts issue number from PR body (`Closes #NUMBER`)
2. Validates issue exists in GitHub
3. Checks issue has acceptance criteria
4. Extracts metrics (coverage, tests) from PR
5. Posts metrics as issue comment
6. Validates PR description completeness
7. Confirms auto-close linkage will work on merge

**Why this approach:**
- ✅ Enforces issue-linked PRs (not optional)
- ✅ Validates acceptance criteria are documented
- ✅ Bridges PR and issue with automatic comments
- ✅ Extracts and posts quality metrics
- ✅ Guides developers to complete PR descriptions
- ✅ Ensures issues auto-close on PR merge

**Value delivered:**
- Prevents untracked PRs (all PRs must link to issues)
- Keeps issues updated with PR metrics automatically
- Validates workflow standards are followed
- Enables accurate issue tracking and closure

---

### 🎯 `issue-workflow-validation.yml` (NEW)
**Trigger:** Issue opened/edited, PR opened/edited
**Purpose:** Enforce quality standards for issues and PRs

**What it does:**

**For Issues:**
1. Validates title has service prefix (`[Service]`)
2. Checks description length
3. Validates acceptance criteria section exists
4. Checks test requirements are specified
5. Confirms coverage target is stated
6. Posts validation report with recommendations

**For PRs:**
1. Requires issue linkage (`Closes #NUMBER`)
2. Checks for implementation section
3. Validates test results section exists
4. Looks for quality checklist section
5. Confirms acceptance criteria listed
6. Posts completeness score and recommendations
7. Validates issue closure will occur on merge

**Why this approach:**
- ✅ Catches incomplete issues before work starts
- ✅ Ensures PRs have full context
- ✅ Validates workflow standards are followed
- ✅ Provides actionable feedback
- ✅ Prevents surprise closures or tracking issues
- ✅ Enforces issue→PR linking

**Value delivered:**
- Issues are properly specified before implementation
- PRs have complete information for reviewers
- Prevents workflow violations
- Ensures metric tracking and reporting
- Guarantees accurate issue tracking

---

### 🔒 `security-scan.yml` - Security Scanning
**Trigger:** Push to `main`, weekly schedule, manual dispatch
**Purpose:** Scan for security vulnerabilities

**What it does:**
- Runs `govulncheck` for Go vulnerability scanning
- Dependency review (on PRs)
- Secret scanning with Gitleaks

**Why kept:**
- Runs on schedule (not blocking)
- Catches real security issues
- No false positives from schema drift

---

### 🏷️ `auto-label.yml` - Automatic PR Labeling
**Trigger:** PR opened/updated
**Purpose:** Auto-label PRs for organization

**What it does:**
- Labels by file paths changed
- Labels by PR size (XS/S/M/L/XL)

**Why kept:**
- Non-blocking (doesn't affect CI status)
- Helpful for PR triage

**⚠️ Configuration Issue:** See below

---

## Workflow Philosophy Evolution

### Before (Obsolete)
- ❌ Auto-created next feature branch based on filename
- ❌ Depended on `.docs/issues/NNN-*.md` convention
- ❌ Didn't link PRs to GitHub issues
- ❌ Created extra automation that added no value

### Now (Current)
- ✅ Enforces GitHub Issues as single source of truth
- ✅ Links PRs to issues automatically
- ✅ Validates workflow standards are followed
- ✅ Posts metrics and status updates to issues
- ✅ Ensures issues auto-close on PR merge
- ✅ Provides actionable feedback on completeness

---

## Known Issues

### ⚠️ labeler.yml Configuration Mismatch

**Problem:**
- `.github/labeler.yml` patterns don't match actual repository structure
- Expected: `apps/platform-*/**`, but actual is `apps/portal/**`
- Labels won't apply correctly to PRs

**Impact:** Low (non-blocking, cosmetic)

**Fix:** Update `.github/labeler.yml` patterns to match actual structure:
```yaml
'app:portal':
  - changed-files:
    - any-glob-to-any-file: 'apps/portal/**'

'app:review':
  - changed-files:
    - any-glob-to-any-file: 'apps/review/**'
```

---

## Disabled Workflows

See `.github/workflows-disabled/` for archived workflows and why they were disabled.

**Summary:**
- `test-and-build.yml` - Caused false failures from database schema drift
- `validate-migrations.yml` - Checked static schemas that diverged from code

Both had **fundamental design flaw**: Static `init-schemas.sql` + evolving code models = false "column doesn't exist" errors.

**To re-enable database tests:** Implement migration system first (`internal/*/db/migrations/*.sql`).

---

## Workflow Permissions

|| Workflow | Permissions |
||----------|-------------|
|| `ci.yml` | `contents: read` |
|| `link-pr-to-issue-and-validate.yml` | `contents: read`, `pull-requests: write`, `issues: write` |
|| `issue-workflow-validation.yml` | `issues: write`, `pull-requests: write` |
|| `security-scan.yml` | `contents: read`, `security-events: write` |
|| `auto-label.yml` | `contents: read`, `pull-requests: write` |

---

## CI Failure Troubleshooting

### Build Failures
**Symptom:** `go build` fails in CI but works locally
**Cause:** Likely used `git commit --no-verify` to bypass pre-commit
**Fix:** Run pre-commit checks locally, fix issues, push again

### Docker Build Failures
**Symptom:** Docker image build fails
**Cause:** Invalid Dockerfile or missing dependencies
**Fix:** Test Docker build locally: `docker build -f cmd/SERVICE/Dockerfile .`

### Lint Failures
**Symptom:** golangci-lint fails in CI but passed locally
**Cause:** Different golangci-lint version or config
**Fix:** Run `golangci-lint run ./...` locally with same version

### PR Validation Failures
**Symptom:** PR failing issue linkage check
**Cause:** PR description doesn't reference "Closes #NUMBER"
**Fix:** Update PR description to include `Closes #ISSUE_NUMBER`

---

## Development Workflow

```
1. Create GitHub issue with acceptance criteria
   ↓
2. Create feature branch (manual or from issue template)
   ↓
3. Work on feature with TDD (RED → GREEN → REFACTOR)
   ↓
4. Commit locally (pre-commit hook validates)
   ↓
5. Push to GitHub
   ↓
6. Create PR linked to issue: "Closes #NUMBER"
   ↓
7. GitHub Actions validates:
   - Issue is properly linked
   - PR has required sections
   - Metrics are reported
   - Issue will auto-close on merge
   ↓
8. Code review
   ↓
9. Merge to development
   ↓
10. GitHub auto-closes linked issue
    ↓
11. Issue is marked as complete in tracking
```

---

## Why GitHub Issues Workflow?

**Before:** Sequential file-based issues + auto-generated branches
**Problems:**
- ❌ Duplicate tracking (`.docs/issues` + GitHub issues)
- ❌ Branch generation was brittle (broke at issue #1000)
- ❌ No real value-add from automation
- ❌ Manual work still required for linking
- ❌ Metrics weren't automatically reported

**Now:** GitHub Issues as source of truth + GitHub Actions for validation
**Benefits:**
- ✅ Single source of truth (GitHub issues)
- ✅ No manual linking required
- ✅ Workflow violations caught early
- ✅ Metrics automatically posted
- ✅ Issues auto-close on PR merge
- ✅ Clear audit trail
- ✅ Less manual work, more automation value

---

**Created:** 2025-10-20
**Last Updated:** 2025-10-25
**Philosophy:** Fail loudly for real problems. Enforce workflow standards. Never fail for configuration drift.
