#!/bin/bash

echo "🛑 Stopping DevSmith Platform services..."

# Kill services by PID
for pidfile in .pid_*; do
  if [ -f "$pidfile" ]; then
    SERVICE=$(basename "$pidfile" | sed 's/.pid_//')
    PID=$(cat "$pidfile")

    if kill -0 "$PID" 2>/dev/null; then
      echo "→ Stopping $SERVICE (PID $PID)..."
      kill "$PID" 2>/dev/null || true
      sleep 1
      # Force kill if still running
      kill -9 "$PID" 2>/dev/null || true
    fi

    rm "$pidfile"
  fi
done

echo "✓ All services stopped"
echo ""
echo "Note: Databases and data are preserved"
echo "      To clean everything: ./teardown.sh --clean"

if [ "$1" == "--clean" ]; then
  echo ""
  echo "🗑️  Cleaning databases..."
  psql -U postgres -c "DROP DATABASE IF EXISTS devsmith_portal" 2>/dev/null || true
  psql -U postgres -c "DROP DATABASE IF EXISTS devsmith_review" 2>/dev/null || true
  psql -U postgres -c "DROP DATABASE IF EXISTS devsmith_logs" 2>/dev/null || true
  psql -U postgres -c "DROP DATABASE IF EXISTS devsmith_analytics" 2>/dev/null || true
  echo "✓ Databases dropped"
fi
