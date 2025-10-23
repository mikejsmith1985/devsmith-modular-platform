#!/bin/bash
# Pre-Build Validation v1.0
# Validates project structure and dependencies BEFORE Docker build attempts
# Designed for autonomous Copilot debugging - outputs structured JSON

set -e

# Parse command line arguments
OUTPUT_FORMAT="human"
AUTO_FIX=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --json) OUTPUT_FORMAT="json"; shift ;;
        --fix) AUTO_FIX=true; shift ;;
        *) shift ;;
    esac
done

# Colors (disabled for JSON output)
if [[ "$OUTPUT_FORMAT" == "json" ]]; then
    RED=''; GREEN=''; YELLOW=''; BLUE=''; CYAN=''; BOLD=''; NC=''
else
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    BOLD='\033[1m'
    NC='\033[0m'
fi

# Structured issue tracking
declare -a ISSUES=()
FAILED=0
START_TIME=$(date +%s)

# Check results tracking
declare -A CHECK_RESULTS=(
    [project_structure]="pending"
    [go_modules]="pending"
    [docker_files]="pending"
    [service_files]="pending"
)

# Helper: Strip ANSI and escape for JSON
strip_ansi() {
    echo "$1" | sed 's/\x1b\[[0-9;]*m//g' | tr -d '\000-\037' | sed 's/\\/\\\\/g' | sed 's/"/\\"/g'
}

# Helper: Add issue to structured array
add_issue() {
    local type="$1"
    local severity="$2"
    local service="$3"
    local file="$4"
    local message=$(strip_ansi "$5")
    local suggestion=$(strip_ansi "$6")
    local auto_fixable="${7:-false}"
    local fix_command=$(strip_ansi "${8:-}")

    local issue_json=$(cat <<EOF
{
    "type": "$type",
    "severity": "$severity",
    "service": "$service",
    "file": "$file",
    "message": "$message",
    "suggestion": "$suggestion",
    "autoFixable": $auto_fixable,
    "fixCommand": "$fix_command"
}
EOF
)
    ISSUES+=("$issue_json")

    if [[ "$severity" == "error" ]]; then
        FAILED=1
    fi
}

# Auto-fix: Create missing service directory and main.go
autofix_missing_service() {
    local service="$1"
    local service_dir="cmd/${service}"

    if [[ "$AUTO_FIX" == true ]]; then
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BLUE}Auto-fixing: Creating ${service} service structure...${NC}"

        # Create directory
        mkdir -p "$service_dir"

        # Create basic main.go
        cat > "${service_dir}/main.go" <<'GOEOF'
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	dbURL := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Register handlers
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", rootHandler)

	log.Printf("Starting service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Health check endpoint (REQUIRED for docker-validate)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
			"checks": map[string]bool{
				"database": false,
			},
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"checks": map[string]bool{
			"database": true,
		},
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"service": "SERVICE_NAME",
		"status":  "running",
	})
}
GOEOF

        # Replace SERVICE_NAME placeholder
        sed -i "s/SERVICE_NAME/${service}/g" "${service_dir}/main.go"

        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${GREEN}✓ Created ${service_dir}/main.go${NC}"
        return 0
    fi

    return 1
}

# Validation: Check project structure
check_project_structure() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[1/4] Checking project structure...${NC}"

    local has_issues=false

    # Check essential directories
    local required_dirs=("cmd" "internal" "docker")
    for dir in "${required_dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            add_issue "missing_directory" "error" "" "$dir" \
                "Required directory '${dir}' not found" \
                "Create directory: mkdir -p ${dir}" \
                false "mkdir -p ${dir}"
            has_issues=true
        fi
    done

    # Check docker-compose.yml exists
    if [[ ! -f "docker-compose.yml" ]]; then
        add_issue "missing_file" "error" "" "docker-compose.yml" \
            "docker-compose.yml not found in project root" \
            "Create docker-compose.yml or run from project root" \
            false ""
        has_issues=true
    fi

    # Check go.mod exists
    if [[ ! -f "go.mod" ]]; then
        add_issue "missing_file" "error" "" "go.mod" \
            "go.mod not found - not a Go module" \
            "Initialize Go module: go mod init <module-name>" \
            false "go mod init github.com/yourorg/yourproject"
        has_issues=true
    fi

    if [[ "$has_issues" == false ]]; then
        CHECK_RESULTS[project_structure]="passed"
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}✓${NC} Project structure valid"
    else
        CHECK_RESULTS[project_structure]="failed"
    fi
}

