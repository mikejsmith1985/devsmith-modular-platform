#!/bin/bash

echo "=== JSON RESPONSE TEST ==="
echo "Testing Preview mode..."

# Test Preview mode
response=$(curl -s -X POST http://localhost:3000/api/review/modes/preview \
  -F "pasted_code=package main
func main() {
    println(\"Hello\")
}" \
  -F "model=qwen2.5-coder:7b")

# Check if response is JSON
if echo "$response" | jq -e . >/dev/null 2>&1; then
    echo "✅ Preview mode returns JSON"
    echo "$response" | jq -r '.summary // "No summary"' | head -3
else
    echo "❌ Preview mode does NOT return JSON"
    echo "Response (first 200 chars):"
    echo "$response" | head -c 200
fi

echo ""
echo "Testing Skim mode..."

# Test Skim mode
response=$(curl -s -X POST http://localhost:3000/api/review/modes/skim \
  -F "pasted_code=package main
func main() {}" \
  -F "model=qwen2.5-coder:7b")

if echo "$response" | jq -e . >/dev/null 2>&1; then
    echo "✅ Skim mode returns JSON"
else
    echo "❌ Skim mode does NOT return JSON"
fi

echo ""
echo "=== TEST COMPLETE ==="
