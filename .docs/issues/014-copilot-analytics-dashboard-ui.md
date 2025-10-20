# Issue #014: [COPILOT] Analytics Service - Dashboard UI

**Type:** Feature (Copilot Implementation)
**Service:** Analytics
**Depends On:** Issue #011 (Analytics Service Foundation), Issue #009 (Logs Service)
**Estimated Duration:** 60-75 minutes

---

## Summary

Create the Analytics Service web UI that visualizes trends, anomalies, and top issues from log data. This dashboard displays statistical insights through charts and tables, allowing developers to understand development patterns over time.

**User Story:**
> As a developer, I want to view analytics about my development activity, including trends in log levels, anomaly detection, and top recurring issues, so I can identify patterns and areas for improvement.

---

## Bounded Context

**Analytics Service Context:**
- **Responsibility:** Statistical analysis and visualization of log data
- **Does NOT:** Store raw logs (Logs service does that), perform code review (Review does that)
- **Boundaries:** Analytics reads from `logs.entries` (READ-ONLY cross-schema access)

**Why This Matters:**
- Analytics aggregates data but doesn't own raw logs
- Cross-schema access is read-only (no writes to logs.entries)
- Analytics computes hourly aggregations stored in `analytics.hourly_aggregates`

---

## Success Criteria

### Must Have (MVP)
- [ ] Landing page at `/` shows analytics dashboard
- [ ] Time range selector (Last 24h, 7d, 30d)
- [ ] Trend charts for log levels over time (INFO, WARN, ERROR)
- [ ] Anomaly detection section showing detected anomalies
- [ ] Top issues table (most frequent errors/warnings)
- [ ] Export data as CSV or JSON
- [ ] Responsive design (desktop and tablet)
- [ ] Loading states for data fetching
- [ ] Error handling with user-friendly messages

### Nice to Have (Post-MVP)
- Real-time updates via WebSocket
- Custom date range picker
- Filtering by service or log level
- Drill-down into specific anomalies

---

## Database Schema

**Uses existing schemas from Issue #011.**

Reads from:
- `analytics.hourly_aggregates` (trend data)
- `analytics.anomalies` (detected anomalies)
- `analytics.top_issues` (most frequent issues)

---

## API Endpoints (Existing - Created in Issue #011)

### GET `/api/v1/analytics/trends`
Query parameters:
- `time_range`: `24h`, `7d`, `30d` (default: `24h`)
- `metric`: `log_count`, `error_rate`, `warn_rate` (default: `log_count`)

Returns hourly aggregates for the specified time range.

### GET `/api/v1/analytics/anomalies`
Query parameters:
- `time_range`: `24h`, `7d`, `30d` (default: `24h`)
- `severity`: `high`, `medium`, `low` (optional)

Returns detected anomalies with timestamps and descriptions.

### GET `/api/v1/analytics/top-issues`
Query parameters:
- `time_range`: `24h`, `7d`, `30d` (default: `24h`)
- `level`: `error`, `warn`, `all` (default: `all`)
- `limit`: number (default: `10`)

Returns most frequent log messages grouped by similarity.

### GET `/api/v1/analytics/export`
Query parameters:
- `format`: `csv`, `json` (required)
- `time_range`: `24h`, `7d`, `30d` (default: `24h`)

Returns analytics data in specified format for download.

---

## File Structure

```
apps/analytics/
├── templates/
│   ├── layout.templ              # NEW - Base layout
│   ├── dashboard.templ           # NEW - Main dashboard
│   └── components/
│       ├── trend_chart.templ     # NEW - Trend visualization
│       ├── anomaly_card.templ    # NEW - Anomaly display
│       └── issues_table.templ    # NEW - Top issues table
├── handlers/
│   ├── ui_handler.go             # NEW - UI route handlers
│   └── [existing API handlers]   # From Issue #011
└── static/
    ├── css/
    │   └── analytics.css         # NEW - Dashboard styles
    └── js/
        ├── analytics.js          # NEW - Main dashboard logic
        └── charts.js             # NEW - Chart rendering (Chart.js)

cmd/analytics/
└── main.go                       # UPDATE - Add UI routes
```

