# GitHub Actions Workflows

This directory contains automated CI/CD workflows for the DevSmith Modular Platform.

## Workflow Overview

### 🔄 Development Workflow Automation

#### `auto-sync-next-issue.yml` - Auto-Sync and Create Next Issue Branch
**Trigger:** PR merge to `development` branch
**Purpose:** Automatically prepares for the next issue after PR merge

**What it does:**
1. Detects completed issue number from merged branch
2. Commits any pending `copilot-activity.md` changes
3. Finds next sequential issue file (e.g., 004 → 005)
4. Creates `feature/NNN-description` branch for next issue
5. Posts comment on merged PR with next steps
6. Notifies if no next issue found

**Benefits:**
- Zero manual work to start next issue
- Automatic activity log merge conflict resolution
- Consistent sequential workflow
- Visible progress tracking

**Required:** `contents: write`, `pull-requests: write` permissions

**Example Flow:**
```
1. Merge PR #4 (feature/004-review-service-preview-mode)
   ↓
2. Workflow runs automatically
   ↓
3. Commits merged activity log
   ↓
4. Creates feature/005-review-skim-mode
   ↓
5. Posts comment: "✅ Issue #004 merged! 🚀 Next: Issue #005"
   ↓
6. Developer: git pull && git checkout feature/005-review-skim-mode
```

---

### ✅ Quality & Testing

#### `pr-checks.yml` - PR Quality Checks
**Trigger:** PR opened/updated to `development` or `main`
**Purpose:** Comprehensive PR validation

**Checks:**
- Conventional commit format
- Issue number reference
- PR description quality
- File change limits (500 lines per file)
- Test coverage requirements
- Code quality metrics

---

#### `test-and-build.yml` - Test and Build
**Trigger:** Push to any branch
**Purpose:** Run tests and build services

**What it does:**
- Runs Go unit tests (70%+ coverage required)
- Builds all service binaries
- Validates Go modules
- Checks for compilation errors

---

#### `validate-migrations.yml` - Database Migration Validation
**Trigger:** PR with `migrations/**` changes
**Purpose:** Ensure database migrations are safe

**Checks:**
- Migration file naming convention
- SQL syntax validation
- Rollback script presence
- Migration order consistency

---

### 🔒 Security & Compliance

#### `security-scan.yml` - Security Scanning
**Trigger:** Push to `development` or `main`
**Purpose:** Scan for security vulnerabilities

**What it does:**
- Runs `gosec` for Go security issues
- Checks dependencies for known vulnerabilities
- Scans for hardcoded secrets
- Reports findings as PR comments

---

### 🚀 Deployment & Preview

#### `pr-preview.yml` - PR Preview Deployment
**Trigger:** PR labeled with `preview`
**Purpose:** Deploy preview environment for testing

**What it does:**
- Builds Docker images for changed services
- Deploys to preview environment
- Posts preview URL in PR comment
- Auto-tears down when PR closed

---

### 🏷️ Organization

#### `auto-label.yml` - Automatic PR Labeling
**Trigger:** PR opened/updated
**Purpose:** Automatically label PRs based on changes

**Labels added based on:**
- File paths (e.g., `backend`, `frontend`, `database`)
- Conventional commit type (e.g., `feature`, `bugfix`, `docs`)
- Issue number (e.g., `issue-004`)

---

## Permissions

All workflows use `GITHUB_TOKEN` with these permissions:

| Workflow | Permissions Required |
|----------|---------------------|
| `auto-sync-next-issue.yml` | `contents: write`, `pull-requests: write` |
| `test-and-build.yml` | `contents: read` |
| `security-scan.yml` | `contents: read`, `security-events: write` |
| `pr-preview.yml` | `contents: read`, `deployments: write` |
| `auto-label.yml` | `contents: read`, `pull-requests: write` |
| `validate-migrations.yml` | `contents: read` |

---

## Workflow Dependencies

```
Feature Branch Push
  ↓
Developer manually creates PR (gh pr create)
  ↓
test-and-build.yml + security-scan.yml + auto-label.yml
  ↓
PR Approved & Merged
  ↓
auto-sync-next-issue.yml (creates next feature branch)
```

---

## Troubleshooting

### Next Issue Branch Not Created
- Verify PR was merged (not just closed)
- Check next issue file exists (e.g., `.docs/issues/005-*.md`)
- Review workflow run logs in Actions tab

### Activity Log Merge Conflicts
- Should be auto-resolved by `auto-sync-next-issue.yml`
- If manual intervention needed, use `scripts/sync-and-start-issue.sh`

---

**Created:** 2025-10-20
**Last Updated:** 2025-10-20
