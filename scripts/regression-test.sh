#!/bin/bash

# DevSmith Regression Test Suite
# Tests all critical user workflows with screenshot capture
# MUST PASS before declaring any work "complete"

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:3000}"
SCREENSHOT_DIR="test-results/regression-$(date +%Y%m%d-%H%M%S)"
RESULTS_FILE="$SCREENSHOT_DIR/results.json"
SUMMARY_FILE="$SCREENSHOT_DIR/SUMMARY.md"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Create screenshot directory
mkdir -p "$SCREENSHOT_DIR"

# Initialize results
echo '{"tests": [], "summary": {}}' > "$RESULTS_FILE"

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Test result tracking
record_test() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    local screenshot="$4"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$status" = "pass" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "$test_name"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "$test_name: $message"
    fi
    
    # Append to results file
    jq --arg name "$test_name" \
       --arg status "$status" \
       --arg message "$message" \
       --arg screenshot "$screenshot" \
       '.tests += [{"name": $name, "status": $status, "message": $message, "screenshot": $screenshot}]' \
       "$RESULTS_FILE" > "$RESULTS_FILE.tmp" && mv "$RESULTS_FILE.tmp" "$RESULTS_FILE"
}

# Check if service is running
check_service() {
    local url="$1"
    local service_name="$2"
    
    # 20s timeout to accommodate slow Ollama health checks
    if timeout 20 curl -sf "$url" > /dev/null 2>&1; then
        return 0
    else
        log_error "Service $service_name not responding at $url"
        return 1
    fi
}

# Take screenshot using Playwright
take_screenshot() {
    local url="$1"
    local filename="$2"
    local description="$3"
    
    log_info "Capturing: $description"
    
    npx playwright screenshot \
        --browser chromium \
        --full-page \
        "$url" \
        "$SCREENSHOT_DIR/$filename" 2>/dev/null || {
            log_warning "Screenshot failed for $filename (continuing...)"
            return 1
        }
    
    return 0
}

# Visual inspection prompt
prompt_visual_inspection() {
    local screenshot="$1"
    local expected="$2"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "VISUAL INSPECTION REQUIRED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Screenshot: $screenshot"
    echo "Expected: $expected"
    echo ""
    echo "Please open the screenshot and verify it matches expectations."
    echo "Press 'p' for PASS, 'f' for FAIL, or 's' to skip"
    read -r -n 1 response
    echo ""
    
    case "$response" in
        p|P) return 0 ;;
        f|F) return 1 ;;
        s|S) return 2 ;;
        *) return 2 ;;
    esac
}

# ============================================================================
# PRE-FLIGHT CHECKS
# ============================================================================

log_info "Starting Regression Test Suite"
log_info "Base URL: $BASE_URL"
log_info "Screenshots: $SCREENSHOT_DIR"
echo ""

# Check Docker services
log_info "Checking Docker services..."
if ! docker-compose ps 2>&1 | grep -q "Up"; then
    log_error "Docker services not running. Start with: docker-compose up -d"
    exit 1
fi

# Check all service health
log_info "Checking service health..."
check_service "$BASE_URL" "Gateway (Nginx)" || exit 1
check_service "$BASE_URL/health" "Portal" || exit 1
check_service "http://localhost:8081/health" "Review" || exit 1
check_service "http://localhost:8082/health" "Logs" || exit 1
check_service "http://localhost:8083/health" "Analytics" || exit 1

log_success "All services healthy"
echo ""

# ============================================================================
# TEST 1: Portal Dashboard
# ============================================================================

log_info "━━━ TEST 1: Portal Dashboard ━━━"

take_screenshot "$BASE_URL" "01-portal-landing.png" "Portal landing page"

if take_screenshot "$BASE_URL" "01-portal-landing.png" "Portal landing page"; then
    record_test "Portal Landing Page Screenshot" "pass" "Screenshot captured" "01-portal-landing.png"
else
    record_test "Portal Landing Page Screenshot" "fail" "Failed to capture screenshot" ""
fi

# Visual check for portal elements
log_info "Checking portal for expected elements..."
PORTAL_HTML=$(curl -s "$BASE_URL")

if echo "$PORTAL_HTML" | grep -q "DevSmith"; then
    record_test "Portal Title Visible" "pass" "DevSmith title found" ""
