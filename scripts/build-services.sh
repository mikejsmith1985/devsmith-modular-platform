#!/bin/bash
set -e

echo "Building Go services..."

# Create bin directory
mkdir -p bin

# Generate Templ templates first
echo "→ Generating Templ templates..."
find apps -name "*.templ" -exec dirname {} \; | sort -u | while read dir; do
  echo "  Generating templates in $dir"
  (cd "$dir" && templ generate) || true
done

# Build each service
echo "→ Building Portal service..."
go build -o bin/portal ./cmd/portal

echo "→ Building Review service..."
go build -o bin/review ./cmd/review

echo "→ Building Logs service..."
go build -o bin/logs ./cmd/logs

echo "→ Building Analytics service..."
go build -o bin/analytics ./cmd/analytics

echo "✓ All services built successfully"
