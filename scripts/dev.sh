#!/bin/bash
# DevSmith Modular Platform Development Script
# NOTE: NOT using 'set -e' to allow graceful error handling for validation failures

echo "[DevSmith] Starting development environment..."
echo ""

# STEP 1: Pre-build validation
echo "[DevSmith] Step 1/3: Pre-build validation..."
./scripts/pre-build-validate.sh
PREBUILD_RESULT=$?

if [ $PREBUILD_RESULT -ne 0 ]; then
    echo ""
    echo "[DevSmith] ❌ Pre-build validation failed!"
    echo ""
    echo "To auto-fix issues: ./scripts/pre-build-validate.sh --fix"
    echo "To see JSON output:  ./scripts/pre-build-validate.sh --json"
    echo ""
    echo "For autonomous debugging, Copilot can run:"
    echo "  1. ./scripts/pre-build-validate.sh --json | jq '.issues[]'"
    echo "  2. ./scripts/pre-build-validate.sh --fix"
    echo "  3. ./scripts/dev.sh"
    echo ""
    exit 1
fi

echo ""
echo "[DevSmith] Step 2/3: Building and starting services..."

# Start services with docker-compose
docker-compose up -d --build
BUILD_RESULT=$?

if [ $BUILD_RESULT -ne 0 ]; then
    echo ""
    echo "[DevSmith] ❌ Docker build failed!"
    echo ""
    echo "Check logs above for build errors."
    echo "This should have been caught by pre-build validation."
    echo ""
    exit 1
fi

echo ""
echo "[DevSmith] Step 3/3: Validating runtime health..."
echo ""

# Wait for all services to be healthy (max 120 seconds)
./scripts/docker-validate.sh --wait --max-wait 120
VALIDATION_RESULT=$?

# If validation failed, try auto-recovery
if [ $VALIDATION_RESULT -ne 0 ]; then
    echo ""
    echo "[DevSmith] Attempting automatic recovery..."
    echo ""

    # Check if it's the common nginx 502 issue
    NGINX_502=$(./scripts/docker-validate.sh --json 2>/dev/null | jq -r '.issues[] | select(.service=="nginx" and .type=="http_5xx") | .service' 2>/dev/null || echo "")

    if [ "$NGINX_502" == "nginx" ]; then
        echo "[DevSmith] Detected nginx 502 - container IPs likely changed after rebuild"
        echo "[DevSmith] Restarting nginx to pick up new IPs..."
        docker-compose restart nginx
        sleep 5

        echo "[DevSmith] Re-validating after nginx restart..."
        ./scripts/docker-validate.sh
        VALIDATION_RESULT=$?
    else
        # Try generic auto-restart for unhealthy services
        echo "[DevSmith] Attempting auto-restart of unhealthy services..."
        ./scripts/docker-validate.sh --auto-restart
        sleep 5

        echo "[DevSmith] Re-validating after auto-restart..."
        ./scripts/docker-validate.sh
        VALIDATION_RESULT=$?
    fi
fi

# Handle final validation result
if [ $VALIDATION_RESULT -eq 0 ]; then
    echo ""
    echo "[DevSmith] All services are healthy! 🎉"
    echo ""
    echo "Services available at:"
    echo "  • Portal:    http://localhost:8080"
    echo "  • Review:    http://localhost:8081"
    echo "  • Logs:      http://localhost:8082"
    echo "  • Analytics: http://localhost:8083"
    echo "  • Gateway:   http://localhost:3000"
    echo ""
    echo "To view logs: docker-compose logs -f [service]"
    echo "To validate:  ./scripts/docker-validate.sh"
    echo "To stop:      docker-compose down"
    echo ""

    # Follow logs
    docker-compose logs -f
else
    echo ""
    echo "[DevSmith] ❌ Runtime validation failed!"
    echo ""
    echo "Validation status saved to: .validation/status.json"
    echo ""
    echo "Next steps:"
    echo "  1. Review issues: cat .validation/status.json | jq '.validation.issues[]'"
    echo "  2. Check logs: docker-compose logs [service]"
    echo "  3. Tell Copilot to read .validation/status.json and fix issues"
    echo ""
    exit 1
fi
