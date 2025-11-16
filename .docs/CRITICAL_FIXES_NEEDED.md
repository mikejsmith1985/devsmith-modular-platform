# CRITICAL FIXES NEEDED - Review App Non-Functional

**Status**: BROKEN - Review app appears functional but does NOT work
**Priority**: P0 - Blocks all user value delivery
**Created**: 2025-11-01
**Token Budget Used**: ~145K of 200K (Haiku)

---

## Executive Summary

The Review service has passing E2E tests but **zero working functionality**. The tests validate HTML element existence, not actual user experience. This represents a complete failure to deliver on the core promise of the platform.

### What Works
- ✅ Services start and show healthy in Docker
- ✅ UI renders all 5 mode buttons (Preview, Skim, Scan, Detailed, Critical)
- ✅ Code textarea renders
- ✅ Dark mode toggle works (vanilla JS implementation)
- ✅ Navigation links work
- ✅ Database connections (when not exhausted)
- ✅ Ollama integration verified (host.docker.internal:11434)

### What's Broken
- ❌ **CRITICAL**: Code submission always returns "Code required" (400)
- ❌ **CRITICAL**: No AI model selector exists in UI
- ❌ **CRITICAL**: Form data binding fails - `c.PostForm()` returns empty string
- ❌ **CRITICAL**: No verification that Ollama actually processes requests
- ❌ Database connection pool exhausts (too many clients)
- ❌ E2E tests validate wrong thing (element existence, not functionality)

---

## Problem 1: Form Data Binding Failure (HIGHEST PRIORITY)

### Current State
File: `apps/review/handlers/ui_handler.go`
Function: `bindCodeRequest()`

**Issue**: When form data is posted with `pasted_code=<code>`, the function returns "Code required" (400).

**Test Command**:
```bash
curl -X POST "http://localhost:8081/api/review/modes/preview" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "pasted_code=func main() { println(\"test\") }"
```

**Expected**: Returns analysis JSON from Ollama
**Actual**: Returns "Code required" (400 error)

### Root Cause Analysis

The current implementation:
```go
func (h *UIHandler) bindCodeRequest(c *gin.Context) (string, bool) {
    if err := c.Request.ParseForm(); err != nil {
        h.logger.Warn("Failed to parse form", "error", err)
    }
    
    code := c.Request.PostFormValue("pasted_code")
    
    if code == "" {
        h.logger.Warn("No code provided in request")
        c.String(http.StatusBadRequest, "Code required. Please paste code in the textarea.")
        return "", false
    }
    return code, true
}
```

**Debug Steps Needed**:
1. Check if `ParseForm()` is being called successfully
2. Verify `c.Request.PostForm` contains the expected key
3. Check if Content-Type header is being respected
4. Verify Gin middleware isn't consuming the body before it reaches the handler

**Likely Causes**:
- Gin's request body may already be consumed by middleware
- Form parsing may need to happen earlier in the request lifecycle
- May need to use `c.ShouldBind()` with a struct instead
- HTMX may be sending data in a different format than expected

### Fix Strategy

**Option A: Use Gin's built-in binding**
```go
type CodeRequest struct {
    PastedCode string `form:"pasted_code" json:"pasted_code" binding:"required"`
}

func (h *UIHandler) bindCodeRequest(c *gin.Context) (string, bool) {
    var req CodeRequest
    
    // Try binding as form first, then JSON
    if err := c.ShouldBindWith(&req, binding.Form); err != nil {
        if err := c.ShouldBindJSON(&req); err != nil {
            h.logger.Warn("Failed to bind code", "error", err)
            c.String(http.StatusBadRequest, "Code required. Please paste code in the textarea.")
            return "", false
        }
    }
    
    return req.PastedCode, true
}
```

