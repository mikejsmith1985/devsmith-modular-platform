#!/bin/bash
# bypass-audit.sh - Comprehensive quality gate bypass audit
# Run before every push: ./scripts/bypass-audit.sh
# Exit code 0 = pass, exit code 1 = fail

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FAILURES=0
WARNINGS=0

echo "=========================================="
echo "  PRE-PUSH BYPASS AUDIT"
echo "=========================================="
echo ""

# ===== DIRECT BYPASSES =====
echo "Checking Direct Bypasses..."

# Check for nolint comments
if git diff HEAD -- '*.go' | grep -E '^\+.*//\s*nolint' > /dev/null 2>&1; then
  echo -e "${RED}✗ FAIL: nolint bypass found${NC}"
  git diff HEAD -- '*.go' | grep -E '^\+.*//\s*nolint' || true
  ((FAILURES++))
else
  echo -e "${GREEN}✓ PASS: No nolint comments added${NC}"
fi

# Check for skip/ignore comments
if git diff HEAD -- '*.go' | grep -E '^\+.*(//skip:|//ignore:)' > /dev/null 2>&1; then
  echo -e "${RED}✗ FAIL: skip/ignore bypass found${NC}"
  git diff HEAD -- '*.go' | grep -E '^\+.*(//skip:|//ignore:)' || true
  ((FAILURES++))
else
  echo -e "${GREEN}✓ PASS: No skip/ignore comments added${NC}"
fi

# Check for suspicious error ignoring
if git diff HEAD -- '*.go' | grep -E '^\+.*_ = ' | grep -v -E 'defer|range|for|New[A-Z]' > /dev/null 2>&1; then
  echo -e "${YELLOW}⚠ WARNING: Check for intentional error ignoring${NC}"
  git diff HEAD -- '*.go' | grep -E '^\+.*_ = ' | grep -v -E 'defer|range|for|New[A-Z]' || true
  ((WARNINGS++))
else
  echo -e "${GREEN}✓ PASS: No suspicious error ignoring${NC}"
fi

echo ""

# ===== INDIRECT BYPASSES =====
echo "Checking Indirect Bypasses..."

# Check for t.Skip() in new test code
if git diff HEAD -- '*_test.go' | grep -E '^\+.*t\.Skip\(\)' > /dev/null 2>&1; then
  echo -e "${RED}✗ FAIL: t.Skip() found in new test code${NC}"
  git diff HEAD -- '*_test.go' | grep -E '^\+.*t\.Skip\(\)' || true
  ((FAILURES++))
else
  echo -e "${GREEN}✓ PASS: No test skips added${NC}"
fi

# Check for problematic comments
if git diff HEAD -- '*.go' '*_test.go' | grep -iE '^\+.*(pre-existing|will fix|beyond scope|legacy|flaky|TODO.*later)' > /dev/null 2>&1; then
  echo -e "${YELLOW}⚠ WARNING: Check for bypass rationalizations in comments${NC}"
  git diff HEAD -- '*.go' '*_test.go' | grep -iE '^\+.*(pre-existing|will fix|beyond scope|legacy|flaky|TODO.*later)' || true
  ((WARNINGS++))
else
  echo -e "${GREEN}✓ PASS: No bypass rationalization comments${NC}"
fi

echo ""

# ===== QUALITY GATES =====
echo "Checking Quality Gates..."

# Build check
if ! go build ./... > /dev/null 2>&1; then
  echo -e "${RED}✗ FAIL: go build failed${NC}"
  ((FAILURES++))
else
  echo -e "${GREEN}✓ PASS: go build successful${NC}"
fi

# Test check
TEST_OUTPUT=$(go test ./... -v 2>&1)
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL" || true)
SKIP_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- SKIP" || true)

if [ "$FAIL_COUNT" -gt 0 ]; then
  echo -e "${RED}✗ FAIL: $FAIL_COUNT test(s) failed${NC}"
  echo "$TEST_OUTPUT" | grep "^--- FAIL" || true
  ((FAILURES++))
else
  echo -e "${GREEN}✓ PASS: All tests passed${NC}"
fi

if [ "$SKIP_COUNT" -gt 0 ]; then
  echo -e "${YELLOW}⚠ WARNING: $SKIP_COUNT test(s) skipped${NC}"
  ((WARNINGS++))
fi

# Race detection
if ! go test ./... -race > /dev/null 2>&1; then
  echo -e "${RED}✗ FAIL: Race detection failed${NC}"
  ((FAILURES++))
else
  echo -e "${GREEN}✓ PASS: No race conditions${NC}"
fi

# Linting
if ! golangci-lint run ./... > /dev/null 2>&1; then
  echo -e "${RED}✗ FAIL: Linting failed${NC}"
  ((FAILURES++))
else
  echo -e "${GREEN}✓ PASS: Linting passed${NC}"
fi

# Vet
if ! go vet ./... > /dev/null 2>&1; then
  echo -e "${RED}✗ FAIL: go vet failed${NC}"
  ((FAILURES++))
else
  echo -e "${GREEN}✓ PASS: go vet passed${NC}"
fi

echo ""

# ===== SUMMARY =====
echo "=========================================="
if [ "$FAILURES" -eq 0 ]; then
  if [ "$WARNINGS" -eq 0 ]; then
    echo -e "${GREEN}✓ ALL CHECKS PASSED - Safe to push${NC}"
    echo "=========================================="
    exit 0
  else
    echo -e "${YELLOW}✓ PASSED with $WARNINGS warning(s) - Review before push${NC}"
    echo "=========================================="
    exit 0
  fi
else
  echo -e "${RED}✗ FAILED - $FAILURES issue(s) found${NC}"
  echo "Fix issues before pushing"
  echo "=========================================="
  exit 1
fi
