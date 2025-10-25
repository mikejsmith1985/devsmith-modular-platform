#!/bin/bash
set -e

echo "Creating databases and schemas..."

# Database names
PORTAL_DB="devsmith_portal"
REVIEW_DB="devsmith_review"
LOGS_DB="devsmith_logs"
ANALYTICS_DB="devsmith_analytics"

# Create databases
psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$PORTAL_DB'" | grep -q 1 || \
  psql -U postgres -c "CREATE DATABASE $PORTAL_DB"

psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$REVIEW_DB'" | grep -q 1 || \
  psql -U postgres -c "CREATE DATABASE $REVIEW_DB"

psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$LOGS_DB'" | grep -q 1 || \
  psql -U postgres -c "CREATE DATABASE $LOGS_DB"

psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$ANALYTICS_DB'" | grep -q 1 || \
  psql -U postgres -c "CREATE DATABASE $ANALYTICS_DB"

echo "✓ Databases created"

# Create schemas
psql -U postgres -d "$PORTAL_DB" -c "CREATE SCHEMA IF NOT EXISTS portal"
psql -U postgres -d "$REVIEW_DB" -c "CREATE SCHEMA IF NOT EXISTS review"
psql -U postgres -d "$LOGS_DB" -c "CREATE SCHEMA IF NOT EXISTS logs"
psql -U postgres -d "$ANALYTICS_DB" -c "CREATE SCHEMA IF NOT EXISTS analytics"

echo "✓ Schemas created"

# Create users (if not exist)
psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='portal_user'" | grep -q 1 || \
  psql -U postgres -c "CREATE USER portal_user WITH PASSWORD 'portal_pass'"

psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='review_user'" | grep -q 1 || \
  psql -U postgres -c "CREATE USER review_user WITH PASSWORD 'review_pass'"

psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='logs_user'" | grep -q 1 || \
  psql -U postgres -c "CREATE USER logs_user WITH PASSWORD 'logs_pass'"

psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='analytics_user'" | grep -q 1 || \
  psql -U postgres -c "CREATE USER analytics_user WITH PASSWORD 'analytics_pass'"

echo "✓ Users created"

# Grant permissions
psql -U postgres -d "$PORTAL_DB" -c "GRANT ALL PRIVILEGES ON SCHEMA portal TO portal_user"
psql -U postgres -d "$REVIEW_DB" -c "GRANT ALL PRIVILEGES ON SCHEMA review TO review_user"
psql -U postgres -d "$LOGS_DB" -c "GRANT ALL PRIVILEGES ON SCHEMA logs TO logs_user"
psql -U postgres -d "$ANALYTICS_DB" -c "GRANT ALL PRIVILEGES ON SCHEMA analytics TO analytics_user"

# Grant analytics READ-ONLY access to logs
psql -U postgres -d "$LOGS_DB" -c "GRANT USAGE ON SCHEMA logs TO analytics_user"
psql -U postgres -d "$LOGS_DB" -c "GRANT SELECT ON ALL TABLES IN SCHEMA logs TO analytics_user"
psql -U postgres -d "$LOGS_DB" -c "ALTER DEFAULT PRIVILEGES IN SCHEMA logs GRANT SELECT ON TABLES TO analytics_user"

echo "✓ Permissions granted"
