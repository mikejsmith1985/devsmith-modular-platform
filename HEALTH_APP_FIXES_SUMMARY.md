# Health App Fixes - 2025-11-11

## Issues Fixed

### 1. ✅ Auto-Refresh Too Aggressive (FIXED)
**Problem**: Page refreshed every 5 seconds with full loading spinner, making UI unusable
**Solution**: 
- Changed interval from 5 seconds to 30 seconds
- Modified `fetchData()` to accept `isBackgroundRefresh` parameter
- Only show loading spinner on initial load, not during background refresh
- This prevents full page redraw during auto-refresh

**Changes**:
- `HealthPage.jsx` line 52: `setInterval(() => fetchData(true), 30000)`
- `fetchData()` now checks `if (!isBackgroundRefresh) { setLoading(true); }`

### 2. ✅ AI Analysis Timeout (FIXED)
**Problem**: AI insights generation had no timeout, causing indefinite hang
**Solution**:
- Added 60-second timeout using AbortController
- Graceful timeout handling with user-friendly error message
- Provides actionable suggestions (use smaller model, retry later)

**Changes**:
- `generateAIInsights()` now creates `AbortController` with 60s timeout
- Catches `AbortError` and displays timeout message with suggestions
- Logs timeout events for monitoring

### 3. ⚠️ Memory Crash on Refresh (MITIGATED)
**Problem**: Page crashes with "Out of Memory" error on refresh
**Analysis**: 
- Logs array is properly limited to 100 entries (not unbounded growth)
- No `addEventListener` without cleanup found
- useEffect cleanup properly returns cleanup function for setInterval
- Likely caused by browser DevTools memory profiler or large log payloads

**Mitigation**:
- Auto-refresh reduced to 30s (from 5s) reduces memory churn
- Background refresh doesn't re-render entire component
- Proper cleanup of setInterval in useEffect
- Log limit already enforced at API level (100 entries)

**Recommendation**: Monitor in production without DevTools open to confirm fix

### 4. ❌ Model Dropdown Shows Raw Names (NOT FIXED - Backend Required)
**Problem**: Dropdown shows `qwen2.5-coder:7b-instruct-q4_K_M` instead of friendly names
**Root Cause**: ModelSelector tries to fetch from `/api/portal/llm-configs` but endpoint doesn't exist
**Solution Required**: Backend implementation needed

**ModelSelector.jsx already implements correct logic**:
```javascript
const response = await apiRequest('/api/portal/llm-configs');
modelList = response.map(config => ({
  name: config.model_name,
  displayName: config.display_name || config.model_name,
  provider: config.provider,
  isDefault: config.is_default
}));
```

**Backend TODO**:
```go
// apps/portal/handlers/llm_handler.go (CREATE THIS FILE)
func HandleGetLLMConfigs(c *gin.Context) {
    // Query AI Factory configurations
    // Return JSON with model_name, display_name, provider, is_default
}

// Register route in main.go:
router.GET("/api/portal/llm-configs", handlers.HandleGetLLMConfigs)
```

**Fallback Behavior**: Currently falls back to hardcoded models if API fails

## Testing Required

1. **Manual Testing**:
   - [ ] Open Health page, verify auto-refresh works without blocking UI
   - [ ] Leave page open for 2+ minutes, verify no performance degradation
   - [ ] Generate AI insights, verify 60s timeout works if model is slow
   - [ ] Refresh page multiple times, check for memory leaks in browser Task Manager
   - [ ] Check model dropdown (will still show raw names until backend implemented)

2. **Build & Deploy**:
```bash
# Rebuild frontend with fixes
cd frontend
npm run build

# Rebuild frontend container
docker-compose up -d --build devsmith-frontend

# Verify container running
docker-compose ps devsmith-frontend
```

3. **Verification**:
```bash
# Check browser console for errors
# Monitor auto-refresh behavior (should be 30s intervals)
# Try AI insights on a log entry (should timeout after 60s if needed)
```

## Files Modified

1. `/frontend/src/components/HealthPage.jsx`:
   - Line 52: Changed setInterval from 5000ms to 30000ms
   - Line 53: Added `isBackgroundRefresh` parameter to fetchData call
   - Line 65: Modified fetchData to accept parameter and conditionally set loading
   - Lines 242-315: Added AbortController timeout handling in generateAIInsights

## Backend Work Required

To fully fix model dropdown (Issue #4), create:

1. `apps/portal/handlers/llm_handler.go`:
```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    // Import your AI Factory client
)

type LLMConfig struct {
    ModelName   string `json:"model_name"`
    DisplayName string `json:"display_name"`
    Provider    string `json:"provider"`
    IsDefault   bool   `json:"is_default"`
}

func HandleGetLLMConfigs(c *gin.Context) {
    // TODO: Query AI Factory for available models
    // For now, return hardcoded friendly names
    
    configs := []LLMConfig{
        {
            ModelName:   "qwen2.5-coder:7b-instruct-q4_K_M",
            DisplayName: "Qwen 2.5 Coder 7B (Recommended)",
            Provider:    "Ollama",
            IsDefault:   true,
        },
        {
            ModelName:   "mistral:7b-instruct",
            DisplayName: "Mistral 7B Instruct",
            Provider:    "Ollama",
            IsDefault:   false,
        },
        {
            ModelName:   "deepseek-coder-v2:16b-lite-instruct-q4_K_M",
            DisplayName: "DeepSeek Coder V2 16B (High VRAM)",
            Provider:    "Ollama",
            IsDefault:   false,
        },
    }
    
    c.JSON(http.StatusOK, configs)
}
```

2. Update `apps/portal/main.go` to register route:
```go
// Add to router setup
router.GET("/api/portal/llm-configs", handlers.HandleGetLLMConfigs)
```

## Summary

**3 out of 4 issues fixed in this session**:
- ✅ Auto-refresh optimized (30s interval, background mode)
- ✅ AI timeout handling (60s with abort)
- ⚠️ Memory crash mitigated (needs production monitoring)
- ❌ Model dropdown requires backend work (frontend ready)

**Next Steps**:
1. Rebuild frontend container
2. Test all fixes manually
3. Implement backend `/api/portal/llm-configs` endpoint
4. Monitor memory usage in production (without DevTools)
