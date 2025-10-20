# Issue #XXX: [COPILOT] <Feature Name>

**Labels:** `copilot`, `<service>`, `<category>`
**Assignee:** Mike (with Copilot assistance)
**Created:** YYYY-MM-DD
**Issue:** #XXX
**Estimated Complexity:** Low/Medium/High
**Target Service:** <service>
**Estimated Time:** XX-XX minutes
**Depends On:** Issue #XXX (if applicable)

---

# 🚨 CRITICAL: FIRST STEP - CREATE FEATURE BRANCH 🚨

**⚠️ DO NOT PROCEED UNTIL YOU COMPLETE THIS STEP ⚠️**

## STEP 0: Verify and Create Feature Branch

**BEFORE doing ANYTHING else (reading specs, planning, writing code, or writing tests):**

### 1. Check Current Branch

Run this command FIRST:
```bash
git branch --show-current
```

**Expected output:** `development`

**If you see anything else (like a feature branch), STOP!**
- You may be continuing work on an existing branch
- Or you're on the wrong branch
- Double-check with the user before proceeding

### 2. Verify Branch is Clean

```bash
git status
```

**Expected output:** `nothing to commit, working tree clean`

**If you see uncommitted changes, STOP!**
- Commit or stash changes before creating feature branch
- Ask the user how to proceed

### 3. Update Development Branch

```bash
git checkout development
git pull origin development
```

This ensures you're branching from the latest code.

### 4. Create Feature Branch

```bash
git checkout -b feature/XXX-<brief-description>
```

**Example:** `git checkout -b feature/005-copilot-logs-service-foundation`

### 5. Verify You're On Feature Branch

```bash
git branch --show-current
```

**Expected output:** `feature/XXX-<brief-description>`

**✅ CHECKPOINT: Only proceed if you see the feature branch name.**

---

# ⚠️ STOP! READ THIS BEFORE CODING! ⚠️

**Now that you're on the correct branch, read these critical reminders:**

**DO NOT work on the `development` branch directly!**

If you started coding/planning/testing on `development`, you will break the workflow and have to redo your work.

**Workflow Order:**
1. ✅ Create feature branch (you just did this above)
2. ✅ Read this entire spec
3. ✅ Plan implementation phases
4. ✅ Write tests (TDD)
5. ✅ Implement code
6. ✅ Commit after EACH phase
7. ✅ Push regularly

---

## Task Description

<Brief description of what this issue accomplishes>

**Why This Task for Copilot:**
- <Reason 1>
- <Reason 2>
- <Reason 3>

---

## IMPORTANT: Commit As You Go

**DO NOT wait until everything is done to commit!**

After completing each PHASE (see Implementation Checklist below):
1. Test that phase: `go test ./...`
2. If tests pass, commit that phase:
   ```bash
   git add <files-for-that-phase>
   git commit -m "feat(<scope>): <brief description of phase>"
   ```

**Example commits:**
```bash
# After Phase 1
git add internal/<service>/models/
git commit -m "feat(<service>): add data models"

# After Phase 2
git add internal/<service>/db/
git commit -m "feat(<service>): implement database layer"

# After Phase 3
git add internal/<service>/services/
git commit -m "feat(<service>): implement service layer"
```

### Push Regularly

After every 2-3 commits:
```bash
git push -u origin feature/XXX-<description>
```

**Why push regularly:**
- Backs up your work
- Triggers CI checks early
- Allows others to see progress
- Enables automatic PR creation

---

## Overview

### Feature Description
<Detailed description>

### User Story
As a <role>, I want to <action> so that <benefit>.

### Success Criteria
- [ ] <Criterion 1>
- [ ] <Criterion 2>
- [ ] <Criterion 3>
- [ ] All tests pass with 70%+ coverage
- [ ] Service health check endpoint works

---

## Context for Cognitive Load Management

### Bounded Context

**Service:** <Service Name>
**Domain:** <Domain Description>
**Related Entities:**
- `<Entity1>` - <Description>
- `<Entity2>` - <Description>

**Context Boundaries:**
- ✅ **Within scope:** <What this service handles>
- ❌ **Out of scope:** <What other services handle>

**Why This Separation:**
<Explanation of bounded context rationale>

---

### Layering

**Primary Layer:** All three layers required (Controller → Orchestration → Data)

#### Controller Layer Files
```
cmd/<service>/handlers/
├── <handler>.go
├── <handler>_test.go
...
```

#### Orchestration Layer Files
```
internal/<service>/services/
├── <service>.go
├── <service>_test.go
...
```

