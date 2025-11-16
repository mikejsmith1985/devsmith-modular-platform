#!/bin/bash
# Build Portal service with version information

set -e

# Get version info
VERSION=$(cat VERSION 2>/dev/null || echo "dev")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_TIMESTAMP=$(date +%s)

echo "Building Portal with:"
echo "  VERSION: $VERSION"
echo "  GIT_COMMIT: $GIT_COMMIT"
echo "  BUILD_TIME: $BUILD_TIME"
echo ""

# Build with version information
docker-compose build \
    --build-arg VERSION="$VERSION" \
    --build-arg GIT_COMMIT="$GIT_COMMIT" \
    --build-arg BUILD_TIME="$BUILD_TIME" \
    --build-arg BUILD_TIMESTAMP="$BUILD_TIMESTAMP" \
    portal

echo ""
echo "✅ Portal built successfully!"
echo "   Version: $VERSION-$GIT_COMMIT"
echo ""
echo "To start Portal:"
echo "  docker-compose up -d portal"
echo ""
echo "To check version:"
echo "  curl http://localhost:3000/api/portal/version | jq"
