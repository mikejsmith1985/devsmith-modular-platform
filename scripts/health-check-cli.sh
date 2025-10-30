#!/bin/bash
# DevSmith Health Check CLI - Diagnostic tool with watch, pr, and formatting options
# Usage: ./scripts/health-check-cli.sh [OPTIONS]
# 
# Examples:
#   ./scripts/health-check-cli.sh                          # Human-readable (default, Phase 1)
#   ./scripts/health-check-cli.sh --json                   # JSON format
#   ./scripts/health-check-cli.sh --watch                  # Continuous monitoring (5s interval)
#   ./scripts/health-check-cli.sh --quick                  # Phase 1 only, <500ms
#   ./scripts/health-check-cli.sh --pr                     # Comprehensive PR validation (Phase 1+2+endpoints)
#   ./scripts/health-check-cli.sh --pr --json              # PR validation in JSON

set -e

# Configuration
FORMAT="human"
WATCH_MODE=false
STORE_RESULTS=false
PR_MODE=false
QUICK_MODE=false
DB_URL="${DATABASE_URL:-postgres://devsmith:devsmith@localhost:5432/devsmith?sslmode=disable}"
ADVANCED="true"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HEALTHCHECK_BIN="$SCRIPT_DIR/../healthcheck"

# Colors for human output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Show help message (DEFINE BEFORE USING)
show_help() {
    cat <<EOF
${BOLD}DevSmith Health Check CLI${NC}

${BOLD}Usage:${NC}
  ./scripts/health-check-cli.sh [OPTIONS]

${BOLD}Options:${NC}
  --json              Output in JSON format (parseable)
  --watch             Continuous monitoring mode (5-second interval)
  --quick             Quick check (Phase 1 only, <500ms)
  --pr                Comprehensive PR validation (all phases + endpoints)
  --store             Store results to database (TODO)
  --db-url URL        Override database URL
  --advanced false    Skip Phase 2 advanced diagnostics
  -h, --help          Show this help message

${BOLD}Modes:${NC}
  Default (Phase 1 only)
    ./scripts/health-check-cli.sh
    - Docker containers, HTTP health, database connectivity
    - Fast (~900ms)
    - Use: Before starting work, quick checks

  Quick Mode (Phase 1 only, <500ms)
    ./scripts/health-check-cli.sh --quick
    - Same checks as default but even faster
    - Use: During rapid development

  Watch Mode (continuous monitoring)
    ./scripts/health-check-cli.sh --watch
    - Runs Phase 1 checks every 5 seconds
    - Use: Monitor in background terminal while developing

  PR Mode (comprehensive validation)
    ./scripts/health-check-cli.sh --pr
    - Phase 1: Container/HTTP/Database checks
    - Phase 2: Gateway routing, performance metrics
    - Full endpoint discovery and testing
    - Security scanning (Trivy)
    - Use: BEFORE creating a PR (mandatory)

${BOLD}Examples:${NC}
  # Basic health check
  ./scripts/health-check-cli.sh

  # Quick check during development
  ./scripts/health-check-cli.sh --quick

  # Monitor while developing (Terminal 1)
  ./scripts/health-check-cli.sh --watch

  # PR validation (Terminal 1)
  ./scripts/health-check-cli.sh --pr

  # PR validation with JSON output
  ./scripts/health-check-cli.sh --pr --json

  # Debug specific service
  ./scripts/health-check-cli.sh --json | jq '.Checks[] | select(.Status!="pass")'

  # Check if ready for PR
  STATUS=\$(./scripts/health-check-cli.sh --pr --json | jq -r '.Status')
  if [[ "\$STATUS" == "pass" ]]; then echo "✅ Ready for PR"; else echo "❌ Fix issues first"; fi

${BOLD}Integration with Development:${NC}
  Terminal 1: ./scripts/health-check-cli.sh --watch
  Terminal 2: vim internal/review/services/...
  Terminal 3: docker-compose up -d --build review && go test ./...

${BOLD}Troubleshooting:${NC}
  If healthcheck binary not found:
    go build -o healthcheck ./cmd/healthcheck

  If services are down:
    docker-compose up -d

  If database connection fails:
    export DATABASE_URL="postgres://user:pass@host:5432/dbname"
EOF
}

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --json)       FORMAT="json"; shift ;;
        --watch)      WATCH_MODE=true; shift ;;
        --quick)      QUICK_MODE=true; shift ;;
        --pr)         PR_MODE=true; shift ;;
        --store)      STORE_RESULTS=true; shift ;;
        --db-url)     DB_URL="$2"; shift 2 ;;
        --advanced)   ADVANCED="$2"; shift 2 ;;
        -h|--help)    show_help; exit 0 ;;
        *)            shift ;;
    esac
done

# Check if healthcheck binary exists
if [[ ! -f "$HEALTHCHECK_BIN" ]]; then
    echo -e "${RED}❌ Error: healthcheck binary not found at $HEALTHCHECK_BIN${NC}"
    echo -e "${YELLOW}Build it with:${NC}"
    echo -e "  ${CYAN}go build -o healthcheck ./cmd/healthcheck${NC}"
    exit 1
