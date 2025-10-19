# DevLog Directory

## Purpose

The devlog (development log) tracks **human-readable session summaries** of work done by all agents (Claude, Aider, Copilot) and team members. Unlike automated recovery logs or session logs, devlogs capture **decisions, rationale, and context** for future reference.

---

## Why Devlogs?

**Problem:** AI agents don't remember previous sessions. Each time you start a new session, the agent has no context about:
- What decisions were made and why
- What problems were discovered
- What solutions were implemented
- What action items are pending

**Solution:** Devlogs provide **shared memory** across sessions and agents.

---

## When to Update

### Claude Code Sessions
**Trigger:** End of every session (especially architecture, planning, or review sessions)

**What to log:**
- Problems discovered during validation/review
- Architectural decisions made
- Specifications created
- Issues restructured or updated
- Action items for next agent (Copilot, Aider, OpenHands)

### Aider Sessions
**Trigger:** Before starting (read latest entry) and after completing (update entry)

**What to log:**
- Implementation notes
- Issues encountered and how they were solved
- Test results
- Files created/modified
- Deviations from spec (if any)

### Copilot Sessions
**Trigger:** After completing scaffolding or boilerplate work

**What to log:**
- What was scaffolded
- Any gaps or issues
- Files created
- Next steps

### Mike (Manual Updates)
**Trigger:** When making decisions outside agent sessions

**What to log:**
- Manual fixes or changes
- Sprint planning decisions
- Priority changes
- Configuration changes (Ollama models, GitHub settings, etc.)

---

## File Naming

**Format:** `YYYY-MM-DD.md`

**Examples:**
- `2025-10-19.md`
- `2025-10-20.md`

**Multiple Sessions Per Day:**
Use session numbers in the file:
```markdown
# DevLog - October 19, 2025

## Session 1: Issue Validation
[content]

## Session 2: PR Review
[content]
```

---

## How Each Agent Uses Devlogs

### 1. Claude Code (Architecture & Review)

**Before starting session:**
```markdown
Claude should read the latest devlog entry to understand:
- Recent decisions
- Current blockers
- Action items assigned to it
```

**During session:**
- Note problems discovered
- Document decisions as they're made
- Track action items for other agents

**At end of session:**
- Create or update today's devlog
- Use `TEMPLATE.md` as guide
- Include clear action items for next agent

**Example prompt to Claude:**
> "Before we start, read `.docs/devlog/2025-10-19.md` to understand what happened last session."

### 2. Aider (Implementation)

**Before starting session:**
Aider should be given this prompt:
```
Read the latest devlog in .docs/devlog/ to understand:
1. What action items are assigned to you
2. What problems were discovered that might affect your work
3. What decisions were made about architecture

Then implement [task description].
```

**After completing work:**
```
Update today's devlog entry with:
- Implementation notes
- Any issues you encountered and how you solved them
- Test results
- Files you created/modified
```

### 3. Copilot (IDE Assistance)

**Indirect usage:**
Mike reads the devlog to understand what Copilot should do, then instructs Copilot accordingly.

After Copilot completes work, Mike or Claude updates the devlog.

---

## Template Usage

Copy from `TEMPLATE.md` when creating new entries:

```bash
cp .docs/devlog/TEMPLATE.md .docs/devlog/2025-10-20.md
# Edit the new file
```

---

## What NOT to Log

❌ **Don't log:**
- Automated tool output (that's what recovery logs are for)
- Line-by-line code changes (that's what git is for)
- Complete file contents (use file paths instead)
- Routine operations with no decisions

✅ **Do log:**
- Why a decision was made
- Problems that required investigation
- Lessons learned
- Context for future sessions

---

## Example Workflow

### Scenario: Claude creates spec, then Aider implements it

**Step 1 - Claude Session:**
```markdown
# DevLog - October 20, 2025

## Session 1: Portal Authentication Spec

**Summary:** Created detailed implementation spec for GitHub OAuth.

**Decisions Made:**
1. Use JWT tokens (not sessions)
   - Rationale: Stateless, works with multiple services
   - Impact: Each service validates tokens independently

**Action Items for Aider:**
- [ ] Implement AuthService with JWT generation
- [ ] Create /auth/github/callback handler
- [ ] Add User model with GitHub fields
- [ ] Write tests for OAuth flow
```

**Step 2 - Aider Session (Next Day):**
Mike starts Aider with:
```bash
aider

# In Aider:
/add .docs/devlog/2025-10-20.md
/add .docs/specs/portal-auth.md

Read the devlog to see your action items, then implement the Portal authentication spec.
```

Aider reads the devlog, implements the feature, then updates:
```markdown
## Session 2: Portal Authentication Implementation

**Status:** ✅ Complete

**Implementation Notes:**
- Created AuthService using golang-jwt library
- Implemented callback handler with proper error handling
- Added 15 tests (100% coverage on auth flow)

**Issues Encountered:**
1. GitHub API rate limiting in tests
   - Solution: Added mock GitHub client for tests
   - File: `internal/portal/services/github_client_mock.go`

**Files Created:**
- `internal/portal/services/auth_service.go`
- `internal/portal/services/auth_service_test.go`
- `cmd/portal/handlers/auth_handler.go`
[... etc ...]
```

---

## Benefits

✅ **Continuity:** Each agent knows what happened before
✅ **Context:** Decisions are explained, not just recorded
✅ **Debugging:** Easy to trace when/why something changed
✅ **Onboarding:** New team members can read history
✅ **Audit Trail:** Complete record of project evolution

---

## Quick Reference

| Agent | When to Read | When to Write |
|-------|--------------|---------------|
| Claude | Start of session | End of session |
| Aider | Before implementation | After completion |
| Copilot | (via Mike) | (via Mike/Claude) |
| Mike | Anytime | When making manual decisions |

---

## Related Files

- **Recovery Logs:** `.claude/recovery-logs/` (automatic crash recovery)
- **Session Logs:** `.claude/hooks/session-logger.sh` (automated event log)
- **Specs:** `.docs/specs/` (implementation specifications)
- **Issues:** `.docs/issues/` (detailed issue specs)
- **Architecture:** `ARCHITECTURE.md` (system design)

---

## Getting Started

1. **Copy template for today:**
   ```bash
   cp .docs/devlog/TEMPLATE.md .docs/devlog/$(date +%Y-%m-%d).md
   ```

2. **Fill in the sections** as you work

3. **At end of session:** Make sure action items are clear for next agent

4. **Start next session:** Read the latest entry first!

---

**Remember:** Devlogs are for **humans and future agents**. Write clearly and explain *why*, not just *what*.
