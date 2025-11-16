#!/usr/bin/env bash
# Simple smoke-check for DevSmith stack via nginx gateway
# Exits non-zero if any check fails
set -euo pipefail
GATEWAY_URL=${1:-http://localhost:3000}
OK=0
FAIL=0
check(){
  local url="$1"
  local expect=${2:-200}
  status=$(curl -s -o /dev/null -w "%{http_code}" "$url" || echo "000")
  if [ "$status" = "$expect" ]; then
    printf "OK   %s -> %s\n" "$url" "$status"
    OK=$((OK+1))
  else
    printf "FAIL %s -> %s (expected %s)\n" "$url" "$status" "$expect"
    FAIL=$((FAIL+1))
  fi
}

printf "Running gateway smoke checks against %s\n" "$GATEWAY_URL"
check "$GATEWAY_URL/" 200
check "$GATEWAY_URL/review" 200
check "$GATEWAY_URL/static/css/review.css" 200
check "$GATEWAY_URL/static/js/review.js" 200
check "$GATEWAY_URL/static/js/analysis.js" 200
check "$GATEWAY_URL/static/dashboard.css" 200

printf "\nSummary: %d OK, %d FAIL\n" "$OK" "$FAIL"
if [ "$FAIL" -ne 0 ]; then
  exit 2
fi
exit 0