# Validation: Check Go modules
check_go_modules() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[2/4] Checking Go modules...${NC}"

    if [[ ! -f "go.mod" ]]; then
        CHECK_RESULTS[go_modules]="skipped"
        return
    fi

    # Check if go.sum exists
    if [[ ! -f "go.sum" ]]; then
        add_issue "missing_file" "warning" "" "go.sum" \
            "go.sum not found - dependencies not downloaded" \
            "Download dependencies: go mod download" \
            true "go mod download"

        if [[ "$AUTO_FIX" == true ]]; then
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BLUE}Auto-fixing: Running go mod download...${NC}"
            go mod download 2>&1 || true
        fi
    fi

    # Check for common required dependencies
    local required_deps=("github.com/lib/pq")
    for dep in "${required_deps[@]}"; do
        if ! grep -q "$dep" go.mod; then
            add_issue "missing_dependency" "warning" "" "go.mod" \
                "Missing common dependency: ${dep}" \
                "Add dependency: go get ${dep}" \
                false "go get ${dep}"
        fi
    done

    CHECK_RESULTS[go_modules]="passed"
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}✓${NC} Go modules valid"
}

# Validation: Check Docker files
check_docker_files() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[3/4] Checking Docker configuration...${NC}"

    local has_issues=false

    # Get services from docker-compose.yml
    if [[ ! -f "docker-compose.yml" ]]; then
        CHECK_RESULTS[docker_files]="skipped"
        return
    fi

    # Check for required Dockerfiles
    local services=$(docker-compose config --services 2>/dev/null | grep -v '^postgres$' | grep -v '^nginx$' || true)

    while IFS= read -r service; do
        if [[ -z "$service" ]]; then
            continue
        fi

        # Check if Dockerfile exists for this service
        local dockerfile="cmd/${service}/Dockerfile"

        if [[ ! -f "$dockerfile" ]]; then
            add_issue "missing_dockerfile" "error" "$service" "$dockerfile" \
                "Dockerfile missing for service '${service}'" \
                "Create Dockerfile at ${dockerfile}" \
                false ""
            has_issues=true
        else
            # Validate Dockerfile references correct path
            if ! grep -q "cmd/${service}" "$dockerfile"; then
                add_issue "dockerfile_path_mismatch" "warning" "$service" "$dockerfile" \
                    "Dockerfile may not reference correct service path" \
                    "Verify COPY and RUN commands reference ./cmd/${service}" \
                    false ""
            fi
        fi
    done <<< "$services"

    if [[ "$has_issues" == false ]]; then
        CHECK_RESULTS[docker_files]="passed"
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}✓${NC} Docker files valid"
    else
        CHECK_RESULTS[docker_files]="failed"
    fi
}

