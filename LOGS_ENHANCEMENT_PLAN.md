# Health App Enhancement Plan

**Document Version:** 2.0  
**Created:** 2025-11-09  
**Updated:** 2025-11-09  
**Status:** 📋 Ready for Implementation  
**Estimated Time:** 2-3 hours total

---

## 🎯 Executive Summary

Transform the **Logs app → Health app** with unified platform observability:
- **Three tabs**: Logs (enhanced), Monitoring (coming soon), Analytics (coming soon)
- **Logs tab**: Card-based layout, AI-powered insights, smart tagging
- **Architecture**: Single bounded context for all observability concerns
- **Rationale**: All platform health/debugging in one place, no context switching

---

## 🏗️ Architectural Changes

### App Rename: Logs → Health
**Bounded Context:** Platform Observability
- **Old:** "Logs" app with monitoring dashboard tab
- **New:** "Health" app with three distinct tabs
- **Navigation:** Portal → Health (replaces "Logs" card)
- **URL:** `/health` (currently `/logs`)
- **Service:** Logs service remains (internal name), but UI shows "Health"

### Three-Tab Structure

**Tab 1: Logs** (Implement now - this document)
- Real-time log stream with card-based layout
- Filtering, search, auto-refresh
- AI insights on-demand
- Smart tagging system

**Tab 2: Monitoring** (Coming soon placeholder)
- Service health status dashboard
- Request rates, error rates
- Resource utilization metrics
- Active incidents/alerts

**Tab 3: Analytics** (Coming soon placeholder)
- Historical log patterns
- Error frequency trends
- Service reliability scores
- AI-powered correlation analysis

### Portal Navigation Changes
**Remove:** Analytics card from dashboard  
**Update:** "Logs" card → "Health" card  
**Benefit:** Clearer mental model (Health = all observability)

---

## 📊 Current State Analysis

### What's Working
✅ Logs table displays data from `logs.entries`  
✅ Basic filtering by level/service/search  
✅ Auto-refresh functionality  
✅ Dark mode toggle in navbar  
✅ Model selector in navbar  

### What Needs Improvement
❌ App is called "Logs" instead of "Health"  
❌ Table rows are plain - need card-based layout  
❌ No AI insights available  
❌ No smart tagging system  
❌ Dark mode theming incomplete (modals, cards)  
❌ Model selector shows alphabetical, not default first  
❌ Monitoring and Analytics tabs show empty content (need "Coming Soon" placeholders)  

---

## 🚀 Implementation Phases

### **Phase 0: App Rename & Navigation**
**Priority:** CRITICAL (Do First)  
**Time:** 15-20 minutes  
**Status:** 🔴 Not Started

#### 0.1 Rename App in Portal Dashboard
**Goal:** Update Portal to show "Health" instead of "Logs"

**File:** `frontend/src/components/Dashboard.jsx`

**Changes:**
1. Update card title: "Logs" → "Health"
2. Update card description: "Real-time log monitoring and analysis" → "Platform health monitoring, logs, and analytics"
3. Update icon: Keep existing or change to `bi-heart-pulse-fill`
4. Update navigation URL: `/logs` → `/health`
5. Remove "Analytics" card entirely from dashboard

**Updated Card:**
```jsx
<div className="col-md-6 col-lg-3">
  <div className="card h-100" onClick={() => navigate('/health')}>
    <div className="card-body text-center">
      <div className="app-icon mb-3">
        <i className="bi bi-heart-pulse-fill" style={{ fontSize: '3rem', color: 'var(--bs-danger)' }}></i>
      </div>
      <h5 className="card-title">Health</h5>
      <p className="card-text text-muted small">
        Platform health monitoring, logs, and analytics
      </p>
      <span className="badge bg-success">Ready</span>
    </div>
  </div>
</div>
```

#### 0.2 Update Routing
**File:** `frontend/src/App.jsx`

**Changes:**
```jsx
// OLD:
<Route path="/logs" element={<LogsPage />} />

// NEW:
<Route path="/health" element={<HealthPage />} />
<Route path="/logs" element={<Navigate to="/health" replace />} /> {/* Redirect for backward compatibility */}
```

#### 0.3 Rename Component
**Action:** Rename `LogsPage.jsx` → `HealthPage.jsx`

**File:** `frontend/src/components/HealthPage.jsx` (renamed from LogsPage.jsx)