**Option B: Read raw body**
```go
func (h *UIHandler) bindCodeRequest(c *gin.Context) (string, bool) {
    bodyBytes, err := io.ReadAll(c.Request.Body)
    if err != nil {
        h.logger.Warn("Failed to read body", "error", err)
        c.String(http.StatusBadRequest, "Invalid request")
        return "", false
    }
    
    // Parse as form data
    values, err := url.ParseQuery(string(bodyBytes))
    if err != nil {
        h.logger.Warn("Failed to parse form", "error", err)
        c.String(http.StatusBadRequest, "Invalid form data")
        return "", false
    }
    
    code := values.Get("pasted_code")
    if code == "" {
        c.String(http.StatusBadRequest, "Code required")
        return "", false
    }
    
    return code, true
}
```

**Option C: Check if middleware is the problem**
1. Add logging middleware that dumps request body before it reaches handlers
2. Verify body isn't being consumed
3. If it is, modify middleware to restore body after reading

### Testing After Fix

**Test 1: Direct curl**
```bash
curl -v -X POST "http://localhost:8081/api/review/modes/preview" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "pasted_code=package main

func main() {
    println(\"hello world\")
}"
```

Expected: 200 OK with analysis JSON

**Test 2: Via browser UI**
1. Navigate to http://localhost:3000/review
2. Paste code in textarea
3. Click "Select Preview" button
4. Verify analysis appears below form

**Test 3: All 5 modes**
```bash
for mode in preview skim scan detailed critical; do
  echo "Testing $mode..."
  curl -X POST "http://localhost:8081/api/review/modes/$mode" \
    -d "pasted_code=test" -w "\nStatus: %{http_code}\n\n"
done
```

Expected: All return 200 with different analysis formats

---

## Problem 2: No AI Model Selector (CRITICAL UX ISSUE)

### Current State
The UI has no way for users to select which Ollama model to use. The model is hardcoded in backend.

**Files Involved**:
- `apps/review/templates/index.templ` - Main review UI
- `apps/review/handlers/ui_handler.go` - Mode handlers
- `internal/ai/client.go` - Ollama client

### What Needs to Be Built

**Step 1: Add Model Selector to UI**

Location: `apps/review/templates/index.templ`
Position: Add after the code textarea, before the mode buttons

```html
<!-- Model Selection -->
<div class="mb-6">
    <label for="model-select" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        AI Model
    </label>
    <select 
        id="model-select" 
        name="model"
        class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
    >
        <option value="mistral:7b-instruct" selected>Mistral 7B (Fast, General)</option>
        <option value="codellama:13b">CodeLlama 13B (Better for code)</option>
        <option value="llama2:13b">Llama 2 13B (Balanced)</option>
        <option value="deepseek-coder:6.7b">DeepSeek Coder (Code specialist)</option>
    </select>
    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Select the AI model for code analysis. Larger models are slower but more accurate.
    </p>
</div>
```

**Step 2: Wire Model Selection to Backend**

Modify form submission to include model:

```javascript
// In apps/review/static/js/review.js
document.querySelectorAll('[data-mode]').forEach(button => {
    button.addEventListener('click', async (e) => {
        const mode = e.target.dataset.mode;
        const code = document.getElementById('code-input').value;
        const model = document.getElementById('model-select').value;
        
        const formData = new FormData();
        formData.append('pasted_code', code);
        formData.append('model', model);
        
        const response = await fetch(`/api/review/modes/${mode}`, {
            method: 'POST',
            body: formData
        });
        
        // Handle response...
    });
});
```

**Step 3: Update Backend to Accept Model Parameter**

Modify `bindCodeRequest` to also extract model:

```go
type CodeRequest struct {
    PastedCode string `form:"pasted_code" json:"pasted_code" binding:"required"`
    Model      string `form:"model" json:"model"`
}

func (h *UIHandler) bindCodeRequest(c *gin.Context) (*CodeRequest, bool) {
    var req CodeRequest
    
    if err := c.ShouldBind(&req); err != nil {
        h.logger.Warn("Failed to bind request", "error", err)
        c.String(http.StatusBadRequest, "Code required")
        return nil, false
    }
    
    // Default model if not provided
    if req.Model == "" {
        req.Model = "mistral:7b-instruct"
    }
    
    return &req, true
}
```