---

## Implementation Details

### 1. Dashboard Template

**File:** `apps/analytics/templates/dashboard.templ`

```go
package templates

templ Dashboard() {
	@Layout("DevSmith Analytics") {
		<div class="analytics-container">
			<header class="analytics-header">
				<h1>📊 DevSmith Analytics</h1>
				<p class="subtitle">Development insights and trends</p>

				<div class="controls">
					@TimeRangeSelector()
					@ExportButton()
				</div>
			</header>

			<main class="analytics-main">
				<div class="dashboard-grid">
					@TrendsSection()
					@AnomaliesSection()
					@TopIssuesSection()
				</div>
			</main>
		</div>
	}
}

templ TimeRangeSelector() {
	<div class="time-range-selector">
		<label for="time-range">Time Range:</label>
		<select id="time-range" name="time_range">
			<option value="24h" selected>Last 24 Hours</option>
			<option value="7d">Last 7 Days</option>
			<option value="30d">Last 30 Days</option>
		</select>
	</div>
}

templ ExportButton() {
	<div class="export-controls">
		<button id="export-csv" class="btn-export">Export CSV</button>
		<button id="export-json" class="btn-export">Export JSON</button>
	</div>
}

templ TrendsSection() {
	<section class="trends-section card">
		<h2>Log Trends</h2>
		<div class="chart-container">
			<canvas id="trends-chart"></canvas>
		</div>
		<div id="trends-loading" class="loading">Loading trends...</div>
	</section>
}

templ AnomaliesSection() {
	<section class="anomalies-section card">
		<h2>Detected Anomalies</h2>
		<div id="anomalies-container">
			<div class="loading">Loading anomalies...</div>
		</div>
	</section>
}

templ TopIssuesSection() {
	<section class="issues-section card">
		<h2>Top Issues</h2>
		<div class="issues-filters">
			<select id="issues-level">
				<option value="all">All Levels</option>
				<option value="error">Errors Only</option>
				<option value="warn">Warnings Only</option>
			</select>
		</div>
		<div id="issues-container">
			<div class="loading">Loading top issues...</div>
		</div>
	</section>
}
```

---

### 2. Anomaly Card Component

**File:** `apps/analytics/templates/components/anomaly_card.templ`

```go
package components

templ AnomalyCard(anomaly Anomaly) {
	<div class={"anomaly-card " + anomaly.Severity}>
		<div class="anomaly-header">
			<span class={"severity-badge " + anomaly.Severity}>{anomaly.Severity}</span>
			<span class="anomaly-time">{anomaly.DetectedAt}</span>
		</div>
		<div class="anomaly-content">
			<h4>{anomaly.Metric}</h4>
			<p class="anomaly-description">{anomaly.Description}</p>
			<div class="anomaly-stats">
				<span>Expected: {fmt.Sprintf("%.2f", anomaly.ExpectedValue)}</span>
				<span>Actual: {fmt.Sprintf("%.2f", anomaly.ActualValue)}</span>
				<span>Deviation: {fmt.Sprintf("%.1f", anomaly.Deviation)}σ</span>
			</div>
		</div>
	</div>
}

type Anomaly struct {
	Severity      string
	DetectedAt    string
	Metric        string
	Description   string
	ExpectedValue float64
	ActualValue   float64
	Deviation     float64
}
```

---

### 3. Top Issues Table Component

**File:** `apps/analytics/templates/components/issues_table.templ`

```go
package components

templ IssuesTable(issues []Issue) {
	<table class="issues-table">
		<thead>
			<tr>
				<th>Rank</th>
				<th>Level</th>
				<th>Message Pattern</th>
				<th>Count</th>
				<th>First Seen</th>
				<th>Last Seen</th>
			</tr>
		</thead>
		<tbody>
			for i, issue := range issues {
				<tr>
					<td class="rank">{fmt.Sprintf("#%d", i+1)}</td>
					<td><span class={"level-badge " + issue.Level}>{issue.Level}</span></td>
					<td class="message">{issue.MessagePattern}</td>
					<td class="count">{fmt.Sprintf("%d", issue.Count)}</td>
					<td class="timestamp">{issue.FirstSeen}</td>
					<td class="timestamp">{issue.LastSeen}</td>
				</tr>
			}
		</tbody>
	</table>
}

type Issue struct {
	Level          string
	MessagePattern string
	Count          int
	FirstSeen      string
	LastSeen       string
}
```