**Changes:**
1. Component name: `function LogsPage()` → `function HealthPage()`
2. Page title: `<h2>Logs</h2>` → `<h2>Health</h2>`
3. Export: `export default LogsPage` → `export default HealthPage`

#### 0.4 Add Tab Navigation
**File:** `frontend/src/components/HealthPage.jsx`

**New State:**
```jsx
const [activeTab, setActiveTab] = useState('logs'); // 'logs' | 'monitoring' | 'analytics'
```

**Tab UI (Add below page title):**
```jsx
{/* Tab Navigation */}
<ul className="nav nav-tabs mb-4">
  <li className="nav-item">
    <button
      className={`nav-link ${activeTab === 'logs' ? 'active' : ''}`}
      onClick={() => setActiveTab('logs')}
    >
      <i className="bi bi-list-ul me-2"></i>
      Logs
    </button>
  </li>
  <li className="nav-item">
    <button
      className={`nav-link ${activeTab === 'monitoring' ? 'active' : ''}`}
      onClick={() => setActiveTab('monitoring')}
    >
      <i className="bi bi-activity me-2"></i>
      Monitoring
    </button>
  </li>
  <li className="nav-item">
    <button
      className={`nav-link ${activeTab === 'analytics' ? 'active' : ''}`}
      onClick={() => setActiveTab('analytics')}
    >
      <i className="bi bi-graph-up me-2"></i>
      Analytics
    </button>
  </li>
</ul>

{/* Tab Content */}
{activeTab === 'logs' && (
  <div className="logs-tab">
    {/* All existing logs content goes here */}
  </div>
)}

{activeTab === 'monitoring' && (
  <div className="monitoring-tab text-center py-5">
    <i className="bi bi-activity" style={{ fontSize: '4rem', opacity: 0.3 }}></i>
    <h3 className="mt-3 text-muted">Monitoring Dashboard</h3>
    <p className="text-muted">Coming Soon</p>
    <p className="small text-muted">
      Real-time service health, request rates, and error tracking.
    </p>
  </div>
)}

{activeTab === 'analytics' && (
  <div className="analytics-tab text-center py-5">
    <i className="bi bi-graph-up" style={{ fontSize: '4rem', opacity: 0.3 }}></i>
    <h3 className="mt-3 text-muted">Analytics Dashboard</h3>
    <p className="text-muted">Coming Soon</p>
    <p className="small text-muted">
      Historical trends, error patterns, and AI-powered insights.
    </p>
  </div>
)}
```

#### 0.5 Remove MonitoringDashboard Component
**File:** `frontend/src/components/HealthPage.jsx`

**Remove import:**
```jsx
// DELETE THIS LINE:
import MonitoringDashboard from './MonitoringDashboard';
```

**Remove usage:** (already handled by tab structure above)

#### 0.6 Update Service Names (Optional)
**Goal:** Internal consistency (logs service can remain "logs" internally)

**Status:** 🔴 TODO - Deferred until service naming patterns are established

**Files to update (if time permits):**
- Backend: `cmd/logs/main.go` (add comment: "// Health app backend")
- Docker: `docker-compose.yml` (add label: "devsmith.app=health")
- No functional changes needed - just documentation

**Note:** This task is logged for future implementation once we determine the service naming conventions across the platform.

---

## 📋 Implementation Clarifications

### Testing Approach (Phase 0)
- **E2E tests**: Playwright tests for navigation changes, tab switching, and "Coming Soon" placeholders
- **Component tests**: React component tests for HealthPage tabs and routing
- **Coverage**: Both E2E and component tests required per TDD guidelines

### Backward Compatibility
- `/logs` URL will redirect to `/health` for bookmarks and deep links
- Existing API endpoints remain at `/api/logs/*` (backend internal naming)
- No breaking changes to backend services

### Git Workflow
- Current branch: `review-rebuild` (documentation only)
- After PR merge: Create `feature/phase0-health-rename` from `development`
- Follow TDD: RED → GREEN → REFACTOR cycle
- Use pre-push hook for validation before pushing

### Implementation Scope
Phase 0 includes tasks 0.1-0.5:
- ✅ 0.1: Rename "Logs" card to "Health" in Dashboard (remove Analytics card)
- ✅ 0.2: Update routing (`/logs` → `/health` with redirect)
- ✅ 0.3: Rename LogsPage.jsx → HealthPage.jsx
- ✅ 0.4: Add tab navigation (Logs/Monitoring/Analytics tabs)
- ✅ 0.5: Remove MonitoringDashboard import
- 🔴 0.6: Deferred - Service name updates (see TODO above)

