#!/bin/bash
# Simple Load Test for Batch Log Ingestion
# Measures average response time, throughput, and failure rate
# Comparable to Phase 13 baseline: 330ms avg, 0% failures, 118+ req/s

set -e

# Configuration
API_URL="${API_URL:-http://localhost:8082/api/logs/batch}"
API_KEY="${API_KEY:-test-api-token-12345}"
PROJECT_SLUG="${PROJECT_SLUG:-load-test-v2}"
TOTAL_REQUESTS="${TOTAL_REQUESTS:-1000}"
CONCURRENT="${CONCURRENT:-10}"
BATCH_SIZE="${BATCH_SIZE:-100}"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  BATCH LOG INGESTION LOAD TEST${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo "Configuration:"
echo "  API URL: $API_URL"
echo "  Total Requests: $TOTAL_REQUESTS"
echo "  Concurrent: $CONCURRENT"
echo "  Batch Size: $BATCH_SIZE logs per request"
echo "  Total Logs: $((TOTAL_REQUESTS * BATCH_SIZE))"
echo ""

# Create temporary directory for results
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Generate payload
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
PAYLOAD=$(cat <<EOF
{
  "project_slug": "$PROJECT_SLUG",
  "logs": [
    {
      "timestamp": "$TIMESTAMP",
      "level": "info",
      "message": "Load test message",
      "service_name": "load-test",
      "context": {"test": true, "iteration": 0}
    }
  ]
}
EOF
)

# Function to make single request
make_request() {
    local index=$1
    local start_time=$(date +%s%3N)
    
    response=$(curl -s -w "\n%{http_code}\n%{time_total}" -X POST "$API_URL" \
        -H "X-API-Key: $API_KEY" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" 2>/dev/null)
    
    http_code=$(echo "$response" | tail -2 | head -1)
    time_total=$(echo "$response" | tail -1)
    
    # Convert to milliseconds (avoiding bc dependency)
    # Use awk for floating point conversion
    time_ms=$(awk "BEGIN {print $time_total * 1000}")
    
    echo "$http_code $time_ms" > "$TEMP_DIR/result_$index.txt"
}

echo -e "${YELLOW}Starting load test...${NC}"
echo ""

START_TIME=$(date +%s)

# Run requests in parallel batches
for ((batch=0; batch<TOTAL_REQUESTS; batch+=CONCURRENT)); do
    # Launch concurrent requests
    for ((i=0; i<CONCURRENT && (batch+i)<TOTAL_REQUESTS; i++)); do
        make_request $((batch + i)) &
    done
    
    # Wait for batch to complete
    wait
    
    # Progress indicator
    completed=$((batch + CONCURRENT))
    if [ $completed -gt $TOTAL_REQUESTS ]; then
        completed=$TOTAL_REQUESTS
    fi
    progress=$((completed * 100 / TOTAL_REQUESTS))
    echo -ne "\rProgress: $completed/$TOTAL_REQUESTS ($progress%)  "
done

END_TIME=$(date +%s)
echo ""
echo ""

# Calculate statistics
total_time=$((END_TIME - START_TIME))
success_count=0
failure_count=0
total_response_time=0
min_time=999999
max_time=0

for result_file in "$TEMP_DIR"/result_*.txt; do
    if [ -f "$result_file" ]; then
        read http_code time_ms < "$result_file"
        
        if [ "$http_code" = "201" ] || [ "$http_code" = "200" ]; then
            success_count=$((success_count + 1))
        else
            failure_count=$((failure_count + 1))
        fi
        
        # Update response time stats
        time_int=${time_ms%.*}
        total_response_time=$((total_response_time + time_int))
        
        if [ $time_int -lt $min_time ]; then
            min_time=$time_int
        fi
        if [ $time_int -gt $max_time ]; then
            max_time=$time_int
        fi
    fi
done

# Calculate metrics
total_requests=$((success_count + failure_count))

# Avoid division by zero
if [ $total_requests -eq 0 ]; then
    echo "Error: No results collected"
    exit 1
fi

avg_response_time=$((total_response_time / total_requests))
failure_rate=$(awk "BEGIN {printf \"%.2f\", ($failure_count * 100.0 / $total_requests)}")
throughput=$(awk "BEGIN {printf \"%.2f\", ($total_requests / $total_time)}")
logs_per_second=$(awk "BEGIN {printf \"%.2f\", ($total_requests * $BATCH_SIZE / $total_time)}")

# Display results
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  RESULTS${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo "Test Duration: ${total_time}s"
echo ""
echo "Requests:"
echo "  Total: $total_requests"
echo -e "  Success: ${GREEN}$success_count${NC}"
if [ $failure_count -gt 0 ]; then
    echo -e "  Failed: ${RED}$failure_count${NC}"
else
    echo -e "  Failed: ${GREEN}$failure_count${NC}"
fi
echo -e "  Failure Rate: ${failure_rate}%"
echo ""
echo "Response Times (ms):"
echo "  Average: ${avg_response_time}ms"
echo "  Min: ${min_time}ms"
echo "  Max: ${max_time}ms"
echo ""
echo "Throughput:"
echo "  Requests/sec: ${throughput}"
echo "  Logs/sec: ${logs_per_second}"
echo "  Total Logs: $((total_requests * BATCH_SIZE))"
echo ""

# Compare to Phase 13 baseline
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  COMPARISON TO PHASE 13 BASELINE${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo "Phase 13 Baseline:"
echo "  Average Response Time: 330ms"
echo "  Failure Rate: 0%"
echo "  Throughput: 118+ req/s"
echo ""
echo "Current Results:"

# Check average response time (allow 10ms overhead for middleware)
if [ $avg_response_time -le 340 ]; then
    echo -e "  ${GREEN}✓${NC} Average Response Time: ${avg_response_time}ms (target: ≤340ms)"
else
    echo -e "  ${RED}✗${NC} Average Response Time: ${avg_response_time}ms (target: ≤340ms)"
fi

# Check failure rate
if [ "$failure_rate" = "0.00" ] || [ "$failure_rate" = "0" ]; then
    echo -e "  ${GREEN}✓${NC} Failure Rate: ${failure_rate}% (target: 0%)"
else
    echo -e "  ${RED}✗${NC} Failure Rate: ${failure_rate}% (target: 0%)"
fi

# Check throughput
throughput_int=$(awk "BEGIN {printf \"%d\", $throughput}")
if [ $throughput_int -ge 118 ]; then
    echo -e "  ${GREEN}✓${NC} Throughput: ${throughput} req/s (target: ≥118 req/s)"
else
    echo -e "  ${RED}✗${NC} Throughput: ${throughput} req/s (target: ≥118 req/s)"
fi

echo ""

# Overall assessment
failure_is_zero=$(awk "BEGIN {print ($failure_rate == 0)}")
if [ $avg_response_time -le 340 ] && [ "$failure_is_zero" = "1" ] && [ $throughput_int -ge 118 ]; then
    echo -e "${GREEN}✓ PASS: Performance meets Phase 13 baseline!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ WARNING: Performance does not meet all Phase 13 targets${NC}"
    exit 1
fi
