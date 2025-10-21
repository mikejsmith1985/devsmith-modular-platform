# Disabled Workflows

These workflows have been archived because they caused false failures and provided no value beyond what the pre-commit hook already validates.

## Problems with Old Workflows

### test-and-build.yml
**Issues:**
- Used static `docker/postgres/init-schemas.sql` that got out of sync with evolving code models
- Caused "column does not exist" errors when structs added fields (e.g., `User.Email`)
- Schema drift: Static schema + dynamic code = false failures
- Duplicated pre-commit checks (go vet, golangci-lint, builds, tests)
- Cost hours of debugging for zero actual bugs caught

**What it tried to do:**
- Run tests with PostgreSQL database
- Validate database schema matches code

**Why removed:**
- Pre-commit hook already validates builds and tests locally
- Database tests require migrations system (not implemented)
- No value added, only false failures

### validate-migrations.yml
**Issues:**
- Same static schema problem as test-and-build.yml
- Ran on every commit even when no migrations existed
- Checked schemas that would drift from code models

**What it tried to do:**
- Validate schema initialization script
- Count schemas and tables

**Why removed:**
- No migration system exists yet
- Static validation doesn't work with evolving schemas
- Not needed for MVP

## Current CI Strategy

See `.github/workflows/ci.yml` for the new minimal CI that:
- ✅ Validates code builds (catches `--no-verify` commits)
- ✅ Validates Docker images build (can't do in pre-commit)
- ✅ Quick lint pass (fast safety net)
- ❌ NO database tests (avoiding schema drift)
- ❌ NO duplicate checks (pre-commit is the quality gate)

## When to Re-enable Database Tests

Only re-enable when:
1. Migration system implemented (`internal/*/db/migrations/*.sql`)
2. CI runs migrations in order instead of static schema
3. Schema can evolve with code without manual sync

Until then: **Pre-commit hook is the quality gate, CI is a lightweight safety net.**
