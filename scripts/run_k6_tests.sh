#!/bin/bash

#
# DevSmith k6 Performance Test Runner
#
# Runs all k6 performance tests and aggregates results.
#
# Usage:
#   ./scripts/run_k6_tests.sh                    # Run all tests
#   ./scripts/run_k6_tests.sh --base-url http://prod:3000
#   ./scripts/run_k6_tests.sh --verbose
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:3000}"
K6_DIR="tests/k6"
RESULTS_DIR="test-results/k6"
VERBOSE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --base-url)
            BASE_URL="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "Usage: ./scripts/run_k6_tests.sh [options]"
            echo ""
            echo "Options:"
            echo "  --base-url URL    Base URL for API (default: http://localhost:3000)"
            echo "  --verbose, -v     Verbose output"
            echo "  --help, -h        Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}❌ k6 is not installed${NC}"
    echo "Install k6: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# Create results directory
mkdir -p "$RESULTS_DIR"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         DevSmith k6 Performance Test Suite                 ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "Base URL: $BASE_URL"
echo "Results: $RESULTS_DIR"
echo ""

# Test counters
PASSED=0
FAILED=0
TOTAL=0

# Array to store test results
declare -a TEST_RESULTS

# Function to run a single test
run_test() {
    local test_file=$1
    local test_name=$(basename "$test_file" .js)
    
    TOTAL=$((TOTAL + 1))
    
    echo -e "${YELLOW}→ Running: $test_name${NC}"
    
    # Build k6 command
    local k6_cmd="k6 run $test_file --env BASE_URL=$BASE_URL"
    
    if [ "$VERBOSE" = true ]; then
        k6_cmd="$k6_cmd -v"
    fi
    
    # Save output to file
    local output_file="$RESULTS_DIR/${test_name}.txt"
    local json_file="$RESULTS_DIR/${test_name}.json"
    
    # Run test
    if output=$($k6_cmd --out json=$json_file 2>&1); then
        echo "$output" > "$output_file"
        echo -e "${GREEN}  ✓ PASSED${NC}"
        PASSED=$((PASSED + 1))
        TEST_RESULTS+=("$test_name:PASS")
    else
        echo "$output" > "$output_file"
        echo -e "${RED}  ✗ FAILED${NC}"
        FAILED=$((FAILED + 1))
        TEST_RESULTS+=("$test_name:FAIL")
    fi
    echo ""
}

# Run all tests
for test_file in "$K6_DIR"/*.js; do
    if [ -f "$test_file" ]; then
        run_test "$test_file"
    fi
done

# Print summary
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    Test Summary                           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo "Total Tests: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""

# Print individual test results
echo "Test Results:"
for result in "${TEST_RESULTS[@]}"; do
    test_name="${result%:*}"
    status="${result#*:}"
    
    if [ "$status" = "PASS" ]; then
        echo -e "  ${GREEN}✓${NC} $test_name"
    else
        echo -e "  ${RED}✗${NC} $test_name"
    fi
done

echo ""
echo "Results saved to: $RESULTS_DIR/"
echo ""

# Exit code
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