---

### 4. UI Handler

**File:** `apps/analytics/handlers/ui_handler.go`

```go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"devsmith/apps/analytics/templates"
)

// DashboardHandler serves the main Analytics dashboard
func DashboardHandler(c *gin.Context) {
	component := templates.Dashboard()
	component.Render(c.Request.Context(), c.Writer)
}
```

---

### 5. Main JavaScript (Chart Rendering)

**File:** `apps/analytics/static/js/analytics.js`

```javascript
let currentTimeRange = '24h';
let trendsChart = null;

// Initialize dashboard
document.addEventListener('DOMContentLoaded', () => {
  loadTrends();
  loadAnomalies();
  loadTopIssues();
});

// Time range selector
document.getElementById('time-range').addEventListener('change', (e) => {
  currentTimeRange = e.target.value;
  loadTrends();
  loadAnomalies();
  loadTopIssues();
});

// Issues level filter
document.getElementById('issues-level').addEventListener('change', () => {
  loadTopIssues();
});

// Export buttons
document.getElementById('export-csv').addEventListener('click', () => {
  exportData('csv');
});

document.getElementById('export-json').addEventListener('click', () => {
  exportData('json');
});

// Load trends data and render chart
async function loadTrends() {
  try {
    const response = await fetch(`/api/v1/analytics/trends?time_range=${currentTimeRange}`);
    const data = await response.json();

    renderTrendsChart(data);
    document.getElementById('trends-loading').style.display = 'none';

  } catch (error) {
    console.error('Failed to load trends:', error);
    showError('trends-chart', 'Failed to load trend data');
  }
}

// Load anomalies
async function loadAnomalies() {
  try {
    const response = await fetch(`/api/v1/analytics/anomalies?time_range=${currentTimeRange}`);
    const anomalies = await response.json();

    renderAnomalies(anomalies);

  } catch (error) {
    console.error('Failed to load anomalies:', error);
    showError('anomalies-container', 'Failed to load anomalies');
  }
}

// Load top issues
async function loadTopIssues() {
  try {
    const level = document.getElementById('issues-level').value;
    const response = await fetch(
      `/api/v1/analytics/top-issues?time_range=${currentTimeRange}&level=${level}&limit=10`
    );
    const issues = await response.json();

    renderTopIssues(issues);

  } catch (error) {
    console.error('Failed to load top issues:', error);
    showError('issues-container', 'Failed to load top issues');
  }
}

// Render trends chart using Chart.js
function renderTrendsChart(data) {
  const ctx = document.getElementById('trends-chart').getContext('2d');

  if (trendsChart) {
    trendsChart.destroy();
  }

  trendsChart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: data.timestamps,
      datasets: [
        {
          label: 'INFO',
          data: data.info_counts,
          borderColor: '#0366d6',
          backgroundColor: 'rgba(3, 102, 214, 0.1)',
          tension: 0.3,
        },
        {
          label: 'WARN',
          data: data.warn_counts,
          borderColor: '#ffc107',
          backgroundColor: 'rgba(255, 193, 7, 0.1)',
          tension: 0.3,
        },
        {
          label: 'ERROR',
          data: data.error_counts,
          borderColor: '#dc3545',
          backgroundColor: 'rgba(220, 53, 69, 0.1)',
          tension: 0.3,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'top',
        },
        title: {
          display: true,
          text: `Log Trends (${currentTimeRange})`,
        },
      },
      scales: {
        y: {
          beginAtZero: true,
          title: {
            display: true,
            text: 'Log Count',
          },
        },
        x: {
          title: {
            display: true,
            text: 'Time',
          },
        },
      },
    },
  });
}

// Render anomalies
function renderAnomalies(anomalies) {
  const container = document.getElementById('anomalies-container');

  if (anomalies.length === 0) {
    container.innerHTML = '<p class="no-data">No anomalies detected in this time range.</p>';
    return;
  }

  container.innerHTML = anomalies.map(anomaly => `
    <div class="anomaly-card ${anomaly.severity}">
      <div class="anomaly-header">
        <span class="severity-badge ${anomaly.severity}">${anomaly.severity}</span>
        <span class="anomaly-time">${formatTimestamp(anomaly.detected_at)}</span>
      </div>
      <div class="anomaly-content">
        <h4>${anomaly.metric}</h4>
        <p class="anomaly-description">${anomaly.description}</p>
        <div class="anomaly-stats">
          <span>Expected: ${anomaly.expected_value.toFixed(2)}</span>
          <span>Actual: ${anomaly.actual_value.toFixed(2)}</span>
          <span>Deviation: ${anomaly.deviation.toFixed(1)}σ</span>
        </div>
      </div>
    </div>
  `).join('');
}

// Render top issues table
function renderTopIssues(issues) {
  const container = document.getElementById('issues-container');

  if (issues.length === 0) {
    container.innerHTML = '<p class="no-data">No issues found in this time range.</p>';
    return;
  }

  container.innerHTML = `
    <table class="issues-table">
      <thead>
        <tr>
          <th>Rank</th>
          <th>Level</th>
          <th>Message Pattern</th>
          <th>Count</th>
          <th>First Seen</th>
          <th>Last Seen</th>
        </tr>
      </thead>
      <tbody>
        ${issues.map((issue, i) => `
          <tr>
            <td class="rank">#${i + 1}</td>
            <td><span class="level-badge ${issue.level}">${issue.level}</span></td>
            <td class="message">${issue.message_pattern}</td>
            <td class="count">${issue.count}</td>
            <td class="timestamp">${formatTimestamp(issue.first_seen)}</td>
            <td class="timestamp">${formatTimestamp(issue.last_seen)}</td>
          </tr>
        `).join('')}
      </tbody>
    </table>
  `;
}

// Export data
async function exportData(format) {
  try {
    const response = await fetch(
      `/api/v1/analytics/export?format=${format}&time_range=${currentTimeRange}`
    );

    if (!response.ok) throw new Error('Export failed');

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `analytics-${currentTimeRange}.${format}`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(url);

  } catch (error) {
    console.error('Export failed:', error);
    alert('Failed to export data');
  }
}

function formatTimestamp(ts) {
  return new Date(ts).toLocaleString();
}

function showError(containerId, message) {
  document.getElementById(containerId).innerHTML =
    `<p class="error-message">${message}</p>`;
}
```