else
    record_test "Portal Title Visible" "fail" "DevSmith title not found" ""
fi

if echo "$PORTAL_HTML" | grep -q "Login\|Sign in"; then
    record_test "Portal Login Button Visible" "pass" "Login button found" ""
else
    record_test "Portal Login Button Visible" "fail" "Login button not found" ""
fi

# ============================================================================
# TEST 2: Review Service UI
# ============================================================================

log_info "━━━ TEST 2: Review Service UI ━━━"

take_screenshot "http://localhost:8081/review" "02-review-landing.png" "Review service landing"

# Review service should respond with either:
# 1. JSON error ({"error":"Authentication required"}) for API requests
# 2. HTML redirect to login for browser requests
# 3. 401 status code
REVIEW_RESPONSE=$(curl -s -w "\n%{http_code}" "http://localhost:8081/review" 2>&1 || echo "")

# Check if service is responding correctly (401 or JSON error or redirect)
if echo "$REVIEW_RESPONSE" | grep -q -i "Authentication required\|401\|302\|Found"; then
    record_test "Review Service Accessible" "pass" "Review service responding correctly (auth required)" "02-review-landing.png"
else
    record_test "Review Service Accessible" "fail" "Review service not responding: $REVIEW_RESPONSE" "02-review-landing.png"
fi

# ============================================================================
# TEST 3: Logs Service UI
# ============================================================================

log_info "━━━ TEST 3: Logs Service UI ━━━"

take_screenshot "http://localhost:8082" "03-logs-landing.png" "Logs service landing"

LOGS_HTML=$(curl -s "http://localhost:8082" || echo "")

if echo "$LOGS_HTML" | grep -q -i "log\|entry\|monitor"; then
    record_test "Logs Service Accessible" "pass" "Logs service responding" "03-logs-landing.png"
else
    record_test "Logs Service Accessible" "fail" "Logs service not responding" "03-logs-landing.png"
fi

# ============================================================================
# TEST 4: Analytics Service UI
# ============================================================================

log_info "━━━ TEST 4: Analytics Service UI ━━━"

take_screenshot "http://localhost:8083" "04-analytics-landing.png" "Analytics service landing"

ANALYTICS_HTML=$(curl -s "http://localhost:8083" || echo "")

if echo "$ANALYTICS_HTML" | grep -q -i "analytic\|metric\|trend"; then
    record_test "Analytics Service Accessible" "pass" "Analytics service responding" "04-analytics-landing.png"
else
    record_test "Analytics Service Accessible" "fail" "Analytics service not responding" "04-analytics-landing.png"
fi

# ============================================================================
# TEST 5: API Health Endpoints
# ============================================================================

log_info "━━━ TEST 5: API Health Endpoints ━━━"

# Portal health
PORTAL_HEALTH=$(curl -s "$BASE_URL/health" || echo '{}')
if echo "$PORTAL_HEALTH" | jq -e '.status' > /dev/null 2>&1; then
    record_test "Portal Health Endpoint" "pass" "Health check passed" ""
else
    record_test "Portal Health Endpoint" "fail" "Health check failed or invalid JSON" ""
fi

# Review health
REVIEW_HEALTH=$(curl -s "http://localhost:8081/health" || echo '{}')
if echo "$REVIEW_HEALTH" | jq -e '.status' > /dev/null 2>&1; then
    record_test "Review Health Endpoint" "pass" "Health check passed" ""
else
    record_test "Review Health Endpoint" "fail" "Health check failed or invalid JSON" ""
fi

# Logs health
LOGS_HEALTH=$(curl -s "http://localhost:8082/health" || echo '{}')
if echo "$LOGS_HEALTH" | jq -e '.status' > /dev/null 2>&1; then
    record_test "Logs Health Endpoint" "pass" "Health check passed" ""
else
    record_test "Logs Health Endpoint" "fail" "Health check failed or invalid JSON" ""
fi

# Analytics health
ANALYTICS_HEALTH=$(curl -s "http://localhost:8083/health" || echo '{}')
if echo "$ANALYTICS_HEALTH" | jq -e '.status' > /dev/null 2>&1; then
    record_test "Analytics Health Endpoint" "pass" "Health check passed" ""
else
    record_test "Analytics Health Endpoint" "fail" "Health check failed or invalid JSON" ""
