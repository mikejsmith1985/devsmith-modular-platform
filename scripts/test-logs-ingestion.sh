#!/usr/bin/env bash
set -euo pipefail

# Simple integration helper to POST a test log and verify ingestion via the Logs service API.
# Usage: LOGS_SERVICE_URL=http://localhost:8082/api/logs ./scripts/test-logs-ingestion.sh

LOGS_URL=${LOGS_SERVICE_URL:-http://localhost:8082/api/logs}
UNIQUE="ingest-$(date +%s)-$RANDOM"
PAYLOAD=$(cat <<EOF
{"service":"tests-ingest","level":"INFO","message":"integration-test-${UNIQUE}","metadata":{}}
EOF
)

echo "Posting test log to ${LOGS_URL}"
curl -s -X POST -H "Content-Type: application/json" -d "$PAYLOAD" "$LOGS_URL" || {
  echo "POST failed" >&2
  exit 2
}

echo "Polling for ingestion (10s max)"
for i in {1..10}; do
  sleep 1
  # Query endpoint — note: server must support `search` query parameter
  if curl -s "${LOGS_URL%/}/?search=integration-test-${UNIQUE}" | grep -q "integration-test-${UNIQUE}"; then
    echo "Found ingested log: integration-test-${UNIQUE}"
    exit 0
  fi
done

echo "Timed out waiting for ingested log" >&2
exit 3
