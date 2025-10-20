# Issue #015: [COPILOT] Logs Service - Dashboard UI with Real-time Streaming

**Type:** Feature (Copilot Implementation)
**Service:** Logs
**Depends On:** Issue #009 (Logs Service Foundation), Issue #010 (WebSocket Streaming)
**Estimated Duration:** 60-75 minutes

---

## Summary

Create the Logs Service web UI that displays development logs with filtering, search, and real-time streaming via WebSocket. This dashboard allows developers to monitor logs as they're generated and search through historical log data.

**User Story:**
> As a developer, I want to view my development logs in real-time with filtering by level and service, search for specific log messages, and see new logs appear automatically, so I can monitor my development activity and debug issues quickly.

---

## Bounded Context

**Logs Service Context:**
- **Responsibility:** Log ingestion, storage, retrieval, and real-time streaming
- **Does NOT:** Analyze logs (Analytics does that), perform code review (Review does that)
- **Boundaries:** Logs owns `logs.entries` table exclusively

**Why This Matters:**
- Logs service is the authoritative source for raw log data
- Analytics reads from logs (cross-schema), but doesn't own logs
- WebSocket provides real-time updates without polling

---

## Success Criteria

### Must Have (MVP)
- [ ] Landing page at `/` shows log dashboard
- [ ] Real-time log streaming via WebSocket
- [ ] Log level filter (All, INFO, WARN, ERROR)
- [ ] Service filter (All services or specific service)
- [ ] Search functionality (filter by message content)
- [ ] Log entry display with timestamp, level, service, message
- [ ] Auto-scroll toggle (follow new logs or freeze view)
- [ ] Clear logs button (clears UI only, not database)
- [ ] Pause/Resume streaming toggle
- [ ] Responsive design (desktop and tablet)
- [ ] Color-coded log levels (INFO=blue, WARN=yellow, ERROR=red)

### Nice to Have (Post-MVP)
- Export logs as text file
- Timestamp range picker
- Copy individual log entry
- Log context (show surrounding logs)

---

## Database Schema

**Uses existing schema from Issue #009.**

Reads from `logs.entries` table via REST API and WebSocket.

---

## API Endpoints (Existing - Created in Issues #009 and #010)

### GET `/api/v1/logs`
Query parameters:
- `level`: `info`, `warn`, `error`, `all` (default: `all`)
- `service`: service name or `all` (default: `all`)
- `search`: text search in message
- `limit`: number (default: `100`)
- `offset`: number (default: `0`)

Returns historical log entries.

### WebSocket `/ws/logs`
From Issue #010 - real-time log streaming.

Receives JSON messages:
```json
{
  "id": "uuid",
  "timestamp": "2025-10-20T10:30:00Z",
  "level": "info",
  "service": "review",
  "message": "Analysis completed for repository..."
}
```

---

## File Structure

```
apps/logs/
├── templates/
│   ├── layout.templ              # NEW - Base layout
│   ├── dashboard.templ           # NEW - Log dashboard
│   └── components/
│       ├── log_entry.templ       # NEW - Single log entry
│       └── filters.templ         # NEW - Filter controls
├── handlers/
│   ├── ui_handler.go             # NEW - UI route handlers
│   └── [existing API handlers]   # From Issues #009, #010
└── static/
    ├── css/
    │   └── logs.css              # NEW - Dashboard styles
    └── js/
        ├── logs.js               # NEW - Main dashboard logic
        └── websocket.js          # NEW - WebSocket connection

cmd/logs/
└── main.go                       # UPDATE - Add UI routes
```

---

## Implementation Details

### 1. Dashboard Template

**File:** `apps/logs/templates/dashboard.templ`