---

### 6. CSS Styles

**File:** `apps/analytics/static/css/analytics.css`

```css
:root {
  --primary: #0366d6;
  --success: #28a745;
  --warning: #ffc107;
  --danger: #dc3545;
  --info: #17a2b8;
  --bg-light: #f6f8fa;
  --border: #e1e4e8;
  --text-primary: #24292e;
  --text-secondary: #586069;
}

.analytics-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem;
}

.analytics-header {
  text-align: center;
  margin-bottom: 2rem;
}

.analytics-header h1 {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.subtitle {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
}

.controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.time-range-selector select,
.issues-filters select {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 1rem;
}

.btn-export {
  padding: 0.5rem 1rem;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-export:hover {
  background: #0256c7;
}

/* Dashboard Grid */
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 1.5rem;
}

.card {
  background: white;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
}

.card h2 {
  margin: 0 0 1rem 0;
  font-size: 1.25rem;
  color: var(--text-primary);
}

/* Trends Chart */
.trends-section {
  grid-column: 1 / -1;
}

.chart-container {
  position: relative;
  height: 400px;
}

/* Anomalies */
.anomalies-section {
  grid-column: 1 / -1;
}

#anomalies-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.anomaly-card {
  border: 2px solid;
  border-radius: 6px;
  padding: 1rem;
}

.anomaly-card.high { border-color: var(--danger); background: #fff5f5; }
.anomaly-card.medium { border-color: var(--warning); background: #fffbf0; }
.anomaly-card.low { border-color: var(--info); background: #f0f8ff; }

.anomaly-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.severity-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.severity-badge.high { background: var(--danger); color: white; }
.severity-badge.medium { background: var(--warning); color: #333; }
.severity-badge.low { background: var(--info); color: white; }

.anomaly-time {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.anomaly-description {
  margin: 0.5rem 0;
  font-size: 0.875rem;
}

.anomaly-stats {
  display: flex;
  gap: 1rem;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

/* Top Issues Table */
.issues-section {
  grid-column: 1 / -1;
}

.issues-filters {
  margin-bottom: 1rem;
}

.issues-table {
  width: 100%;
  border-collapse: collapse;
}

.issues-table th,
.issues-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

.issues-table th {
  background: var(--bg-light);
  font-weight: 600;
  color: var(--text-primary);
}

.issues-table .rank {
  font-weight: 600;
  color: var(--text-secondary);
}

.level-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.level-badge.error { background: #ffd6d9; color: #d73a49; }
.level-badge.warn { background: #fff5b1; color: #b08800; }
.level-badge.info { background: #d1ecf1; color: #0c5460; }

.issues-table .message {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.issues-table .count {
  font-weight: 600;
  color: var(--primary);
}

.issues-table .timestamp {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

/* Loading & Error States */
.loading {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.no-data {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
  font-style: italic;
}

.error-message {
  text-align: center;
  padding: 2rem;
  color: var(--danger);
}

@media (max-width: 768px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }

  .chart-container {
    height: 300px;
  }

  .issues-table {
    font-size: 0.75rem;
  }

  .issues-table .message {
    max-width: 200px;
  }
}
```