# Validation: Check service files
check_service_files() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[4/4] Checking service files...${NC}"

    local has_issues=false

    # Get services from docker-compose.yml
    if [[ ! -f "docker-compose.yml" ]]; then
        CHECK_RESULTS[service_files]="skipped"
        return
    fi

    local services=$(docker-compose config --services 2>/dev/null | grep -v '^postgres$' | grep -v '^nginx$' || true)

    while IFS= read -r service; do
        if [[ -z "$service" ]]; then
            continue
        fi

        local service_dir="cmd/${service}"
        local main_file="${service_dir}/main.go"

        # Check if service directory exists
        if [[ ! -d "$service_dir" ]]; then
            add_issue "missing_service_dir" "error" "$service" "$service_dir" \
                "Service directory '${service_dir}' does not exist" \
                "Create service: mkdir -p ${service_dir} && create main.go" \
                true "mkdir -p ${service_dir}"

            if autofix_missing_service "$service"; then
                [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}✓${NC} Auto-created ${service} service"
                continue
            else
                has_issues=true
                continue
            fi
        fi

        # Check if main.go exists
        if [[ ! -f "$main_file" ]]; then
            add_issue "missing_main_go" "error" "$service" "$main_file" \
                "No main.go found in ${service_dir}" \
                "Create main.go with main() function and /health endpoint" \
                true "touch ${main_file}"

            if autofix_missing_service "$service"; then
                [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}✓${NC} Auto-created ${main_file}"
                continue
            else
                has_issues=true
                continue
            fi
        fi

        # Check for Go files
        local go_files=$(find "$service_dir" -name "*.go" 2>/dev/null | wc -l)
        if [[ $go_files -eq 0 ]]; then
            add_issue "no_go_files" "error" "$service" "$service_dir" \
                "No Go files found in ${service_dir} (would cause: 'no Go files' build error)" \
                "Add Go source files to ${service_dir}" \
                false ""
            has_issues=true
            continue
        fi

        # Check if main.go has package main
        if ! grep -q "^package main" "$main_file" 2>/dev/null; then
            add_issue "wrong_package" "error" "$service" "$main_file" \
                "main.go must declare 'package main'" \
                "Change first line to: package main" \
                false ""
            has_issues=true
        fi

        # Check if main.go has main() function
        if ! grep -q "func main()" "$main_file" 2>/dev/null; then
            add_issue "missing_main_func" "error" "$service" "$main_file" \
                "main.go must have main() function" \
                "Add: func main() { ... }" \
                false ""
            has_issues=true
        fi

        # Check if health endpoint is implemented
        if ! grep -q "health" "$main_file" 2>/dev/null; then
            add_issue "missing_health_endpoint" "warning" "$service" "$main_file" \
                "No /health endpoint detected (required for docker-validate)" \
                "Add health handler: http.HandleFunc(\"/health\", healthHandler)" \
                false ""
        fi

        [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$has_issues" == false ]] && echo -e "  ${GREEN}✓${NC} ${service}: valid"

    done <<< "$services"

    if [[ "$has_issues" == false ]]; then
        CHECK_RESULTS[service_files]="passed"
    else
        CHECK_RESULTS[service_files]="failed"
    fi
}

# Priority grouping
group_issues_by_priority() {
    local high_priority=()
    local medium_priority=()
    local low_priority=()

    for issue in "${ISSUES[@]}"; do
        local severity=$(echo "$issue" | jq -r '.severity')
        local auto_fixable=$(echo "$issue" | jq -r '.autoFixable')

        if [[ "$severity" == "error" ]]; then
            high_priority+=("$issue")
        elif [[ "$auto_fixable" == "true" ]]; then
            medium_priority+=("$issue")
        else
            low_priority+=("$issue")
        fi
    done

    echo '{"high":'"$(printf '%s\n' "${high_priority[@]}" | jq -s '.')"',"medium":'"$(printf '%s\n' "${medium_priority[@]}" | jq -s '.')"',"low":'"$(printf '%s\n' "${low_priority[@]}" | jq -s '.')"'}'
}

# Output formatters
output_human() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    local grouped=$(group_issues_by_priority)
    local high_count=$(echo "$grouped" | jq '.high | length')
    local medium_count=$(echo "$grouped" | jq '.medium | length')
    local low_count=$(echo "$grouped" | jq '.low | length')

    echo ""
    echo -e "${BOLD}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}🔍 PRE-BUILD VALIDATION SUMMARY${NC} ${CYAN}(completed in ${duration}s)${NC}"
    echo -e "${BOLD}════════════════════════════════════════════════════════════════${NC}"
    echo ""

    # Show check results
    echo -e "${BOLD}CHECK RESULTS:${NC}"
    for check in project_structure go_modules docker_files service_files; do
        local label="${check//_/ }"
        if [[ "${CHECK_RESULTS[$check]}" == "passed" ]]; then
            echo -e "  ${GREEN}✓${NC} $(printf '%-25s' "$label") passed"
        elif [[ "${CHECK_RESULTS[$check]}" == "failed" ]]; then
            echo -e "  ${RED}✗${NC} $(printf '%-25s' "$label") failed"
        elif [[ "${CHECK_RESULTS[$check]}" == "skipped" ]]; then
            echo -e "  ${YELLOW}○${NC} $(printf '%-25s' "$label") skipped"
        fi
    done
    echo ""

    # Priority-based output
    if [[ $high_count -gt 0 ]]; then
        echo -e "${RED}${BOLD}HIGH PRIORITY (Blocking builds):${NC} $high_count issue(s)"
        echo "$grouped" | jq -r '.high[] | "  • [\(.type)] \(.service) - \(.file)\n    \(.message)\n    → \(.suggestion)"' 2>/dev/null || echo "  (Issues detected)"
        echo ""
    fi

    if [[ $medium_count -gt 0 ]]; then
        echo -e "${YELLOW}${BOLD}MEDIUM PRIORITY (Auto-fixable):${NC} $medium_count issue(s)"
        echo "$grouped" | jq -r '.medium[] | "  • [\(.type)] \(.service) - \(.message)\n    → \(.suggestion)"' 2>/dev/null
        echo ""
    fi

    if [[ $low_count -gt 0 ]]; then
        echo -e "${BLUE}${BOLD}LOW PRIORITY:${NC} $low_count issue(s)"
        echo "$grouped" | jq -r '.low[] | "  • [\(.type)] \(.service) - \(.message)"' 2>/dev/null
        echo ""
    fi

    # Quick fixes
    echo -e "${BLUE}${BOLD}QUICK FIXES:${NC}"
    echo -e "  • Auto-fix issues:     ${CYAN}$0 --fix${NC}"
    echo -e "  • View JSON output:    ${CYAN}$0 --json${NC}"
    echo ""

    echo -e "${BOLD}════════════════════════════════════════════════════════════════${NC}"
    echo ""
}

