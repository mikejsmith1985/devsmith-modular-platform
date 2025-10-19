# Issue #2: [COPILOT] CI/CD Pipeline Setup

**Labels:** `copilot`, `ci-cd`, `infrastructure`, `automation`
**Assignee:** Mike (with Copilot assistance)
**Estimated Time:** 60-90 minutes
**Complexity:** Medium
**Depends On:** Issue #001

---

## Task Description

Set up GitHub Actions CI/CD pipeline to automatically test, build, and validate code changes on every push and pull request. This ensures code quality is maintained and provides fast feedback to developers and AI agents.

**Why This Task for Copilot:**
- Standard GitHub Actions patterns
- Well-documented YAML syntax
- Copilot excels at DevOps configurations
- Clear acceptance criteria
- Mike can review and test in GitHub UI

---

## Overview

### What We're Building

A multi-workflow CI/CD system that:
1. **Tests** all Go code on every push/PR
2. **Builds** Docker images to verify they work
3. **Lints** code for quality and style
4. **Validates** database migrations
5. **Reports** coverage and test results
6. **Provides fast feedback** (< 5 minutes for most workflows)

### Why This Matters

- **Catches errors early**: Before they reach production
- **AI agent validation**: Aider/OpenHands work is automatically tested
- **Code quality**: Enforces standards without manual review
- **Deployment confidence**: Know that merged code works

---

## Workflows to Create

### 1. Main Test & Build Workflow

**File:** `.github/workflows/test-and-build.yml`

**Triggers:**
- Every push to `development`, `main`
- Every pull request to `development`, `main`

**Jobs:**
1. **Go Tests**: Run all tests with coverage
2. **Go Build**: Build all services
3. **Docker Build**: Verify Dockerfiles work
4. **Lint**: Run golangci-lint
5. **Coverage Report**: Upload to Codecov (optional)

**Estimated Runtime:** 3-5 minutes

---

### 2. Database Migration Validation

**File:** `.github/workflows/validate-migrations.yml`

**Triggers:**
- Every push to `development`, `main`
- Every pull request to `development`, `main`
- Changes to `**/migrations/*.sql` or `docker/postgres/*.sql`

**Jobs:**
1. Start PostgreSQL
2. Run init-schemas.sql
3. Run all migrations in order
4. Verify schema integrity

**Estimated Runtime:** 1-2 minutes

---

### 3. Security Scanning

**File:** `.github/workflows/security-scan.yml`

**Triggers:**
- Every push to `main`
- Weekly schedule (Monday 2am)
- Manual trigger

**Jobs:**
1. Go vulnerability check (govulncheck)
2. Dependency scanning (Dependabot)
3. Secret scanning

**Estimated Runtime:** 2-3 minutes

---

### 4. PR Preview Comments

**File:** `.github/workflows/pr-preview.yml`

**Triggers:**
- Pull request opened/updated

**Jobs:**
1. Comment on PR with test results
2. Comment with coverage diff
3. Comment with build status
4. Add labels based on results (✅ tests-passing, ❌ tests-failing)

**Estimated Runtime:** < 1 minute

---

## Detailed Specifications

### .github/workflows/test-and-build.yml