**Step 4: Pass Model to Ollama Client**

Update mode handlers:

```go
func (h *UIHandler) HandlePreview(c *gin.Context) {
    req, ok := h.bindCodeRequest(c)
    if !ok {
        return
    }

    // Override Ollama client model for this request
    ctx := context.WithValue(c.Request.Context(), "model", req.Model)
    
    result, err := h.previewService.Analyze(ctx, req.PastedCode)
    if err != nil {
        h.logger.Error("Preview analysis failed", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed"})
        return
    }

    h.marshalAndFormat(c, result, "Preview Analysis", "bg-blue-50")
}
```

**Step 5: Update Ollama Client to Use Context Model**

Modify `internal/ai/client.go`:

```go
func (c *Client) Generate(ctx context.Context, prompt string) (*Response, error) {
    // Check if model override in context
    model := c.model
    if ctxModel := ctx.Value("model"); ctxModel != nil {
        if m, ok := ctxModel.(string); ok && m != "" {
            model = m
        }
    }
    
    req := &Request{
        Model:  model,
        Prompt: prompt,
        Stream: false,
        Options: map[string]interface{}{
            "temperature": 0.7,
        },
    }
    
    // ... rest of implementation
}
```

**Step 6: Add Model Validation**

Query Ollama for available models:

```bash
curl http://host.docker.internal:11434/api/tags
```

Add validation endpoint:

```go
// apps/review/handlers/ui_handler.go
func (h *UIHandler) GetAvailableModels(c *gin.Context) {
    models, err := h.reviewService.GetAvailableModels(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"models": models})
}
```

**Step 7: Dynamic Model Dropdown**

Instead of hardcoding models, fetch from Ollama:

```javascript
// apps/review/static/js/review.js
async function loadAvailableModels() {
    try {
        const response = await fetch('/api/review/models');
        const data = await response.json();
        
        const select = document.getElementById('model-select');
        select.innerHTML = '';
        
        data.models.forEach(model => {
            const option = document.createElement('option');
            option.value = model.name;
            option.textContent = `${model.name} (${model.size})`;
            select.appendChild(option);
        });
    } catch (error) {
        console.error('Failed to load models:', error);
    }
}

// Call on page load
document.addEventListener('DOMContentLoaded', loadAvailableModels);
```

### Testing After Implementation

**Test 1: UI renders selector**
1. Navigate to http://localhost:3000/review
2. Verify model dropdown is visible
3. Verify dropdown contains at least one model
4. Verify default selection is "mistral:7b-instruct"

**Test 2: Model selection persists**
1. Select a different model
2. Submit code for analysis
3. Verify logs show selected model was used

**Test 3: Model override works**
```bash
curl -X POST "http://localhost:8081/api/review/modes/preview" \
  -d "pasted_code=test" \
  -d "model=codellama:13b"
```

Verify logs show: `"model":"codellama:13b"`

---

## Problem 3: Database Connection Pool Exhaustion

### Current State
Services frequently fail with:
```
FATAL: sorry, too many clients already (SQLSTATE 53300)
```

### Root Cause
- PostgreSQL default max_connections: 100
- Each service creates connection pool
- Services don't close connections properly
- Health checks create new connections frequently

### Fix Strategy

**Step 1: Increase PostgreSQL max_connections**

Edit `docker-compose.yml`:

```yaml
postgres:
  image: postgres:15-alpine
  container_name: devsmith-postgres
  environment:
    POSTGRES_USER: devsmith
    POSTGRES_PASSWORD: devsmith_password
    POSTGRES_DB: devsmith
  command: 
    - "postgres"
    - "-c"
    - "max_connections=200"
    - "-c"
    - "shared_buffers=256MB"
  # ... rest of config
```

**Step 2: Configure Connection Pooling in Services**

Each service should use proper pooling:

```go
// Example from apps/review/main.go
func initDB() (*sql.DB, error) {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        return nil, err
    }
    
    // Configure pool
    db.SetMaxOpenConns(10)        // Max 10 connections per service
    db.SetMaxIdleConns(5)         // Keep 5 idle
    db.SetConnMaxLifetime(time.Hour)
    db.SetConnMaxIdleTime(10 * time.Minute)
    
    return db, nil
}
```

**Step 3: Ensure Proper Connection Cleanup**

Review all DB queries to ensure:
```go
defer rows.Close()
defer stmt.Close()
```

**Step 4: Reduce Health Check Connection Usage**

Health checks should reuse existing pool, not create new connections.

```go
func healthCheckHandler(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        
        if err := db.PingContext(ctx); err != nil {
            c.JSON(500, gin.H{"status": "unhealthy", "error": err.Error()})
            return
        }
        
        c.JSON(200, gin.H{"status": "healthy"})
    }
}
```

### Testing After Fix

**Test 1: Monitor connection count**
```bash
docker exec -it devsmith-postgres psql -U devsmith -c \
  "SELECT count(*) FROM pg_stat_activity WHERE datname='devsmith';"
```

Should stay under 50 connections even with all services running.

**Test 2: Stress test**
```bash
for i in {1..100}; do
  curl -s http://localhost:3000/health &
done
wait
```

No "too many clients" errors should appear.

**Test 3: Connection leak detection**
Run for 5 minutes, monitor connection count every 30 seconds:
```bash
while true; do
  docker exec devsmith-postgres psql -U devsmith -c \
    "SELECT count(*) FROM pg_stat_activity WHERE datname='devsmith';"
  sleep 30
done
```

Connection count should remain stable, not grow continuously.

---

## Problem 4: E2E Tests Validate Wrong Thing

### Current State
Tests check for HTML element existence, not actual functionality.

Example from `tests/e2e/smoke/ollama-integration/review-loads.spec.ts`:
```typescript
await expect(page.locator('button:has-text("Select Preview")')).toBeVisible();
```

This passes even when clicking the button does nothing.

### What Tests Should Actually Do

**Test 1: Full Preview Mode Flow**
```typescript
test('Preview mode analyzes code end-to-end', async ({ page }) => {
  await page.goto('/review');
  
  // Enter code
  await page.fill('#code-input', 'func main() { println("test") }');
  
  // Select model
  await page.selectOption('#model-select', 'mistral:7b-instruct');
  
  // Click Preview
  await page.click('button:has-text("Select Preview")');
  
  // Wait for analysis to complete
  await page.waitForSelector('.analysis-result', { timeout: 15000 });
  
  // Verify analysis contains expected content
  const result = await page.textContent('.analysis-result');
  expect(result).toContain('Structure');
  expect(result).toContain('Components');
  expect(result).toContain('Dependencies');
  
  // Verify Ollama was actually called
  const logs = await getDockerLogs('review');
  expect(logs).toContain('Calling Ollama');
  expect(logs).toContain('Analysis complete');
});
```

**Test 2: All 5 Modes Work**
```typescript
test('All reading modes return different analyses', async ({ page }) => {
  const modes = ['preview', 'skim', 'scan', 'detailed', 'critical'];
  const results = [];
  
  for (const mode of modes) {
    await page.goto('/review');
    await page.fill('#code-input', TEST_CODE);
    await page.click(`button[data-mode="${mode}"]`);
    await page.waitForSelector('.analysis-result');
    
    const text = await page.textContent('.analysis-result');
    results.push(text);
  }
  
  // Verify all results are different
  const uniqueResults = new Set(results);
  expect(uniqueResults.size).toBe(5);
  
  // Verify each mode has expected keywords
  expect(results[0]).toContain('Structure');  // Preview
  expect(results[1]).toContain('Abstractions'); // Skim
  expect(results[2]).toContain('Search'); // Scan
  expect(results[3]).toContain('Line-by-line'); // Detailed
  expect(results[4]).toContain('Quality'); // Critical
});
```

