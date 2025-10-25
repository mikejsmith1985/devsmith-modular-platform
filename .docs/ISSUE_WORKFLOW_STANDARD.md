# Issue Workflow Standard Process

## Overview
This document establishes the standard workflow for implementing GitHub issues using TDD and ensuring proper tracking through GitHub's issue system.

## Workflow Steps

### 1. Issue Analysis & Planning
- [ ] Read the full issue requirements and acceptance criteria
- [ ] Check for referenced documentation (ARCHITECTURE.md, TDD docs, etc.)
- [ ] Break requirements into TDD phases: RED → GREEN → REFACTOR
- [ ] Identify all acceptance criteria that must be met

### 2. Implementation (TDD Phases)
- [ ] **RED Phase:** Write failing tests for requirements
- [ ] **GREEN Phase:** Implement code to make tests pass
- [ ] **REFACTOR Phase:** Improve code quality, error handling, documentation
- [ ] Follow project coding standards and conventions

### 3. Quality Assurance
- [ ] Run all tests locally
- [ ] Verify coverage meets or exceeds requirements
- [ ] Run pre-commit checks (fmt, vet, lint, coverage, security)
- [ ] Test integration with existing systems
- [ ] Document any design decisions

### 4. Git Workflow
- [ ] Create feature branch: `git checkout -b feature/issue-{NUMBER}-{DESCRIPTION}`
- [ ] Commit using conventional commits: `feat(issue-{NUMBER}): description`
- [ ] Push feature branch to GitHub
- [ ] Create PR with comprehensive description including:
  - Issue number
  - Acceptance criteria checklist
  - Test results and coverage
  - Any breaking changes

### 5. Issue & PR Linking (NEW STANDARD)
- [ ] **Update the GitHub issue** with completion details:
  - Check off all acceptance criteria (✅ marks)
  - Add implementation summary
  - Link to PR number
  - Note any metrics (coverage %, tests passed, etc.)
  
- [ ] **Add issue comment** with status update:
  - Link to PR
  - Summarize test results
  - List what this unblocks
  - Next steps
  
- [ ] **Link PR to issue:**
  - Reference issue in PR body: "Closes #20"
  - Use `gh pr edit` to ensure linkage
  
- [ ] **Close the issue** once PR is created:
  ```bash
  gh issue close {ISSUE_NUMBER}
  ```
  The PR will reopen it automatically if needed during review.

### 6. Code Review & Merge
- [ ] Request review from team
- [ ] Address review comments
- [ ] Ensure all checks pass
- [ ] Merge to development branch
- [ ] Verify CI/CD passes

### 7. Documentation
- [ ] Update any architecture documentation
- [ ] Add inline code comments for complex logic
- [ ] Update README if adding new features
- [ ] Document API changes if applicable

## Acceptance Criteria Checklist Template

When updating issues, use this format for acceptance criteria:

```markdown
## ✅ Acceptance Criteria - ALL MET

- [x] First criterion
  - Implementation detail
  - Implementation detail

- [x] Second criterion
  - Implementation detail

- [x] Quality metrics
  - Coverage: X%
  - Tests: Y/Z passing
  - Quality checks: all pass
```

## PR Description Template

```markdown
## Issue
Review #{NUMBER} - [Title]

## Implementation
- Feature 1: Description
- Feature 2: Description

## Test Results
- Unit Tests: X/Y passing
- Integration Tests: A/B passing
- Coverage: Z%

## Quality
✅ fmt pass
✅ vet pass
✅ lint pass
✅ All security checks pass

## Acceptance Criteria Met
- [x] Criterion 1
- [x] Criterion 2
- [x] Criterion 3

Closes #{NUMBER}
```

## GitHub CLI Commands Quick Reference

```bash
# View issue details
gh issue view {NUMBER}

# Update issue description
gh issue edit {NUMBER} --body "new body"

# Add comment to issue
gh issue comment {NUMBER} --body "comment text"

# Close issue
gh issue close {NUMBER}

# Create PR linked to issue
gh pr create --base development --head feature-branch --title "Title" --body "Closes #{NUMBER}"

# View PR
gh pr view {NUMBER}
```

## Standard Metrics to Track

For each issue, document:
- ✅ **Test Coverage:** Report actual % achieved
- ✅ **Tests Passing:** List number of integration/unit tests
- ✅ **Quality Checks:** fmt, vet, lint status
- ✅ **Acceptance Criteria:** All checkboxes completed
- ✅ **Blockers Unblocked:** List what this enables

## Common Mistakes to Avoid

❌ **Creating PR without updating the issue**
- Always update the issue with completion status

❌ **Forgetting to link PR to issue**
- Use "Closes #NUMBER" in PR description
- Update issue with PR link in comments

❌ **Not documenting acceptance criteria completion**
- Check off each criterion with implementation details
- Add metrics and test results

❌ **Closing issue before PR review**
- Close issue only after PR is created and linked
- GitHub will handle re-opening if needed

✅ **CORRECT PROCESS:**
1. Implement feature with TDD
2. Create feature branch & commit
3. Push branch to GitHub
4. Update GitHub issue with completion details
5. Create PR linked to issue
6. Add completion comment to issue
7. Close issue
8. Request review and merge

## Example Workflow

```bash
# 1. Create feature branch
git checkout -b feature/review-020-database-persistence

# 2. Implement and test (TDD cycle)
# ... RED-GREEN-REFACTOR ...
go test ./...

# 3. Commit work
git add -A
git commit -m "feat(review-020): database persistence layer

- ReadingSessionRepository CRUD
- CriticalIssuesRepository CRUD
- 9 integration tests passing
- 71.3% coverage"

# 4. Push branch
git push origin feature/review-020-database-persistence

# 5. Update issue
gh issue edit 20 --body "$(cat updated-body.md)"
gh issue comment 20 --body "Implementation complete. See PR #40"

# 6. Create PR
gh pr create --base development \
  --head feature/review-020-database-persistence \
  --title "Review #20 - Database Persistence" \
  --body "Closes #20"

# 7. Close issue
gh issue close 20

# 8. Request review (via GitHub UI or)
gh pr edit 40 --add-reviewer @reviewer

# 9. Merge when approved
gh pr merge 40 --squash
```

## Team Agreement

✅ This workflow is now the **STANDARD PROCESS** for all GitHub issues

✅ All future issues will follow this structure

✅ PR descriptions will be comprehensive with acceptance criteria

✅ Issues will be properly closed/linked to track progress
