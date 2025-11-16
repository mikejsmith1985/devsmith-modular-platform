#!/bin/bash
set -euo pipefail

echo "🚀 Starting atomic Portal deployment..."

# Check if we're in the right directory
if [[ ! -f "docker-compose.yml" ]]; then
    echo "❌ Error: Must run from project root (where docker-compose.yml exists)"
    exit 1
fi

# Build frontend inside Docker (no local node_modules needed)
echo "📦 Building Portal with embedded frontend..."
docker-compose build --no-cache portal

# Deploy all required services (Traefik, Portal, dependencies)
echo "🔄 Deploying all services..."
docker-compose up -d

# Wait for startup
echo "⏳ Waiting for services to start..."
sleep 15

# Verify Traefik is running
echo "✅ Verifying Traefik..."
if docker-compose ps traefik | grep -q "Up"; then
    echo "✅ Traefik is running"
else
    echo "❌ Traefik is not running!"
    docker-compose logs traefik --tail=50
    exit 1
fi

# Verify Portal health via Traefik (port 3000)
echo "✅ Verifying Portal via Traefik (port 3000)..."
if curl -f http://localhost:3000/health &>/dev/null || curl -f http://localhost:3000/ &>/dev/null; then
    echo "✅ Portal accessible via Traefik!"
else
    echo "❌ Portal not accessible via Traefik! Checking logs..."
    echo "--- Traefik logs ---"
    docker-compose logs traefik --tail=30
    echo "--- Portal logs ---"
    docker-compose logs portal --tail=30
    exit 1
fi

# Verify Portal backend API
echo "✅ Verifying Portal backend API..."
if curl -f http://localhost:3001/api/portal/health &>/dev/null; then
    echo "✅ Portal backend API is healthy"
else
    echo "❌ Portal backend API failed!"
    docker-compose logs portal --tail=50
    exit 1
fi

# Check bundle version
NEW_BUNDLE=$(curl -s http://localhost:3000/ | grep -o 'index-[^.]*\.js' || echo "unknown")
echo "✅ Portal deployed successfully! New bundle: $NEW_BUNDLE"

# Show container status
echo "📊 Container status:"
docker-compose ps

echo "🎉 Atomic deployment complete!"