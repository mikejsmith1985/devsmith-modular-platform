#!/bin/bash
# DevSmith Modular Platform Test Script
set -e

echo "[DevSmith] Running tests..."

# Example: run Go tests for all services
for svc in portal review logs analytics; do
  if [ -d "cmd/$svc" ]; then
    echo "[DevSmith] Testing $svc..."
    (cd cmd/$svc && go test ./...)
  fi
done

# Include tests from the apps directory
for app in portal review; do
  if [ -d "apps/$app" ]; then
    echo "[DevSmith] Testing $app..."
    (cd apps/$app && go test ./...)
  fi
done

echo "[DevSmith] All tests complete."
