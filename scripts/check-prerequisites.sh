#!/bin/bash
set -e

MISSING=0

# Check Go
if ! command -v go &> /dev/null; then
  echo "❌ Go is not installed (required: >= 1.23)"
  MISSING=1
else
  GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
  echo "✓ Go $GO_VERSION"
fi

# Check PostgreSQL
if ! command -v psql &> /dev/null; then
  echo "❌ PostgreSQL is not installed (required: >= 14)"
  MISSING=1
else
  PG_VERSION=$(psql --version | awk '{print $3}')
  echo "✓ PostgreSQL $PG_VERSION"
fi

# Check Docker
if ! command -v docker &> /dev/null; then
  echo "⚠️  Docker is not installed (optional but recommended)"
else
  DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
  echo "✓ Docker $DOCKER_VERSION"
fi

# Check Ollama
if ! command -v ollama &> /dev/null; then
  echo "❌ Ollama is not installed (required for AI features)"
  MISSING=1
else
  echo "✓ Ollama installed"
fi

# Check Node.js (for development)
if ! command -v node &> /dev/null; then
  echo "⚠️  Node.js not found (optional for development)"
else
  NODE_VERSION=$(node --version)
  echo "✓ Node.js $NODE_VERSION"
fi

# Check Templ CLI
if ! command -v templ &> /dev/null; then
  echo "📦 Installing Templ CLI..."
  go install github.com/a-h/templ/cmd/templ@latest
  echo "✓ Templ CLI installed"
else
  echo "✓ Templ CLI installed"
fi

# Check curl (for health checks)
if ! command -v curl &> /dev/null; then
  echo "❌ curl is not installed (required for health checks)"
  MISSING=1
else
  echo "✓ curl installed"
fi

if [ $MISSING -eq 1 ]; then
  echo ""
  echo "❌ Missing required dependencies. Please install them first."
  echo ""
  echo "Installation instructions:"
  echo "  Go: https://go.dev/dl/"
  echo "  PostgreSQL: https://www.postgresql.org/download/"
  echo "  Ollama: https://ollama.ai/"
  echo "  curl: apt-get install curl (or brew install curl)"
  exit 1
fi

echo "✓ All prerequisites met"