#### Data Layer Files
```
internal/<service>/db/
├── <repository>.go
├── <repository>_test.go
├── migrations/
    └── <timestamp>_<description>.sql
```

**Cross-Layer Rules:**
- ✅ Handlers call services
- ✅ Services call repositories
- ❌ Handlers MUST NOT call repositories directly
- ❌ Repositories MUST NOT import service/handler packages
- ❌ No circular dependencies

---

## Implementation Specification

### Phase 1: <Phase Name>

<Detailed specification with complete code examples>

**Files to create:**
- `path/to/file.go`

**Complete code:**
```go
// Full, working code example
package example

// ...
```

**Commit after this phase:**
```bash
git add <files>
git commit -m "feat(<scope>): <phase description>"
```

---

### Phase 2: <Phase Name>

<Continue with more phases...>

---

## Implementation Checklist

### Phase 0: Branch Setup ✅ (ALREADY DONE)
- [x] Verified on development branch
- [x] Pulled latest changes
- [x] Created feature branch: `feature/XXX-<description>`
- [x] Verified on feature branch

### Phase 1: <Phase Name> ✅
- [ ] Create files listed above
- [ ] Run: `go test ./...`
- [ ] Commit: `git add <files> && git commit -m "feat(<scope>): <description>"`

### Phase 2: <Phase Name> ✅
- [ ] Create files listed above
- [ ] Run: `go test ./...`
- [ ] Commit: `git add <files> && git commit -m "feat(<scope>): <description>"`

### Phase 3: Push Progress ✅
- [ ] Push: `git push -u origin feature/XXX-<description>`

### Phase N: Final Testing ✅
- [ ] Run full test suite: `make test`
- [ ] Run linting: `make lint`
- [ ] Verify all services start: `make dev`
- [ ] Test endpoints manually

### Phase N+1: Final Push ✅
- [ ] Review all commits: `git log development..HEAD --oneline`
- [ ] Push final changes: `git push`
- [ ] **Wait for automatic PR creation** (GitHub Actions will create it)
- [ ] Verify CI passes on PR
- [ ] Tag @Claude for code review

---

## Environment Variables

Add to `.env.example` if needed:

```bash
# <Service> Configuration
<SERVICE>_PORT=808X
<SERVICE>_CONFIG=value
```

---

## Testing Strategy

### Unit Tests (70%+ coverage required)

**Test Coverage Targets:**
- Models: 80%+
- Repositories: 75%+
- Services: 80%+
- Handlers: 70%+

**Key Test Cases:**
1. ✅ <Test case 1>
2. ✅ <Test case 2>
3. ✅ <Test case 3>

---

## Success Metrics

This issue is complete when:

1. ✅ All database migrations run successfully
2. ✅ Service starts without errors
3. ✅ Health check endpoint returns 200 OK
4. ✅ All acceptance criteria met
5. ✅ All unit tests pass with 70%+ coverage
6. ✅ No linting errors
7. ✅ CI/CD pipeline passes
8. ✅ PR created automatically
9. ✅ Claude review completed

---

## Cognitive Load Optimization Notes

### For Intrinsic Complexity (Simplify)
- <How complexity is encapsulated>
- <Clear naming conventions>
- <Separation of concerns>

### For Extraneous Load (Reduce)
- <No magic strings/numbers>
- <Explicit error messages>
- <Consistent naming patterns>
- <No global state>

### For Germane Load (Maximize)
- <Follows established patterns>
- <Respects bounded contexts>
- <Uses platform idioms>
- <Interfaces enable testing>

---

## Questions and Clarifications

### Before Starting
- [x] Feature branch created (STEP 0 above)
- [x] Bounded context clear
- [x] Dependencies understood
- [x] Testing strategy defined

### During Implementation

If you encounter:
- **<Issue>** → <Solution>
- **<Issue>** → <Solution>

---

## References

- `ARCHITECTURE.md` - <Service> specification (lines XXX-XXX)
- `Requirements.md` - <Feature> requirements (lines XXX-XXX)
- `DevSmithTDD.md` - Testing strategy
- <External docs links>

---

**Next Steps (For Copilot):**
1. ✅ Feature branch already created (STEP 0 above)
2. Read this spec completely (entire document)
3. Follow implementation checklist phase by phase
4. **Commit after each phase** (X commits expected)
5. Test after each phase: `go test ./...`
6. Push regularly: `git push` after every 2-3 commits
7. Wait for automatic PR creation (GitHub Actions)
8. Tag Claude for architecture review

**Estimated Time:** XX-XX minutes
**Test Coverage Target:** 70%+ (aim for 75%+)
**Success Metric:** <Brief success description>
**Depends On:** Issue #XXX (if applicable)
