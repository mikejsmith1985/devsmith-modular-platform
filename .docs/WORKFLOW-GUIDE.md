# DevSmith Development Workflow Guide

**Purpose:** Step-by-step guide for working with issues, branches, commits, PRs, and CI/CD.

---

## Core Principle

**One Issue = One Branch = One Feature**

- Each GitHub issue gets its own feature branch
- Branch contains changes for that issue ONLY
- No mixing of multiple features on one branch
- Different agents can work on different issues in parallel (separate windows)

---

## Working on an Issue (Step-by-Step)

### Step 1: Pick an Issue

Check `.docs/issues/` for your next issue:
```bash
ls -la .docs/issues/
```

**Example:**
- `001-copilot-project-scaffolding.md` → Issue #001
- `002-copilot-cicd-setup.md` → Issue #002 for Copilot
- `002-openhands-portal-authentication.md` → Issue #002 for OpenHands

### Step 2: Create Feature Branch

```bash
# From development branch
git checkout development
git pull origin development

# Create feature branch (naming: feature/###-brief-description)
git checkout -b feature/001-copilot-project-scaffolding
```

**Branch Naming Convention:**
- `feature/###-brief-description` (e.g., `feature/001-copilot-project-scaffolding`)
- `###` = Issue number (left-padded with zeros to 3 digits)
- Brief description in kebab-case

### Step 3: Work on the Issue

**For Copilot:**
1. Open issue file in Copilot chat
2. Copilot implements according to spec
3. Copilot should commit after each logical change (but may need reminders)

**For Aider:**
1. Read latest devlog entry first: `.docs/devlog/YYYY-MM-DD.md`
2. Point Aider to issue spec with line numbers
3. Let Aider work autonomously
4. Monitor progress periodically

**For Claude:**
1. Read issue and codebase
2. Provide architecture guidance
3. Create/update specifications
4. Review and validate changes
5. Update devlog at end of session

### Step 4: Check Your Work

Before committing, verify what changed:

```bash
# See all modified files
git status

# See all changes in detail
git diff

# See staged changes only
git diff --cached
```

**Common Issues to Check:**
- Are there files from OTHER issues? (workflow violation!)
- Are there sensitive files? (.env, credentials, etc.)
- Are there debug/temp files? (*.log, *.tmp, etc.)

### Step 5: Stage and Commit Changes

**Commit Early and Often!** Don't wait until everything is done.

```bash
# Stage specific files (preferred)
git add path/to/file1 path/to/file2

# OR stage all changes (use carefully)
git add .

# Commit with conventional commit message
git commit -m "$(cat <<'EOF'
feat(scope): brief description

Longer explanation of what changed and why.
Include context that would help reviewers.

Bullet points if multiple changes:
- First change
- Second change
- Third change

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

**Conventional Commit Format:**
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, config, etc.)
- `ci`: CI/CD changes

**Scopes (examples):**
- `infra`: Infrastructure (Docker, nginx, etc.)
- `portal`: Portal service
- `review`: Review service
- `auth`: Authentication
- `api`: API changes
- `db`: Database schemas/migrations

### Step 6: View Commit History

```bash
# See recent commits (one line each)
git log --oneline -10

# See detailed commit with files changed
git show HEAD

# See specific commit
git show <commit-hash>

# See all commits for this branch (compared to development)
git log development..HEAD
```

---

## Creating a Pull Request

### Step 1: Push Your Branch

```bash
# First push (sets upstream)
git push -u origin feature/001-copilot-project-scaffolding

