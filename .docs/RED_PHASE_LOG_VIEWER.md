# RED PHASE: Log Viewer UI with Virtual Scrolling

## Overview
Complete RED phase test suite for Log Viewer UI feature (Issue #XX)
All tests written to FAIL FIRST per TDD strict practice

## Test File Location
`tests/e2e/logs_viewer_complete.spec.ts`

## Test Statistics
- **Total Tests**: 31
- **Test Categories**: 9
- **Test Status**: ALL WRITTEN TO FAIL (RED PHASE)

## Acceptance Criteria Coverage

### 1. ✅ Templ Template for Log Viewer Page
- ✓ Template exists at `apps/logs/templates/dashboard.templ`
- ✓ Layout template at `apps/logs/templates/layout.templ`
- ✗ **RED tests validate** template includes all new components

### 2. ✅ Virtual Scrolling or Pagination
**Tests (3 total):**
- `RED: Virtual scrolling - should render 10,000 logs without lag`
  - Validates scroll performance < 200ms on 10k items
- `RED: Virtual scrolling - should only render visible viewport (windowing)`
  - Tests DOM windowing (expects <200 items in DOM when virtualized)
  - **Currently FAILS**: All 10k items in DOM (not virtualized)
- `RED: Virtual scrolling - should lazy-load as user scrolls`
  - Tests scroll-triggered batch loading

### 3. ✅ Color-Coded Log Levels
- ✓ CSS colors implemented (ERROR=red, WARN=yellow, INFO=blue)
- ✗ **RED tests validate** accessibility of colors

### 4. ❌ Expand/Collapse for Stack Traces
**Tests (4 total):**
- `RED: Expandable details - should show expand button on log entry`
  - **Currently FAILS**: No expand button element
- `RED: Expandable details - should expand to show stack trace`
  - **Currently FAILS**: No stack-trace element
- `RED: Expandable details - should show context metadata`
  - **Currently FAILS**: No metadata/context display
- `RED: Expandable details - should show full error message when expanded`
  - **Currently FAILS**: Messages not expandable

### 5. ✅ WebSocket Integration for Real-Time Updates
- ✓ WebSocket client implemented
- ✓ Real-time log streaming working
- ✗ **RED tests validate** expanded details stream in real-time

**Tests (2 total):**
- `RED: WebSocket - should receive expanded log details in real-time`
  - Tests new logs with stack traces/context arrive
- `RED: WebSocket - expanded details should update in real-time`
  - Tests live updates to expanded content

### 6. ❌ Filter UI with Service/Level/Date Dropdowns
**Tests (3 total):**
- `RED: Date picker - should have date range input fields`
  - **Currently FAILS**: No date inputs (#date-from, #date-to)
- `RED: Date picker - should filter logs by date range`
  - **Currently FAILS**: Date filtering not implemented
- `RED: Date picker - should persist date filters in URL params`
  - **Currently FAILS**: No URL param persistence for dates

### 7. ❌ Search Bar with Debounced Input
**Tests (2 total):**
- `RED: Search debouncing - should wait 300ms before filtering`
  - **Currently FAILS**: No debounce (immediate filtering)
- `RED: Search debouncing - should only call API once for multi-character input`
  - **Currently FAILS**: API called 5 times for 5 characters (no debounce)

### 8. ❌ Toast Notifications for New Errors
**Tests (4 total):**
- `RED: Toast notifications - should show when new error log received`
  - **Currently FAILS**: No toast system
- `RED: Toast notifications - should auto-dismiss after 5 seconds`
  - **Currently FAILS**: No toast component
- `RED: Toast notifications - should allow manual dismiss`
  - **Currently FAILS**: No close button on toast
- `RED: Toast notifications - should show for warnings`
  - **Currently FAILS**: Only errors shown, not warnings

### 9. ❌ Copy Log Entry to Clipboard
**Tests (3 total):**
- `RED: Copy clipboard - should have copy button on log entry`
  - **Currently FAILS**: No copy button
- `RED: Copy clipboard - should copy full log entry to clipboard`
  - **Currently FAILS**: Copy function not implemented
- `RED: Copy clipboard - should show confirmation toast on copy`
  - **Currently FAILS**: No confirmation feedback

### 10. ✅ Responsive Design (Mobile-Friendly)
**Tests (3 total):**
- `RED: Responsive design - should adapt to mobile (375px)`
  - ✓ CSS media queries implemented
  - Tests expand button visible on mobile
- `RED: Responsive design - stack trace should be readable on mobile`
  - Tests stack trace display on small screens
- `RED: Responsive design - copy button should work on mobile`
  - Tests copy button has minimum touch target (30x30px)

### 11. ✅ Accessibility (WCAG 2.1 AA)
**Tests (5 total):**
- `RED: Accessibility - should have proper ARIA labels`
  - Tests aria-label or associated <label> on inputs
- `RED: Accessibility - should support keyboard navigation`
  - Tests Tab key navigation works
- `RED: Accessibility - should have sufficient color contrast`
  - Tests computed colors meet contrast requirements
- `RED: Accessibility - should include role attributes`
  - Tests role="list" or role="region" on containers
- `RED: Accessibility - should use semantic HTML`
  - Tests presence of <nav>, <main>, <section>

### 12. ✅ E2E Tests with Playwright
**Test File**: `tests/e2e/logs_viewer_complete.spec.ts`
- All 31 tests using Playwright
- Browser automation (Chrome default)
- Network monitoring built in

### 13. ✅ Integration Tests
**Tests (2 total):**
- `RED: Integration - expand + copy should work together`
  - Tests expand then copy preserves stack traces
- `RED: Integration - date filter + search + copy should work together`
  - Tests all filters + copy work simultaneously

---

## Current Implementation Status

### COMPLETE (No Changes Needed)
- ✅ Templ templates exist
- ✅ WebSocket real-time streaming
- ✅ Color-coded levels
- ✅ Responsive CSS
- ✅ Basic accessibility
- ✅ E2E test framework

### MISSING (RED Phase Failing)
- ❌ Virtual scrolling / windowing
- ❌ Expandable log details
- ❌ Date range picker
- ❌ Search debouncing
- ❌ Toast notifications
- ❌ Copy to clipboard
- ❌ Full accessibility compliance
- ❌ Real-time expanded details

---

## RED Phase Test Execution

```bash
# Run all RED phase tests
npx playwright test tests/e2e/logs_viewer_complete.spec.ts

# Run specific test category
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Virtual scrolling"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Expandable details"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Date picker"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Copy clipboard"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Toast"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Search debouncing"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Accessibility"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Responsive"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "WebSocket"
npx playwright test tests/e2e/logs_viewer_complete.spec.ts -g "Integration"

# Run with detailed output
npx playwright test tests/e2e/logs_viewer_complete.spec.ts --reporter=html

# Run in headed mode to see browser
npx playwright test tests/e2e/logs_viewer_complete.spec.ts --headed
```

---

## Next Steps (GREEN Phase)

1. **Virtual Scrolling** - Implement windowing library or custom virtualization
2. **Expandable Details** - Add expand button, stack trace/context display
3. **Date Picker** - Add date input fields and filtering logic
4. **Search Debounce** - Add 300ms debounce to search input
5. **Toast System** - Implement toast notification component
6. **Copy Clipboard** - Add copy button and clipboard API integration
7. **Enhanced WebSocket** - Stream expanded details in real-time
8. **Full A11y** - Add ARIA labels, keyboard navigation, contrast compliance

---

## RED Phase Philosophy

All 31 tests are written to **FAIL FIRST**. This is intentional.

**Why?** Because TDD requires:
1. Write tests that fail (RED)
2. Write minimal code to pass (GREEN)
3. Refactor for clarity (REFACTOR)

By completing RED first, we:
- ✅ Define exact acceptance criteria
- ✅ Prove requirements aren't met
- ✅ Create a roadmap for implementation
- ✅ Ensure tests are robust (not written after code)

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `tests/e2e/logs_viewer_complete.spec.ts` | 650+ | 31 RED phase tests |
| `.docs/RED_PHASE_LOG_VIEWER.md` | This doc | Test coverage matrix |

---

## Acceptance Criteria: RED Phase ✅ COMPLETE

- [x] 31 comprehensive E2E tests written
- [x] All acceptance criteria covered
- [x] All tests failing (confirming requirements unmet)
- [x] Ready for GREEN phase implementation

**RED Phase Status**: ✅ COMPLETE
