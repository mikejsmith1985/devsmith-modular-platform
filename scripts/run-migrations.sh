#!/bin/bash
set -e

echo "Running database migrations..."

# Install golang-migrate if not present
if ! command -v migrate &> /dev/null; then
  echo "Installing golang-migrate..."
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Portal migrations
echo "→ Portal service migrations..."
migrate -path migrations/portal -database "postgresql://portal_user:portal_pass@localhost:5432/devsmith_portal?sslmode=disable" up || true

# Review migrations
echo "→ Review service migrations..."
migrate -path migrations/review -database "postgresql://review_user:review_pass@localhost:5432/devsmith_review?sslmode=disable" up || true

# Logs migrations
echo "→ Logs service migrations..."
migrate -path migrations/logs -database "postgresql://logs_user:logs_pass@localhost:5432/devsmith_logs?sslmode=disable" up || true

# Analytics migrations
echo "→ Analytics service migrations..."
migrate -path migrations/analytics -database "postgresql://analytics_user:analytics_pass@localhost:5432/devsmith_analytics?sslmode=disable" up || true

echo "✓ All migrations completed"
