#!/bin/bash
set -e

# Load environment variables from .env if present
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Best practice: Do NOT commit .env files with secrets. Ensure .env is in .gitignore.
set -e

echo "Creating databases and schemas..."

DB_USER="${POSTGRES_USER:-devsmith}"
DB_PASS="${POSTGRES_PASSWORD:-devsmith}"
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
MAIN_DB="${POSTGRES_DB:-devsmith}"



PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$MAIN_DB'" | grep -q 1 || \
  PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d postgres -c "CREATE DATABASE $MAIN_DB"

echo "✓ Main database created"

PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$MAIN_DB" -c "CREATE SCHEMA IF NOT EXISTS portal"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$MAIN_DB" -c "CREATE SCHEMA IF NOT EXISTS review"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$MAIN_DB" -c "CREATE SCHEMA IF NOT EXISTS logs"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$MAIN_DB" -c "CREATE SCHEMA IF NOT EXISTS analytics"

echo "✓ Schemas created"


PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -tc "SELECT 1 FROM pg_roles WHERE rolname='portal_user'" | grep -q 1 || \
  PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -c "CREATE USER portal_user WITH PASSWORD 'portal_pass'"

PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -tc "SELECT 1 FROM pg_roles WHERE rolname='review_user'" | grep -q 1 || \
  PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -c "CREATE USER review_user WITH PASSWORD 'review_pass'"

PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -tc "SELECT 1 FROM pg_roles WHERE rolname='logs_user'" | grep -q 1 || \
  PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -c "CREATE USER logs_user WITH PASSWORD 'logs_pass'"

PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -tc "SELECT 1 FROM pg_roles WHERE rolname='analytics_user'" | grep -q 1 || \
  PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -c "CREATE USER analytics_user WITH PASSWORD 'analytics_pass'"

echo "✓ Users created"

PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$PORTAL_DB" -c "GRANT ALL PRIVILEGES ON SCHEMA portal TO portal_user"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$REVIEW_DB" -c "GRANT ALL PRIVILEGES ON SCHEMA review TO review_user"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$LOGS_DB" -c "GRANT ALL PRIVILEGES ON SCHEMA logs TO logs_user"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$ANALYTICS_DB" -c "GRANT ALL PRIVILEGES ON SCHEMA analytics TO analytics_user"

# Grant analytics READ-ONLY access to logs
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$LOGS_DB" -c "GRANT USAGE ON SCHEMA logs TO analytics_user"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$LOGS_DB" -c "GRANT SELECT ON ALL TABLES IN SCHEMA logs TO analytics_user"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$LOGS_DB" -c "ALTER DEFAULT PRIVILEGES IN SCHEMA logs GRANT SELECT ON TABLES TO analytics_user"

echo "✓ Permissions granted"

PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$MAIN_DB" -c "CREATE TABLE IF NOT EXISTS logs.health_policies (service VARCHAR(50) PRIMARY KEY, policy_json JSONB NOT NULL DEFAULT '{}');"
PGPASSWORD="$DB_PASS" psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$MAIN_DB" -c "INSERT INTO logs.health_policies (service, policy_json) VALUES ('portal', '{}'), ('review', '{}'), ('logs', '{}'), ('analytics', '{}') ON CONFLICT (service) DO NOTHING;"

echo "✓ Health policies seeded"
