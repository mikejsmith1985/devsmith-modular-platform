# Phase 2: GitHub Integration - Implementation Status

**Goal:** Enable Review app to analyze entire GitHub repositories, folders, and multiple files

**Duration:** 3-4 weeks (estimated)  
**Branch:** `feature/phase2-github-integration`  
**Status:** 🔵 IN PROGRESS

---

## Overview

Phase 2 extends the Review application to handle full GitHub repositories, enabling code review at scale through:
- GitHub API integration for fetching repository structures
- File tree UI for navigation
- Multi-tab system for viewing multiple files simultaneously
- Cross-file analysis for architectural insights

---

## Acceptance Criteria Checklist

### GitHub API Client (`internal/review/github/client.go`)
- [ ] Method: `GetRepoTree(owner, repo, branch)` returns hierarchical file structure
- [ ] Method: `GetFileContent(owner, repo, path, branch)` returns file content (decoded)
- [ ] Method: `GetPullRequest(owner, repo, prNum)` returns PR metadata
- [ ] Method: `GetPRFiles(owner, repo, prNum)` returns changed files + diffs
- [ ] Authentication: Uses `GITHUB_TOKEN` from env or user-provided token
- [ ] Rate limiting: Respects GitHub API limits (5000/hour authenticated)

### Session Management Enhancement
- [ ] New session type: `SessionTypeGitHub` with fields:
  - `github_url`, `owner`, `repo`, `branch`, `file_tree` (JSONB cached)
- [ ] Database table: `reviews.github_sessions`
- [ ] Endpoint: `POST /review/sessions/github` creates GitHub session
- [ ] Endpoint: `GET /review/sessions/:id/tree` returns cached or fresh tree

### UI - File Tree Viewer (LEFT PANE)
- [ ] Recursive tree component (`apps/review/templates/components/file_tree.templ`)
- [ ] Click folder → expand/collapse
- [ ] Click file → open in new tab (loads content via htmx)
- [ ] File icons by extension (.go, .js, .md, .yaml, etc.)
- [ ] Breadcrumb navigation (e.g., `repo/src/handlers/auth.go`)

### UI - Multi-Tab System
- [ ] Tab bar above code pane with `+ New Tab` button
- [ ] Each tab: unique `tab_id` (UUID), filename label, close button
- [ ] Active tab highlighted
- [ ] Click tab → switches active pane
- [ ] Close tab → `hx-confirm="Discard changes?"` if unsaved analysis exists
- [ ] Tab state persisted in session storage (survive page refresh)

### Multi-File Analysis
- [ ] Endpoint: `POST /review/analyze-multiple`
- [ ] Input: `{session_id, file_ids[], mode}`
- [ ] Process:
  1. Fetch all file contents
  2. Concatenate with separators: `=== FILE: {path} ===`
  3. Send to Ollama with cross-file context prompt
  4. Return unified analysis
- [ ] Output: Dependencies between files, shared abstractions, architecture patterns

### Testing
- [ ] Unit tests: GitHub API client methods (mocked responses)
- [ ] Integration test: Full flow (paste GitHub URL → tree loads → file opens → analyze)
- [ ] UI test (Playwright): Navigate tree, open 3 files in tabs, run multi-file analysis

---

## Progress Tracking

### Session 1: 2025-11-04
**Focus:** Extend GitHub Client for Phase 2 (Tree & PR APIs)

**Context:**
- GitHub client already exists from Issue #27 with basic functionality:
  - `FetchCode()` - Retrieve code from repository
  - `GetRepoMetadata()` - Get repository information
  - `ValidateURL()` - Parse and validate GitHub URLs
  - `GetRateLimit()` - Check API rate limits
- Need to extend with Phase 2 methods:
  - `GetRepoTree()` - Fetch repository file tree structure
  - `GetFileContent()` - Get individual file content (decoded)
  - `GetPullRequest()` - Fetch PR metadata
  - `GetPRFiles()` - Get changed files in PR with diffs

**Completed:**
- [x] Analyzed existing GitHub client infrastructure
- [x] Identified Phase 2 extension points
- [x] Add GetRepoTree() method (TDD: RED+GREEN phases complete)
- [x] Add GetFileContent() method (TDD: RED+GREEN phases complete)
- [x] Add GetPullRequest() method (TDD: RED+GREEN phases complete)
- [x] Add GetPRFiles() method (TDD: RED+GREEN phases complete)
- [x] Extended ClientInterface with 4 new methods
- [x] Added Phase 2 types: TreeNode, RepoTree, FileContent, PullRequest, PRFile
- [x] Wrote 12 comprehensive tests (RED phase)
- [x] Implemented stub methods in DefaultClient (GREEN phase)
- [x] All 22/22 tests passing

**Next Steps:**
- [ ] Database migration for reviews.github_sessions table
- [ ] Session service for storing GitHub repository metadata
- [ ] API endpoints: POST /sessions, GET /sessions/:id, DELETE /sessions/:id
- [ ] Add rate limiting for new methods (when implementing real API calls)

---

## Technical Notes

### GitHub API Rate Limiting
- Authenticated: 5000 requests/hour
- Unauthenticated: 60 requests/hour
- Strategy: Cache tree structures, use conditional requests

### Database Schema Design
```sql
-- Table: reviews.github_sessions
-- Purpose: Store GitHub repository session metadata and cached tree
-- Relationships: Many-to-one with reviews.sessions

-- Table: reviews.open_files
-- Purpose: Track files opened in multi-tab UI
-- Relationships: Many-to-one with reviews.sessions
```

### Multi-File Analysis Prompt Strategy
- Concatenate files with clear separators
- Provide cross-file context to AI
- Extract: shared abstractions, dependencies, architecture patterns
- Detect: inconsistencies, refactoring opportunities

---

## Success Metrics (from Roadmap)
- Repo tree loads in <5 seconds for repos with <10,000 files
- Multi-file analysis completes in <30 seconds for 10 files
- Tree navigation feels instant (<100ms per click)
- Cache hit rate: >70% for repeated tree requests

---

## References
- **Implementation Roadmap:** `.docs/IMPLEMENTATION_ROADMAP.md` (Phase 2, lines 303-467)
- **TDD Guide:** `DevsmithTDD.md`
- **Architecture:** `ARCHITECTURE.md`
- **Phase 1 Status:** `.docs/PHASE1_STATUS.md` (completed example)
