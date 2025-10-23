#!/bin/bash
# Docker Container & Service Validation v2.0 - Dynamic Endpoint Discovery
# Validates Docker containers are running, healthy, and serving traffic correctly
# Dynamically discovers endpoints from: nginx.conf, docker-compose.yml, Go route registrations

set -e

# Parse command line arguments
MODE="standard"
OUTPUT_FORMAT="human"
AUTO_RESTART=false
WAIT_FOR_HEALTHY=false
MAX_WAIT_TIME=120  # seconds
RETEST_FAILED=false
SHOW_DIFF=false
PROGRESSIVE=false
AUTO_FIX=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --json) OUTPUT_FORMAT="json"; shift ;;
        --quick) MODE="quick"; shift ;;
        --thorough) MODE="thorough"; shift ;;
        --auto-restart) AUTO_RESTART=true; shift ;;
        --wait) WAIT_FOR_HEALTHY=true; shift ;;
        --max-wait) MAX_WAIT_TIME="$2"; shift 2 ;;
        --retest-failed) RETEST_FAILED=true; shift ;;
        --diff) SHOW_DIFF=true; shift ;;
        --progressive) PROGRESSIVE=true; shift ;;
        --auto-fix) AUTO_FIX=true; shift ;;
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

# Project name (from docker-compose)
PROJECT_NAME="devsmith-modular-platform"

# Discovered endpoints will be stored here
declare -A DISCOVERED_SERVICES=()
declare -A DISCOVERED_ENDPOINTS=()
declare -A NGINX_ROUTES=()
declare -a ALL_ENDPOINTS=()

# Structured issue tracking
declare -a ISSUES=()
FAILED=0
START_TIME=$(date +%s)

# Check results tracking
declare -A CHECK_RESULTS=(
    [containers_running]="pending"
    [health_checks]="pending"
    [endpoint_discovery]="pending"
    [http_endpoints]="pending"
    [blank_pages]="pending"
    [port_bindings]="pending"
)

# ============================================================================
# DYNAMIC ENDPOINT DISCOVERY
# ============================================================================