```yaml
name: Test and Build

on:
  push:
    branches: [development, main]
  pull_request:
    branches: [development, main]

jobs:
  test:
    name: Run Tests
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_DB: devsmith_test
          POSTGRES_USER: devsmith
          POSTGRES_PASSWORD: test_password
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Verify dependencies
        run: go mod verify

      - name: Initialize database
        env:
          PGPASSWORD: test_password
        run: |
          psql -h localhost -U devsmith -d devsmith_test -f docker/postgres/init-schemas.sql

      - name: Run tests
        env:
          DATABASE_URL: postgres://devsmith:test_password@localhost:5432/devsmith_test?sslmode=disable
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Display coverage
        run: |
          go tool cover -func=coverage.out | grep total:

      - name: Upload coverage to Codecov (optional)
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
          flags: unittests
          name: codecov-devsmith
        continue-on-error: true

  build:
    name: Build Services
    runs-on: ubuntu-latest
    needs: test

    strategy:
      matrix:
        service: [portal, review, logs, analytics]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Build ${{ matrix.service }}
        run: |
          go build -v -o bin/${{ matrix.service }} ./cmd/${{ matrix.service }}

      - name: Verify binary
        run: |
          test -f bin/${{ matrix.service }}
          file bin/${{ matrix.service }}

  docker-build:
    name: Docker Build
    runs-on: ubuntu-latest
    needs: test

    strategy:
      matrix:
        service: [portal, review, logs, analytics]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build ${{ matrix.service }} Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./cmd/${{ matrix.service }}/Dockerfile
          push: false
          tags: devsmith-${{ matrix.service }}:test
          cache-from: type=gha
          cache-to: type=gha,mode=max

  lint:
    name: Lint Code
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

  summary:
    name: CI Summary
    runs-on: ubuntu-latest
    needs: [test, build, docker-build, lint]
    if: always()

    steps:
      - name: Check results
        run: |
          echo "Test: ${{ needs.test.result }}"
          echo "Build: ${{ needs.build.result }}"
          echo "Docker: ${{ needs.docker-build.result }}"
          echo "Lint: ${{ needs.lint.result }}"

          if [ "${{ needs.test.result }}" != "success" ] || \
             [ "${{ needs.build.result }}" != "success" ] || \
             [ "${{ needs.docker-build.result }}" != "success" ] || \
             [ "${{ needs.lint.result }}" != "success" ]; then
            echo "❌ CI failed"
            exit 1
          fi

          echo "✅ All CI checks passed"
```

---

### .github/workflows/validate-migrations.yml

```yaml
name: Validate Database Migrations

on:
  push:
    branches: [development, main]
    paths:
      - '**/migrations/*.sql'
      - 'docker/postgres/*.sql'
  pull_request:
    branches: [development, main]
    paths:
      - '**/migrations/*.sql'
      - 'docker/postgres/*.sql'

jobs:
  validate:
    name: Validate Migrations
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_DB: devsmith
          POSTGRES_USER: devsmith
          POSTGRES_PASSWORD: test_password
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run schema initialization
        env:
          PGPASSWORD: test_password
        run: |
          psql -h localhost -U devsmith -d devsmith -f docker/postgres/init-schemas.sql

      - name: Verify schemas created
        env:
          PGPASSWORD: test_password
        run: |
          SCHEMA_COUNT=$(psql -h localhost -U devsmith -d devsmith -t -c "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name IN ('portal', 'reviews', 'logs', 'analytics');" | tr -d ' ')

          if [ "$SCHEMA_COUNT" != "4" ]; then
            echo "❌ Expected 4 schemas, found $SCHEMA_COUNT"
            exit 1
          fi

          echo "✅ All schemas created successfully"

      - name: Run migrations (if any)
        env:
          PGPASSWORD: test_password
        run: |
          # Run migrations for each service
          # Example: for f in internal/portal/db/migrations/*.sql; do psql ... -f "$f"; done
          echo "No migrations yet, but this will run them when they exist"

      - name: List all tables
        env:
          PGPASSWORD: test_password
        run: |
          psql -h localhost -U devsmith -d devsmith -c "
            SELECT schemaname, tablename
            FROM pg_tables
            WHERE schemaname IN ('portal', 'reviews', 'logs', 'analytics')
            ORDER BY schemaname, tablename;
          "
```

---

### .github/workflows/security-scan.yml

```yaml
name: Security Scan

on:
  push:
    branches: [main]
  schedule:
    # Run every Monday at 2am UTC
    - cron: '0 2 * * 1'
  workflow_dispatch:

jobs:
  govulncheck:
    name: Go Vulnerability Check
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Run govulncheck
        run: govulncheck ./...

  dependency-review:
    name: Dependency Review
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Dependency Review
        uses: actions/dependency-review-action@v4

  secret-scan:
    name: Secret Scanning
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### .github/workflows/pr-preview.yml

```yaml
name: PR Preview

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  preview:
    name: Generate PR Preview
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Get changed files
        id: changed-files
        uses: tj-actions/changed-files@v41
        with:
          files: |
            **/*.go
            **/go.mod
            **/go.sum

      - name: Comment on PR
        uses: actions/github-script@v7
        with:
          script: |
            const changedFiles = '${{ steps.changed-files.outputs.all_changed_files }}'.split(' ');
            const goFiles = changedFiles.filter(f => f.endsWith('.go'));

            const body = `
            ## 🔍 PR Preview

            **Changed Files:** ${changedFiles.length} files
            **Go Files Changed:** ${goFiles.length} files

            ### 📊 What's Next
            - ⏳ Running tests...
            - ⏳ Building Docker images...
            - ⏳ Running linters...

            Results will be updated shortly.

            ---
            🤖 Generated by GitHub Actions
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });
```