fi

# ============================================================================
# TEST 6: Database Connectivity
# ============================================================================

log_info "━━━ TEST 6: Database Connectivity ━━━"

# Check if logs.entries table exists with AI columns
DB_CHECK=$(docker-compose exec -T postgres psql -U devsmith -d devsmith -c "\d logs.entries" 2>&1 || echo "")

if echo "$DB_CHECK" | grep -q "issue_type"; then
    record_test "Phase 1 AI Columns Exist" "pass" "issue_type column found in logs.entries" ""
else
    record_test "Phase 1 AI Columns Exist" "fail" "issue_type column not found" ""
fi

if echo "$DB_CHECK" | grep -q "ai_analysis"; then
    record_test "AI Analysis Column Exists" "pass" "ai_analysis JSONB column found" ""
else
    record_test "AI Analysis Column Exists" "fail" "ai_analysis column not found" ""
fi

if echo "$DB_CHECK" | grep -q "severity_score"; then
    record_test "Severity Score Column Exists" "pass" "severity_score column found" ""
else
    record_test "Severity Score Column Exists" "fail" "severity_score column not found" ""
fi

# ============================================================================
# TEST 7: Nginx Gateway Routing
# ============================================================================

log_info "━━━ TEST 7: Nginx Gateway Routing ━━━"

# Test gateway routes to each service
GATEWAY_PORTAL=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/")
if [ "$GATEWAY_PORTAL" = "200" ] || [ "$GATEWAY_PORTAL" = "302" ]; then
    record_test "Gateway Routes to Portal" "pass" "HTTP $GATEWAY_PORTAL" ""
else
    record_test "Gateway Routes to Portal" "fail" "HTTP $GATEWAY_PORTAL" ""
fi

# ============================================================================
# GENERATE SUMMARY
# ============================================================================

log_info "━━━ Generating Summary ━━━"

# Update results summary
jq --argjson total "$TOTAL_TESTS" \
   --argjson passed "$PASSED_TESTS" \
   --argjson failed "$FAILED_TESTS" \
   '.summary = {"total": $total, "passed": $passed, "failed": $failed, "pass_rate": (($passed / $total) * 100 | floor)}' \
   "$RESULTS_FILE" > "$RESULTS_FILE.tmp" && mv "$RESULTS_FILE.tmp" "$RESULTS_FILE"

# Generate markdown summary
cat > "$SUMMARY_FILE" << EOF
# Regression Test Summary

**Date**: $(date '+%Y-%m-%d %H:%M:%S')  
**Base URL**: $BASE_URL  
**Branch**: $(git branch --show-current)  
**Commit**: $(git rev-parse --short HEAD)

## Results

- **Total Tests**: $TOTAL_TESTS
- **Passed**: $PASSED_TESTS ✓
- **Failed**: $FAILED_TESTS ✗
- **Pass Rate**: $(( (PASSED_TESTS * 100) / TOTAL_TESTS ))%

## Test Details

EOF

# Append test results to summary
jq -r '.tests[] | "### \(.name)\n- **Status**: \(.status)\n- **Message**: \(.message)\n- **Screenshot**: \(.screenshot)\n"' "$RESULTS_FILE" >> "$SUMMARY_FILE"

# Display final results
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "REGRESSION TEST RESULTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Total Tests:  $TOTAL_TESTS"
echo "Passed:       $GREEN$PASSED_TESTS ✓$NC"
echo "Failed:       $RED$FAILED_TESTS ✗$NC"
echo "Pass Rate:    $(( (PASSED_TESTS * 100) / TOTAL_TESTS ))%"
echo ""
echo "Results saved to: $SCREENSHOT_DIR"
echo "Summary: $SUMMARY_FILE"
echo "JSON: $RESULTS_FILE"
echo ""

# Exit code based on pass/fail
if [ "$FAILED_TESTS" -gt 0 ]; then
    log_error "REGRESSION TESTS FAILED"
    echo ""
    echo "❌ DO NOT proceed with PR creation or declaring work 'complete'"
    echo "❌ Fix failing tests and re-run"
    exit 1
else
    log_success "ALL REGRESSION TESTS PASSED"
    echo ""
    echo "✅ OK to proceed with PR creation"
    exit 0
fi