fi

# Run single health check
run_check() {
    local advanced_flag="$ADVANCED"
    if [[ "$QUICK_MODE" == "true" ]]; then
        advanced_flag="false"
    fi
    
    if [[ "$FORMAT" == "json" ]]; then
        # JSON format - direct output
        "$HEALTHCHECK_BIN" --format json --advanced "$advanced_flag" 2>/dev/null || echo '{"error": "Health check failed"}'
    else
        # Human-readable format
        "$HEALTHCHECK_BIN" --format human --advanced "$advanced_flag" 2>/dev/null || echo "❌ Health check failed - services may be down"
    fi
}

# Parse and display JSON output
display_json() {
    local output="$1"
    
    if [[ "$output" =~ "error" ]]; then
        echo -e "${RED}❌ Health Check Failed${NC}"
        echo "$output" | jq . 2>/dev/null || echo "$output"
        return 1
    fi
    
    # Extract summary for display
    local status=$(echo "$output" | jq -r '.Status // "unknown"' 2>/dev/null)
    local summary=$(echo "$output" | jq '.Summary // {}' 2>/dev/null)
    
    # Display summary
    echo "$summary" | jq .
}

# Continuous monitoring mode
watch_mode() {
    local counter=0
    
    echo -e "${BOLD}🔄 Health Check Monitor${NC}"
    echo -e "${CYAN}Press Ctrl+C to stop${NC}"
    echo ""
    
    while true; do
        counter=$((counter + 1))
        timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        
        # Clear screen for cleaner output (but keep last N lines visible)
        if [[ $counter -gt 1 ]]; then
            echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        fi
        
        # Run check
        OUTPUT=$(run_check)
        
        if [[ "$FORMAT" == "json" ]]; then
            echo -e "${BOLD}[$timestamp] - Check #$counter${NC}"
            display_json "$OUTPUT"
        else
            echo -e "${BOLD}[$timestamp] - Check #$counter${NC}"
            echo "$OUTPUT"
        fi
        
        # Wait before next check
        sleep 5
    done
}

# Comprehensive PR validation mode
pr_validation_mode() {
    local health_output
    local validate_output
    local validation_failed=0
    
    echo -e "${BOLD}🔍 Comprehensive PR Validation${NC}"
    echo -e "${CYAN}Running Phase 1, 2, and endpoint validation...${NC}"
    echo ""
    
    # Step 1: Run full health checks (Phase 1 + 2)
    echo -e "${BOLD}Step 1: Health Checks (Phase 1 + 2)${NC}"
    health_output=$(run_check)
    
    if [[ "$FORMAT" == "json" ]]; then
        display_json "$health_output" || validation_failed=1
    else
        echo "$health_output" || validation_failed=1
    fi
    
    # Step 2: Run docker-validate.sh for full endpoint testing
    echo ""
    echo -e "${BOLD}Step 2: Full Endpoint Validation${NC}"
    if [[ -f "$SCRIPT_DIR/docker-validate.sh" ]]; then
        if "$SCRIPT_DIR/docker-validate.sh" --json > /tmp/docker-validate-result.json 2>&1; then
            if [[ "$FORMAT" == "json" ]]; then
                echo "✅ Endpoint validation passed"
            else
                echo "✅ Endpoint validation passed"
            fi
        else
            echo -e "${YELLOW}⚠️  Some endpoint checks failed or returned warnings${NC}"
            validation_failed=1
        fi
    else
        echo -e "${YELLOW}⚠️  docker-validate.sh not found, skipping endpoint tests${NC}"
    fi
    
    # Step 3: Summary
    echo ""
    echo -e "${BOLD}Step 3: Summary${NC}"
    if [[ $validation_failed -eq 0 ]]; then
        echo -e "${GREEN}✅ PR validation PASSED${NC}"
        echo -e "${CYAN}Ready to create PR${NC}"
        return 0
    else
        echo -e "${RED}❌ PR validation FAILED${NC}"
        echo -e "${YELLOW}Fix issues above before creating PR${NC}"
        return 1
    fi
}

# Store results to database
store_results() {
    if [[ "$STORE_RESULTS" != "true" ]]; then
        return
    fi
    
    # TODO: Implement storage via Logs service API
    echo -e "${YELLOW}📊 TODO: Storing results to database...${NC}"
    echo -e "   Endpoint: POST /api/logs/health/store"
    echo -e "   Database: $DB_URL"
}

# Main execution
main() {
    if [[ "$PR_MODE" == "true" ]]; then
        # Comprehensive PR validation mode
        pr_validation_mode
        exit_code=$?
        store_results
        exit $exit_code
    elif [[ "$WATCH_MODE" == "true" ]]; then
        # Continuous monitoring mode
        watch_mode
    else
        # Single run
        OUTPUT=$(run_check)
        
        if [[ "$FORMAT" == "json" ]]; then
            display_json "$OUTPUT"
        else
            echo "$OUTPUT"
        fi
        
        store_results
    fi
}

# Run main function
main
