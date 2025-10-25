#!/bin/bash

echo "🔍 DevSmith Platform - Setup Verification"
echo "==========================================="
echo ""

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

echo "📊 Service Health:"
check_service "Portal" "http://localhost:8080/health"
check_service "Review" "http://localhost:8081/health"
check_service "Logs" "http://localhost:8082/health"
check_service "Analytics" "http://localhost:8083/health"

echo ""
echo "🗄️  Database Connections:"

# Check databases
for db in devsmith_portal devsmith_review devsmith_logs devsmith_analytics; do
  if psql -U postgres -lqt | cut -d \| -f 1 | grep -qw "$db"; then
    echo "✓ Database $db exists"
  else
    echo "❌ Database $db missing"
    FAILED=1
  fi
done

echo ""
echo "🌐 Platform URLs:"
echo "   Portal:    http://localhost:8080"
echo "   Review:    http://localhost:8081"
echo "   Logs:      http://localhost:8082"
echo "   Analytics: http://localhost:8083"

if [ $FAILED -eq 1 ]; then
  echo ""
  echo "❌ Some checks failed"
  exit 1
fi

echo ""
echo "✅ All systems operational!"