**Test 3: Model Selection Works**
```typescript
test('Different models produce different output', async ({ page }) => {
  const models = ['mistral:7b-instruct', 'codellama:13b'];
  const results = [];
  
  for (const model of models) {
    await page.goto('/review');
    await page.fill('#code-input', TEST_CODE);
    await page.selectOption('#model-select', model);
    await page.click('button[data-mode="preview"]');
    await page.waitForSelector('.analysis-result');
    
    const text = await page.textContent('.analysis-result');
    results.push(text);
  }
  
  // Results should be different (different models analyze differently)
  expect(results[0]).not.toBe(results[1]);
});
```

**Test 4: Error Handling**
```typescript
test('Empty code shows validation error', async ({ page }) => {
  await page.goto('/review');
  await page.click('button[data-mode="preview"]');
  
  await expect(page.locator('.error-message')).toContainText('Code required');
});

test('Ollama failure shows user-friendly error', async ({ page }) => {
  // Stop Ollama container
  await exec('docker stop devsmith-ollama');
  
  await page.goto('/review');
  await page.fill('#code-input', TEST_CODE);
  await page.click('button[data-mode="preview"]');
  
  await expect(page.locator('.error-message')).toContainText('AI service unavailable');
  
  // Restart Ollama
  await exec('docker start devsmith-ollama');
});
```

### Rewrite Required Files

1. `tests/e2e/smoke/ollama-integration/review-loads.spec.ts` → `review-functionality.spec.ts`
2. Add: `tests/e2e/smoke/ollama-integration/all-modes-work.spec.ts`
3. Add: `tests/e2e/smoke/ollama-integration/model-selection.spec.ts`
4. Add: `tests/e2e/smoke/ollama-integration/error-handling.spec.ts`

---

## Problem 5: No Verification Ollama Actually Processes Requests

### Current State
We assume Ollama works because health check passes. We never verify:
- Requests actually reach Ollama
- Ollama returns valid analysis
- Analysis is displayed to user

### What Needs to Be Added

**Step 1: Request/Response Logging**

Add to `internal/ai/client.go`:

```go
func (c *Client) Generate(ctx context.Context, prompt string) (*Response, error) {
    reqID := uuid.New().String()
    
    c.logger.Info("Sending request to Ollama",
        "request_id", reqID,
        "model", c.model,
        "prompt_length", len(prompt))
    
    start := time.Now()
    
    // ... make request ...
    
    c.logger.Info("Received response from Ollama",
        "request_id", reqID,
        "duration_ms", time.Since(start).Milliseconds(),
        "response_length", len(resp.Response))
    
    return resp, nil
}
```

**Step 2: Add Tracing Header**

```go
func (h *UIHandler) HandlePreview(c *gin.Context) {
    traceID := c.GetHeader("X-Trace-ID")
    if traceID == "" {
        traceID = uuid.New().String()
    }
    
    ctx := context.WithValue(c.Request.Context(), "trace_id", traceID)
    
    h.logger.Info("Starting analysis",
        "trace_id", traceID,
        "mode", "preview")
    
    // ... analysis ...
    
    c.Header("X-Trace-ID", traceID)
}
```

**Step 3: Add Metrics Endpoint**

```go
// apps/review/handlers/metrics.go
type Metrics struct {
    TotalRequests    int64
    SuccessfulCalls  int64
    FailedCalls      int64
    AverageDuration  float64
}

func (h *UIHandler) GetMetrics(c *gin.Context) {
    metrics := h.reviewService.GetMetrics()
    c.JSON(200, metrics)
}
```

**Step 4: Test Verification Script**

Create `scripts/verify-ollama-integration.sh`:

```bash
#!/bin/bash

set -e

echo "=== Ollama Integration Verification ==="

# Test 1: Ollama is reachable
echo "1. Testing Ollama connectivity..."
curl -f http://localhost:11434/api/tags || {
    echo "FAIL: Cannot reach Ollama"
    exit 1
}
echo "PASS"

# Test 2: Review service can call Ollama
echo "2. Testing Review → Ollama integration..."
TRACE_ID=$(uuidgen)
RESPONSE=$(curl -s -H "X-Trace-ID: $TRACE_ID" \
  -X POST http://localhost:8081/api/review/modes/preview \
  -d "pasted_code=func main() {}")

if echo "$RESPONSE" | grep -q "Structure"; then
    echo "PASS"
else
    echo "FAIL: No analysis returned"
    echo "Response: $RESPONSE"
    exit 1
fi

# Test 3: Verify logs show Ollama call
echo "3. Checking logs for Ollama interaction..."
docker logs devsmith-review-1 2>&1 | grep -q "Sending request to Ollama" || {
    echo "FAIL: No Ollama request logged"
    exit 1
}
echo "PASS"

# Test 4: Verify response time is reasonable
echo "4. Testing response time..."
START=$(date +%s%N)
curl -s -X POST http://localhost:8081/api/review/modes/preview \
  -d "pasted_code=func main() {}" > /dev/null
END=$(date +%s%N)
DURATION=$(( (END - START) / 1000000 ))

if [ $DURATION -lt 10000 ]; then
    echo "PASS (${DURATION}ms)"
else
    echo "FAIL: Too slow (${DURATION}ms)"
    exit 1
fi

echo "=== All Tests Passed ==="
```

---

## Action Plan for AI Agent

**Priority Order**:

### Phase 1: Fix Core Functionality (BLOCKING)
1. ✅ **Fix form data binding** (Problem 1)
   - Try Option A first (Gin's ShouldBind)
   - If that fails, try Option B (raw body reading)
   - Test with curl after each attempt
   - Verify logs show code being received

2. ✅ **Fix database connections** (Problem 3)
   - Update docker-compose.yml with max_connections=200
   - Add connection pooling to all services
   - Restart all services
   - Monitor connection count

3. ✅ **Verify Ollama integration** (Problem 5)
   - Add request/response logging
   - Run verification script
   - Confirm analysis actually happens
   - Test all 5 modes return different results

### Phase 2: Add Model Selector (HIGH PRIORITY)
4. ✅ **UI implementation**
   - Add dropdown to index.templ
   - Wire to form submission
   - Test dropdown renders and populates

5. ✅ **Backend wiring**
   - Update bindCodeRequest to accept model parameter
   - Pass model to Ollama client
   - Add model validation
   - Test model override works

6. ✅ **Dynamic model loading**
   - Add /api/review/models endpoint
   - Fetch available models from Ollama
   - Populate dropdown dynamically
   - Handle Ollama unreachable gracefully

### Phase 3: Fix E2E Tests (QUALITY GATE)
7. ✅ **Rewrite tests**
   - Create review-functionality.spec.ts
   - Test actual code analysis happens
   - Verify different modes return different results
   - Test error handling

8. ✅ **Add integration verification**
   - Create verify-ollama-integration.sh
   - Add to pre-push hook
   - Run in CI/CD
   - Document expected behavior

### Phase 4: Validation (DONE CRITERIA)
9. ✅ **Manual testing**
   - Submit code in UI
   - Verify analysis appears
   - Test all 5 modes
   - Change model and verify different output
   - Test error cases

10. ✅ **Automated testing**
    - All E2E tests pass
    - Verification script passes
    - No database connection errors
    - Response times < 5 seconds

---

## Success Criteria

The Review app is considered **FUNCTIONAL** when:

1. ✅ User can paste code in textarea
2. ✅ User can select AI model from dropdown
3. ✅ Clicking any mode button triggers actual analysis
4. ✅ Analysis result appears in UI within 5 seconds
5. ✅ All 5 modes return different analyses
6. ✅ Different models return different analyses
7. ✅ Empty code shows validation error
8. ✅ Ollama failure shows user-friendly error
9. ✅ No database connection errors for 10 minutes of use
10. ✅ E2E tests validate actual functionality, not just HTML

---

## Current Code State

### Files Modified Today (2025-11-01)
- `apps/review/handlers/ui_handler.go` - Multiple attempts to fix binding (STILL BROKEN)
- `apps/review/templates/index.templ` - HTMX form structure
- `docker/nginx/conf.d/default.conf` - Fixed proxy_pass trailing slash
- `apps/logs/handlers/ui_handler.go` - Added /logs route
- `apps/analytics/handlers/ui_handler.go` - Added /analytics route
- Various E2E test files - Updated selectors (BUT TESTS DON'T VALIDATE FUNCTIONALITY)

### Files That Need Changes
- `apps/review/handlers/ui_handler.go` - Fix bindCodeRequest (CRITICAL)
- `apps/review/templates/index.templ` - Add model selector UI
- `apps/review/static/js/review.js` - Wire model selection to form
- `internal/ai/client.go` - Support model override from context
- `docker-compose.yml` - Increase postgres max_connections
- All E2E test files in `tests/e2e/smoke/ollama-integration/` - Rewrite to test functionality
- `scripts/verify-ollama-integration.sh` - Create new verification script

### Files That Work (Don't Break These)
- `internal/ui/components/nav/nav.templ` - Dark mode toggle (vanilla JS)
- `docker/nginx/conf.d/default.conf` - Routing (fixed today)
- `playwright.config.ts` - Test configuration
- `.git/hooks/pre-push` - Quality gates
- All authentication code - Works correctly

---

## Known Gotchas

1. **Templ Stripping**: Templ strips Alpine.js directives. Use vanilla JS or HTMX only.

2. **Docker Networking**: Use `host.docker.internal:11434` for Ollama, not `localhost`.

3. **Gin Body Consumption**: Request body can only be read once. Use `c.ShouldBind()` or copy body before reading.

4. **Database Pool**: Default 100 max_connections is too low for 5 services + health checks.

5. **HTMX Form Data**: HTMX sends `application/x-www-form-urlencoded` by default, not JSON.

6. **E2E Test Authentication**: Must POST to `/auth/test-login` before accessing authenticated routes.

7. **Mode Buttons**: Currently configured with HTMX but form binding prevents reaching service layer.

8. **Service Restart Loop**: If DB connection fails, service crashes and restarts, exhausting connections faster.

---

## Debugging Commands

```bash
# Check service health
docker-compose ps

# View service logs
docker-compose logs -f review

# Test endpoint directly
curl -v -X POST http://localhost:8081/api/review/modes/preview \
  -d "pasted_code=test"

# Check database connections
docker exec devsmith-postgres psql -U devsmith -c \
  "SELECT count(*) FROM pg_stat_activity WHERE datname='devsmith';"

# Test Ollama directly
curl http://localhost:11434/api/generate -d '{
  "model": "mistral:7b-instruct",
  "prompt": "Say hello",
  "stream": false
}'

# Run E2E tests
cd tests/e2e && npm test

# Check pre-push hook
.git/hooks/pre-push

# Monitor response times
time curl -X POST http://localhost:8081/api/review/modes/preview \
  -d "pasted_code=test"
```

---

## Token Budget Remaining

**Used**: ~145K Haiku tokens
**Remaining**: ~55K Haiku tokens

**Estimated Work Required**:
- Fix form binding: ~10K tokens
- Add model selector: ~15K tokens  
- Fix database pooling: ~5K tokens
- Rewrite E2E tests: ~20K tokens
- Testing and validation: ~10K tokens
**Total**: ~60K tokens

**Recommendation**: Continue with remaining budget. The work is well-scoped now.

---

## Questions for AI Agent

Before starting, verify:
1. Can you access all files mentioned above?
2. Can you run Docker commands?
3. Can you execute curl commands for testing?
4. Can you read Docker logs?
5. Do you have access to create new files?

If yes to all, proceed with Phase 1, Step 1.

If no to any, document what you cannot do so user can assist.

---

**END OF DOCUMENT**

