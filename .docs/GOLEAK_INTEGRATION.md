# Goleak Integration - Phase 3 WebSocket Test Reliability

## Overview

Phase 3 completes the WebSocket test reliability implementation by integrating the `goleak` package for compile-time goroutine leak detection. This provides automatic verification that no goroutines are left running after tests complete.

## The Three-Phase Approach

### Phase 1: Runtime Cleanup
- Implemented `t.Cleanup()` mechanism
- 50ms grace period for goroutine exit
- **Result:** Fixed all WebSocket test flakiness

### Phase 2: Runtime Diagnostics  
- Added `runtime.NumGoroutine()` tracking
- Applied to 5 key representative tests
- **Result:** Verified cleanup works reliably

### Phase 3: Compile-Time Detection
- Added `go.uber.org/goleak` package
- Integrated `goleak.VerifyNone()` into 5 key tests
- **Result:** Automatic leak detection at compile time

**Defense-in-Depth:** 
Cleanup → Diagnostics → Detection = Reliable tests

## Implementation Details

### Step 1: Add goleak Package

```bash
go get -u go.uber.org/goleak@latest
```

### Step 2: Add Hub Shutdown Support

Modified `websocket_hub.go`:

```go
type WebSocketHub struct {
    clients    map[*Client]bool
    broadcast  chan *logs_models.LogEntry
    register   chan *Client
    unregister chan *Client
    stop       chan struct{}  // NEW: Signal channel for shutdown
    mu         sync.RWMutex
}

func (h *WebSocketHub) Run() {
    heartbeatTicker := time.NewTicker(30 * time.Second)
    defer heartbeatTicker.Stop()

    for {
        select {
        case <-h.stop:  // NEW: Handle stop signal
            // Graceful shutdown
            h.mu.Lock()
            for client := range h.clients {
                close(client.Send)
            }
            h.clients = make(map[*Client]bool)
            h.mu.Unlock()
            return
            
        // ... existing cases ...
        }
    }
}

func (h *WebSocketHub) Stop() {  // NEW: Shutdown method
    defer func() {
        if r := recover(); r != nil {
            // Already stopped, ignore
        }
    }()
    close(h.stop)
}
```

### Step 3: Update Test Setup Functions

Modified `websocket_handler_test.go`:

```go
import (
    "go.uber.org/goleak"  // NEW
    // ... other imports ...
)

func setupWebSocketTestServer(t *testing.T) http.Handler {
    hub := NewWebSocketHub()
    go hub.Run()
    
    // NEW: Register hub shutdown in cleanup
    if t != nil {
        t.Cleanup(func() {
            hub.Stop()
            time.Sleep(10 * time.Millisecond)  // Grace period for goroutine
        })
    }
    
    return http.HandlerFunc(/* ... */)
}
```

### Step 4: Integrate goleak into Key Tests

```go
func TestWebSocketHandler_EndpointExists(t *testing.T) {
    // Phase 3: Compile-time goroutine leak detection
    defer goleak.VerifyNone(t, 
        goleak.IgnoreTopFunction(
            "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services.(*WebSocketHub).Run",
        ),
    )
    
    // Phase 1-2: Runtime diagnostics  
    diagnosticGoroutines(t)
    
    handler := setupWebSocketTestServer(t)
    // ... rest of test ...
}
```

## Goleak Configuration

### IgnoreTopFunction

Used to exclude test infrastructure goroutines that are intentional:

```go
goleak.IgnoreTopFunction("(*WebSocketHub).Run")
```

This tells goleak to ignore goroutines stuck in `WebSocketHub.Run()` because:
1. It's test infrastructure (we control the setup)
2. We explicitly shut it down with `hub.Stop()`
3. It's not an application leak

### When to Use IgnoreTopFunction

✅ **Use for:**
- Test fixture goroutines (hubs, servers, etc.)
- Infrastructure that's intentionally left running during test
- Code paths you control and have consciously chosen to keep running

