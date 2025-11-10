# Phase 2 Backend Infrastructure - Complete

## Summary
Successfully created the backend infrastructure for AI-powered log insights as part of Phase 2 of LOGS_ENHANCEMENT_PLAN.md (Tasks 2.1-2.4). The system is now ready for database migration and frontend integration.

## What Was Created

### 1. Database Schema Migration ✅
**File:** `internal/logs/db/migrations/20251110_001_add_ai_insights.sql`
- Creates `logs.ai_insights` table with columns:
  - `id` (serial primary key)
  - `log_id` (foreign key to logs.entries, UNIQUE constraint)
  - `analysis` (TEXT) - AI-generated analysis of log entry
  - `root_cause` (TEXT) - Root cause explanation
  - `suggestions` (JSONB) - Array of actionable suggestions
  - `model_used` (VARCHAR) - Which AI model generated insights
  - `generated_at` (TIMESTAMP) - When insights were generated
  - `created_at` (TIMESTAMP) - When record was created
- Indexes on `log_id` and `generated_at`
- UNIQUE constraint ensures one insight per log (regenerate overwrites)

### 2. Data Model ✅
**File:** `internal/logs/models/ai_insight.go`
- Defines `AIInsight` struct with JSON and DB tags
- Located in models package to avoid import cycles
- Used by both services and repository layers

### 3. Repository Layer ✅
**File:** `internal/logs/db/ai_insights_repository.go`
- Package: `logs_db` (following existing convention)
- Struct: `AIInsightsRepository` with database connection
- Methods:
  - `NewAIInsightsRepository(db *sql.DB)` - Constructor
  - `GetByLogID(ctx, logID)` - Retrieve cached insights
  - `Upsert(ctx, insight)` - Insert or update with ON CONFLICT
- Uses JSONB marshaling/unmarshaling for suggestions array
- Returns nil (not error) when no insights found

### 4. Service Layer ✅
**File:** `internal/logs/services/ai_insights_service.go`
- Package: `logs_services` (following existing convention)
- Struct: `AIInsightsService` with dependencies:
  - `aiClient` (AIProvider interface)
  - `logRepo` (LogRepository interface)
  - `insightsRepo` (AIInsightsRepository interface)
- Methods:
  - `NewAIInsightsService()` - Constructor with dependency injection
  - `GenerateInsights(ctx, logID, model)` - Main orchestration:
    1. Fetch log entry from database
    2. Build analysis prompt with log details
    3. Call AI model via provider interface
    4. Parse JSON response
    5. Upsert to database
  - `GetInsights(ctx, logID)` - Retrieve cached insights
  - `buildAnalysisPrompt(log)` - Format prompt for AI (private)
  - `parseAIResponse(content)` - Parse AI JSON response (private)

### 5. API Handlers ✅
**File:** `internal/logs/handlers/ai_insights_handler.go`
- Package: `internal_logs_handlers` (following existing convention)
- Struct: `AIInsightsHandler` with service dependency
- Endpoints:
  - `POST /api/logs/:id/insights` - Generate/regenerate insights
    - Request: `{"model": "deepseek-coder:6.7b"}`
    - Response: Complete AIInsight JSON
  - `GET /api/logs/:id/insights` - Fetch cached insights
    - Response: AIInsight JSON or 404 if not found
- Error handling for invalid log IDs, missing model, service failures

## Architectural Decisions

### Import Cycle Resolution
**Problem:** Repository needed AIInsight type, but importing services package created cycle.  
**Solution:** Moved AIInsight model to `internal/logs/models/ai_insight.go`
- Services import models (logs_models alias)
- Repository imports models (logs_models alias)
- No cycle, clean separation of concerns

### Package Naming Consistency
All new files follow existing conventions:
- `logs_services` (with underscore) for services package
- `logs_db` for database/repository package
- `internal_logs_handlers` for handlers package
- `logs_models` alias when importing models package

### Interface-Based Design
Service layer defines interfaces for dependencies:
- `AIProvider` - For calling AI models (will be implemented separately)
- `LogRepository` - For fetching log entries (already exists)
- `AIInsightsRepository` - For database operations (implemented)

This enables:
- Easy testing with mocks
- Swappable AI providers (Ollama, OpenAI, Claude, etc.)
- Clean dependency injection

## Testing Approach

### Interfaces to Mock
```go
// For testing AIInsightsService
type MockAIProvider struct {
    mock.Mock
}
func (m *MockAIProvider) Generate(ctx, req) (*AIResponse, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*AIResponse), args.Error(1)
}

type MockLogRepository struct {
    mock.Mock
}
// Similar mock implementation...

type MockAIInsightsRepository struct {
    mock.Mock
}
// Similar mock implementation...
```

### Test Cases to Add
1. **Service Layer Tests:**
   - GenerateInsights with valid log ID
   - GenerateInsights with invalid log ID (404)
   - GenerateInsights with AI failure (error handling)
   - GetInsights with cached data
   - GetInsights with no cached data (404)

2. **Repository Layer Tests:**
   - Upsert new insight
   - Upsert existing insight (updates, not duplicates)
   - GetByLogID with existing data
   - GetByLogID with no data (returns nil)

3. **Handler Layer Tests:**
   - POST with valid model parameter
   - POST with missing model parameter (400)
   - POST with invalid log ID (400)
   - GET with existing insights (200)
   - GET with no insights (404)

## Next Steps

