#!/bin/bash

# Phase 6 Verification Script
# Runs all new testing features to verify completion

set -e

echo "=========================================="
echo "Phase 6 Verification Script"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

# Function to print info
print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Check prerequisites
echo "Step 1: Checking Prerequisites"
echo "----------------------------------------------"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running${NC}"
    exit 1
fi
print_status 0 "Docker is running"

# Check if services are up
if ! docker-compose ps | grep -q "Up"; then
    print_info "Services not running. Starting platform..."
    docker-compose up -d
    sleep 10
fi
print_status 0 "Platform services running"

echo ""

# Test 1: Healthcheck (Nginx Routing)
echo "Step 2: Testing Nginx Routing Fix"
echo "----------------------------------------------"

# Rebuild healthcheck binary
go build -o healthcheck ./cmd/healthcheck 2>/dev/null
print_status $? "Healthcheck binary rebuilt"

# Run health check
ROUTES_COUNT=$(./scripts/health-check-cli.sh --quick 2>&1 | grep -o '[0-9]* routes discovered' | awk '{print $1}')

if [ "$ROUTES_COUNT" -eq 9 ]; then
    print_status 0 "Nginx routing: 9 routes discovered"
else
    print_status 1 "Nginx routing: Expected 9 routes, found $ROUTES_COUNT"
fi

echo ""

# Test 2: Swagger UI
echo "Step 3: Testing Swagger UI"
echo "----------------------------------------------"

# Check if swagger-ui container is running
if docker-compose ps swagger-ui | grep -q "Up"; then
    print_status 0 "Swagger UI container running"
else
    print_info "Starting Swagger UI..."
    docker-compose up -d swagger-ui
    sleep 5
fi

# Test direct access
if curl -s http://localhost:8090 > /dev/null; then
    print_status 0 "Swagger UI accessible (direct: port 8090)"
else
    print_status 1 "Swagger UI not accessible (direct)"
fi

# Test proxied access
if curl -s http://localhost:3000/docs/review > /dev/null; then
    print_status 0 "Swagger UI accessible (nginx: /docs/review)"
else
    print_status 1 "Swagger UI not accessible (nginx)"
fi

echo ""

# Test 3: E2E Tests
echo "Step 4: Running E2E Tests"
echo "----------------------------------------------"

print_info "Running 15 E2E tests (all reading modes)..."

if npx playwright test tests/e2e/review/all-reading-modes.spec.ts --quiet; then
    print_status 0 "E2E tests passed (15 tests)"
else
    print_status 1 "E2E tests failed"
    print_info "View report: npx playwright show-report"
fi

echo ""

# Test 4: Accessibility Tests
echo "Step 5: Running Accessibility Tests"
echo "----------------------------------------------"

print_info "Running 8 accessibility tests (WCAG 2.1 Level AA)..."

if npx playwright test tests/e2e/review/accessibility.spec.ts --quiet; then
    print_status 0 "Accessibility tests passed (8 tests)"
else
    print_status 1 "Accessibility tests failed"
    print_info "View report: npx playwright show-report"
fi

echo ""

# Test 5: K6 Load Test (Optional)
echo "Step 6: K6 Load Test (Optional)"
echo "----------------------------------------------"

if command -v k6 > /dev/null 2>&1; then
    print_info "k6 found. Running load test (10 VUs, 100 iterations)..."
    
    if k6 run tests/k6/review-load.js --quiet; then
        print_status 0 "k6 load test passed"
        print_info "Fill baseline report: .docs/perf/review-k6-baseline.md"
    else
        print_status 1 "k6 load test failed"
    fi
else
    print_info "k6 not installed. Skipping load test."
    print_info "Install: https://k6.io/docs/get-started/installation/"
fi

echo ""

# Test 6: CI/CD Workflow
echo "Step 7: CI/CD Workflow Validation"
echo "----------------------------------------------"

if [ -f .github/workflows/quality-performance.yml ]; then
    print_status 0 "quality-performance.yml exists"
    
    # Check job definitions
    if grep -q "tests:" .github/workflows/quality-performance.yml; then
        print_status 0 "Tests job defined"
    fi
    
    if grep -q "benchmarks:" .github/workflows/quality-performance.yml; then
        print_status 0 "Benchmarks job defined"
    fi
    
    if grep -q "e2e-smoke:" .github/workflows/quality-performance.yml; then
        print_status 0 "E2E smoke job defined"
    fi
    
    if grep -q "accessibility:" .github/workflows/quality-performance.yml; then
        print_status 0 "Accessibility job defined"
    fi
    
    if grep -q "openapi:" .github/workflows/quality-performance.yml; then
        print_status 0 "OpenAPI job defined"
    fi
    
    if grep -q "quality-gate:" .github/workflows/quality-performance.yml; then
        print_status 0 "Quality gate job defined"
    fi
else
    print_status 1 "quality-performance.yml not found"
fi

echo ""

# Summary
echo "=========================================="
echo "Phase 6 Verification Complete"
echo "=========================================="
echo ""
echo "Next Steps:"
echo "1. Review any failed tests above"
echo "2. Fill k6 baseline report: .docs/perf/review-k6-baseline.md"
echo "3. Commit changes: git add . && git commit -m 'feat(phase6): Complete backlog'"
echo "4. Push to trigger CI: git push origin development"
echo "5. Monitor CI workflow: gh run watch"
echo ""
echo "Documentation: .docs/PHASE-6-COMPLETION-SUMMARY.md"
echo ""
