#!/bin/bash

FAILED=0

check_service() {
  SERVICE=$1
  URL=$2

  if curl -f -s "$URL" > /dev/null 2>&1; then
    echo "✓ $SERVICE is healthy"
  else
    echo "❌ $SERVICE is NOT responding at $URL"
    FAILED=1
  fi
}

echo "Checking service health..."

check_service "Portal" "http://localhost:8080/health"
check_service "Review" "http://localhost:8081/health"
check_service "Logs" "http://localhost:8082/health"
check_service "Analytics" "http://localhost:8083/health"

if [ $FAILED -eq 1 ]; then
  echo ""
  echo "❌ Some services failed health checks"
  echo "   Check logs in logs/ directory"
  exit 1
fi

echo ""
echo "✓ All services are healthy"
