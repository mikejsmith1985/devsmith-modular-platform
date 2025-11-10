# Phase 2 Testing Guide

## Quick Start

### 1. Access Health App
```bash
# Open browser
http://localhost:3000/health
```

### 2. Create Test Log (if needed)
```bash
curl -X POST http://localhost:8082/api/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "level": "error",
    "service": "review",
    "message": "Failed to connect to AI service: connection timeout after 30s",
    "metadata": {
      "endpoint": "http://host.docker.internal:11434",
      "model": "qwen2.5-coder:7b-instruct-q4_K_M",
      "timeout": "30s",
      "retry_count": 3
    }
  }'
```

### 3. Test AI Insights Flow

**Step 1:** Click on error log card in Health app  
**Step 2:** Log detail modal opens  
**Step 3:** Scroll to "AI Insights" section  
**Step 4:** Click "Generate Insights" button  
**Step 5:** Wait ~15 seconds (AI generation)  
**Step 6:** View analysis, root cause, suggestions  

### 4. Test Cached Insights

**Step 1:** Close modal  
**Step 2:** Click same log card again  
**Step 3:** Insights load instantly (cached)  
**Step 4:** Click "Regenerate" to create new analysis  

## Backend API Testing

### Generate Insights
```bash
curl -X POST http://localhost:8082/api/logs/1/insights \
  -H "Content-Type: application/json" \
  -d '{"model": "qwen2.5-coder:7b-instruct-q4_K_M"}' \
  -s | jq .
```

Expected response time: 10-20 seconds

### Retrieve Cached Insights
```bash
curl -s http://localhost:8082/api/logs/1/insights | jq .
```

Expected response time: <100ms

### Verify Database
```bash
docker exec devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith \
  -c "SELECT id, log_id, substring(analysis, 1, 100), 
      jsonb_array_length(suggestions) as suggestion_count 
      FROM logs.ai_insights;"
```

## Expected AI Response Quality

For the timeout error example:

**Analysis:** 
- "The log indicates an error in the 'review' service where a connection attempt to an AI service timed out after 30 seconds."

**Suggestions (5):**
1. Check network connectivity between services
2. Verify AI service is up and accessible
3. Increase timeout setting if acceptable
4. Review server logs for related errors
5. Implement retries with exponential backoff

## Troubleshooting

### No logs visible in Health app
```bash
# Check if logs exist
docker exec devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith \
  -c "SELECT COUNT(*) FROM logs.entries;"

# If 0, create test log (see Step 2 above)
```

### "Generate Insights" button doesn't respond
```bash
# Check logs service is running
docker ps | grep logs

# Check Ollama is accessible
curl http://localhost:11434/api/tags
```

### AI generation takes too long (>30s)
- **Cause:** Ollama model loading or slow inference
- **Solution:** Wait for first generation, subsequent calls faster (model cached)

### Error: "Failed to generate insights"
```bash
# Check logs service errors
docker logs devsmith-modular-platform-logs-1 --tail 50

# Common issues:
# - Ollama not running: Start with `ollama serve`
# - Model not available: Run `ollama pull qwen2.5-coder:7b-instruct-q4_K_M`
# - Network issues: Check Docker networking
```

## Performance Benchmarks

### First Generation (Cold Start)
- **AI Model Loading:** 5-10 seconds
- **Inference Time:** 10-15 seconds
- **Database Write:** <100ms
- **Total:** 15-25 seconds

### Cached Retrieval
- **Database Query:** <5ms
- **Network Transfer:** <50ms
- **Total:** <100ms

### Regeneration (Warm)
- **AI Model Loaded:** 0 seconds (cached)
- **Inference Time:** 10-15 seconds
- **Database Upsert:** <100ms
- **Total:** 10-16 seconds

## Success Criteria

- [x] Health app displays logs in card layout
- [x] Clicking card opens detail modal
- [x] "Generate Insights" button visible
- [x] Button shows spinner during generation
- [x] Insights display with analysis, suggestions
- [x] Cached insights load instantly on reopen
- [x] "Regenerate" button creates new analysis
- [x] Model selector changes AI model used
- [x] Error handling shows user-friendly messages
- [x] Database stores insights with JSONB

## Next Steps

After verifying Phase 2 works correctly, proceed to Phase 3:

**Phase 3: Smart Tagging System**
- Automatic content-based tag extraction
- Manual tag management in modal
- Tag-based filtering in log display
- GIN index for fast tag queries

See LOGS_ENHANCEMENT_PLAN.md for Phase 3 details.