---

### **Phase 1: Core UX Improvements (Logs Tab)** 
**Priority:** CRITICAL  
**Time:** 30-40 minutes  
**Status:** 🔴 Not Started

#### 1.1 Card-Based Layout
**Goal:** Replace plain table rows with modern card design

**Current Structure:**
```jsx
<table className="table table-hover">
  <tbody>
    <tr>
      <td>{level badge}</td>
      <td>{service}</td>
      <td>{timestamp}</td>
      <td>{message}</td>
    </tr>
  </tbody>
</table>
```

**New Structure:**
```jsx
<div className="log-cards-container">
  {filteredLogs.map(log => (
    <div 
      key={log.id}
      className="log-card"
      onClick={() => openDetailModal(log)}
    >
      <div className="log-card-row">
        <div className="log-card-col level">
          <span className={`badge bg-${getLevelColor(log.level)}`}>
            {log.level}
          </span>
        </div>
        <div className="log-card-col service">
          {log.service}
        </div>
        <div className="log-card-col timestamp">
          {formatTimestamp(log.created_at)}
        </div>
        <div className="log-card-col message">
          {log.message}
        </div>
        <div className="log-card-col tags">
          {log.tags?.map(tag => (
            <span className="badge bg-secondary me-1">{tag}</span>
          ))}
        </div>
      </div>
    </div>
  ))}
</div>
```

**CSS Requirements:**
```css
/* Card-based table layout */
.log-cards-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.log-card {
  background: var(--bs-card-bg);
  border: 1px solid var(--bs-border-color);
  border-radius: 0.5rem;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.log-card:hover {
  border-color: var(--bs-primary);
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.15);
  transform: translateY(-1px);
}

.log-card-row {
  display: grid;
  grid-template-columns: 100px 150px 180px 1fr 150px;
  gap: 1rem;
  align-items: center;
}

.log-card-col {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.log-card-col.message {
  white-space: normal;
  line-height: 1.4;
  max-height: 2.8em;
  overflow: hidden;
}

/* Dark mode specific */
[data-bs-theme="dark"] .log-card {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}

[data-bs-theme="dark"] .log-card:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(99, 102, 241, 0.5);
}
```

#### 1.2 Detail Modal
**Goal:** Show full log details with AI insights option