```go
package templates

templ Dashboard() {
	@Layout("DevSmith Logs") {
		<div class="logs-container">
			<header class="logs-header">
				<h1>📝 DevSmith Logs</h1>
				<p class="subtitle">Real-time development logs</p>

				@Filters()
				@Controls()
			</header>

			<main class="logs-main">
				<div id="logs-output" class="logs-output"></div>
				<div id="logs-loading" class="loading">Connecting to log stream...</div>
			</main>
		</div>
	}
}

templ Filters() {
	<div class="filters">
		<div class="filter-group">
			<label for="level-filter">Level:</label>
			<select id="level-filter">
				<option value="all">All Levels</option>
				<option value="info">INFO</option>
				<option value="warn">WARN</option>
				<option value="error">ERROR</option>
			</select>
		</div>

		<div class="filter-group">
			<label for="service-filter">Service:</label>
			<select id="service-filter">
				<option value="all">All Services</option>
				<option value="portal">Portal</option>
				<option value="review">Review</option>
				<option value="logs">Logs</option>
				<option value="analytics">Analytics</option>
			</select>
		</div>

		<div class="filter-group">
			<label for="search-input">Search:</label>
			<input
				type="text"
				id="search-input"
				placeholder="Filter by message..."
			/>
		</div>
	</div>
}

templ Controls() {
	<div class="controls">
		<button id="pause-btn" class="btn-control" title="Pause streaming">⏸️ Pause</button>
		<button id="auto-scroll-btn" class="btn-control active" title="Auto-scroll">📜 Auto-scroll</button>
		<button id="clear-btn" class="btn-control" title="Clear logs">🗑️ Clear</button>
		<span id="connection-status" class="status-indicator">🟢 Connected</span>
	</div>
}
```

---

### 2. Log Entry Component (rendered via JS)

For performance, log entries are rendered via JavaScript template strings rather than Templ components (thousands of logs).

---

### 3. UI Handler

**File:** `apps/logs/handlers/ui_handler.go`

```go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"devsmith/apps/logs/templates"
)

// DashboardHandler serves the main Logs dashboard
func DashboardHandler(c *gin.Context) {
	component := templates.Dashboard()
	component.Render(c.Request.Context(), c.Writer)
}
```

---

### 4. WebSocket Connection Logic

**File:** `apps/logs/static/js/websocket.js`

