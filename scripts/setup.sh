#!/bin/bash
# DevSmith Modular Platform Setup Script
set -e

echo "[DevSmith] Starting setup..."

# Build all Docker images
DOCKER_BUILDKIT=1 docker-compose build

echo "[DevSmith] Setup complete. Use ./scripts/dev.sh to start development."
