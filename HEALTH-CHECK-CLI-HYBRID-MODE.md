# Health Check CLI - Hybrid Mode Implementation

**Date:** October 30, 2025  
**Status:** ✅ Complete  
**Version:** 2.0 (Hybrid with PR Mode)

## What Changed

Added **`--pr` mode** to the health-check CLI that provides comprehensive PR validation, eliminating the need for Copilot to use `docker-validate.sh` separately.

## The Four Modes

```bash
# Development (Phase 1 only, ~900ms)
./scripts/health-check-cli.sh

# Quick development (Phase 1 only, <500ms)
./scripts/health-check-cli.sh --quick

# Real-time monitoring (continuous, 5s intervals)
./scripts/health-check-cli.sh --watch

# PR validation (Phase 1 + 2 + endpoints, comprehensive)
./scripts/health-check-cli.sh --pr
```

## When Copilot Should Use Each

### Before Starting Work
```bash
./scripts/health-check-cli.sh
# Quick verification that system is ready
```

### While Developing
```bash
# Terminal 1: Monitor health
./scripts/health-check-cli.sh --watch

# Terminal 2: Make changes
# Terminal 3: Run tests
```

### For Quick Feedback During Rapid Changes
```bash
./scripts/health-check-cli.sh --quick
# Even faster than default
```

### BEFORE CREATING PR (MANDATORY)
```bash
./scripts/health-check-cli.sh --pr
# Comprehensive validation:
# - Phase 1: Container/HTTP/Database
# - Phase 2: Gateway/Performance/Dependencies
# - Endpoint discovery & testing
# - Security scanning
```

## What --pr Mode Does

```
🔍 Comprehensive PR Validation
Running Phase 1, 2, and endpoint validation...

Step 1: Health Checks (Phase 1 + 2)
✅ All services healthy

Step 2: Full Endpoint Validation
✅ Endpoint validation passed

Step 3: Summary
✅ PR validation PASSED
Ready to create PR
```

## Key Benefits

✅ **Single Tool:** Copilot only needs to know `./scripts/health-check-cli.sh`  
✅ **Four Modes:** Dev, Quick, Watch, PR - each optimized for its purpose  
✅ **No docker-validate.sh Needed:** `--pr` mode is the new gate  
✅ **Fast Feedback:** Quick/default modes under 1 second  
✅ **Real-time Monitoring:** Watch mode shows health continuously  
✅ **Comprehensive Validation:** PR mode includes all checks  

## Copilot Instructions

Tell Copilot to use this pattern:

```
HEALTH CHECK CLI USAGE (REQUIRED)
==================================

1. BEFORE STARTING WORK:
   ./scripts/health-check-cli.sh
   
   If healthy, proceed to 2.

2. WHILE DEVELOPING:
   Terminal 1: ./scripts/health-check-cli.sh --watch
   Terminal 2: vim internal/review/services/...
   Terminal 3: docker-compose up -d --build review && go test ./...

3. QUICK CHECKS DURING RAPID DEVELOPMENT:
   ./scripts/health-check-cli.sh --quick
   
4. BEFORE CREATING PR (MANDATORY):
   ./scripts/health-check-cli.sh --pr
   
   Do NOT create PR unless this passes.
   If failures, fix and run again.

5. IF PR VALIDATION FAILS:
   ./scripts/health-check-cli.sh --pr --json
   
   Use JSON output to debug specific failures.

REMEMBER:
- Use health-check-cli.sh for ALL diagnostics
- Do NOT use docker-validate.sh directly
- Always run --pr before creating PR
- Monitor with --watch while developing
```

## File Changes

### Updated Files
1. `./scripts/health-check-cli.sh`
   - Added `--quick` flag (Phase 1 only, fast)
   - Added `--pr` flag (comprehensive validation)
   - Added `pr_validation_mode()` function
   - Enhanced help with mode descriptions