---

### .golangci.yml (Linter Configuration)

**File:** `.golangci.yml`

```yaml
# golangci-lint configuration
# See https://golangci-lint.run/usage/configuration/

run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly

linters:
  enable:
    - errcheck      # Check for unchecked errors
    - gosimple      # Simplify code
    - govet         # Vet examines Go source code
    - ineffassign   # Detect ineffectual assignments
    - staticcheck   # Advanced static analysis
    - unused        # Check for unused code
    - gofmt         # Check formatting
    - goimports     # Check imports
    - misspell      # Check for misspelled English words
    - revive        # Replacement for golint
    - gocritic      # Opinionated linter
    - gosec         # Security issues
    - bodyclose     # Check HTTP response body closed
    - nilerr        # Find code returning nil even if err != nil

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true

  govet:
    check-shadowing: true

  revive:
    rules:
      - name: exported
        severity: warning
        disabled: false
        arguments:
          - "checkPrivateReceivers"
          - "sayRepetitiveInsteadOfStutters"

  gosec:
    excludes:
      - G404 # Weak random number generator (OK for non-crypto use)

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0

  exclude-rules:
    # Exclude test files from some checks
    - path: _test\.go
      linters:
        - errcheck
        - gosec

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
```

---

### Codecov Configuration (Optional)

**File:** `.codecov.yml`

```yaml
# Codecov configuration
# See https://docs.codecov.io/docs/codecov-yaml

coverage:
  status:
    project:
      default:
        target: 70%          # Project must maintain 70% coverage
        threshold: 5%        # Allow 5% drop before failing
    patch:
      default:
        target: 70%          # New code must have 70% coverage

comment:
  layout: "header, diff, files, footer"
  behavior: default
  require_changes: false

ignore:
  - "**/*_test.go"
  - "**/mocks/**"
  - "cmd/*/main.go"        # Entry points don't need test coverage
```

---

## Acceptance Criteria

- [ ] `.github/workflows/test-and-build.yml` created
- [ ] `.github/workflows/validate-migrations.yml` created
- [ ] `.github/workflows/security-scan.yml` created
- [ ] `.github/workflows/pr-preview.yml` created
- [ ] `.golangci.yml` linter configuration created
- [ ] `.codecov.yml` created (optional)
- [ ] All workflows trigger on appropriate events
- [ ] Test workflow runs Go tests with coverage
- [ ] Build workflow compiles all services
- [ ] Docker workflow builds all Dockerfiles
- [ ] Lint workflow runs golangci-lint
- [ ] Migration workflow validates database schemas
- [ ] All workflows complete in < 5 minutes
- [ ] Workflows show status badges (can be added to README)
- [ ] PR preview comments appear on pull requests

---

## Testing the Workflows

### Local Testing with Act

