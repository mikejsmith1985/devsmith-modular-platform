#!/bin/bash
set -euo pipefail

echo "🔧 Running post-deployment database fixes..."

# Check if AI insights table exists
AI_INSIGHTS_EXISTS=$(docker-compose exec -T postgres psql -U devsmith -d devsmith -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'logs' AND table_name = 'ai_insights');" | xargs)

if [ "$AI_INSIGHTS_EXISTS" = "f" ]; then
    echo "📝 Creating missing AI insights table..."
    docker-compose exec -T postgres psql -U devsmith -d devsmith -c "$(cat internal/logs/db/migrations/20251110_001_add_ai_insights.sql)"
    echo "✅ AI insights table created"
else
    echo "✅ AI insights table already exists"
fi

# Test AI insights endpoint
echo "🧪 Testing AI insights endpoint..."
if curl -f -X POST "http://localhost:3000/api/logs/1/insights" -H "Content-Type: application/json" -d '{"model":"qwen2.5-coder:7b-instruct-q4_K_M"}' &>/dev/null; then
    echo "✅ AI insights endpoint is working"
else
    echo "❌ AI insights endpoint test failed"
    exit 1
fi

echo "🎉 Database fixes complete!"