2. `./.github/copilot-instructions.md`
   - Section 3.5: Added PR validation workflow
   - Added "PR Validation Workflow (MANDATORY)" section
   - Added "Complete Development + PR Workflow" example
   - Updated examples and troubleshooting

## Test Commands

```bash
# Test basic mode (Phase 1)
./scripts/health-check-cli.sh

# Test quick mode
./scripts/health-check-cli.sh --quick

# Test PR mode
./scripts/health-check-cli.sh --pr

# Test PR mode with JSON
./scripts/health-check-cli.sh --pr --json

# Test watch mode (Ctrl+C to exit)
./scripts/health-check-cli.sh --watch

# Get help
./scripts/health-check-cli.sh --help
```

## Migration Path

### For Copilot
- ✅ **NOW:** Use `./scripts/health-check-cli.sh --pr` before creating PRs
- ✅ **NOW:** Use `./scripts/health-check-cli.sh` for quick checks
- ✅ **NOW:** Use `./scripts/health-check-cli.sh --watch` for monitoring
- ❌ **STOP:** Using `./scripts/docker-validate.sh` directly

### For docker-validate.sh
- Status: Still available as fallback
- Used internally by `--pr` mode for endpoint testing
- Can be deprecated once `--pr` mode is fully tested
- Kept for now as comprehensive fallback

## Why This Works

1. **Single command for Copilot** - No need to juggle two different scripts
2. **Mode-based design** - Each mode optimized for its use case
3. **PR gate in health-check-cli** - `--pr` is the mandatory checkpoint
4. **No disruption** - docker-validate.sh still works if needed
5. **Progressive consolidation** - Can deprecate docker-validate.sh later

## What Copilot Needs to Know

Print this for Copilot:

```
============================================================
HEALTH CHECK CLI - COPILOT REFERENCE
============================================================

Use ./scripts/health-check-cli.sh for ALL diagnostics

Modes:
  (default)    Phase 1 only, quick (~900ms)
  --quick      Phase 1 only, super fast (<500ms)
  --watch      Continuous monitoring (5s intervals)
  --pr         ⭐ REQUIRED BEFORE PR - Full validation

Workflow:
  1. ./scripts/health-check-cli.sh          (start)
  2. ./scripts/health-check-cli.sh --watch  (dev terminal 1)
  3. [implement feature]                    (dev terminal 2)
  4. ./scripts/health-check-cli.sh --pr     (validation before PR)
  5. gh pr create (if step 4 passes)

DO NOT:
  - Use docker-validate.sh directly
  - Create PR without running --pr mode first
  - Ignore --pr validation failures
============================================================
```

## Architecture

```
health-check-cli.sh
├── Default Mode (Phase 1)
│   └── healthcheck binary
├── Quick Mode (Phase 1, optimized)
│   └── healthcheck binary
├── Watch Mode (continuous)
│   └── healthcheck binary (every 5s)
└── PR Mode (comprehensive)
    ├── healthcheck binary (Phase 1+2)
    └── docker-validate.sh (endpoint testing)
```

## Future Enhancements

1. **Deprecate docker-validate.sh** - Once `--pr` is battle-tested
2. **Standalone --pr** - Embed endpoint testing directly (no docker-validate.sh dependency)
3. **Store to database** - Implement `--store` flag for historical tracking
4. **CI/CD integration** - `--pr --json` output feeds into CI pipelines
5. **Performance trends** - Track `--pr` validation times over days/weeks

## Success Criteria

✅ Copilot can run all diagnostics with single tool  
✅ `--pr` mode is mandatory gate before PR creation  
✅ Documentation updated in copilot-instructions.md  
✅ Help system shows all four modes  
✅ Backward compatible (docker-validate.sh still works)  
✅ Fast feedback loops for development  

## Status: COMPLETE ✅

All modes implemented and tested. Ready for Copilot to use.

---

**For Copilot Setup:** Use the "HEALTH CHECK CLI - COPILOT REFERENCE" section above when giving Copilot instructions.

**For Documentation:** Reference this file and `.github/copilot-instructions.md` Section 3.5.
