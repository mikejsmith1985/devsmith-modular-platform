## Purge sensitive env files from repository history

### Summary
This PR rewrites repository history to remove the tracked sensitive files `.env.playwright` and `.env.test`, and updates `.gitignore` plus example env files.

### Actions taken
- Created a backup branch `backup-before-purge` and a bundle `backup-before-purge.bundle` before any destructive changes.
- Untracked `.env.playwright` and `.env.test` and added `/.env.*` ignore patterns.
- Added `.env.playwright.example` and `.env.test.example` to provide safe templates.
- Rewrote history (used `git filter-branch` fallback) to remove the sensitive files from all commits and force-pushed rewritten branches and tags to `origin`.
- Cleaned up temporary refs and ran `git gc --prune=now --aggressive`.

### Verification performed
- Confirmed `.env.playwright` and `.env.test` do not appear in `git log --all --name-only`; only the example files remain present in history.

### Important notes for reviewers and the team
- This is a history rewrite. All collaborators must re-sync their local clones. The easiest option for most users is to re-clone the repository.
- If those files contained real secrets, those credentials must be rotated immediately (API keys, tokens, CI secrets, etc.).
- Open PRs may need to be recreated or rebased against the rewritten branches.

### Suggested next steps for the team
1. Rotate any exposed credentials (list service tokens, CI variables, etc.).
2. Team members re-clone or reset their local clones:
```bash
# simplest (recommended)
git clone https://github.com/mikejsmith1985/devsmith-modular-platform.git

# advanced (if you must keep an existing clone — careful: this destroys uncommitted work)
git fetch origin --prune
git checkout development
git reset --hard origin/development
```
3. Verify CI status and reopen/recreate any PRs that show conflicts.

---

If you'd like, I can also (A) add a short Slack/GitHub message for teammates, or (B) list likely secrets to rotate (common places: GitHub Actions secrets, third-party API keys, cloud creds).