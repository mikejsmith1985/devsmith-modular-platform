# Review Service Caching Layer

## Overview

The caching layer provides high-performance caching for analysis results in the Review Service, reducing AI API calls and improving response times for recurring queries.

**Issue #26 Implementation**: Performance Optimization & Caching

## Components

### CacheInterface

The `CacheInterface` defines the contract for all cache implementations, allowing for future Redis or Memcached backends:

```go
type CacheInterface interface {
    Get(ctx context.Context, reviewID int64, mode string) (*review_models.AnalysisResult, error)
    Set(ctx context.Context, reviewID int64, mode string, result *review_models.AnalysisResult, ttl time.Duration) error
    Delete(ctx context.Context, reviewID int64, mode string) error
    Clear(ctx context.Context) error
    Stats(ctx context.Context) *CacheStats
}
```

### InMemoryCache

Thread-safe in-memory implementation with:
- TTL-based automatic expiration
- Hit/miss tracking
- Background cleanup goroutine
- Per-mode caching per session

**Performance Characteristics:**
- **Hit:** O(1) lookup + statistics update
- **Miss:** O(1) lookup + statistics update  
- **Set:** O(1) insertion + expiration tracking
- **Memory:** ~1KB per cached analysis result

### CacheStats

Tracks performance metrics:
- `Hits`: Successful cache retrievals
- `Misses`: Failed lookups
- `Evictions`: Expired entries removed
- `HitRate`: Percentage of successful hits

## Usage

### Basic Usage

```go
// Create a new cache
cache := cache.NewInMemoryCache()

// Store an analysis result
err := cache.Set(ctx, sessionID, "skim", result, 1*time.Hour)
if err != nil {
    log.Fatalf("Failed to cache: %v", err)
}

// Retrieve from cache
cached, err := cache.Get(ctx, sessionID, "skim")
if cached != nil {
    // Use cached result
    return cached, nil
}

// Cache miss - call AI API
result := callOllamaAPI(...)
cache.Set(ctx, sessionID, "skim", result, 1*time.Hour)
```

### Monitoring

```go
stats := cache.Stats(ctx)
fmt.Printf("Cache Hit Rate: %.2f%% (%d hits, %d misses)\n", 
    stats.HitRate, stats.Hits, stats.Misses)
fmt.Printf("Current Size: %d entries\n", stats.CurrentSize)
```

## Performance Impact

### Issue #26 Requirements Met

✅ **Response Time**: <500ms for cached queries
- In-memory lookup: <1ms
- Expiration check: <1ms
- Lock contention: Minimal (RWMutex)

✅ **Cache Hit/Miss Tracking**
- All operations recorded
- Hit rate calculation
- Eviction counter

✅ **TTL Support**
- Default: 1 hour for analysis results
- Configurable per Set() call
- Automatic cleanup every minute

✅ **Multiple Analysis Modes**
- Separate cache entries per mode per session
- Example: Session #1 has skim, scan, detailed, critical cached independently

## Cache Keys

Format: `analysis:{reviewID}:{mode}`

Examples:
- `analysis:123:skim` - Skim mode analysis for review #123
- `analysis:123:scan` - Scan mode analysis for review #123
- `analysis:456:critical` - Critical mode analysis for review #456

## Future Enhancements

1. **Redis Backend**: `RedisCache` implementation for distributed caching
2. **Cache Warming**: Pre-populate hot data on startup
3. **Metrics Export**: Prometheus metrics for monitoring
4. **Selective Invalidation**: Invalidate cache for specific users/repos
5. **Compression**: Compress large analysis results in cache

## Testing

Run cache tests:
```bash
go test -v ./internal/review/cache/...
```

Coverage includes:
- Set/Get operations
- TTL expiration
- Deletion
- Statistics
- Context cancellation
- Error handling
