# Pull Request

## 🎯 Issue Reference (REQUIRED - DO NOT SKIP!)

**Closes #<!-- issue number goes here - EXAMPLE: Closes #36 -->**

<!-- 
🚨 CRITICAL: Replace the example above with your actual issue number!
This MUST be filled in for:
1. Issue to auto-link to PR
2. Issue to auto-close when PR merges

Correct format: "Closes #36" (where 36 is your issue number)
Wrong formats that WON'T work:
  ❌ Feature #36
  ❌ Issue 36
  ❌ #36
  ❌ Closes Feature 36

GitHub keywords that work:
✓ Closes #36
✓ Close #36  
✓ Closed #36
✓ Fixes #36
✓ Fixed #36
✓ Fix #36
✓ Resolves #36
✓ Resolved #36
✓ Resolve #36
-->

## Feature/Fix Description

<!-- Brief description of what this PR implements -->

## Implementation Details

<!-- List key changes made -->
-
-
-

## Testing

### Automated Testing
- [ ] All unit tests pass
- [ ] Unit test coverage >= 70%
- [ ] Critical path coverage >= 90%
- [ ] All linting checks pass
- [ ] Docker build succeeds

### Manual Testing Checklist
- [ ] Feature works as expected in browser
- [ ] No console errors or warnings
- [ ] All related features still work (regression check)
- [ ] Works in both light and dark mode (if applicable)
- [ ] Responsive design works on mobile/tablet (if applicable)
- [ ] Works through nginx gateway (http://localhost:3000)
- [ ] Authentication persists across apps (if applicable)
- [ ] WebSocket connections work (if applicable)
- [ ] Hot module reload (HMR) works during development

### Test Results
```
<!-- Paste test output showing coverage -->
Unit Test Coverage: X%
Integration Tests: X/X passing
```

## Standards Compliance
- [ ] Follows file organization (ARCHITECTURE.md Section 13)
- [ ] Follows naming conventions (ARCHITECTURE.md Section 13)
- [ ] React components follow standard template
- [ ] API calls follow error handling pattern
- [ ] Backend endpoints follow Go patterns
- [ ] No hardcoded URLs, ports, or credentials
- [ ] All configuration in environment variables
- [ ] .env.example updated (if new variables added)
- [ ] Error messages are user-friendly
- [ ] Loading states present
- [ ] Comprehensive logging with context

## Acceptance Criteria

<!-- Copy acceptance criteria from the GitHub issue and check them off -->

From Issue #XXX:
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3
- [ ] Unit tests achieve 70%+ coverage
- [ ] Integration test covers critical path
- [ ] No hardcoded values
- [ ] Error messages are user-friendly

**ALL criteria must be checked before PR can be approved.**

## Changelog
<!-- Activity logging is now automatic - no manual changelog needed! -->
Activity logged automatically via git hooks.

## Screenshots
<!-- If UI changes, include before/after screenshots -->

### Before
<!-- Screenshot or N/A -->

### After
<!-- Screenshot or N/A -->

## Additional Notes
<!-- Any additional context, decisions made, or things reviewers should know -->

---

## Reviewer Checklist (for Claude)

### Architecture Compliance
- [ ] Follows gateway-first design
- [ ] Respects service boundaries
- [ ] Database schema changes backward compatible
- [ ] No new dependencies without design discussion
- [ ] Fits into overall system architecture

### Code Quality
- [ ] Tests written BEFORE implementation (TDD)
- [ ] Error handling comprehensive
- [ ] No error strings returned as data
- [ ] Logging includes context

### Acceptance Criteria
- [ ] Every criterion from issue is met
- [ ] No partial implementations
- [ ] Feature is 100% complete
- [ ] Issue will auto-close on merge (has "Closes #XXX")

**Recommendation:** ☐ Approve  ☐ Request Changes

**Reasoning:**
<!-- Claude's review comments -->
