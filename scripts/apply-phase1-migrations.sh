#!/bin/bash

# Script to apply Phase 1 Multi-LLM migrations
# Usage: bash scripts/apply-phase1-migrations.sh

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Phase 1 Multi-LLM Platform Migrations"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Database connection settings
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-devsmith}"
DB_NAME="${DB_NAME:-devsmith}"
DB_PASSWORD="${DB_PASSWORD:-devsmith}"

PSQL_CMD="docker-compose exec -T postgres psql -U $DB_USER -d $DB_NAME"

echo "Step 1: Checking database connection..."
if ! $PSQL_CMD -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ Error: Cannot connect to database"
    echo "Make sure PostgreSQL is running: docker-compose up -d postgres"
    exit 1
fi
echo "✓ Database connection OK"
echo ""

echo "Step 2: Applying migration 20251108_001_prompt_templates.sql..."
if $PSQL_CMD < db/migrations/20251108_001_prompt_templates.sql; then
    echo "✓ Migration 20251108_001 applied successfully"
else
    echo "❌ Error applying migration 20251108_001"
    exit 1
fi
echo ""

echo "Step 3: Applying migration 20251108_002_llm_configs.sql..."
if $PSQL_CMD < db/migrations/20251108_002_llm_configs.sql; then
    echo "✓ Migration 20251108_002 applied successfully"
else
    echo "❌ Error applying migration 20251108_002"
    exit 1
fi
echo ""

echo "Step 4: Applying seed data 20251108_001_default_prompts.sql..."
if $PSQL_CMD < db/seeds/20251108_001_default_prompts.sql; then
    echo "✓ Seed data applied successfully"
else
    echo "❌ Error applying seed data"
    exit 1
fi
echo ""

echo "Step 5: Verifying migrations..."
echo ""

# Verify prompt_templates table
echo "Checking prompt_templates table..."
PROMPT_COUNT=$($PSQL_CMD -t -c "SELECT COUNT(*) FROM review.prompt_templates WHERE is_default = true;" | xargs)
if [ "$PROMPT_COUNT" -eq "15" ]; then
    echo "✓ Found 15 default prompts (5 modes × 3 user levels)"
else
    echo "❌ Expected 15 default prompts, found $PROMPT_COUNT"
    exit 1
fi

# Verify prompt_executions table
echo "Checking prompt_executions table..."
if $PSQL_CMD -c "\d review.prompt_executions" > /dev/null 2>&1; then
    echo "✓ prompt_executions table exists"
else
    echo "❌ prompt_executions table not found"
    exit 1
fi

# Verify llm_configs table
echo "Checking llm_configs table..."
if $PSQL_CMD -c "\d portal.llm_configs" > /dev/null 2>&1; then
    echo "✓ llm_configs table exists"
else
    echo "❌ llm_configs table not found"
    exit 1
fi

# Verify app_llm_preferences table
echo "Checking app_llm_preferences table..."
if $PSQL_CMD -c "\d portal.app_llm_preferences" > /dev/null 2>&1; then
    echo "✓ app_llm_preferences table exists"
else
    echo "❌ app_llm_preferences table not found"
    exit 1
fi

# Verify llm_usage_logs table
echo "Checking llm_usage_logs table..."
if $PSQL_CMD -c "\d portal.llm_usage_logs" > /dev/null 2>&1; then
    echo "✓ llm_usage_logs table exists"
else
    echo "❌ llm_usage_logs table not found"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Phase 1 Migrations Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Summary:"
echo "  ✓ review.prompt_templates table created"
echo "  ✓ review.prompt_executions table created"
echo "  ✓ portal.llm_configs table created"
echo "  ✓ portal.app_llm_preferences table created"
echo "  ✓ portal.llm_usage_logs table created"
echo "  ✓ 15 default prompts seeded"
echo ""
echo "Next steps:"
echo "  1. Run integration tests: cd tests/db && go test -v"
echo "  2. Proceed to Phase 2: Backend Services - Prompt Management"
echo ""