# Subsequent pushes (after more commits)
git push
```

### Step 2: Create PR via GitHub Web Interface

1. Go to: https://github.com/your-org/DevSmith-Modular-Platform
2. Click "Pull requests" tab
3. Click "New pull request" button
4. Set:
   - **Base:** `development` (NOT main!)
   - **Compare:** `feature/001-copilot-project-scaffolding`
5. Click "Create pull request"
6. Fill in title and description
7. Click "Create pull request"

**PR Title Format:**
```
[Issue #001] Project Scaffolding - Complete Docker Infrastructure
```

**PR Description Template:**
```markdown
## Issue
Closes #001

## Summary
Brief description of what this PR does (2-3 sentences).

## Changes
- First major change
- Second major change
- Third major change

## Testing
- [ ] Local testing completed
- [ ] All services start successfully
- [ ] Health checks pass
- [ ] Tests pass locally

## Checklist
- [ ] Follows conventional commit format
- [ ] Updated relevant documentation
- [ ] No sensitive data committed
- [ ] Devlog updated
```

### Step 3: Wait for CI/CD Checks

GitHub Actions will automatically run when you create the PR:

**Workflows that run:**
1. **Test and Build** (`.github/workflows/test-and-build.yml`)
   - Runs `make test`
   - Runs `make build`
   - Checks code coverage

2. **Validate Migrations** (`.github/workflows/validate-migrations.yml`)
   - Checks database migration files
   - Ensures migrations are idempotent

3. **Security Scan** (`.github/workflows/security-scan.yml`)
   - Runs `gosec` for security issues
   - Scans dependencies for vulnerabilities

4. **PR Preview** (`.github/workflows/pr-preview.yml`)
   - Builds Docker images
   - Runs docker-compose up
   - Verifies all services healthy

### Step 4: Check CI Results

On the PR page, scroll down to see checks:

- ✅ **Green check** = Passed
- ❌ **Red X** = Failed
- 🟡 **Yellow circle** = In progress

**If checks fail:**
1. Click on the failed check name
2. Click "Details" to see logs
3. Identify the error
4. Fix the issue locally
5. Commit and push the fix
6. CI will re-run automatically

### Step 5: Review and Merge

**Claude's Role:**
1. Review the PR for architecture compliance
2. Validate against issue acceptance criteria
3. Check commit messages follow conventions
4. Approve PR if all looks good

**Mike's Role:**
1. Final review and approval
2. Merge PR via GitHub interface
3. Delete feature branch after merge

**Merge Process:**
```bash
# On GitHub PR page:
1. Click "Squash and merge" button
2. Edit commit message if needed
3. Click "Confirm squash and merge"
4. Click "Delete branch" button
```

---

## Multi-Window Workflow (Parallel Development)

You can have multiple agents working on different issues simultaneously:

**Window 1: Issue #001 (Copilot)**
```bash
git checkout feature/001-copilot-project-scaffolding
# Work on Issue #001
```

**Window 2: Issue #002 (Copilot)**
```bash
git checkout feature/002-copilot-cicd-setup
# Work on Issue #002
```

**Window 3: Issue #002 (Aider)**
```bash
git checkout feature/002-openhands-portal-authentication
# Work on different Issue #002
```

**CRITICAL RULES:**
- Each window stays on its own branch
- Never switch branches while agent is working
- Don't mix changes from different branches
- Commit in each window independently

---

## Reviewing Commits (Teaching Yourself)

### Basic Review

```bash
# See what files changed
git show --stat HEAD

# See the actual changes
git show HEAD

# See multiple commits
git log -p -3  # Shows last 3 commits with diffs
```

### Detailed Review Checklist

**1. Commit Message Quality**
```bash
git log --oneline -10
```

Check:
- [ ] Follows conventional commit format? (type(scope): description)
- [ ] Description is clear and concise?
- [ ] Body explains "why" not "what"?
- [ ] References issue number if applicable?

**2. Code Changes Review**
```bash
git show HEAD
```

Check:
- [ ] Changes match commit message description?
- [ ] No unrelated changes included?
- [ ] No commented-out code left behind?
- [ ] No debug print statements?
- [ ] No TODO comments without issue references?

**3. File Types Review**
```bash
git show --stat HEAD
```

Check:
- [ ] No sensitive files? (.env, secrets, credentials)
- [ ] No binary files? (unless intentional like images)
- [ ] No generated files? (vendor/, node_modules/)
- [ ] No IDE-specific files? (.vscode/, .idea/)

**4. Diff Size Review**

```bash
git show --stat HEAD
```

Check:
- [ ] Commit size reasonable? (< 500 lines ideal)
- [ ] If large, can it be split into multiple commits?
- [ ] Changes focused on one logical change?

### Review All Commits Before PR

```bash
# See all commits in this branch
git log development..HEAD

# Review each commit
git show <commit-hash>

# Or review all at once
git log -p development..HEAD | less
```

---

## Common Git Commands Reference

### Status and Info
```bash
git status                    # What's changed
git log --oneline -10        # Recent commits
git branch                   # List branches
git branch -a                # List all branches (including remote)
git diff                     # Unstaged changes
git diff --cached            # Staged changes
git show HEAD                # Last commit details
```

### Branching
```bash
git checkout development              # Switch to development
git checkout -b feature/001-thing     # Create and switch to new branch
git branch -d feature/001-thing       # Delete local branch (after merge)
git push origin --delete feature/001  # Delete remote branch
```

### Staging and Committing
```bash
git add file1 file2           # Stage specific files
git add .                     # Stage all changes
git reset HEAD file1          # Unstage a file
git commit -m "message"       # Commit staged changes
git commit --amend            # Modify last commit (use carefully!)
```

### Pushing and Pulling
```bash
git push                              # Push commits to remote
git push -u origin feature/001        # First push (sets upstream)
git pull                              # Pull changes from remote
git fetch origin                      # Fetch without merging
```

### Undoing Changes
```bash
git restore file.txt          # Discard uncommitted changes
git restore --staged file.txt # Unstage file
git reset --soft HEAD~1       # Undo last commit, keep changes
git reset --hard HEAD~1       # Undo last commit, discard changes (DANGEROUS!)
```

---

## Troubleshooting

### Problem: Mixed Changes from Multiple Issues

**Symptom:**
```bash
git status
# Shows files from Issue #001 AND Issue #002
```

**Solution:**
```bash
# Option 1: Commit everything together (simple but violates workflow)
git add .
git commit -m "feat: complete multiple changes"

# Option 2: Commit selectively (preferred)
git add file1 file2 file3  # Only Issue #001 files
git commit -m "feat(issue-001): complete scaffolding"

git add file4 file5        # Only Issue #002 files
git commit -m "feat(issue-002): add CI/CD"
```

### Problem: Accidentally on Wrong Branch

**Symptom:**
```bash
git branch
# * development  ← Should be on feature branch!
```

**Solution:**
```bash
# Move uncommitted changes to correct branch
git stash                           # Save changes
git checkout feature/001-thing      # Switch to correct branch
git stash pop                       # Restore changes
```

### Problem: Commit to Wrong Branch

**Symptom:**
```bash
git log --oneline -1
# abc1234 feat: my change
git branch
# * development  ← Oops! Should be on feature branch
```

**Solution:**
```bash
# Move commit to correct branch
git checkout feature/001-thing         # Switch to correct branch
git cherry-pick abc1234                # Copy commit here
git checkout development               # Go back
git reset --hard HEAD~1                # Remove commit from development
```

### Problem: Need to Fix PR After Review

**Symptom:** Reviewer requested changes on your PR.

**Solution:**
```bash
# Make the fixes locally
git checkout feature/001-thing
# ... edit files ...
git add .
git commit -m "fix(review): address PR feedback"
git push
# PR will automatically update with new commit
```

### Problem: Merge Conflicts

**Symptom:**
```bash
git pull origin development
# CONFLICT (content): Merge conflict in file.txt
```

**Solution:**
```bash
# Open file.txt and look for conflict markers:
# <<<<<<< HEAD
# Your changes
# =======
# Their changes
# >>>>>>> development

# Edit file to resolve conflict, then:
git add file.txt
git commit -m "fix: resolve merge conflict"
```

---

## Quick Reference Card

### Starting New Issue
```bash
git checkout development
git pull origin development
git checkout -b feature/###-description
# Work on issue
git add .
git commit -m "feat(scope): description"
git push -u origin feature/###-description
# Create PR on GitHub
```

### Checking Your Work
```bash
git status                  # What changed?
git diff                    # Show changes
git log --oneline -5        # Recent commits
git show HEAD               # Last commit details
```

### Reviewing Before PR
```bash
git log development..HEAD           # All commits in this branch
git diff development...HEAD         # All changes vs development
git show --stat HEAD                # Files changed in last commit
```

### After PR Merged
```bash
git checkout development
git pull origin development
git branch -d feature/###-thing     # Delete local branch
```

---

## Tips for Success

1. **Commit early, commit often** - Don't wait until everything is done
2. **Read your diffs before committing** - `git diff` is your friend
3. **Write clear commit messages** - Future you will thank you
4. **One branch per issue** - Never mix features
5. **Keep branches short-lived** - Merge within 1-2 days if possible
6. **Pull development frequently** - Stay up to date
7. **Run tests before pushing** - `make test` should pass
8. **Review your own PR first** - Catch obvious issues before others review
9. **Update devlog at end of session** - Helps next session pick up context
10. **Don't force push to shared branches** - Especially main/development

---

## Learning Resources

**Understanding Git:**
- `git help <command>` - Built-in help (e.g., `git help commit`)
- https://git-scm.com/docs - Official Git documentation

**Conventional Commits:**
- https://www.conventionalcommits.org/

**GitHub Flow:**
- https://guides.github.com/introduction/flow/

**DevSmith-Specific:**
- `.docs/devlog/README.md` - Devlog system guide
- `ARCHITECTURE.md` - System architecture
- `DevSmithRoles.md` - Agent roles and workflow
- `.docs/issues/` - Issue specifications

---

## When to Ask for Help

**Ask Claude when:**
- Unsure about architecture decisions
- Need specification clarified
- Want commit message reviewed
- Need help with git commands
- PR review feedback unclear

**Ask Copilot when:**
- Implementation details unclear
- Need code examples
- Want to explore alternatives

**Check devlog when:**
- Starting new session
- Wondering what happened before
- Looking for past decisions
- Need context on issues

---

**Last Updated:** 2025-10-19
**Version:** 1.0
