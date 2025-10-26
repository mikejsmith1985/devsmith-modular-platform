# PR-to-Issue Linking & Auto-Closing Guide

## Overview

GitHub automatically links PRs to issues and **auto-closes issues when PRs are merged**, but ONLY if you use the correct syntax in your PR description.

This document explains how to do it correctly.

## The One Required Thing

In your PR description, include this line:

```
Closes #36
```

Replace `36` with your actual issue number.

## Valid Keywords

GitHub recognizes these keywords (case-insensitive):

```
✓ Closes / Close / Closed
✓ Fixes / Fixed / Fix
✓ Resolves / Resolved / Resolve
```

All of these work the same way:
- Link PR to Issue
- Auto-close Issue when PR merges

Pick one and use it consistently.

## ✅ Correct Examples

```
Closes #36

Fixes #42

Resolves #50
```

## ❌ WRONG - These Don't Work

```
Feature #36           ❌ (no keyword)
Issue 36              ❌ (no # symbol)
#36                   ❌ (no keyword)
Closes Feature #36    ❌ (extra words)
Closes: #36           ❌ (colon after keyword)
```

## How It Works

### Step 1: PR Created with Closes #36
```
PR Description:
  Closes #36
  
  Implementation of log retention feature...
```

### Step 2: Automatic Linking
- GitHub sees "Closes #36"
- PR becomes linked to Issue #36
- Visible in Issue #36's UI: "PR #50 wants to merge and close this"

### Step 3: PR Merged
- Developer/Reviewer merges PR #50
- GitHub automatically closes Issue #36
- Both PR and Issue show as completed

## Where to Put It

**Location in PR:** Near the top in "Issue Reference" section

**PR Template Location:** `.github/PULL_REQUEST_TEMPLATE/pull_request_template.md`

```markdown
## 🎯 Issue Reference (REQUIRED)

**Closes #<!-- issue number here -->**
```

## Validation

A GitHub Actions workflow validates every PR:

**Workflow File:** `.github/workflows/validate-pr-issue-link.yml`

**What It Checks:**
✓ PR description contains valid "Closes #XXX" or similar
✓ Issue number format is correct
✓ Fails PR check if linking is missing

**If Linking is Missing:**
- ❌ GitHub Actions check fails
- Message shows: "PR must contain 'Closes #XXX' or similar"
- PR cannot be merged until fixed

## Quick Checklist

Before creating a PR:

- [ ] Read your GitHub issue number (e.g., #36)
- [ ] In PR description, add: `Closes #36` (replace 36)
- [ ] Make sure "Closes" is at start of line or after punctuation
- [ ] No extra words before/after issue number
- [ ] Use "Closes", "Fixes", or "Resolves" (pick one)

## Troubleshooting

### "PR doesn't have issue link" message

**Problem:** Workflow is complaining about missing link

**Solution:**
1. Check your PR description
2. Look for "Closes #XXX" line
3. Verify format exactly matches (case doesn't matter, but spacing does)
4. If not there, edit PR and add it

### Issue didn't close automatically

**Problem:** You merged PR but issue stayed open

**Possible causes:**
- ❌ PR description didn't have "Closes #XXX"
- ❌ Issue number was wrong (e.g., #99 when you meant #36)
- ❌ Format was wrong (e.g., "Closes: #36" with colon)

**Manual fix:** Close issue manually and add comment linking to PR

### Issue exists but linking still fails

**Problem:** Issue #36 exists, PR has "Closes #36", but linking won't work

**Common causes:**
- PR is against wrong branch (must target `development`)
- Issue was deleted/archived
- GitHub sync delay (wait 30 seconds and refresh)

**Solution:**
1. Verify target branch is `development` (not `main`)
2. Verify issue exists: `gh issue view 36`
3. Verify PR description has "Closes #36" exactly
4. Check GitHub Actions workflow: did it pass?

## Automation Benefits

With proper linking:

✅ **Developers:**
- Issues auto-close (no manual cleanup)
- PR-Issue relationship visible everywhere
- Easy to track what's in progress

✅ **Reviewers (Claude):**
- Automatic connection between code and requirements
- Can see all related PRs on issue
- Can see issue context when reviewing PR

✅ **Project Management:**
- Velocity tracking (closed issues)
- Release notes (all closed issues since v1.0)
- Dependency mapping (which issues depend on others)

## GitHub Documentation

For more details, see [GitHub's guide on linking issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue)

Key points from GitHub:
- Linking only works on PRs to the same repository
- Linking happens when PR is created, not merged
- Auto-closing happens when PR is merged
- All 9 keywords work the same (closes, fixes, resolves, etc.)

## Quick Test

Test the linking system:

```bash
# 1. Create a test issue
gh issue create --title "Test Issue" --body "This is a test"
# Returns: Issue #XXX

# 2. Create a test branch
git checkout -b test-linking

# 3. Make trivial change (e.g., update README)
echo "Test" >> README.md
git add README.md
git commit -m "test: linking test"

# 4. Create PR with Closes
gh pr create \
  --title "Test PR" \
  --body "Closes #XXX" \
  --base development \
  --head test-linking

# 5. Watch GitHub Actions run validation workflow
# Should pass because PR has valid "Closes #XXX"

# 6. Merge PR (if tests pass)
gh pr merge --squash

# 7. Verify: Issue #XXX should now be CLOSED
gh issue view #XXX
```

## Reference

- **Template:** `.github/PULL_REQUEST_TEMPLATE/pull_request_template.md`
- **Workflow:** `.github/workflows/validate-pr-issue-link.yml`
- **Current PRs:** `gh pr list`
- **Current Issues:** `gh issue list`

---

**Last Updated:** 2025-10-26
**Status:** Active - All new PRs must follow this linking format