---

## TDD Workflow

### TDD Workflow for This Issue

**Step 1: RED PHASE (Write Failing Tests) - DO THIS FIRST!**

Create test files BEFORE implementation:

```go
// apps/analytics/handlers/ui_handler_test.go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDashboardHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "DevSmith Analytics")
	assert.Contains(t, w.Body.String(), "trends-chart")
	assert.Contains(t, w.Body.String(), "anomalies-container")
}

// apps/analytics/static/js/analytics.test.js (if using Jest)
describe('Analytics Dashboard', () => {
  test('loadTrends fetches and renders chart', async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        json: () => Promise.resolve({
          timestamps: ['10:00', '11:00'],
          info_counts: [10, 15],
          warn_counts: [2, 3],
          error_counts: [1, 0],
        }),
      })
    );

    await loadTrends();
    expect(fetch).toHaveBeenCalledWith('/api/v1/analytics/trends?time_range=24h');
    expect(document.getElementById('trends-chart')).toBeDefined();
  });

  test('filterAnomaliesBySeverity works', () => {
    const anomalies = [
      { severity: 'high', metric: 'error_rate' },
      { severity: 'low', metric: 'log_count' },
    ];

    const filtered = anomalies.filter(a => a.severity === 'high');
    expect(filtered).toHaveLength(1);
    expect(filtered[0].metric).toBe('error_rate');
  });

  test('exportData triggers download', async () => {
    const createElementSpy = jest.spyOn(document, 'createElement');

    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        blob: () => Promise.resolve(new Blob(['test data'])),
      })
    );

    await exportData('csv');

    expect(fetch).toHaveBeenCalledWith('/api/v1/analytics/export?format=csv&time_range=24h');
    expect(createElementSpy).toHaveBeenCalledWith('a');
  });
});
```

**Run tests (should FAIL):**
```bash
go test ./apps/analytics/handlers/...
# Expected: FAIL - DashboardHandler undefined

# For JavaScript (if using Jest):
npm test -- analytics.test.js
# Expected: FAIL - loadTrends is not defined
```

**Commit failing tests:**
```bash
git add apps/analytics/handlers/ui_handler_test.go
git add apps/analytics/static/js/analytics.test.js
git commit -m "test(analytics): add failing tests for dashboard UI (RED phase)"
```

**Step 2: GREEN PHASE - Implement to Pass Tests**

