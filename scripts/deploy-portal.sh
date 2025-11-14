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

# Deploy
echo "🔄 Deploying Portal..."
docker-compose up -d portal

# Wait for startup
echo "⏳ Waiting for Portal to start..."
sleep 10

# Verify health
echo "✅ Verifying deployment..."
if curl -f http://localhost:3000/health &>/dev/null; then
    echo "✅ Health check passed!"
else
    echo "❌ Health check failed! Checking logs..."
    docker-compose logs portal --tail=50
    exit 1
fi

# Check bundle version
NEW_BUNDLE=$(curl -s http://localhost:3000/ | grep -o 'index-[^.]*\.js' || echo "unknown")
echo "✅ Portal deployed successfully! New bundle: $NEW_BUNDLE"

# Show container status
echo "📊 Container status:"
docker-compose ps portal

echo "🎉 Atomic deployment complete!"