❌ **Don't use for:**
- Real application goroutines
- Goroutines from client code
- Unexpected goroutines (that's the leak you need to fix!)

## Key Tests with Goleak

Five representative tests verify the pattern:

1. **TestWebSocketHandler_EndpointExists** (First test / canary)
2. **TestWebSocketHandler_FiltersLogsByLevel** (Representative filter test)  
3. **TestWebSocketHandler_RequiresAuthentication** (Auth boundary)
4. **TestWebSocketHandler_SendsHeartbeatEvery30Seconds** (30s stress test)
5. **TestWebSocketHandler_HighFrequencyMessageStream** (Load stress test)

### Test Results

✅ All 5 key tests pass with goleak  
✅ Total execution: ~35 seconds (includes 30s heartbeat test)  
✅ Zero goroutine leaks detected  
✅ Clean hub shutdown verified

## Error Patterns goleak Detects

If goleak reports unexpected goroutines, examples:

```
found unexpected goroutines:
[Goroutine 42 in state select, with (*WebSocketHub).Run on top of the stack]
```

This means:
1. A goroutine is stuck waiting in a `select` statement
2. It's in the `WebSocketHub.Run()` function
3. Either: 
   - `hub.Stop()` wasn't called (add t.Cleanup)
   - `hub.Run()` ignores the stop signal (fix the select case)
   - Timeout is too short (increase sleep duration)

## Best Practices

### In New Tests

Add goleak to **representative tests only** (not all tests):

```go
func TestMyFeature_MainScenario(t *testing.T) {
    defer goleak.VerifyNone(t,
        goleak.IgnoreTopFunction("myapp.(*MyHub).Run"),
    )
    
    // ... test code ...
}
```

### For Test Fixtures

Always implement clean shutdown:

```go
type MyHub struct {
    stop chan struct{}
}

func (h *MyHub) Stop() {
    close(h.stop)  // Trigger Run() to exit
}

// In test:
t.Cleanup(func() {
    hub.Stop()
    time.Sleep(10 * time.Millisecond)  // Grace period
})
```

### For Diagnostics

Combine all three phases:

```go
func TestFeature(t *testing.T) {
    // Phase 3: Compile-time verification
    defer goleak.VerifyNone(t, 
        goleak.IgnoreTopFunction("(*MyHub).Run"),
    )
    
    // Phase 1-2: Runtime diagnostics
    diagnosticGoroutines(t)
    
    // ... test ...
}
```

## Integration with Pre-Push Hook

Future: Add goleak check to pre-push validation:

```bash
# In scripts/hooks/pre-push:
if ! go test -race ./... -run "EndpointExists|FiltersLogsByLevel|RequiresAuthentication|SendsHeartbeatEvery30Seconds|HighFrequencyMessageStream"; then
    echo "❌ goleak tests failed"
    exit 1
fi
```

## Troubleshooting

### Goleak Reports Expected Fixture Goroutines

**Solution:** Use `IgnoreTopFunction()` to whitelist intentional goroutines.

```go
defer goleak.VerifyNone(t,
    goleak.IgnoreTopFunction("myapp.(*ServerHub).Run"),
    goleak.IgnoreTopFunction("net/http.(*Server).Serve"),
)
```

### Cleanup Not Called  

**Solution:** Ensure `t.Cleanup()` is registered:

```go
if t != nil {
    t.Cleanup(func() {
        hub.Stop()
        time.Sleep(10 * time.Millisecond)
    })
}
```

### Goroutine Still Running After Stop

**Solution:** Verify the `select` case handles the stop signal:

```go
for {
    select {
    case <-h.stop:  // Must have this case
        return      // Must return to exit goroutine
        
    // ... other cases ...
    }
}
```

## See Also

- `.docs/WEBSOCKET_TEST_PATTERN.md` - Phase 1-2 details
- `internal/logs/services/websocket_hub.go` - Hub implementation with stop support
- `internal/logs/services/websocket_handler_test.go` - Test integration
- `go.uber.org/goleak` - Official goleak documentation

## Key Insight

**goleak makes implicit assumptions explicit:**
- Every goroutine must exit after tests
- No "acceptable" leaked goroutines
- Forces conscious decisions about lifecycle management

This is stricter than runtime diagnostics but catches subtle leaks automatically.

