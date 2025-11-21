#!/usr/bin/env bash
set -euo pipefail

# Validate Traefik routers and priorities via Traefik dashboard API
# This script will:
#  - Query Traefik API for routers
#  - Print router name, rule, priority, entrypoint, and service
#  - Check for common issues: missing portal-ui, missing api routers

TRAEFIK_DASHBOARD_URL=${TRAEFIK_DASHBOARD_URL:-http://localhost:8090}

echo "Querying Traefik dashboard at ${TRAEFIK_DASHBOARD_URL}"
ROUTERS_JSON=$(curl -sS "${TRAEFIK_DASHBOARD_URL}/api/http/routers")
if [ -z "${ROUTERS_JSON}" ]; then
  echo "ERROR: Traefik dashboard returned empty response. Is Traefik running on ${TRAEFIK_DASHBOARD_URL}?"
  exit 1
fi

echo "Routers found (name : rule : priority : entrypoints : service):"
echo "---------------------------------------------------------------"
echo "${ROUTERS_JSON}" | jq -r '.[] | "\(.name) : \(.rule) : \(.priority) : \(.entrypoints | join(",")) : \(.service)"'

echo "\nChecking portal UI router..."
if echo "${ROUTERS_JSON}" | jq -e '.[] | select(.name == "portal-ui")' >/dev/null; then
  echo "  portal-ui found"
else
  echo "  WARNING: portal-ui route not found. This will cause UI paths to route incorrectly."
fi

echo "\nChecking API routers priorities (portal-api, review, logs, analytics)..."
for r in portal-api review logs analytics; do
  if echo "${ROUTERS_JSON}" | jq -e --arg name "$r" '.[] | select(.name == $name)' >/dev/null; then
    pr=$(echo "${ROUTERS_JSON}" | jq -r --arg name "$r" '.[] | select(.name == $name) | .priority')
    echo "  $r present with priority $pr"
  else
    echo "  WARNING: $r router missing"
  fi
done

echo "\nDone. If UI routes like /logs or /analytics 404 via Traefik, ensure portal-ui router has PathPrefix('/') and that portal is healthy and its index.html is accessible."
echo "Also check for any other services that define PathPrefix('/') with higher priority than portal-ui."

exit 0