**Modal Structure:**
```jsx
<Modal show={showDetailModal} onHide={() => setShowDetailModal(false)} size="lg">
  <Modal.Header closeButton className={theme === 'dark' ? 'bg-dark text-light' : ''}>
    <Modal.Title>
      <span className={`badge bg-${getLevelColor(selectedLog.level)} me-2`}>
        {selectedLog.level}
      </span>
      Log Details
    </Modal.Title>
  </Modal.Header>
  <Modal.Body className={theme === 'dark' ? 'bg-dark text-light' : ''}>
    {/* Key Info Section */}
    <div className="row mb-3">
      <div className="col-md-6">
        <strong>Service:</strong> {selectedLog.service}
      </div>
      <div className="col-md-6">
        <strong>Timestamp:</strong> {selectedLog.created_at}
      </div>
    </div>

    {/* Message Section */}
    <div className="mb-3">
      <strong>Message:</strong>
      <div className="mt-2 p-3 rounded" style={{
        backgroundColor: theme === 'dark' ? 'rgba(0,0,0,0.3)' : 'rgba(0,0,0,0.05)',
        border: `1px solid ${theme === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)'}`
      }}>
        {selectedLog.message}
      </div>
    </div>

    {/* Metadata Section */}
    {selectedLog.metadata && (
      <div className="mb-3">
        <strong>Metadata:</strong>
        <pre className="mt-2 p-3 rounded" style={{
          backgroundColor: theme === 'dark' ? 'rgba(0,0,0,0.3)' : 'rgba(0,0,0,0.05)',
          color: theme === 'dark' ? 'var(--bs-gray-300)' : 'var(--bs-gray-800)',
          border: `1px solid ${theme === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)'}`
        }}>
          {JSON.stringify(selectedLog.metadata, null, 2)}
        </pre>
      </div>
    )}

    {/* Tags Section */}
    <div className="mb-3">
      <strong>Tags:</strong>
      <div className="mt-2">
        {selectedLog.tags?.map((tag, idx) => (
          <span key={idx} className="badge bg-secondary me-2">{tag}</span>
        ))}
      </div>
    </div>

    {/* AI Insights Section */}
    <div className="mb-3">
      <div className="d-flex justify-content-between align-items-center mb-2">
        <strong>AI Insights:</strong>
        <button
          className="btn btn-primary btn-sm"
          onClick={() => generateAIInsights(selectedLog.id)}
          disabled={loadingInsights}
        >
          {loadingInsights ? (
            <>
              <span className="spinner-border spinner-border-sm me-2"></span>
              Analyzing...
            </>
          ) : aiInsights ? (
            <>
              <i className="bi bi-arrow-clockwise me-2"></i>
              Regenerate
            </>
          ) : (
            <>
              <i className="bi bi-stars me-2"></i>
              Generate Insights
            </>
          )}
        </button>
      </div>
      {aiInsights && (
        <div className="p-3 rounded" style={{
          backgroundColor: theme === 'dark' ? 'rgba(99,102,241,0.1)' : 'rgba(99,102,241,0.05)',
          border: '1px solid rgba(99,102,241,0.3)'
        }}>
          <div className="mb-2">
            <strong>Analysis:</strong>
            <p className="mb-2">{aiInsights.analysis}</p>
          </div>
          {aiInsights.root_cause && (
            <div className="mb-2">
              <strong>Root Cause:</strong>
              <p className="mb-2">{aiInsights.root_cause}</p>
            </div>
          )}
          {aiInsights.suggestions && aiInsights.suggestions.length > 0 && (
            <div>
              <strong>Suggestions:</strong>
              <ul className="mb-0">
                {aiInsights.suggestions.map((suggestion, idx) => (
                  <li key={idx}>{suggestion}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  </Modal.Body>
  <Modal.Footer className={theme === 'dark' ? 'bg-dark border-secondary' : ''}>
    <button className="btn btn-secondary" onClick={() => setShowDetailModal(false)}>
      Close
    </button>
  </Modal.Footer>
</Modal>
```

#### 1.3 Fix Default Model Selection (GLOBAL)
**Bug:** ModelSelector shows models alphabetically instead of respecting `is_default: true`

**Affected Components:**
- `frontend/src/components/ModelSelector.jsx`
- Used in: ReviewPage, LogsPage, AnalyticsPage (if applicable)

**Current Code (BROKEN):**
```jsx
// ModelSelector.jsx - Around line 144
modelList = response.map(config => ({
  name: config.model_name,
  provider: config.provider,
  displayName: config.display_name || config.model_name,
  isDefault: config.is_default
}));
setModels(modelList); // Just sets in alphabetical order from API
```

**Fixed Code:**
```jsx
// Transform and sort by default first
modelList = response
  .map(config => ({
    name: config.model_name,
    provider: config.provider,
    displayName: config.display_name || config.model_name,
    isDefault: config.is_default
  }))
  .sort((a, b) => {
    // Default model comes first
    if (a.isDefault && !b.isDefault) return -1;
    if (!a.isDefault && b.isDefault) return 1;
    // Then alphabetically by provider
    return a.provider.localeCompare(b.provider);
  });

setModels(modelList);

// Auto-select default model
const defaultModel = modelList.find(m => m.isDefault);
if (defaultModel && !selectedModel) {
  console.log('Auto-selecting default model:', defaultModel.name);
  onModelSelect(defaultModel.name);
}
```

**Testing:**
1. Set Ollama DeepSeek as default in AI Factory
2. Open Review page → Should show Ollama selected
3. Open Logs page → Should show Ollama selected
4. Change default in AI Factory → All pages should update

---

### **Phase 2: AI Insights Integration**
**Priority:** HIGH  
**Time:** 20-30 minutes  
**Status:** 🔴 Not Started

#### 2.1 Backend API Endpoint
**File:** `cmd/logs/main.go` (add new route)

**Endpoint:** `POST /api/logs/:id/insights`

**Request Body:**
```json
{
  "model": "ollama/deepseek-coder-v2:16b"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 123,
    "log_id": 456,
    "analysis": "This error indicates a database connection timeout...",
    "root_cause": "Connection pool exhausted due to high traffic",
    "suggestions": [
      "Increase connection pool size in database config",
      "Add connection retry logic with exponential backoff",
      "Monitor connection pool metrics"
    ],
    "model_used": "ollama/deepseek-coder-v2:16b",
    "generated_at": "2025-11-09T16:30:00Z"
  }
}
```

#### 2.2 Database Schema
**Migration:** `internal/logs/db/migrations/20251109_001_add_ai_insights.sql`

```sql
-- AI Insights table
CREATE TABLE IF NOT EXISTS logs.ai_insights (
    id SERIAL PRIMARY KEY,
    log_id BIGINT NOT NULL REFERENCES logs.entries(id) ON DELETE CASCADE,
    analysis TEXT NOT NULL,
    root_cause TEXT,
    suggestions JSONB,
    model_used VARCHAR(255) NOT NULL,
    generated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(log_id) -- One insight per log (regenerate overwrites)
);

CREATE INDEX idx_ai_insights_log_id ON logs.ai_insights(log_id);
CREATE INDEX idx_ai_insights_generated_at ON logs.ai_insights(generated_at DESC);
```

#### 2.3 AI Service Implementation
**File:** `internal/logs/services/ai_insights_service.go` (new file)

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
)

type AIInsightsService struct {
    aiClient AIProvider
    repo     AIInsightsRepository
}

type AIInsight struct {
    ID          int       `json:"id"`
    LogID       int64     `json:"log_id"`
    Analysis    string    `json:"analysis"`
    RootCause   string    `json:"root_cause,omitempty"`
    Suggestions []string  `json:"suggestions,omitempty"`
    ModelUsed   string    `json:"model_used"`
    GeneratedAt time.Time `json:"generated_at"`
}

func (s *AIInsightsService) GenerateInsights(ctx context.Context, logID int64, model string) (*AIInsight, error) {
    // 1. Fetch log entry
    log, err := s.logRepo.GetByID(ctx, logID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch log: %w", err)
    }

    // 2. Build prompt
    prompt := s.buildAnalysisPrompt(log)

    // 3. Call AI
    response, err := s.aiClient.Generate(ctx, &AIRequest{
        Model:  model,
        Prompt: prompt,
    })
    if err != nil {
        return nil, fmt.Errorf("AI generation failed: %w", err)
    }

    // 4. Parse response
    insight, err := s.parseAIResponse(response.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse AI response: %w", err)
    }

    insight.LogID = logID
    insight.ModelUsed = model

    // 5. Save to database (upsert)
    savedInsight, err := s.repo.Upsert(ctx, insight)
    if err != nil {
        return nil, fmt.Errorf("failed to save insight: %w", err)
    }

    return savedInsight, nil
}

func (s *AIInsightsService) buildAnalysisPrompt(log *LogEntry) string {
    return fmt.Sprintf(`Analyze this log entry and provide insights:

Level: %s
Service: %s
Message: %s
Timestamp: %s
Metadata: %s

Please provide:
1. Analysis: What does this log indicate?
2. Root Cause: What likely caused this? (if it's an error/warning)
3. Suggestions: How to fix or prevent this? (3-5 actionable items)

Format your response as JSON:
{
  "analysis": "...",
  "root_cause": "...",
  "suggestions": ["...", "..."]
}`,
        log.Level,
        log.Service,
        log.Message,
        log.CreatedAt,
        log.Metadata,
    )
}
```

#### 2.4 Frontend Integration
**File:** `frontend/src/components/LogsPage.jsx`

**State Management:**
```jsx
const [aiInsights, setAIInsights] = useState(null);
const [loadingInsights, setLoadingInsights] = useState(false);

const generateAIInsights = async (logId) => {
  setLoadingInsights(true);
  setAIInsights(null);
  
  try {
    const response = await fetch(`/api/logs/${logId}/insights`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('devsmith_token')}`
      },
      body: JSON.stringify({
        model: selectedModel
      })
    });

    if (!response.ok) {
      throw new Error('Failed to generate insights');
    }

    const result = await response.json();
    setAIInsights(result.data);
  } catch (err) {
    console.error('AI insights error:', err);
    setError('Failed to generate AI insights');
  } finally {
    setLoadingInsights(false);
  }
};

// When modal opens, check if insights exist
const openDetailModal = (log) => {
  setSelectedLog(log);
  setShowDetailModal(true);
  setAIInsights(null);
  
  // Fetch cached insights if they exist
  fetchExistingInsights(log.id);
};

const fetchExistingInsights = async (logId) => {
  try {
    const response = await fetch(`/api/logs/${logId}/insights`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('devsmith_token')}`
      }
    });

    if (response.ok) {
      const result = await response.json();
      setAIInsights(result.data);
    }
  } catch (err) {
    // No cached insights - that's fine
    console.log('No cached insights available');
  }
};
```

---

### **Phase 3: Smart Tagging System**
**Priority:** MEDIUM  
**Time:** 20-25 minutes  
**Status:** 🔴 Not Started

#### 3.1 Auto-Tagging Logic
**Implementation:** Backend service that analyzes logs and adds tags

**Tag Sources:**

**A. Service-Based Tags (Automatic):**
- `log.service` → direct tag (e.g., "portal", "review", "logs")

**B. Level-Based Tags (Automatic):**
- `log.level` → "error", "warning", "info", "debug", "critical"

**C. Content-Based Tags (Automatic via keyword matching):**
```go
var tagKeywords = map[string][]string{
    "traefik":  {"traefik", "gateway", "routing", "proxy"},
    "docker":   {"docker", "container", "image", "build"},
    "frontend": {"react", "vite", "npm", "javascript", "jsx"},
    "backend":  {"golang", "gin", "api", "handler"},
    "database": {"postgres", "sql", "migration", "query"},
    "auth":     {"oauth", "jwt", "token", "login", "authentication"},
    "ai":       {"ollama", "anthropic", "openai", "claude", "model"},
}

func extractContentTags(message string, metadata map[string]interface{}) []string {
    lowerMsg := strings.ToLower(message)
    tags := []string{}
    
    for tag, keywords := range tagKeywords {
        for _, keyword := range keywords {
            if strings.Contains(lowerMsg, keyword) {
                tags = append(tags, tag)
                break
            }
        }
    }
    
    return tags
}
```

**D. Manual Tags (User-Added):**
- UI button: "Add Tag" in detail modal
- Allows custom tags like "investigated", "resolved", "needs-attention"

#### 3.2 Database Schema Update
**Migration:** `internal/logs/db/migrations/20251109_002_add_tags_column.sql`

```sql
-- Add tags column to logs.entries
ALTER TABLE logs.entries 
ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';

-- Index for tag filtering
CREATE INDEX IF NOT EXISTS idx_logs_entries_tags ON logs.entries USING GIN(tags);

-- Function to auto-generate tags (called on insert/update)
CREATE OR REPLACE FUNCTION logs.auto_generate_tags()
RETURNS TRIGGER AS $$
BEGIN
    -- Service tag
    NEW.tags := array_append(NEW.tags, NEW.service);
    
    -- Level tag
    NEW.tags := array_append(NEW.tags, lower(NEW.level));
    
    -- Content-based tags (simple keyword matching)
    IF NEW.message ~* 'traefik|gateway|routing|proxy' THEN
        NEW.tags := array_append(NEW.tags, 'traefik');
    END IF;
    
    IF NEW.message ~* 'docker|container|image|build' THEN
        NEW.tags := array_append(NEW.tags, 'docker');
    END IF;
    
    IF NEW.message ~* 'react|vite|npm|javascript|jsx' THEN
        NEW.tags := array_append(NEW.tags, 'frontend');
    END IF;
    
    IF NEW.message ~* 'golang|gin|api|handler' THEN
        NEW.tags := array_append(NEW.tags, 'backend');
    END IF;
    
    IF NEW.message ~* 'postgres|sql|migration|query' THEN
        NEW.tags := array_append(NEW.tags, 'database');
    END IF;
    
    IF NEW.message ~* 'oauth|jwt|token|login|authentication' THEN
        NEW.tags := array_append(NEW.tags, 'auth');
    END IF;
    
    IF NEW.message ~* 'ollama|anthropic|openai|claude|model' THEN
        NEW.tags := array_append(NEW.tags, 'ai');
    END IF;
    
    -- Remove duplicates
    NEW.tags := array(SELECT DISTINCT unnest(NEW.tags));
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for auto-tagging
DROP TRIGGER IF EXISTS trigger_auto_generate_tags ON logs.entries;
CREATE TRIGGER trigger_auto_generate_tags
    BEFORE INSERT OR UPDATE ON logs.entries
    FOR EACH ROW
    EXECUTE FUNCTION logs.auto_generate_tags();
```

#### 3.3 Tag Filtering UI
**Component:** `TagFilter.jsx` (new component)

```jsx
function TagFilter({ availableTags, selectedTags, onTagToggle }) {
  const { isDarkMode } = useTheme();
  
  return (
    <div className="mb-3">
      <label className="form-label">Filter by Tags:</label>
      <div className="d-flex flex-wrap gap-2">
        {availableTags.map(tag => {
          const isSelected = selectedTags.includes(tag);
          return (
            <button
              key={tag}
              className={`btn btn-sm ${
                isSelected 
                  ? 'btn-primary' 
                  : isDarkMode 
                    ? 'btn-outline-light' 
                    : 'btn-outline-secondary'
              }`}
              onClick={() => onTagToggle(tag)}
            >
              {isSelected && <i className="bi bi-check-circle-fill me-1"></i>}
              {tag}
              <span className="badge bg-secondary ms-2">
                {getTagCount(tag)}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
```

**Integration in LogsPage:**
```jsx
const [selectedTags, setSelectedTags] = useState([]);
const [availableTags, setAvailableTags] = useState([]);

// Fetch available tags
useEffect(() => {
  const tags = new Set();
  logs.forEach(log => {
    log.tags?.forEach(tag => tags.add(tag));
  });
  setAvailableTags(Array.from(tags).sort());
}, [logs]);

// Apply tag filtering
useEffect(() => {
  let filtered = logs;
  
  // Existing filters (level, service, search)
  if (filters.level !== 'all') {
    filtered = filtered.filter(log => log.level === filters.level);
  }
  if (filters.service !== 'all') {
    filtered = filtered.filter(log => log.service === filters.service);
  }
  if (filters.search) {
    filtered = filtered.filter(log =>
      log.message.toLowerCase().includes(filters.search.toLowerCase())
    );
  }
  
  // NEW: Tag filtering (OR logic - show logs with ANY selected tag)
  if (selectedTags.length > 0) {
    filtered = filtered.filter(log =>
      log.tags?.some(tag => selectedTags.includes(tag))
    );
  }
  
  setFilteredLogs(filtered);
}, [logs, filters, selectedTags]);

const handleTagToggle = (tag) => {
  setSelectedTags(prev =>
    prev.includes(tag)
      ? prev.filter(t => t !== tag)
      : [...prev, tag]
  );
};
```

---

## 🧪 Testing Checklist

### Phase 0 Testing (App Rename)
- [ ] Portal dashboard shows "Health" card (not "Logs")
- [ ] Analytics card removed from dashboard
- [ ] Health card description mentions "logs and analytics"
- [ ] Clicking Health card navigates to `/health`
- [ ] Old `/logs` URL redirects to `/health`
- [ ] Health page shows three tabs: Logs, Monitoring, Analytics
- [ ] Tab navigation works correctly
- [ ] Logs tab shows enhanced content (after Phase 1)
- [ ] Monitoring tab shows "Coming Soon" placeholder
- [ ] Analytics tab shows "Coming Soon" placeholder
- [ ] No console errors on tab switching
- [ ] Dark mode works on all tabs

### Phase 1 Testing (Logs Tab)
- [ ] Card layout displays correctly in light mode
- [ ] Card layout displays correctly in dark mode
- [ ] Cards have hover effect
- [ ] Clicking card opens detail modal
- [ ] Modal displays all log information
- [ ] Modal is themed correctly (dark/light)
- [ ] Default model is selected first in dropdown (not alphabetical)
- [ ] Default model works on Review page
- [ ] Default model works on Health page
- [ ] Changing default in AI Factory updates all pages

### Phase 2 Testing (AI Insights)
- [ ] "Generate Insights" button appears in modal
- [ ] Button shows loading state while processing
- [ ] AI insights display correctly after generation
- [ ] Insights are cached in database
- [ ] Cached insights load when reopening modal
- [ ] "Regenerate" button works
- [ ] Different models produce different insights
- [ ] Error handling works (AI service down, invalid response)

### Phase 3 Testing (Smart Tags)
- [ ] Tags auto-generate on log creation
- [ ] Service tag appears (e.g., "portal")
- [ ] Level tag appears (e.g., "error")
- [ ] Content tags appear (e.g., "docker", "traefik")
- [ ] Tag filter UI displays all unique tags
- [ ] Clicking tag filters logs correctly
- [ ] Multiple tag selection works (OR logic)
- [ ] Tag count displays correctly
- [ ] Manual tag addition works (if implemented)
- [ ] Existing logs get tags retroactively (migration)

---

## 📁 Files to Modify/Create

### Phase 0: App Rename
**Modified Files:**
- `frontend/src/components/Dashboard.jsx` (rename "Logs" → "Health", remove "Analytics" card)
- `frontend/src/App.jsx` (update routes: `/logs` → `/health`)
- `frontend/src/components/LogsPage.jsx` → **RENAME TO** `HealthPage.jsx` (component name, page title, tab structure)

**Deleted Files:**
- `frontend/src/components/MonitoringDashboard.jsx` (replaced by inline "Coming Soon" in tabs)

### Phase 1-3: Logs Tab Enhancements
**Backend - New Files:**
- `internal/logs/services/ai_insights_service.go`
- `internal/logs/repositories/ai_insights_repository.go`
- `internal/logs/db/migrations/20251109_001_add_ai_insights.sql`
- `internal/logs/db/migrations/20251109_002_add_tags_column.sql`

**Backend - Modified Files:**
- `cmd/logs/main.go` (add AI insights routes)
- `internal/logs/handlers/log_handler.go` (add insights endpoint)

**Frontend - New Files:**
- `frontend/src/components/TagFilter.jsx`
- `frontend/src/components/LogDetailModal.jsx` (extract from HealthPage)

**Frontend - Modified Files:**
- `frontend/src/components/HealthPage.jsx` (card layout, modal, tag filtering, tabs)
- `frontend/src/components/ModelSelector.jsx` (fix default selection bug)

---

## 🎯 Success Criteria

### User Experience
✅ "Health" app clearly communicates unified observability  
✅ Three tabs provide logical organization (Logs, Monitoring, Analytics)  
✅ No context switching needed for debugging workflows  
✅ Logs are easy to scan visually (card layout)  
✅ Detailed information is one click away (modal)  
✅ AI provides actionable insights on demand  
✅ Tags make filtering intuitive and fast  
✅ Dark mode is fully functional everywhere  
✅ Default model selection works correctly  

### Performance
✅ Tab switching is instant (<100ms)  
✅ Card rendering is smooth (virtual scrolling if >100 logs)  
✅ AI insights generate in <5 seconds  
✅ Tag filtering updates instantly  
✅ Modal opens without lag  

### Data Quality
✅ Tags are accurate and helpful  
✅ AI insights are relevant and actionable  
✅ All log metadata is preserved and accessible  

### Architecture
✅ Single bounded context for observability  
✅ Logs service backend remains unchanged (internal name)  
✅ Frontend clearly shows "Health" branding  
✅ Easy to add future tabs (Traces, Alerts, Incidents)  

---

## 🚀 Next Steps

1. **Review this document** in the editor
2. **Start new chat** with context:
   - "Implement Phase 0 of LOGS_ENHANCEMENT_PLAN.md (Health app rename)"
   - Or "Implement Phase 1 of LOGS_ENHANCEMENT_PLAN.md (Card layout)"
   - Or "Implement all phases of LOGS_ENHANCEMENT_PLAN.md"

3. **Implementation order:**
   - **Phase 0** (App rename and tab structure) - MUST DO FIRST
   - **Phase 1** (Core UX - card layout, modal, default model fix)
   - **Phase 2** (AI insights)
   - **Phase 3** (Smart tags)

4. **After Phase 0:** User should see:
   - Portal dashboard with "Health" card (no "Analytics" card)
   - Health page with three tabs
   - Monitoring and Analytics tabs showing "Coming Soon"
   - Logs tab showing existing functionality (table view until Phase 1)

---

## 📝 Notes

- **App Name**: "Health" (user-facing), "Logs" (internal service name)
- **URL Change**: `/logs` → `/health` (with redirect for backward compatibility)
- **Bounded Context**: All observability concerns unified
- **Portal Change**: Remove "Analytics" card, rename "Logs" → "Health"
- **Tab Structure**: Logs (enhanced), Monitoring (coming soon), Analytics (coming soon)
- **AI Model**: Use selected model from navbar dropdown
- **Caching**: AI insights stored in `logs.ai_insights` table
- **Tags**: Mix of auto-generated and user-added
- **Dark Mode**: All new components must support theme switching
- **Default Model Bug**: Fix globally affects Review, Health pages
- **Future Tabs**: Easy to add Traces, Alerts, Incidents as platform grows

---

**Document Status:** ✅ Complete - Ready for Implementation  
**Architecture:** Health App (3 tabs: Logs, Monitoring, Analytics)  
**Next Action:** Review in editor, then implement Phase 0 (rename) first