Now implement the templates, handlers, and JavaScript. See Implementation section above.

**After implementation, run tests:**
```bash
go test ./apps/analytics/...
# Expected: PASS

npm test
# Expected: PASS
```

**Step 3: Verify Build**
```bash
templ generate apps/analytics/templates/*.templ
go build -o /dev/null ./cmd/analytics
```

**Step 4: Manual Testing**

Follow the manual testing checklist below.

**Step 5: Commit Implementation**
```bash
git add apps/analytics/
git commit -m "feat(analytics): implement dashboard UI with trends, anomalies, and exports (GREEN phase)"
```

**Step 6: REFACTOR PHASE (Optional)**

If needed, refactor for:
- Chart rendering performance (virtualization for large datasets)
- Reusable chart components
- Error handling improvements
- Accessibility (ARIA labels, keyboard navigation)

**Commit refactors:**
```bash
git add apps/analytics/
git commit -m "refactor(analytics): improve chart performance and accessibility"
```

**Reference:** DevsmithTDD.md lines 15-36, 38-86 (RED-GREEN-REFACTOR)

**Key TDD Principles for UI:**
1. **Test HTML structure exists** (dashboard container, chart canvas, buttons)
2. **Test event handlers work** (time range change, export buttons)
3. **Test data fetching** (API calls with correct parameters)
4. **Test rendering logic** (anomalies display correctly, table populates)
5. **Test error states** (failed fetch shows error message)

**Coverage Target:** 70%+ for Go handlers, 60%+ for JavaScript logic

---

## Testing Requirements

### Manual Testing Checklist

- [ ] Navigate to `http://localhost:8083/`
- [ ] Verify trends chart displays with 3 lines (INFO, WARN, ERROR)
- [ ] Change time range selector (24h → 7d → 30d)
- [ ] Verify chart updates with new data
- [ ] Check anomalies section shows detected anomalies
- [ ] Verify anomaly severity colors (high=red, medium=yellow, low=blue)
- [ ] Check top issues table displays with correct ranking
- [ ] Filter issues by level (All → Errors → Warnings)
- [ ] Click "Export CSV" - verify file downloads
- [ ] Click "Export JSON" - verify file downloads
- [ ] Test responsive design on mobile viewport
- [ ] Verify no console errors

---

## Configuration

**File:** `.env` (Analytics service)

```bash
# Existing from Issue #011
DATABASE_URL=postgresql://analytics_user:analytics_pass@localhost:5432/devsmith_analytics
LOGS_DATABASE_URL=postgresql://analytics_user:readonly_pass@localhost:5432/devsmith_logs

# Analytics settings
AGGREGATION_INTERVAL=1h
ANOMALY_THRESHOLD=2.0
```

---

## Dependencies

**Add to `go.mod`:**
```
github.com/gomarkdown/markdown v0.0.0-20231222211730-1d6d20845b47
```

**Frontend (via CDN in layout.templ):**
```html
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
```

---

## Acceptance Criteria

Before marking this issue complete, verify:

- [x] Dashboard loads at `http://localhost:8083/`
- [x] Trends chart renders with Chart.js
- [x] Time range selector updates all sections
- [x] Anomalies display with correct severity colors
- [x] Top issues table renders with ranking
- [x] Issues level filter works
- [x] Export CSV downloads file
- [x] Export JSON downloads file
- [x] Responsive design works on desktop and tablet
- [x] Loading states show during data fetching
- [x] Error handling displays user-friendly messages
- [x] No console errors
- [x] Manual testing checklist complete

---

## Branch Naming

```bash
feature/014-analytics-dashboard-ui
```

---

## Notes

- Uses Chart.js via CDN for trend visualization
- Analytics reads from `logs.entries` (cross-schema READ-ONLY)
- Aggregation job runs hourly (configured in Issue #011)
- For MVP, anomaly detection uses 2σ threshold
- Export functionality uses existing `/api/v1/analytics/export` endpoint
- No real-time updates in MVP (future enhancement with WebSocket)

---

**Created:** 2025-10-20
**For:** Copilot Implementation
**Estimated Time:** 60-75 minutes
