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

### 🔄 `auto-sync-next-issue.yml` - Auto-Create Next Issue Branch
**Trigger:** PR merge to `development`
**Purpose:** Automatically prepare for next sequential issue

**What it does:**
1. Detects completed issue number from merged branch
2. Commits any pending `copilot-activity.md` changes
3. Finds next sequential issue file (e.g., `004 → 005`)
4. Creates `feature/NNN-description` branch for next issue
5. Posts comment on merged PR with next steps

**Benefits:**
- Zero manual work to start next issue
- Consistent sequential workflow
- Automatic activity log merge conflict resolution

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

| Workflow | Permissions |
|----------|-------------|
| `ci.yml` | `contents: read` |
| `auto-sync-next-issue.yml` | `contents: write`, `pull-requests: write` |
| `security-scan.yml` | `contents: read`, `security-events: write` |
| `auto-label.yml` | `contents: read`, `pull-requests: write` |

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

---

## Development Workflow

```
1. Work on feature branch
   ↓
2. Commit (pre-commit hook validates locally)
   ↓
3. Push to GitHub
   ↓
4. Create PR (gh pr create)
   ↓
5. CI runs (validates Docker + builds)
   ↓
6. Review & merge
   ↓
7. auto-sync-next-issue creates next branch
```

---

## Why No Database Tests in CI?

**Problem:** Static schema file (`docker/postgres/init-schemas.sql`) gets out of sync with evolving code models.

**Example from PR #9:**
```
1. Developer adds User.Email field to struct
2. Updates repository queries to use email
3. Tests pass locally (local DB has email column)
4. CI fails: "column email does not exist"
5. Developer spends hours debugging
6. Root cause: init-schemas.sql missing email column
```

**This is not a bug catch - it's a false failure from schema drift.**

**Solution:** Pre-commit hook runs tests against local database (which stays in sync through development). CI skips database tests until migration system exists.

---

**Created:** 2025-10-20
**Last Updated:** 2025-10-21
**Philosophy:** Fail loudly for real problems. Never fail for configuration drift.