# Parse docker-compose.yml for service ports
discover_services_from_compose() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[1/8] Discovering services from docker-compose.yml...${NC}"

    local compose_file="docker-compose.yml"
    if [[ ! -f "$compose_file" ]]; then
        add_issue "discovery_failed" "error" "system" \
            "docker-compose.yml not found" \
            "Ensure you're running from project root directory" \
            false ""
        return 1
    fi

    # Extract service names and ports using grep and awk
    local current_service=""
    while IFS= read -r line; do
        # Match service definition (e.g., "  portal:")
        if [[ "$line" =~ ^[[:space:]]{2}([a-z_]+):[[:space:]]*$ ]]; then
            current_service="${BASH_REMATCH[1]}"
        fi

        # Match port mapping (e.g., "      - "8080:8080"")
        if [[ -n "$current_service" ]] && [[ "$line" =~ [[:space:]]*-[[:space:]]*\"?([0-9]+):([0-9]+)\"? ]]; then
            local host_port="${BASH_REMATCH[1]}"
            local container_port="${BASH_REMATCH[2]}"
            DISCOVERED_SERVICES[$current_service]="$host_port"
            [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                echo -e "  ${GREEN}✓${NC} Found service: ${current_service} on port ${host_port}"
        fi
    done < "$compose_file"

    CHECK_RESULTS[endpoint_discovery]="passed"
}

# Parse nginx.conf for location blocks
discover_nginx_routes() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[2/8] Discovering routes from nginx.conf...${NC}"

    local nginx_conf="docker/nginx/nginx.conf"
    if [[ ! -f "$nginx_conf" ]]; then
        add_issue "discovery_warning" "warning" "nginx" \
            "nginx.conf not found at docker/nginx/nginx.conf" \
            "Skipping nginx route discovery" \
            false ""
        return 0
    fi

    local in_server_block=false
    while IFS= read -r line; do
        # Detect server block
        if [[ "$line" =~ ^[[:space:]]*server[[:space:]]*\{ ]]; then
            in_server_block=true
        fi

        # Parse location blocks
        if [[ "$in_server_block" == true ]] && [[ "$line" =~ ^[[:space:]]*location[[:space:]]+([^[:space:]]+)[[:space:]]+ ]]; then
            local route="${BASH_REMATCH[1]}"

            # Determine target service from proxy_pass
            local next_line
            read -r next_line
            if [[ "$next_line" =~ proxy_pass[[:space:]]+http://([a-z_]+) ]]; then
                local target_service="${BASH_REMATCH[1]}"
                NGINX_ROUTES[$route]="$target_service"
                [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                    echo -e "  ${GREEN}✓${NC} Found route: ${route} → ${target_service}"
            fi
        fi
    done < "$nginx_conf"
}

# Parse Go files for route registrations
discover_go_routes() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[3/8] Discovering routes from services (runtime)...${NC}"

    # Query each service's /debug/routes endpoint for runtime route discovery
    for service_name in "${!DISCOVERED_SERVICES[@]}"; do
        # Debug: Always show what service we're checking (remove after debugging)
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${CYAN}Checking service: ${service_name}${NC}"
        local service_port="${DISCOVERED_SERVICES[$service_name]}"
        [[ -z "$service_port" ]] && continue

        # Skip nginx (it's the gateway, not an app service)
        [[ "$service_name" == "nginx" ]] && continue

        # Debug: Show which service we're querying
        [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
            echo -e "  ${CYAN}→${NC} Querying ${service_name} debug endpoint..."

        # Query the debug endpoint
        local debug_url="http://localhost:${service_port}/debug/routes"
        local routes_json=$(curl -s -m 2 "$debug_url" 2>/dev/null)

        # Check if debug endpoint is available
        if [[ -z "$routes_json" ]] || [[ "$routes_json" == *"404"* ]] || [[ "$routes_json" == *"Not Found"* ]]; then
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${YELLOW}⚠${NC} Debug endpoint unavailable for ${service_name}, skipping runtime discovery"
            continue
        fi

        # Parse JSON response and extract routes
        local route_count=$(echo "$routes_json" | jq -r '.count // 0' 2>/dev/null)
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${CYAN}Route count for ${service_name}: ${route_count}${NC}"
        if [[ "$route_count" -eq 0 ]]; then
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${YELLOW}⚠${NC} No routes found for ${service_name}"
            continue
        fi

        [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
            echo -e "  ${GREEN}✓${NC} Found ${route_count} routes in ${service_name} (runtime)"

        # Extract each route from JSON
        local routes=$(echo "$routes_json" | jq -r '.routes[] | "\(.method)|\(.path)"' 2>/dev/null)
        while IFS='|' read -r method path; do
            [[ -z "$method" ]] || [[ -z "$path" ]] && continue

            # Skip debug endpoint itself
            [[ "$path" == "/debug/routes" ]] && continue

            local url="http://localhost:${service_port}${path}"

            # IMPORTANT: Only test health checks on direct service ports
            # All other routes should be accessed through the nginx gateway (port 3000)
            # This prevents "running services locally" confusion

            # Only add health check endpoints from services (tested on direct port)
            if [[ "$path" == "/health" ]]; then
                local endpoint_key="${service_name}_${path//\//_}_${method}"
                DISCOVERED_ENDPOINTS[$endpoint_key]="$url|$method|$service_name|runtime"

                [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                    echo -e "  ${GREEN}✓${NC} Found health check: ${service_name}${path}"
            else
                # Check if this route is exposed through nginx
                # Pattern matching for nginx proxy routes
                local nginx_path=""
                local nginx_port="${DISCOVERED_SERVICES[nginx]}"

                # Map service API paths to nginx paths based on nginx.conf patterns
                case "$service_name" in
                    analytics)
                        # nginx: /analytics/ -> /api/analytics/
                        if [[ "$path" =~ ^/api/analytics/(.+)$ ]]; then
                            nginx_path="/analytics/${BASH_REMATCH[1]}"
                        fi
                        ;;
                    review)
                        # nginx: /review/ -> /api/reviews/
                        if [[ "$path" =~ ^/api/review[s]?/(.+)$ ]]; then
                            nginx_path="/review/${BASH_REMATCH[1]}"
                        fi
                        ;;
                    logs)
                        # nginx: /logs/ -> (direct proxy)
                        if [[ "$path" =~ ^/(.+)$ ]] && [[ "$path" != "/" ]]; then
                            nginx_path="/logs/${BASH_REMATCH[1]}"
                        fi
                        ;;
                    portal)
                        # Portal routes are typically accessed directly through nginx root
                        # Already handled by nginx route discovery
                        ;;
                esac

                # If route is exposed through nginx, test it there
                if [[ -n "$nginx_path" ]] && [[ -n "$nginx_port" ]]; then
                    local gateway_url="http://localhost:${nginx_port}${nginx_path}"
                    local endpoint_key="gateway_${service_name}_${nginx_path//\//_}_${method}"
                    DISCOVERED_ENDPOINTS[$endpoint_key]="$gateway_url|$method|nginx|gateway_proxy"

                    [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                        echo -e "  ${GREEN}✓${NC} Will test ${method} ${nginx_path} through gateway (${service_name} route)"
                else
                    # Non-exposed route, just document it
                    [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                        echo -e "  ${CYAN}ℹ${NC} Discovered ${method} ${path} on ${service_name} (internal only)"
                fi
            fi
        done <<< "$routes" || true
    done
}

# Build comprehensive endpoint list
build_endpoint_list() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[4/8] Building comprehensive endpoint test list...${NC}"

    # If retest-failed mode, load previously failed endpoints
    if [[ "$RETEST_FAILED" == true ]] && [[ -f ".validation/status.json" ]]; then
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${CYAN}Loading previously failed endpoints...${NC}"

        # Extract failed endpoint URLs from previous run
        local failed_urls=$(cat .validation/status.json | jq -r '.validation.issues[] | select(.type | startswith("http_")) | .message' | grep -oP 'http://[^ ]+' || echo "")

        if [[ -z "$failed_urls" ]]; then
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}No previously failed endpoints found. Running full validation.${NC}"
            RETEST_FAILED=false
        else
            # Build limited endpoint list from failed URLs
            while IFS= read -r url; do
                [[ -z "$url" ]] && continue
                # Determine service and method from URL
                local service="unknown"
                local method="GET"
                if [[ "$url" =~ :3000 ]]; then
                    service="nginx"
                elif [[ "$url" =~ :8080 ]]; then
                    service="portal"
                elif [[ "$url" =~ :8081 ]]; then
                    service="review"
                elif [[ "$url" =~ :8082 ]]; then
                    service="logs"
                elif [[ "$url" =~ :8083 ]]; then
                    service="analytics"
                fi
                ALL_ENDPOINTS+=("$url|$method|$service|retest")
            done <<< "$failed_urls"

            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${CYAN}Re-testing ${#ALL_ENDPOINTS[@]} previously failed endpoints${NC}"
            return
        fi
    fi

    # Standard discovery if not in retest mode
    # Add all discovered endpoints
    for key in "${!DISCOVERED_ENDPOINTS[@]}"; do
        ALL_ENDPOINTS+=("${DISCOVERED_ENDPOINTS[$key]}")
    done

    # Add nginx routes through gateway
    # Skip directory-only routes (like /review/, /analytics/) that don't have handlers
    local nginx_port="${DISCOVERED_SERVICES[nginx]}"
    if [[ -n "$nginx_port" ]]; then
        for route in "${!NGINX_ROUTES[@]}"; do
            local target_service="${NGINX_ROUTES[$route]}"

            # Only test routes that are likely to have handlers
            # Skip trailing slash routes unless they're root
            if [[ "$route" == "/" ]] || [[ "$route" =~ /health$ ]] || [[ ! "$route" =~ /$ ]]; then
                local url="http://localhost:${nginx_port}${route}"
                ALL_ENDPOINTS+=("$url|GET|nginx|gateway")
            fi
        done
    fi

    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${CYAN}Total endpoints discovered: ${#ALL_ENDPOINTS[@]}${NC}"
}

# ============================================================================
# HELPER FUNCTIONS
# ============================================================================

# Helper: Strip ANSI control characters and escape for JSON
strip_ansi() {
    echo "$1" | sed 's/\x1b\[[0-9;]*m//g' | tr -d '\000-\037' | sed 's/\\/\\\\/g' | sed 's/"/\\"/g'
}

# Helper: Escape string for JSON (more robust)
json_escape() {
    local input="$1"
    # Remove control characters, escape backslashes, then escape quotes, then escape newlines
    echo "$input" | tr -d '\000-\011\013-\037' | sed 's/\\/\\\\/g' | sed 's/"/\\"/g' | sed ':a;N;$!ba;s/\n/\\n/g'
}

# Helper: Extract code context from file (Phase 2)
extract_code_context() {
    local file="$1"
    local line_num="${2:-0}"
    local context_lines="${3:-3}"

    # Return empty if file doesn't exist or line is 0
    if [[ ! -f "$file" ]] || [[ "$line_num" -eq 0 ]]; then
        echo '{"lineNumber":0,"beforeCode":"","afterCode":"","currentLine":""}'
        return
    fi

    # Read file and extract context
    local start_line=$((line_num - context_lines))
    [[ $start_line -lt 1 ]] && start_line=1
    local end_line=$((line_num + context_lines))

    local current_line=$(sed -n "${line_num}p" "$file" 2>/dev/null)
    local before_code=$(sed -n "${start_line},$((line_num - 1))p" "$file" 2>/dev/null | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')
    local after_code=$(sed -n "$((line_num + 1)),${end_line}p" "$file" 2>/dev/null | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')

    # Escape for JSON
    current_line=$(json_escape "$current_line")
    before_code=$(json_escape "$before_code")
    after_code=$(json_escape "$after_code")

    # Build JSON
    cat <<EOF
{"lineNumber":$line_num,"beforeCode":"$before_code","currentLine":"$current_line","afterCode":"$after_code"}
EOF
}

# Helper: Generate test command for issue (Phase 2)
generate_test_command() {
    local type="$1"
    local service="$2"
    local url="$3"

    case "$type" in
        http_404|http_5xx|http_timeout|http_unexpected)
            echo "curl -v $url"
            ;;
        container_*)
            echo "docker-compose ps $service && docker-compose logs $service --tail=20"
            ;;
        health_*)
            echo "docker inspect ${PROJECT_NAME}-${service}-1 | jq '.[0].State.Health'"
            ;;
        *)
            echo "docker-compose logs $service --tail=20"
            ;;
    esac
}

# Helper: Determine fix priority order (Phase 3)
get_fix_priority() {
    local file="$1"
    local type="$2"

    # Priority 1: Gateway/nginx issues (must work before services)
    if [[ "$file" =~ nginx.conf ]]; then
        echo 1
        return
    fi

    # Priority 2: docker-compose issues (infrastructure)
    if [[ "$file" =~ docker-compose.yml ]]; then
        echo 2
        return
    fi

    # Priority 3: Service code issues
    if [[ "$file" =~ cmd/.*/main.go ]]; then
        echo 3
        return
    fi

    # Priority 4: Dockerfile issues
    if [[ "$file" =~ Dockerfile ]]; then
        echo 4
        return
    fi

    # Priority 5: Everything else
    echo 5
}

# Helper: Add issue to structured array (Phase 3 Enhanced)
add_issue() {
    local type="$1"
    local severity="$2"
    local service="$3"
    local message=$(strip_ansi "$4")
    local suggestion=$(strip_ansi "$5")
    local auto_fixable="${6:-false}"
    local fix_command=$(strip_ansi "${7:-}")
    local file="${8:-unknown}"
    local requires_rebuild="${9:-false}"
    local line_number="${10:-0}"
    local url="${11:-}"

    # Determine restart command
    local restart_command=""
    if [[ "$requires_rebuild" == "true" ]]; then
        restart_command="docker-compose up -d --build ${service}"
    else
        restart_command="docker-compose restart ${service}"
    fi

    # Extract code context if file exists and line number provided
    local code_context='{}'
    if [[ -f "$file" ]] && [[ "$line_number" -gt 0 ]]; then
        code_context=$(extract_code_context "$file" "$line_number" 3)
    fi

    # Generate test command
    local test_command=$(generate_test_command "$type" "$service" "$url")

    # Get fix priority
    local priority=$(get_fix_priority "$file" "$type")

    # Determine dependencies
    local depends_on=""
    if [[ "$file" =~ cmd/.*/main.go ]]; then
        depends_on="nginx.conf,docker-compose.yml"
    elif [[ "$file" =~ Dockerfile ]]; then
        depends_on="docker-compose.yml"
    fi

    local issue_json=$(cat <<EOF
{
    "type": "$type",
    "severity": "$severity",
    "service": "$service",
    "file": "$file",
    "lineNumber": $line_number,
    "priority": $priority,
    "dependsOn": "$depends_on",
    "message": "$message",
    "suggestion": "$suggestion",
    "codeContext": $code_context,
    "context": "Docker container issue - services are running in containers, not locally",
    "troubleshooting": "Check Docker container logs and configuration files, do not run services locally",
    "testCommand": "$test_command",
    "verifyCommand": "curl -I ${url:-http://localhost:3000/}",
    "requiresRebuild": $requires_rebuild,
    "fastCommand": "$([ "$requires_rebuild" == "false" ] && echo "$restart_command" || echo "")",
    "slowCommand": "$([ "$requires_rebuild" == "true" ] && echo "$restart_command" || echo "")",
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

# Helper: Get container status
get_container_status() {
    local service="$1"
    local container_name="${PROJECT_NAME}-${service}-1"

    local status=$(docker ps -a --filter "name=${container_name}" --format "{{.Status}}" 2>/dev/null || echo "")

    if [[ -z "$status" ]]; then
        container_name="${PROJECT_NAME}_${service}_1"
        status=$(docker ps -a --filter "name=${container_name}" --format "{{.Status}}" 2>/dev/null || echo "")
    fi

    if [[ -z "$status" ]]; then
        status=$(docker ps -a --filter "name=${service}" --format "{{.Status}}" 2>/dev/null | head -1 || echo "")
    fi

    echo "$status"
}

# Helper: Get container health
get_container_health() {
    local service="$1"
    local container_name="${PROJECT_NAME}-${service}-1"

    local health=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "none")

    if [[ "$health" == "none" ]] || [[ -z "$health" ]]; then
        container_name="${PROJECT_NAME}_${service}_1"
        health=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "none")
    fi

    if [[ "$health" == "none" ]] || [[ -z "$health" ]]; then
        health=$(docker ps --filter "name=${service}" --format "{{.State}}" 2>/dev/null | head -1)
        if [[ "$health" == "running" ]]; then
            health="none"
        fi
    fi

    echo "$health"
}

# Helper: Check HTTP endpoint with detailed response info
check_http_endpoint_detailed() {
    local url="$1"
    local method="${2:-GET}"
    local timeout=5

    # Get status code, response size, and content-type
    local response=$(curl -s -X "$method" -o /tmp/response_body_$$ -w "%{http_code}|%{size_download}|%{content_type}" \
        --max-time "$timeout" "$url" 2>/dev/null || echo "000|0|unknown")

    local http_code=$(echo "$response" | cut -d'|' -f1)
    local size=$(echo "$response" | cut -d'|' -f2)
    local content_type=$(echo "$response" | cut -d'|' -f3)

    # Check if response is blank (HTML but very small)
    local is_blank="false"
    if [[ "$http_code" == "200" ]] && [[ "$content_type" =~ html ]] && [[ "$size" -lt 100 ]]; then
        is_blank="true"
    fi

    rm -f /tmp/response_body_$$ 2>/dev/null

    echo "$http_code|$size|$content_type|$is_blank"
}

# Helper: Wait for service to be healthy
wait_for_healthy() {
    local service="$1"
    local max_wait="$2"
    local waited=0

    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BLUE}Waiting for ${service} to be healthy (max ${max_wait}s)...${NC}"

    while [[ $waited -lt $max_wait ]]; do
        local health=$(get_container_health "$service")

        if [[ "$health" == "healthy" ]]; then
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${GREEN}✓ ${service} is healthy${NC}"
            return 0
        fi

        sleep 2
        waited=$((waited + 2))
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -ne "${CYAN}  Waited ${waited}s... (status: ${health})${NC}\r"
    done

    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "\n${RED}✗ ${service} did not become healthy within ${max_wait}s${NC}"
    return 1
}

# Helper: Restart container
restart_container() {
    local service="$1"
    local container_name="${PROJECT_NAME}-${service}-1"

    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BLUE}Restarting ${service}...${NC}"

    docker restart "$container_name" >/dev/null 2>&1 || {
        container_name="${PROJECT_NAME}_${service}_1"
        docker restart "$container_name" >/dev/null 2>&1
    }

    sleep 5
    return 0
}

# ============================================================================
# VALIDATION CHECKS
# ============================================================================

# Validation: Check containers are running
check_containers_running() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[5/8] Checking container status...${NC}"

    local all_running=true

    for service in "${!DISCOVERED_SERVICES[@]}"; do
        local status=$(get_container_status "$service")

        if [[ -z "$status" ]]; then
            add_issue "container_missing" "error" "$service" \
                "Container not found" \
                "Run: docker-compose up -d ${service}" \
                false "docker-compose up -d ${service}"
            all_running=false

        elif [[ ! "$status" =~ ^Up ]]; then
            add_issue "container_stopped" "error" "$service" \
                "Container is not running (status: ${status})" \
                "Run: docker-compose start ${service}" \
                true "docker-compose start ${service}"
            all_running=false

            if [[ "$AUTO_RESTART" == true ]]; then
                restart_container "$service"
                [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${GREEN}✓ Restarted ${service}${NC}"
            fi
        else
            [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && echo -e "  ${GREEN}✓${NC} ${service}: running"
        fi
    done

    if [[ "$all_running" == true ]]; then
        CHECK_RESULTS[containers_running]="passed"
    else
        CHECK_RESULTS[containers_running]="failed"
    fi
}

# Validation: Check health checks
check_health_status() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[6/8] Checking health checks...${NC}"

    local all_healthy=true

    for service in "${!DISCOVERED_SERVICES[@]}"; do
        # Skip postgres and nginx from health checks (they have their own mechanisms)
        [[ "$service" == "postgres" ]] && continue

        local health=$(get_container_health "$service")

        if [[ "$health" == "healthy" ]]; then
            [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && echo -e "  ${GREEN}✓${NC} ${service}: healthy"

        elif [[ "$health" == "starting" ]]; then
            if [[ "$WAIT_FOR_HEALTHY" == true ]]; then
                if wait_for_healthy "$service" "$MAX_WAIT_TIME"; then
                    continue
                fi
            fi

            add_issue "health_starting" "warning" "$service" \
                "Health check still starting" \
                "Wait for service to initialize or check logs: docker-compose logs ${service}" \
                false ""
            all_healthy=false

        elif [[ "$health" == "unhealthy" ]]; then
            add_issue "health_unhealthy" "error" "$service" \
                "Health check failed - service not responding correctly" \
                "Check logs: docker-compose logs ${service} | Review health endpoint implementation" \
                true "docker-compose restart ${service}"
            all_healthy=false

            if [[ "$AUTO_RESTART" == true ]]; then
                restart_container "$service"
            fi

        elif [[ "$health" == "none" ]]; then
            add_issue "health_missing" "warning" "$service" \
                "No health check configured" \
                "Add HEALTHCHECK to Dockerfile or healthcheck to docker-compose.yml" \
                false ""
            all_healthy=false
        fi
    done

    if [[ "$all_healthy" == true ]]; then
        CHECK_RESULTS[health_checks]="passed"
    else
        CHECK_RESULTS[health_checks]="failed"
    fi
}

# Validation: Check all discovered HTTP endpoints (Phase 3: Progressive Support)
check_all_endpoints() {
    if [[ "$PROGRESSIVE" == true ]]; then
        check_endpoints_progressive
        return
    fi

    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[7/8] Testing all discovered endpoints (${#ALL_ENDPOINTS[@]} total)...${NC}"

    local all_responsive=true
    local tested_count=0
    local passed_count=0

    for endpoint_data in "${ALL_ENDPOINTS[@]}"; do
        IFS='|' read -r url method service source <<< "$endpoint_data"

        tested_count=$((tested_count + 1))

        # Get detailed response info
        local response=$(check_http_endpoint_detailed "$url" "$method")
        IFS='|' read -r http_code size content_type is_blank <<< "$response"

        if [[ "$http_code" == "200" ]]; then
            passed_count=$((passed_count + 1))

            # Check for blank page
            if [[ "$is_blank" == "true" ]]; then
                add_issue "blank_page" "warning" "$service" \
                    "Endpoint $url returns 200 but appears blank (${size} bytes)" \
                    "Check if this endpoint should return content. May be intentional for some routes." \
                    false ""
                all_responsive=false
            fi

            [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                echo -e "  ${GREEN}✓${NC} ${method} ${url} → 200 OK (${size}B, ${content_type})"

        elif [[ "$http_code" == "401" ]] || [[ "$http_code" == "403" ]]; then
            # Authentication required - this is expected for protected routes
            passed_count=$((passed_count + 1))
            [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                echo -e "  ${CYAN}✓${NC} ${method} ${url} → ${http_code} (protected)"

        elif [[ "$http_code" == "404" ]]; then
            # Provide Docker-specific troubleshooting
            local docker_suggestion="DOCKER ISSUE: Check if ${service} container has the route configured. "
            local affected_file=""
            local requires_rebuild=false
            local line_num=0

            if [[ "$source" == "static_file" ]]; then
                docker_suggestion+="Verify static files were copied to container in Dockerfile. Check: docker exec <container> ls -la /path/to/static"
                affected_file="cmd/${service}/Dockerfile"
                requires_rebuild=true
                # Find COPY line in Dockerfile
                line_num=$(grep -n "COPY.*static" "$affected_file" 2>/dev/null | head -1 | cut -d: -f1 || echo "0")
            elif [[ "$source" == "gateway" ]]; then
                docker_suggestion+="Verify nginx.conf location block is correct and nginx container restarted after changes."
                affected_file="docker/nginx/nginx.conf"
                requires_rebuild=false  # nginx config just needs restart
                # Find location block for this route
                local route_path=$(echo "$url" | sed 's|http://[^/]*/||' | sed 's|/.*|/|')
                line_num=$(grep -n "location.*${route_path}" "$affected_file" 2>/dev/null | head -1 | cut -d: -f1 || echo "0")
            else
                docker_suggestion+="Verify route is registered in cmd/${service}/main.go and container was rebuilt after changes."
                affected_file="cmd/${service}/main.go"
                requires_rebuild=true  # code changes need rebuild
                # Find router registration for this path
                local route_path=$(echo "$url" | sed 's|http://[^/]*/||' | cut -d'?' -f1)
                line_num=$(grep -n "router.*${route_path}" "$affected_file" 2>/dev/null | head -1 | cut -d: -f1 || echo "0")
            fi

            add_issue "http_404" "error" "$service" \
                "Endpoint ${method} ${url} returned 404 Not Found" \
                "$docker_suggestion" \
                false "" "$affected_file" "$requires_rebuild" "$line_num" "$url"
            all_responsive=false

        elif [[ "$http_code" == "500" ]] || [[ "$http_code" == "502" ]] || [[ "$http_code" == "503" ]]; then
            local affected_file="cmd/${service}/main.go"
            local line_num=$(grep -n "func main" "$affected_file" 2>/dev/null | head -1 | cut -d: -f1 || echo "0")
            add_issue "http_5xx" "error" "$service" \
                "Endpoint ${method} ${url} returned ${http_code} (server error)" \
                "Check application logs: docker-compose logs ${service}. Verify dependencies and configuration." \
                false "" "$affected_file" true "$line_num" "$url"
            all_responsive=false

        elif [[ "$http_code" == "000" ]]; then
            local affected_file="docker-compose.yml"
            local line_num=$(grep -n "^\s*${service}:" "$affected_file" 2>/dev/null | head -1 | cut -d: -f1 || echo "0")
            add_issue "http_timeout" "error" "$service" \
                "Endpoint ${method} ${url} timed out or refused connection" \
                "Verify service is running and port binding is correct." \
                false "" "$affected_file" false "$line_num" "$url"
            all_responsive=false

        else
            # Skip 400 errors for routes that need parameters or request body
            if [[ "$http_code" == "400" ]] && ( [[ "$url" =~ :id|:slug|\{.*\} ]] || [[ "$method" == "POST" ]] || [[ "$method" == "PUT" ]] || [[ "$method" == "PATCH" ]] ); then
                # This is expected - route needs parameters or request body
                passed_count=$((passed_count + 1))
                [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                    echo -e "  ${CYAN}✓${NC} ${method} ${url} → 400 (expected - needs data)"
            else
                local affected_file="cmd/${service}/main.go"
                local line_num=$(grep -n "func main" "$affected_file" 2>/dev/null | head -1 | cut -d: -f1 || echo "0")
                add_issue "http_unexpected" "warning" "$service" \
                    "Endpoint ${method} ${url} returned unexpected code: ${http_code}" \
                    "DOCKER ISSUE: Check ${service} container logs with: docker-compose logs ${service}" \
                    false "" "$affected_file" true "$line_num" "$url"
                all_responsive=false
            fi
        fi
    done

    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${CYAN}Results: ${passed_count}/${tested_count} endpoints passed${NC}"

    if [[ "$all_responsive" == true ]]; then
        CHECK_RESULTS[http_endpoints]="passed"
    else
        CHECK_RESULTS[http_endpoints]="failed"
    fi
}

# Helper: Progressive endpoint testing (Phase 3)
check_endpoints_progressive() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[7/8] Progressive validation (layer-by-layer)...${NC}"

    # Layer 1: Gateway health
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${BOLD}Layer 1: Gateway (nginx)${NC}"
    local nginx_port="${DISCOVERED_SERVICES[nginx]}"
    if [[ -n "$nginx_port" ]]; then
        local response=$(check_http_endpoint_detailed "http://localhost:${nginx_port}/" "GET")
        local http_code=$(echo "$response" | cut -d'|' -f1)

        if [[ "$http_code" == "502" ]] || [[ "$http_code" == "000" ]]; then
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${RED}✗${NC} Gateway not responding - stopping validation"
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${YELLOW}Fix gateway before testing services${NC}"
            add_issue "gateway_down" "error" "nginx" \
                "Gateway not responding (502/timeout)" \
                "Fix nginx configuration and ensure portal service is running. Gateway must work before services." \
                false "" "docker/nginx/nginx.conf" false 0 "http://localhost:${nginx_port}/"
            CHECK_RESULTS[http_endpoints]="failed"
            return 1
        fi
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}✓${NC} Gateway responding"
    fi

    # Layer 2: Service health checks
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${BOLD}Layer 2: Service Health${NC}"
    local healthy_services=0
    local total_services=0
    for service in portal review logs analytics; do
        total_services=$((total_services + 1))
        local port="${DISCOVERED_SERVICES[$service]}"
        [[ -z "$port" ]] && continue

        local response=$(check_http_endpoint_detailed "http://localhost:${port}/health" "GET")
        local http_code=$(echo "$response" | cut -d'|' -f1)

        if [[ "$http_code" == "200" ]]; then
            healthy_services=$((healthy_services + 1))
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}✓${NC} ${service} healthy"
        else
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${RED}✗${NC} ${service} unhealthy (${http_code})"
        fi
    done

    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${CYAN}Services healthy: ${healthy_services}/${total_services}${NC}"

    # Layer 3: Full endpoint testing (reuse existing logic)
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${BOLD}Layer 3: Full Endpoints${NC}"

    # Continue with normal endpoint testing
    local all_responsive=true
    local tested_count=0
    local passed_count=0

    for endpoint_data in "${ALL_ENDPOINTS[@]}"; do
        IFS='|' read -r url method service source <<< "$endpoint_data"
        tested_count=$((tested_count + 1))

        local response=$(check_http_endpoint_detailed "$url" "$method")
        IFS='|' read -r http_code size content_type is_blank <<< "$response"

        if [[ "$http_code" == "200" ]] || [[ "$http_code" == "401" ]] || [[ "$http_code" == "403" ]]; then
            passed_count=$((passed_count + 1))
        fi
    done

    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${CYAN}Results: ${passed_count}/${tested_count} endpoints passed${NC}"

    if [[ "$all_responsive" == true ]]; then
        CHECK_RESULTS[http_endpoints]="passed"
    else
        CHECK_RESULTS[http_endpoints]="failed"
    fi
}

# Validation: Check port bindings
check_port_bindings() {
    [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${BOLD}[8/8] Checking port bindings...${NC}"

    local all_bound=true

    for service in "${!DISCOVERED_SERVICES[@]}"; do
        local port="${DISCOVERED_SERVICES[$service]}"
        local container_name="${PROJECT_NAME}-${service}-1"

        local published=$(docker port "$container_name" 2>/dev/null | grep -w "$port" || {
            container_name="${PROJECT_NAME}_${service}_1"
            docker port "$container_name" 2>/dev/null | grep -w "$port" || echo ""
        })

        if [[ -z "$published" ]]; then
            local status=$(get_container_status "$service")
            if [[ -n "$status" ]]; then
                add_issue "port_not_bound" "error" "$service" \
                    "Port ${port} is not bound to host" \
                    "Check docker-compose.yml ports section for ${service}." \
                    false ""
                all_bound=false
            fi
        else
            [[ "$OUTPUT_FORMAT" != "json" ]] && [[ "$MODE" == "thorough" ]] && \
                echo -e "  ${GREEN}✓${NC} ${service}: port ${port} → ${published}"
        fi
    done

    if [[ "$all_bound" == true ]]; then
        CHECK_RESULTS[port_bindings]="passed"
    else
        CHECK_RESULTS[port_bindings]="failed"
    fi
}

# ============================================================================
# OUTPUT FORMATTERS
# ============================================================================

# Priority grouping
group_issues_by_priority() {
    local high_priority=()
    local medium_priority=()
    local low_priority=()

    for issue in "${ISSUES[@]}"; do
        local severity=$(echo "$issue" | jq -r '.severity')
        local type=$(echo "$issue" | jq -r '.type')

        if [[ "$severity" == "error" ]]; then
            high_priority+=("$issue")
        elif [[ "$type" =~ ^(health_starting|blank_page)$ ]]; then
            medium_priority+=("$issue")
        else
            low_priority+=("$issue")
        fi
    done

    echo '{"high":'"$(printf '%s\n' "${high_priority[@]}" | jq -s '.')"',"medium":'"$(printf '%s\n' "${medium_priority[@]}" | jq -s '.')"',"low":'"$(printf '%s\n' "${low_priority[@]}" | jq -s '.')"'}'
}

# Output human-readable format
output_human() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    local grouped=$(group_issues_by_priority)
    local high_count=$(echo "$grouped" | jq '.high | length')
    local medium_count=$(echo "$grouped" | jq '.medium | length')
    local low_count=$(echo "$grouped" | jq '.low | length')

    echo ""
    echo -e "${BOLD}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}🐳 DOCKER VALIDATION SUMMARY${NC} ${CYAN}(completed in ${duration}s)${NC}"
    echo -e "${BOLD}════════════════════════════════════════════════════════════════${NC}"
    echo ""

    # Discovery stats
    echo -e "${BOLD}DISCOVERY:${NC}"
    echo -e "  Services found:    ${CYAN}${#DISCOVERED_SERVICES[@]}${NC}"
    echo -e "  Endpoints tested:  ${CYAN}${#ALL_ENDPOINTS[@]}${NC}"
    echo ""

    # Show check results
    echo -e "${BOLD}CHECK RESULTS:${NC}"
    for check in containers_running health_checks endpoint_discovery http_endpoints port_bindings; do
        local label="${check//_/ }"
        if [[ "${CHECK_RESULTS[$check]}" == "passed" ]]; then
            echo -e "  ${GREEN}✓${NC} $(printf '%-25s' "$label") passed"
        elif [[ "${CHECK_RESULTS[$check]}" == "failed" ]]; then
            echo -e "  ${RED}✗${NC} $(printf '%-25s' "$label") failed"
        fi
    done
    echo ""

    # Priority-based output
    if [[ $high_count -gt 0 ]]; then
        echo -e "${RED}${BOLD}HIGH PRIORITY (Blocking):${NC} $high_count issue(s)"
        echo "$grouped" | jq -r '.high[] | "  • [\(.type)] \(.service) - \(.message)\n    → \(.suggestion)"' 2>/dev/null || echo "  (Issues detected but formatting failed)"
        echo ""
    fi

    if [[ $medium_count -gt 0 ]]; then
        echo -e "${YELLOW}${BOLD}MEDIUM PRIORITY (Should fix):${NC} $medium_count issue(s)"
        echo "$grouped" | jq -r '.medium[] | "  • [\(.type)] \(.service) - \(.message)\n    → \(.suggestion)"' 2>/dev/null || echo "  (Issues detected but formatting failed)"
        echo ""
    fi

    if [[ $low_count -gt 0 ]]; then
        echo -e "${BLUE}${BOLD}LOW PRIORITY (Can defer):${NC} $low_count issue(s)"
        echo "$grouped" | jq -r '.low[] | "  • [\(.type)] \(.service) - \(.message)"' 2>/dev/null || echo "  (Issues detected but formatting failed)"
        echo ""
    fi

    # Quick fixes
    echo -e "${BLUE}${BOLD}NEXT STEPS (Docker Troubleshooting):${NC}"
    echo -e "  • View details:  ${CYAN}cat .validation/status.json | jq '.validation.issues[]'${NC}"
    echo -e "  • Fix order:     ${CYAN}cat .validation/status.json | jq '.validation.issuesByFixOrder'${NC}"
    echo -e "  • Show progress: ${CYAN}./scripts/docker-validate.sh --diff${NC}"
    echo -e "  • Check logs:    ${CYAN}docker-compose logs [service]${NC}"
    echo -e "  • Restart:       ${CYAN}docker-compose restart [service]${NC}"
    echo -e "  • Rebuild:       ${CYAN}docker-compose up -d --build [service]${NC}"
    echo -e ""

    # Show diff if requested
    if [[ "$SHOW_DIFF" == true ]] && [[ -f ".validation/status.json" ]]; then
        local prev_total=$(cat .validation/status.json | jq '.validation.summary.total' 2>/dev/null || echo "0")
        local current_total=${#ISSUES[@]}
        local fixed=$((prev_total - current_total))

        if [[ $fixed -gt 0 ]]; then
            echo -e "${GREEN}${BOLD}📊 PROGRESS:${NC}"
            echo -e "  ✅ Fixed: ${GREEN}${fixed}${NC} issue(s)"
            echo -e "  ⏳ Remaining: ${YELLOW}${current_total}${NC} issue(s)"
            echo -e "  📈 Progress: ${GREEN}$((100 - (current_total * 100 / prev_total)))%${NC}"
            echo -e ""
        elif [[ $fixed -lt 0 ]]; then
            local new_issues=$((-fixed))
            echo -e "${YELLOW}${BOLD}⚠️  NEW ISSUES:${NC}"
            echo -e "  🆕 New: ${YELLOW}${new_issues}${NC} issue(s) appeared"
            echo -e "  ⏳ Total: ${YELLOW}${current_total}${NC} issue(s)"
            echo -e ""
        else
            echo -e "${CYAN}${BOLD}📊 STATUS:${NC}"
            echo -e "  Same as previous run: ${current_total} issue(s)"
            echo -e ""
        fi
    fi

    echo -e "${YELLOW}${BOLD}⚠️  IMPORTANT FOR COPILOT:${NC}"
    echo -e "  Services are running in Docker containers, not locally."
    echo -e "  Do NOT run 'go run' or bind to ports locally."
    echo -e "  Fix Docker configs and rebuild containers."
    echo -e "  Fix in priority order: nginx → docker-compose → services → Dockerfiles"
    echo ""

    echo -e "${BLUE}${BOLD}DOCUMENTATION:${NC}"
    echo -e "  • Validation guide: ${CYAN}.docs/DOCKER-VALIDATION.md${NC}"
    echo -e "${BOLD}════════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Group issues by file for faster Copilot workflow
group_issues_by_file() {
    declare -A file_groups
    local all_files=()

    # Group issues by file
    for issue in "${ISSUES[@]}"; do
        local file=$(echo "$issue" | jq -r '.file')
        if [[ -z "${file_groups[$file]}" ]]; then
            file_groups[$file]="$issue"
            all_files+=("$file")
        else
            file_groups[$file]="${file_groups[$file]},${issue}"
        fi
    done

    # Build JSON output
    local json="{"
    local first=true
    for file in "${all_files[@]}"; do
        [[ "$first" == false ]] && json+=","
        first=false

        local issues="${file_groups[$file]}"
        local rebuild_needed=$(echo "[$issues]" | jq '[.[] | select(.requiresRebuild==true)] | length > 0')
        local restart_cmd=""
        local rebuild_cmd=""

        # Determine service from first issue
        local service=$(echo "$issues" | jq -s '.[0].service' | tr -d '"')

        if [[ "$rebuild_needed" == "true" ]]; then
            rebuild_cmd="docker-compose up -d --build ${service}"
        else
            restart_cmd="docker-compose restart ${service}"
        fi

        json+="\"$file\":{\"issues\":[$issues],\"requiresRebuild\":$rebuild_needed,\"restartCommand\":\"$restart_cmd\",\"rebuildCommand\":\"$rebuild_cmd\"}"
    done
    json+="}"

    echo "$json"
}

# Helper: Calculate diff from previous run (Phase 3)
calculate_diff() {
    local previous_file=".validation/status.json"

    if [[ ! -f "$previous_file" ]]; then
        echo '{"isFirstRun":true,"fixed":0,"new":0,"remaining":0,"progress":"N/A"}'
        return
    fi

    # Get previous issues
    local prev_total=$(cat "$previous_file" | jq '.validation.summary.total' 2>/dev/null || echo "0")
    local current_total=${#ISSUES[@]}

    # Calculate changes
    local fixed=$((prev_total - current_total))
    [[ $fixed -lt 0 ]] && fixed=0
    local new_issues=$((current_total - prev_total))
    [[ $new_issues -lt 0 ]] && new_issues=0

    # Progress percentage
    local progress="N/A"
    if [[ $prev_total -gt 0 ]]; then
        progress=$((100 - (current_total * 100 / prev_total)))
    fi

    cat <<EOF
{
    "isFirstRun": false,
    "previousTotal": $prev_total,
    "currentTotal": $current_total,
    "fixed": $fixed,
    "new": $new_issues,
    "remaining": $current_total,
    "progress": "$progress%"
}
EOF
}

# Helper: Group issues by fix order (Phase 3)
group_issues_by_fix_order() {
    local priority1=()
    local priority2=()
    local priority3=()
    local priority4=()
    local priority5=()

    for issue in "${ISSUES[@]}"; do
        local priority=$(echo "$issue" | jq -r '.priority')
        case "$priority" in
            1) priority1+=("$issue") ;;
            2) priority2+=("$issue") ;;
            3) priority3+=("$issue") ;;
            4) priority4+=("$issue") ;;
            *) priority5+=("$issue") ;;
        esac
    done

    cat <<EOF
{
    "fixOrder": [
        {
            "priority": 1,
            "name": "Gateway/Nginx Issues",
            "reason": "Must fix gateway routing before testing services",
            "count": ${#priority1[@]},
            "issues": $(printf '%s\n' "${priority1[@]}" | jq -s '.' 2>/dev/null || echo '[]')
        },
        {
            "priority": 2,
            "name": "Infrastructure Issues",
            "reason": "Fix docker-compose configuration before services",
            "count": ${#priority2[@]},
            "issues": $(printf '%s\n' "${priority2[@]}" | jq -s '.' 2>/dev/null || echo '[]')
        },
        {
            "priority": 3,
            "name": "Service Code Issues",
            "reason": "Fix service implementations after infrastructure",
            "count": ${#priority3[@]},
            "issues": $(printf '%s\n' "${priority3[@]}" | jq -s '.' 2>/dev/null || echo '[]')
        },
        {
            "priority": 4,
            "name": "Build Issues",
            "reason": "Fix Dockerfiles last",
            "count": ${#priority4[@]},
            "issues": $(printf '%s\n' "${priority4[@]}" | jq -s '.' 2>/dev/null || echo '[]')
        }
    ]
}
EOF
}

# Output JSON format (Phase 3 Enhanced)
output_json() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    local grouped=$(group_issues_by_priority)
    local by_file=$(group_issues_by_file)
    local by_order=$(group_issues_by_fix_order)
    local diff=$(calculate_diff)

    # Build discovered services JSON
    local services_json="{"
    local first=true
    for service in "${!DISCOVERED_SERVICES[@]}"; do
        [[ "$first" == false ]] && services_json+=","
        services_json+="\"$service\":${DISCOVERED_SERVICES[$service]}"
        first=false
    done
    services_json+="}"

    # Build discovered endpoints JSON
    local endpoints_json="["
    first=true
    for endpoint_data in "${ALL_ENDPOINTS[@]}"; do
        IFS='|' read -r url method service source <<< "$endpoint_data"
        [[ "$first" == false ]] && endpoints_json+=","
        endpoints_json+="{\"url\":\"$url\",\"method\":\"$method\",\"service\":\"$service\",\"source\":\"$source\"}"
        first=false
    done
    endpoints_json+="]"

    cat <<EOF
{
  "status": "$([ $FAILED -eq 0 ] && echo "passed" || echo "failed")",
  "duration": $duration,
  "mode": "$MODE",
  "retestMode": $RETEST_FAILED,
  "progressiveMode": $PROGRESSIVE,
  "diffMode": $SHOW_DIFF,
  "discovery": {
    "servicesFound": ${#DISCOVERED_SERVICES[@]},
    "endpointsDiscovered": ${#ALL_ENDPOINTS[@]},
    "services": $services_json,
    "endpoints": $endpoints_json
  },
  "issues": $(printf '%s\n' "${ISSUES[@]}" | jq -s '.'),
  "issuesByFile": $by_file,
  "issuesByFixOrder": $by_order,
  "grouped": $grouped,
  "diff": $diff,
  "checkResults": {
    "containersRunning": "${CHECK_RESULTS[containers_running]}",
    "healthChecks": "${CHECK_RESULTS[health_checks]}",
    "endpointDiscovery": "${CHECK_RESULTS[endpoint_discovery]}",
    "httpEndpoints": "${CHECK_RESULTS[http_endpoints]}",
    "portBindings": "${CHECK_RESULTS[port_bindings]}"
  },
  "summary": {
    "total": ${#ISSUES[@]},
    "errors": $(printf '%s\n' "${ISSUES[@]}" | jq -s '[.[] | select(.severity=="error")] | length'),
    "warnings": $(printf '%s\n' "${ISSUES[@]}" | jq -s '[.[] | select(.severity=="warning")] | length'),
    "autoFixable": $(printf '%s\n' "${ISSUES[@]}" | jq -s '[.[] | select(.autoFixable==true)] | length'),
    "requiresRebuild": $(printf '%s\n' "${ISSUES[@]}" | jq -s '[.[] | select(.requiresRebuild==true)] | length'),
    "requiresRestartOnly": $(printf '%s\n' "${ISSUES[@]}" | jq -s '[.[] | select(.requiresRebuild==false)] | length')
  }
}
EOF
}

# ============================================================================
# AUTO-FIX FUNCTIONALITY
# ============================================================================

# Run auto-fixable commands
run_auto_fixes() {
    # Only run if --auto-fix flag is set
    [[ "$AUTO_FIX" != "true" ]] && return

    # Load status.json to get auto-fixable issues
    if [[ ! -f ".validation/status.json" ]]; then
        [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "${YELLOW}⚠${NC} No status.json file found, skipping auto-fix"
        return
    fi

    # Get unique auto-fixable commands
    local fix_commands=$(cat .validation/status.json | jq -r '.validation.issues[] | select(.autoFixable == true) | .fixCommand' 2>/dev/null | sort -u)

    if [[ -z "$fix_commands" ]]; then
        [[ "$OUTPUT_FORMAT" != "json" ]] && {
            echo ""
            echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
            echo -e "${GREEN}✓ No auto-fixable issues found${NC}"
            echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
        }
        return
    fi

    # Count commands
    local count=$(echo "$fix_commands" | wc -l)

    [[ "$OUTPUT_FORMAT" != "json" ]] && {
        echo ""
        echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
        echo -e "${BOLD}🔧 AUTO-FIX MODE${NC}"
        echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
        echo ""
        echo -e "${BOLD}Found ${count} auto-fixable command(s):${NC}"
        echo "$fix_commands" | while read -r cmd; do
            echo -e "  ${CYAN}→${NC} $cmd"
        done
        echo ""
    }

    # Execute each command
    local success_count=0
    local fail_count=0

    while read -r cmd; do
        [[ -z "$cmd" ]] && continue

        [[ "$OUTPUT_FORMAT" != "json" ]] && {
            echo -e "${BOLD}Executing:${NC} $cmd"
        }

        # Run the command
        if eval "$cmd" > /tmp/autofix-output.log 2>&1; then
            ((success_count++))
            [[ "$OUTPUT_FORMAT" != "json" ]] && echo -e "  ${GREEN}✓${NC} Success"
        else
            ((fail_count++))
            [[ "$OUTPUT_FORMAT" != "json" ]] && {
                echo -e "  ${RED}✗${NC} Failed"
                echo -e "  ${YELLOW}Output:${NC}"
                cat /tmp/autofix-output.log | sed 's/^/    /'
            }
        fi
        echo ""
    done <<< "$fix_commands"

    [[ "$OUTPUT_FORMAT" != "json" ]] && {
        echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
        echo -e "${BOLD}AUTO-FIX SUMMARY${NC}"
        echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
        echo ""
        echo -e "  ${GREEN}✓${NC} Successful: $success_count"
        echo -e "  ${RED}✗${NC} Failed: $fail_count"
        echo ""

        if [[ $success_count -gt 0 ]]; then
            echo -e "${YELLOW}Recommendation:${NC} Re-run validation to check if issues are resolved:"
            echo -e "  ${CYAN}./scripts/docker-validate.sh --retest-failed${NC}"
        fi
        echo ""
        echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
    }

    # Clean up
    rm -f /tmp/autofix-output.log
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

run_validation() {
    # Check if docker-compose is running
    local running_containers=$(docker-compose ps --services 2>/dev/null | wc -l)

    if [[ "$running_containers" -eq 0 ]]; then
        [[ "$OUTPUT_FORMAT" != "json" ]] && {
            echo -e "${RED}✗ No containers running${NC}"
            echo -e "Start services with: ${CYAN}docker-compose up -d${NC}"
        }
        exit 1
    fi

    [[ "$OUTPUT_FORMAT" != "json" ]] && {
        echo "🐳 Docker validation with dynamic endpoint discovery ($MODE mode)..."
        echo ""
        echo "📦 Project: $PROJECT_NAME"
        echo ""
    }

    # Discovery phase
    discover_services_from_compose
    discover_nginx_routes
    discover_go_routes
    build_endpoint_list

    # Validation phase
    if [[ "$MODE" == "quick" ]]; then
        check_containers_running
        check_health_status

    elif [[ "$MODE" == "thorough" ]]; then
        check_containers_running
        check_health_status
        check_all_endpoints
        check_port_bindings

    else  # standard
        check_containers_running
        check_health_status
        check_all_endpoints
    fi
}

# Run validation
run_validation

# Ensure FAILED is set if any check failed
for check in containers_running health_checks endpoint_discovery http_endpoints port_bindings; do
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

# Always save JSON output to single status file for Copilot access (overwrites previous)
mkdir -p .validation 2>/dev/null || true
{
    echo "{"
    echo "  \"timestamp\": \"$(date -Iseconds)\","
    echo "  \"phase\": \"runtime\","
    echo "  \"validation\": $(output_json)"
    echo "}"
} > .validation/status.json 2>/dev/null || true

# Run auto-fixes if --auto-fix flag is set
run_auto_fixes

# Exit with failure if issues found
if [[ $FAILED -eq 1 ]]; then
    [[ "$OUTPUT_FORMAT" != "json" ]] && {
        echo -e "${RED}================================================${NC}"
        echo -e "${RED}✗ Docker validation FAILED${NC}"
        echo -e "${RED}================================================${NC}"
    }
    exit 1
fi

[[ "$OUTPUT_FORMAT" != "json" ]] && {
    echo -e "${GREEN}================================================${NC}"
    echo -e "${GREEN}✅ Docker validation PASSED${NC}"
    echo -e "${GREEN}================================================${NC}"
}

exit 0