Install [act](https://github.com/nektos/act) to test workflows locally:

```bash
# Install act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Test the main workflow
act push -W .github/workflows/test-and-build.yml

# Test PR workflow
act pull_request -W .github/workflows/pr-preview.yml
```

### Testing on GitHub

1. **Create test branch:**
```bash
git checkout -b feature/003-cicd-setup
```

2. **Add workflow files:**
```bash
git add .github/workflows/*.yml .golangci.yml .codecov.yml
git commit -m "feat(ci): add GitHub Actions CI/CD workflows"
```

3. **Push and create PR:**
```bash
git push origin feature/002-cicd-setup
gh pr create --title "feat(ci): add CI/CD workflows" --base development
```

4. **Watch Actions tab:**
   - Go to https://github.com/mikejsmith1985/devsmith-modular-platform/actions
   - See workflows running
   - Check for errors

5. **Verify:**
   - ✅ All checks pass
   - ✅ PR comment appears with preview
   - ✅ Coverage report generated
   - ✅ Linter runs without errors

---

## Status Badges for README

Add these to README.md after workflows are set up:

```markdown
# DevSmith Modular Platform

[![Test and Build](https://github.com/mikejsmith1985/devsmith-modular-platform/actions/workflows/test-and-build.yml/badge.svg)](https://github.com/mikejsmith1985/devsmith-modular-platform/actions/workflows/test-and-build.yml)
[![Security Scan](https://github.com/mikejsmith1985/devsmith-modular-platform/actions/workflows/security-scan.yml/badge.svg)](https://github.com/mikejsmith1985/devsmith-modular-platform/actions/workflows/security-scan.yml)
[![codecov](https://codecov.io/gh/mikejsmith1985/devsmith-modular-platform/branch/main/graph/badge.svg)](https://codecov.io/gh/mikejsmith1985/devsmith-modular-platform)
```

---

## Implementation Steps

### 1. Create Feature Branch

```bash
git checkout development
git pull origin development
git checkout -b feature/002-cicd-setup
```

### 2. Create Workflow Files

Create each workflow file in `.github/workflows/` following the specifications above.

**Order:**
1. `test-and-build.yml` (most important)
2. `validate-migrations.yml`
3. `security-scan.yml`
4. `pr-preview.yml`
5. `.golangci.yml` (in project root)
6. `.codecov.yml` (in project root, optional)

### 3. Test Locally (with act)

```bash
# Test main workflow
act push -W .github/workflows/test-and-build.yml
```

### 4. Commit and Push

```bash
git add .github/workflows/*.yml .golangci.yml .codecov.yml
git commit -m "feat(ci): add GitHub Actions CI/CD workflows

- Add test-and-build workflow (tests, build, Docker, lint)
- Add database migration validation workflow
- Add security scanning workflow (govulncheck, Dependabot, secrets)
- Add PR preview comments workflow
- Add golangci-lint configuration
- Add Codecov configuration

All workflows complete in < 5 minutes
Provides fast feedback for AI agents and developers

Implements .docs/issues/002-copilot-cicd-setup.md

🤖 Generated with Copilot assistance

Co-Authored-By: GitHub Copilot <noreply@github.com>"

git push origin feature/002-cicd-setup
```

### 5. Create Pull Request

```bash
gh pr create --title "feat(ci): add CI/CD workflows" --body "
## Summary
Complete CI/CD pipeline with automated testing, building, linting, and security scanning.

## Implementation
Implements \`.docs/issues/002-copilot-cicd-setup.md\`

## Workflows Added
- ✅ Test and Build (tests, build, Docker, lint)
- ✅ Validate Migrations (database schema verification)
- ✅ Security Scan (vulnerabilities, dependencies, secrets)
- ✅ PR Preview (automated comments on PRs)

## Acceptance Criteria
- [x] All workflow files created
- [x] Linter configuration added
- [x] Codecov configuration added
- [x] Workflows trigger on appropriate events
- [x] All checks pass in < 5 minutes
- [x] PR comments work

## Testing
Created test PR and verified all workflows pass.

## Next Steps
After merge, Aider will implement Issue #003 (Portal Authentication).
"
```

### 6. Watch Actions Tab

Go to GitHub Actions tab and verify all workflows run successfully.

### 7. After Merge

```bash
git checkout development
git pull origin development
git branch -d feature/002-cicd-setup
```

---

## Context and References

- **ARCHITECTURE.md** - See "Development Workflow" section
- **DevSmithTDD.md** - CI enforces 70%+ test coverage
- **Issue #002** - Completes infrastructure setup

---

## Common Issues and Solutions

### Issue: PostgreSQL not ready in time

**Solution:** Increase health check retries in workflow:
```yaml
options: >-
  --health-cmd pg_isready
  --health-interval 10s
  --health-timeout 5s
  --health-retries 10  # Increased from 5
```

### Issue: Go module cache not working

**Solution:** Ensure `cache: true` in `setup-go` action:
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.22'
    cache: true  # Important!
```

### Issue: Docker build too slow

**Solution:** Use BuildKit cache:
```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

### Issue: Linter finds too many issues

**Solution:** Add exceptions to `.golangci.yml`:
```yaml
issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

---

## Success Metrics

This issue is complete when:

1. ✅ All 4 workflows created and configured
2. ✅ Workflows run successfully on push/PR
3. ✅ Test coverage reported (even if 0%)
4. ✅ All services build successfully
5. ✅ Linter runs without errors
6. ✅ Database validation passes
7. ✅ PR comments appear automatically
8. ✅ All workflows complete in < 5 minutes
9. ✅ Status badges can be added to README

---

## Next Steps After Completion

1. Merge this PR to `development` branch
2. Add status badges to README.md
3. Move to Issue #003: Portal Authentication (Aider implementation)
4. CI will automatically validate all future PRs
