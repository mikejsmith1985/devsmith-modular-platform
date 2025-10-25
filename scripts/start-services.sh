#!/bin/bash
set -e

echo "Starting services..."

# Create logs directory
mkdir -p logs

# Start services in background
echo "→ Starting Portal service (port 8080)..."
./bin/portal > logs/portal.log 2>&1 &
PORTAL_PID=$!
echo $PORTAL_PID > .pid_portal
sleep 1

echo "→ Starting Review service (port 8081)..."
./bin/review > logs/review.log 2>&1 &
REVIEW_PID=$!
echo $REVIEW_PID > .pid_review
sleep 1

echo "→ Starting Logs service (port 8082)..."
./bin/logs > logs/logs.log 2>&1 &
LOGS_PID=$!
echo $LOGS_PID > .pid_logs
sleep 1

echo "→ Starting Analytics service (port 8083)..."
./bin/analytics > logs/analytics.log 2>&1 &
ANALYTICS_PID=$!
echo $ANALYTICS_PID > .pid_analytics
sleep 1

echo "✓ All services started"
echo "  Portal PID:    $PORTAL_PID"
echo "  Review PID:    $REVIEW_PID"
echo "  Logs PID:      $LOGS_PID"
echo "  Analytics PID: $ANALYTICS_PID"
echo "  Logs available in logs/ directory"