### Step 1: Run Database Migration
```bash
# Connect to database container
docker exec -it devsmith-logs-db psql -U devsmith -d devsmith

# Run migration
\i /migrations/20251110_001_add_ai_insights.sql

# Verify table created
\d logs.ai_insights

# Check indexes
\di logs.ai_insights_*
```

### Step 2: Wire Handlers into Router
File: `internal/logs/logs.go` (or wherever routes are defined)

```go
// Add to initialization
aiInsightsRepo := logs_db.NewAIInsightsRepository(db)
// TODO: Create AI provider implementation (e.g., OllamaProvider)
aiInsightsService := logs_services.NewAIInsightsService(aiProvider, logRepo, aiInsightsRepo)
aiInsightsHandler := internal_logs_handlers.NewAIInsightsHandler(aiInsightsService)

// Add routes
logsGroup.POST("/logs/:id/insights", aiInsightsHandler.GenerateInsights)
logsGroup.GET("/logs/:id/insights", aiInsightsHandler.GetInsights)
```

### Step 3: Implement AI Provider
File: `internal/logs/services/ollama_provider.go` (new file needed)

```go
package logs_services

type OllamaProvider struct {
    baseURL string
    client  *http.Client
}

func NewOllamaProvider(baseURL string) *OllamaProvider {
    return &OllamaProvider{
        baseURL: baseURL,
        client:  &http.Client{Timeout: 60 * time.Second},
    }
}

func (p *OllamaProvider) Generate(ctx context.Context, req *AIRequest) (*AIResponse, error) {
    // Call Ollama API
    // POST {baseURL}/api/generate
    // Body: {"model": req.Model, "prompt": req.Prompt}
    // Parse response and return
}
```

### Step 4: Frontend Integration
File: `frontend/src/components/HealthPage.jsx`

Replace placeholder `generateAIInsights` function:

```javascript
const generateAIInsights = async (logId) => {
    try {
        setLoadingInsights(true);
        
        // Call backend API
        const response = await fetch(`/api/logs/${logId}/insights`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ model: selectedModel })
        });
        
        if (!response.ok) throw new Error('Failed to generate insights');
        
        const insights = await response.json();
        
        // Update selectedLog with real insights
        setSelectedLog(prev => ({
            ...prev,
            aiInsights: {
                analysis: insights.analysis,
                root_cause: insights.root_cause,
                suggestions: insights.suggestions,
                generated_at: insights.generated_at,
                model_used: insights.model_used
            }
        }));
        
        setLoadingInsights(false);
    } catch (error) {
        console.error('AI insights error:', error);
        setLoadingInsights(false);
        // Show error in UI
    }
};
```

Also add function to fetch existing insights when modal opens:

```javascript
const fetchExistingInsights = async (logId) => {
    try {
        const response = await fetch(`/api/logs/${logId}/insights`);
        if (response.ok) {
            const insights = await response.json();
            return insights;
        }
        return null; // No cached insights
    } catch (error) {
        console.error('Error fetching insights:', error);
        return null;
    }
};

// In handleCardClick, check for existing insights:
const handleCardClick = async (log) => {
    const existingInsights = await fetchExistingInsights(log.id);
    setSelectedLog({
        ...log,
        aiInsights: existingInsights || {
            analysis: "No AI insights yet. Click 'Generate AI Insights' below to analyze this log entry."
        }
    });
    setShowModal(true);
};
```

### Step 5: Test End-to-End
1. Open Health app
2. Click on a log card
3. Modal should show cached insights (if any)
4. Click "Generate AI Insights"
5. Should show loading state
6. Should display real AI analysis after ~5-10 seconds
7. Verify suggestions array displays correctly
8. Close and reopen modal - should show cached insights

## Files Modified/Created

### New Files (5 total)
1. `internal/logs/db/migrations/20251110_001_add_ai_insights.sql`
2. `internal/logs/models/ai_insight.go`
3. `internal/logs/db/ai_insights_repository.go`
4. `internal/logs/services/ai_insights_service.go`
5. `internal/logs/handlers/ai_insights_handler.go`

### Files to Modify
1. `internal/logs/logs.go` - Wire handlers into router
2. `frontend/src/components/HealthPage.jsx` - Replace placeholder
3. Need to create: `internal/logs/services/ollama_provider.go`

## Dependencies Needed

### For OllamaProvider
- Already have: `net/http`, `encoding/json`, `context`
- May need: Ollama API endpoint (default: http://localhost:11434)

### For Testing
- Already have: `github.com/stretchr/testify/mock`
- Already have: `context`, `testing`

## Completion Status

✅ **Task 2.1:** Database migration created  
✅ **Task 2.2:** AI insights service implemented  
✅ **Task 2.3:** Repository layer implemented  
✅ **Task 2.4:** API handlers implemented (partially - needs wiring)

⏸️ **Task 2.5:** Frontend integration (next step)  
⏸️ **Additional:** OllamaProvider implementation needed  
⏸️ **Additional:** Router wiring needed

## Estimated Remaining Time

- Step 1 (Run migration): 5 minutes
- Step 2 (Wire handlers): 10 minutes
- Step 3 (Implement OllamaProvider): 30 minutes
- Step 4 (Frontend integration): 20 minutes
- Step 5 (E2E testing): 15 minutes

**Total:** ~1.5 hours to complete Phase 2

## Ready for User Approval

The backend infrastructure is complete and ready for:
1. Database migration execution
2. AI provider implementation (Ollama)
3. Router configuration
4. Frontend integration
5. End-to-end testing

All code follows existing patterns, uses proper package naming, avoids import cycles, and is ready for production use after testing.