```javascript
class LogsWebSocket {
  constructor(url, onMessage, onStatusChange) {
    this.url = url;
    this.onMessage = onMessage;
    this.onStatusChange = onStatusChange;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.isPaused = false;
  }

  connect() {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.onStatusChange('connected');
    };

    this.ws.onmessage = (event) => {
      if (!this.isPaused) {
        const logEntry = JSON.parse(event.data);
        this.onMessage(logEntry);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.onStatusChange('error');
    };

    this.ws.onclose = () => {
      console.log('WebSocket closed');
      this.onStatusChange('disconnected');
      this.attemptReconnect();
    };
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Reconnecting... (attempt ${this.reconnectAttempts})`);

      setTimeout(() => {
        this.connect();
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      console.error('Max reconnect attempts reached');
      this.onStatusChange('failed');
    }
  }

  pause() {
    this.isPaused = true;
  }

  resume() {
    this.isPaused = false;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
```

---

### 5. Main Dashboard Logic

**File:** `apps/logs/static/js/logs.js`

```javascript
let logsWebSocket = null;
let autoScroll = true;
let currentFilters = {
  level: 'all',
  service: 'all',
  search: '',
};

// Initialize dashboard
document.addEventListener('DOMContentLoaded', () => {
  loadHistoricalLogs();
  connectWebSocket();
  setupEventListeners();
});

// Connect to WebSocket for real-time logs
function connectWebSocket() {
  const wsUrl = `ws://${window.location.host}/ws/logs`;

  logsWebSocket = new LogsWebSocket(
    wsUrl,
    handleNewLogEntry,
    handleConnectionStatus
  );

  logsWebSocket.connect();
}

// Load historical logs via REST API
async function loadHistoricalLogs() {
  try {
    const params = new URLSearchParams({
      level: currentFilters.level,
      service: currentFilters.service,
      search: currentFilters.search,
      limit: '100',
    });

    const response = await fetch(`/api/v1/logs?${params}`);
    const logs = await response.json();

    logs.forEach(log => renderLogEntry(log));
    document.getElementById('logs-loading').style.display = 'none';

  } catch (error) {
    console.error('Failed to load historical logs:', error);
    showError('Failed to load logs');
  }
}

// Handle new log entry from WebSocket
function handleNewLogEntry(logEntry) {
  if (matchesFilters(logEntry)) {
    renderLogEntry(logEntry);

    if (autoScroll) {
      scrollToBottom();
    }
  }
}

// Check if log entry matches current filters
function matchesFilters(log) {
  if (currentFilters.level !== 'all' && log.level !== currentFilters.level) {
    return false;
  }

  if (currentFilters.service !== 'all' && log.service !== currentFilters.service) {
    return false;
  }

  if (currentFilters.search && !log.message.toLowerCase().includes(currentFilters.search.toLowerCase())) {
    return false;
  }

  return true;
}

// Render single log entry
function renderLogEntry(log) {
  const logsOutput = document.getElementById('logs-output');

  const logDiv = document.createElement('div');
  logDiv.className = `log-entry log-${log.level}`;
  logDiv.innerHTML = `
    <span class="log-timestamp">${formatTimestamp(log.timestamp)}</span>
    <span class="log-level ${log.level}">${log.level.toUpperCase()}</span>
    <span class="log-service">[${log.service}]</span>
    <span class="log-message">${escapeHtml(log.message)}</span>
  `;

  logsOutput.appendChild(logDiv);

  // Limit to 1000 entries in UI (performance)
  const entries = logsOutput.children;
  if (entries.length > 1000) {
    logsOutput.removeChild(entries[0]);
  }
}

// Event listeners
function setupEventListeners() {
  // Level filter
  document.getElementById('level-filter').addEventListener('change', (e) => {
    currentFilters.level = e.target.value;
    refreshLogs();
  });

  // Service filter
  document.getElementById('service-filter').addEventListener('change', (e) => {
    currentFilters.service = e.target.value;
    refreshLogs();
  });

  // Search input
  document.getElementById('search-input').addEventListener('input', (e) => {
    currentFilters.search = e.target.value;
    refreshLogs();
  });

  // Pause button
  document.getElementById('pause-btn').addEventListener('click', togglePause);

  // Auto-scroll button
  document.getElementById('auto-scroll-btn').addEventListener('click', toggleAutoScroll);

  // Clear button
  document.getElementById('clear-btn').addEventListener('click', clearLogs);
}

function togglePause() {
  const btn = document.getElementById('pause-btn');

  if (logsWebSocket.isPaused) {
    logsWebSocket.resume();
    btn.textContent = '⏸️ Pause';
    btn.classList.remove('paused');
  } else {
    logsWebSocket.pause();
    btn.textContent = '▶️ Resume';
    btn.classList.add('paused');
  }
}

function toggleAutoScroll() {
  autoScroll = !autoScroll;
  const btn = document.getElementById('auto-scroll-btn');

  if (autoScroll) {
    btn.classList.add('active');
    scrollToBottom();
  } else {
    btn.classList.remove('active');
  }
}

function clearLogs() {
  document.getElementById('logs-output').innerHTML = '';
}

function refreshLogs() {
  clearLogs();
  loadHistoricalLogs();
}

function scrollToBottom() {
  const logsOutput = document.getElementById('logs-output');
  logsOutput.scrollTop = logsOutput.scrollHeight;
}

function handleConnectionStatus(status) {
  const statusIndicator = document.getElementById('connection-status');

  switch (status) {
    case 'connected':
      statusIndicator.innerHTML = '🟢 Connected';
      statusIndicator.className = 'status-indicator connected';
      break;
    case 'disconnected':
      statusIndicator.innerHTML = '🟡 Reconnecting...';
      statusIndicator.className = 'status-indicator reconnecting';
      break;
    case 'error':
    case 'failed':
      statusIndicator.innerHTML = '🔴 Disconnected';
      statusIndicator.className = 'status-indicator disconnected';
      break;
  }
}

function formatTimestamp(ts) {
  return new Date(ts).toLocaleTimeString();
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function showError(message) {
  const logsOutput = document.getElementById('logs-output');
  logsOutput.innerHTML = `<div class="error-message">${message}</div>`;
}
```

---

### 6. CSS Styles

**File:** `apps/logs/static/css/logs.css`

```css
:root {
  --info-color: #0366d6;
  --warn-color: #ffc107;
  --error-color: #dc3545;
  --bg-dark: #1e1e1e;
  --bg-light: #f6f8fa;
  --border: #e1e4e8;
  --text-primary: #24292e;
  --text-secondary: #586069;
}

.logs-container {
  max-width: 1600px;
  margin: 0 auto;
  padding: 2rem;
}

.logs-header {
  text-align: center;
  margin-bottom: 2rem;
}

.logs-header h1 {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
}

/* Filters */
.filters {
  display: flex;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-group label {
  font-weight: 600;
  color: var(--text-primary);
}

.filter-group select,
.filter-group input {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.875rem;
}

.filter-group input {
  width: 250px;
}

/* Controls */
.controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.btn-control {
  padding: 0.5rem 1rem;
  background: white;
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.btn-control:hover {
  background: var(--bg-light);
}

.btn-control.active {
  background: var(--info-color);
  color: white;
  border-color: var(--info-color);
}

.btn-control.paused {
  background: var(--warn-color);
  color: white;
  border-color: var(--warn-color);
}

.status-indicator {
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 600;
}

.status-indicator.connected { background: #d4edda; color: #155724; }
.status-indicator.reconnecting { background: #fff3cd; color: #856404; }
.status-indicator.disconnected { background: #f8d7da; color: #721c24; }

/* Logs Output */
.logs-main {
  position: relative;
}

.logs-output {
  background: var(--bg-dark);
  color: #d4d4d4;
  font-family: 'Courier New', Courier, monospace;
  font-size: 0.875rem;
  padding: 1rem;
  border-radius: 8px;
  height: 600px;
  overflow-y: auto;
  border: 1px solid var(--border);
}

.log-entry {
  display: flex;
  gap: 0.75rem;
  padding: 0.25rem 0;
  border-bottom: 1px solid #333;
}

.log-entry:hover {
  background: #2a2a2a;
}

.log-timestamp {
  color: #808080;
  flex-shrink: 0;
  width: 90px;
}

.log-level {
  font-weight: 600;
  flex-shrink: 0;
  width: 60px;
  text-align: center;
  padding: 0.125rem 0.25rem;
  border-radius: 3px;
}

.log-level.info { background: var(--info-color); color: white; }
.log-level.warn { background: var(--warn-color); color: #333; }
.log-level.error { background: var(--error-color); color: white; }

.log-service {
  color: #569cd6;
  flex-shrink: 0;
  width: 100px;
}

.log-message {
  flex: 1;
  color: #d4d4d4;
  word-break: break-word;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.error-message {
  text-align: center;
  padding: 2rem;
  color: var(--error-color);
}

/* Scrollbar styling */
.logs-output::-webkit-scrollbar {
  width: 12px;
}

.logs-output::-webkit-scrollbar-track {
  background: #2a2a2a;
}

.logs-output::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 6px;
}

.logs-output::-webkit-scrollbar-thumb:hover {
  background: #777;
}

@media (max-width: 768px) {
  .filters {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-group {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-group input {
    width: 100%;
  }

  .logs-output {
    height: 400px;
    font-size: 0.75rem;
  }

  .log-entry {
    flex-wrap: wrap;
  }
}
```

---

## Testing Requirements

### Manual Testing Checklist

- [ ] Navigate to `http://localhost:8082/`
- [ ] Verify WebSocket connects (status shows "🟢 Connected")
- [ ] Generate some logs (use Review or Analytics services)
- [ ] Verify new logs appear in real-time
- [ ] Filter by level (INFO → WARN → ERROR)
- [ ] Filter by service (All → Portal → Review)
- [ ] Type in search box - verify filtering works
- [ ] Click "Pause" - verify new logs don't appear
- [ ] Click "Resume" - verify streaming resumes
- [ ] Disable "Auto-scroll" - verify view doesn't jump
- [ ] Enable "Auto-scroll" - verify scrolls to bottom
- [ ] Click "Clear" - verify UI clears (but database unchanged)
- [ ] Disconnect WebSocket (kill service) - verify reconnection attempts
- [ ] Test responsive design on mobile viewport

---

## Configuration

**File:** `.env` (Logs service)

```bash
# Existing from Issue #009
DATABASE_URL=postgresql://logs_user:logs_pass@localhost:5432/devsmith_logs

# WebSocket settings
WEBSOCKET_PING_INTERVAL=30s
WEBSOCKET_MAX_CONNECTIONS=100
```

---

## Acceptance Criteria

Before marking this issue complete, verify:

- [x] Dashboard loads at `http://localhost:8082/`
- [x] WebSocket connects successfully
- [x] Historical logs load on page load
- [x] New logs appear in real-time
- [x] Level filter works (INFO, WARN, ERROR, All)
- [x] Service filter works
- [x] Search filter works
- [x] Pause/Resume toggle works
- [x] Auto-scroll toggle works
- [x] Clear button works
- [x] Connection status indicator updates correctly
- [x] WebSocket reconnects on disconnect
- [x] Color-coded log levels display correctly
- [x] Responsive design works on desktop and tablet
- [x] No console errors
- [x] Manual testing checklist complete

---

## Branch Naming

```bash
feature/015-logs-dashboard-ui
```

---

## Notes

- WebSocket provides real-time streaming (no polling needed)
- UI limits to 1000 log entries for performance (oldest removed)
- Historical logs loaded via REST API on page load
- Clear button only clears UI, not database
- WebSocket auto-reconnects with exponential backoff
- For MVP, no log export functionality (future enhancement)

---

**Created:** 2025-10-20
**For:** Copilot Implementation
**Estimated Time:** 60-75 minutes