output_json() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    local grouped=$(group_issues_by_priority)

    cat <<EOF
{
  "status": "$([ $FAILED -eq 0 ] && echo "passed" || echo "failed")",
  "duration": $duration,
  "phase": "pre-build",
  "issues": $(printf '%s\n' "${ISSUES[@]}" | jq -s '.'),
  "grouped": $grouped,
  "checkResults": {
    "projectStructure": "${CHECK_RESULTS[project_structure]}",
    "goModules": "${CHECK_RESULTS[go_modules]}",
    "dockerFiles": "${CHECK_RESULTS[docker_files]}",
    "serviceFiles": "${CHECK_RESULTS[service_files]}"
  },
  "summary": {
    "total": ${#ISSUES[@]},
    "errors": $(printf '%s\n' "${ISSUES[@]}" | jq -s '[.[] | select(.severity=="error")] | length'),
    "warnings": $(printf '%s\n' "${ISSUES[@]}" | jq -s '[.[] | select(.severity=="warning")] | length'),
    "autoFixable": $(printf '%s\n' "${ISSUES[@]}" | jq -s '[.[] | select(.autoFixable==true)] | length')
  },
  "nextSteps": [
    "Run: $0 --fix (to auto-fix issues)",
    "Run: docker-compose up -d --build (after fixing)",
    "Run: ./scripts/docker-validate.sh (to validate runtime)"
  ]
}
EOF
}

# Main validation logic
run_validation() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && {
        echo "🔍 Pre-build validation..."
        echo ""
    }

    check_project_structure
    check_go_modules
    check_docker_files
    check_service_files
}

# Run validation
run_validation

# Ensure FAILED is set if any check failed
for check in project_structure go_modules docker_files service_files; do
    if [[ "${CHECK_RESULTS[$check]}" == "failed" ]]; then
        FAILED=1
        break
    fi
done

# Output results
if [[ "$OUTPUT_FORMAT" == "json" ]]; then
    output_json
else
    output_human
fi

# Always save JSON output to single status file for AI assistant access (overwrites previous)
mkdir -p .validation 2>/dev/null || true
{
    echo "{"
    echo "  \"timestamp\": \"$(date -Iseconds)\","
    echo "  \"phase\": \"pre-build\","
    echo "  \"validation\": $(output_json)"
    echo "}"
} > .validation/status.json 2>/dev/null || true

# Exit with failure if issues found
if [[ $FAILED -eq 1 ]]; then
    [[ "$OUTPUT_FORMAT" != "json" ]] && {
        echo -e "${RED}================================================${NC}"
        echo -e "${RED}✗ Pre-build validation FAILED${NC}"
        echo -e "${RED}================================================${NC}"
        echo -e "${RED}Fix issues above before running docker-compose${NC}"
        echo ""
    }
    exit 1
fi

[[ "$OUTPUT_FORMAT" != "json" ]] && {
    echo -e "${GREEN}================================================${NC}"
    echo -e "${GREEN}✅ Pre-build validation PASSED${NC}"
    echo -e "${GREEN}================================================${NC}"
    echo -e "${GREEN}Safe to run: docker-compose up -d --build${NC}"
}

exit 0
