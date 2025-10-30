#!/usr/bin/env bash
set -euo pipefail

# Enhanced docker validator for DevSmith platform
# - Finds postgres container
# - Prints container health
# - Queries pg_stat_activity counts and max_connections (via docker exec psql)
# - Shows recent Postgres logs for errors
# - Shows top processes connecting to Postgres

COMPOSE_FILE=${COMPOSE_FILE:-docker-compose.yml}
SERVICE_NAME=${1:-postgres}
VERBOSE=${VERBOSE:-0}

echo "== DevSmith enhanced docker validation =="
echo "Project compose file: $COMPOSE_FILE"

# Get container id via docker-compose if available
container_id=""
if command -v docker-compose >/dev/null 2>&1; then
  container_id=$(docker-compose -f "$COMPOSE_FILE" ps -q "$SERVICE_NAME" 2>/dev/null || true)
fi
if [ -z "$container_id" ]; then
  # fallback: try to find a container with postgres image or name
  container_id=$(docker ps --filter "ancestor=postgres:15" --format '{{.ID}}' | head -n1 || true)
fi
if [ -z "$container_id" ]; then
  container_id=$(docker ps --filter "name=postgres" --format '{{.ID}}' | head -n1 || true)
fi

if [ -z "$container_id" ]; then
  echo "ERROR: Could not find postgres container. Is docker-compose up?"
  exit 2
fi

echo "Postgres container: $container_id"

# Inspect health
health_status=$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container_id" 2>/dev/null || true)
echo "Health status: ${health_status:-unknown}"

# Tail last 200 lines of postgres logs and grep for problems
echo
echo "-- Recent Postgres logs (last 200 lines; filtered for FATAL/ERROR/warning) --"
# use docker logs
docker logs --tail 200 "$container_id" 2>&1 | sed -n '1,200p' | egrep --color=always -i "fatal|error|warning|too many clients" || true

# Try to run psql queries inside the container
# Determine psql command and credentials from env if present
PGUSER=devsmith
PGDB=devsmith
PGPORT=5432

echo
echo "-- Querying Postgres stats via docker exec (may fail if DB not ready) --"

set +e
# count connections
docker exec -i "$container_id" psql -U "$PGUSER" -d "$PGDB" -c "SELECT count(*) AS total_connections FROM pg_stat_activity;" 2>/dev/null
rc_conn=$?
# show max_connections
docker exec -i "$container_id" psql -U "$PGUSER" -d "$PGDB" -c "SHOW max_connections;" 2>/dev/null
rc_max=$?
# show connections by user
docker exec -i "$container_id" psql -U "$PGUSER" -d "$PGDB" -c "SELECT usename, state, count(*) FROM pg_stat_activity GROUP BY usename, state ORDER BY 3 DESC;" 2>/dev/null
rc_byuser=$?

set -e

if [ $rc_conn -ne 0 ] || [ $rc_max -ne 0 ]; then
  echo
  echo "WARNING: Could not run pg_stat queries inside Postgres container. DB may be unavailable or credentials differ."
  echo "Try: docker exec -it $container_id psql -U <user> -d <db>"
fi

# Check docker stats for container (quick view)
echo
echo "-- Docker stats (one-shot) --"
docker stats --no-stream --format "table {{.Name}}	{{.CPUPerc}}	{{.MemUsage}}" "$container_id" || true

# Show number of containers and status summary
echo
echo "-- docker-compose ps --"
docker-compose -f "$COMPOSE_FILE" ps || true

# Quick check for other containers with many restarts or unhealthy
echo
echo "-- Other services health summary --"
if command -v docker-compose >/dev/null 2>&1; then
  docker-compose -f "$COMPOSE_FILE" ps --services | while read -r svc; do
    id=$(docker-compose -f "$COMPOSE_FILE" ps -q "$svc" 2>/dev/null || true)
    if [ -n "$id" ]; then
      hs=$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$id" 2>/dev/null || true)
      echo "$svc -> $hs"
    fi
  done
fi

# If we saw 'too many clients' in logs, recommend actions
if docker logs --tail 200 "$container_id" 2>&1 | egrep -qi "too many clients"; then
  echo
  echo "ACTION: Postgres reports 'too many clients'. Recommended fixes:"
  echo "  - Restart the postgres container to clear client connections: docker restart $container_id"
  echo "  - Increase max_connections in Postgres config (postgresql.conf) if you need more concurrent clients"
  echo "  - Audit services opening connections but not closing; reduce pool sizes"
fi

echo
echo "Enhanced validation completed." 
