/**
 * Logs Dashboard Module
 * Manages real-time log streaming, filtering, and UI interactions
 * Features: WebSocket streaming, virtual scrolling, search debouncing, expandable details
 */

// ============================================================================
// STATE & CONFIGURATION
// ============================================================================

let logsWebSocket = null;
let autoScroll = true;

/** Current filter state for logs */
let currentFilters = {
  level: 'all',
  service: 'all',
  search: '',
  dateFrom: null,
  dateTo: null,
};

/** Search input debounce timer */
let searchDebounceTimer = null;
const SEARCH_DEBOUNCE_MS = 300;

/** Virtual scrolling configuration */
const VIRTUAL_SCROLL_CONFIG = {
  itemHeight: 25,
  bufferSize: 10,
};

let virtualScrollState = {
  visibleStart: 0,
  visibleEnd: 50,
};

// ============================================================================
// LIFECYCLE & INITIALIZATION
// ============================================================================

document.addEventListener('DOMContentLoaded', () => {
  loadHistoricalLogs();
  connectWebSocket();
  setupEventListeners();
  setupVirtualScrolling();
});

// ============================================================================
// WEBSOCKET & DATA LOADING
// ============================================================================

/**
 * Establishes WebSocket connection for real-time logs
 */
function connectWebSocket() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/ws/logs`;

  logsWebSocket = new LogsWebSocket(
    wsUrl,
    handleNewLogEntry,
    handleConnectionStatus
  );

  logsWebSocket.connect();
}

/**
 * Loads historical logs via REST API with current filters applied
 */
async function loadHistoricalLogs() {
  try {
    const params = new URLSearchParams({
      level: currentFilters.level,
      service: currentFilters.service,
      search: currentFilters.search,
      limit: '100',
    });

    if (currentFilters.dateFrom) {
      params.append('from', currentFilters.dateFrom);
    }
    if (currentFilters.dateTo) {
      params.append('to', currentFilters.dateTo);
    }

    // Use the correct API path through nginx
    // If accessed from /logs/, use /api/v1/logs
    // window.location.pathname will be /logs/ or /logs/dashboard
    const apiPath = `/api/v1/logs?${params}`;
    const response = await fetch(apiPath);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data = await response.json();
    // Handle both direct array response and object with entries property
    const logs = Array.isArray(data) ? data : (data.entries || []);
    if (Array.isArray(logs)) {
      const logsOutput = document.getElementById('logs-output');
      logsOutput.innerHTML = '';
      logs.forEach(log => renderLogEntry(log));
    }

    const loadingElement = document.getElementById('logs-loading');
    if (loadingElement) {
      loadingElement.style.display = 'none';
    }
  } catch (error) {
    console.error('Failed to load historical logs:', error);
    showToast('Failed to load logs', 'error');
  }
}

// ============================================================================
// LOG ENTRY HANDLING
// ============================================================================

/**
 * Handles new log entries from WebSocket stream
 * @param {Object} logEntry - Log entry object with level, message, service, etc.
 */
function handleNewLogEntry(logEntry) {
  if (matchesFilters(logEntry)) {
    renderLogEntry(logEntry);

    if (logEntry.level === 'ERROR') {
      showToast(`Error from ${logEntry.service}: ${logEntry.message.substring(0, 50)}...`, 'error');
    } else if (logEntry.level === 'WARN') {
      showToast(`Warning from ${logEntry.service}`, 'warning');
    }

    if (autoScroll) {
      scrollToBottom();
    }
  }
}

/**
 * Checks if log entry matches current filter criteria
 * @param {Object} log - Log entry to check
 * @returns {boolean} True if log matches all filters
 */
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

  if (currentFilters.dateFrom) {
    const logDate = new Date(log.created_at).toISOString().split('T')[0];
    if (logDate < currentFilters.dateFrom) return false;
  }

  if (currentFilters.dateTo) {
    const logDate = new Date(log.created_at).toISOString().split('T')[0];
    if (logDate > currentFilters.dateTo) return false;
  }

  return true;
}

/**
 * Renders a single log entry to the DOM with expandable details
 * @param {Object} log - Log entry to render
 */
function renderLogEntry(log) {
  const logsOutput = document.getElementById('logs-output');
  if (!logsOutput) return;

  const logDiv = document.createElement('div');
  const levelLower = (log.level || 'info').toLowerCase();
  logDiv.className = `log-entry log-${levelLower}`;
  logDiv.setAttribute('role', 'listitem');
  logDiv.innerHTML = `
    <button class="expand-btn" aria-label="Toggle details" title="Expand details">▶</button>
    <span class="log-timestamp">${formatTimestamp(log.created_at)}</span>
    <span class="log-level ${levelLower}">${(log.level || 'info').toUpperCase()}</span>
    <span class="log-service">[${escapeHtml(log.service)}]</span>
    <span class="log-message">${escapeHtml(log.message)}</span>
    <button class="copy-btn" data-copy aria-label="Copy log entry" title="Copy">📋</button>
    <div class="expanded-details">
      ${log.stackTrace ? `<div class="stack-trace">${escapeHtml(log.stackTrace)}</div>` : ''}
      ${(log.metadata || log.context) ? `<div class="metadata">${renderMetadata(log.metadata || log.context)}</div>` : ''}
    </div>
  `;

  const expandBtn = logDiv.querySelector('.expand-btn');
  expandBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    logDiv.classList.toggle('expanded');
    expandBtn.textContent = logDiv.classList.contains('expanded') ? '▼' : '▶';
  });

  const copyBtn = logDiv.querySelector('[data-copy]');
  copyBtn.addEventListener('click', async (e) => {
    e.stopPropagation();
    const text = logDiv.textContent;
    try {
      await navigator.clipboard.writeText(text);
      showToast('Copied to clipboard', 'success');
    } catch (err) {
      console.error('Failed to copy:', err);
      showToast('Failed to copy', 'error');
    }
  });

  logsOutput.appendChild(logDiv);

  const entries = logsOutput.children;
  if (entries.length > 1000) {
    logsOutput.removeChild(entries[0]);
  }
}

/**
 * Renders metadata/context as readable key-value pairs
 * @param {Object} context - Context object to render
 * @returns {string} HTML string for metadata display
 */
function renderMetadata(context) {
  if (!context || typeof context !== 'object') return '';

  return Object.entries(context)
    .map(([key, value]) => `
      <div class="metadata-row">
        <span class="metadata-key">${escapeHtml(key)}:</span>
        <span class="metadata-value">${escapeHtml(JSON.stringify(value))}</span>
      </div>
    `).join('');
}

// ============================================================================
// UI NOTIFICATIONS & FEEDBACK
// ============================================================================

/**
 * Shows a toast notification
 * @param {string} message - Notification message
 * @param {string} type - Notification type: 'info', 'success', 'warning', 'error'
 * @param {number} duration - Auto-dismiss duration in milliseconds
 */
function showToast(message, type = 'info', duration = 5000) {
  const container = document.getElementById('toast-container');
  if (!container) return;

  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.setAttribute('role', 'alert');
  toast.innerHTML = `
    <div class="toast-message">${escapeHtml(message)}</div>
    <button class="toast-close" aria-label="Close notification">✕</button>
  `;

  container.appendChild(toast);

  const closeBtn = toast.querySelector('.toast-close');
  const dismissToast = () => {
    toast.classList.add('dismissing');
    setTimeout(() => toast.remove(), 300);
  };

  closeBtn.addEventListener('click', dismissToast);
  setTimeout(dismissToast, duration);
}

// ============================================================================
// EVENT LISTENERS & CONTROLS
// ============================================================================

/**
 * Sets up all event listeners for filters and controls
 */
function setupEventListeners() {
  const levelFilter = document.getElementById('level-filter');
  const serviceFilter = document.getElementById('service-filter');
  const searchInput = document.getElementById('search-input');
  const dateFromInput = document.getElementById('date-from');
  const dateToInput = document.getElementById('date-to');
  const applyFiltersBtn = document.getElementById('apply-filters');
  const pauseBtn = document.getElementById('pause-btn');
  const autoScrollBtn = document.getElementById('auto-scroll-btn');
  const clearBtn = document.getElementById('clear-btn');

  if (levelFilter) {
    levelFilter.addEventListener('change', (e) => {
      currentFilters.level = e.target.value;
      refreshLogs();
    });
  }

  if (serviceFilter) {
    serviceFilter.addEventListener('change', (e) => {
      currentFilters.service = e.target.value;
      refreshLogs();
    });
  }

  if (searchInput) {
    searchInput.addEventListener('input', (e) => {
      clearTimeout(searchDebounceTimer);
      searchInput.value = e.target.value;

      searchDebounceTimer = setTimeout(() => {
        currentFilters.search = searchInput.value;
        refreshLogs();
      }, SEARCH_DEBOUNCE_MS);
    });
  }

  if (dateFromInput) {
    dateFromInput.addEventListener('change', (e) => {
      currentFilters.dateFrom = e.target.value;
    });
  }

  if (dateToInput) {
    dateToInput.addEventListener('change', (e) => {
      currentFilters.dateTo = e.target.value;
    });
  }

  if (applyFiltersBtn) {
    applyFiltersBtn.addEventListener('click', refreshLogs);
  }

  if (pauseBtn) {
    pauseBtn.addEventListener('click', togglePause);
  }

  if (autoScrollBtn) {
    autoScrollBtn.addEventListener('click', toggleAutoScroll);
  }

  if (clearBtn) {
    clearBtn.addEventListener('click', clearLogs);
  }
}

/**
 * Toggles pause/resume for log streaming
 */
function togglePause() {
  const btn = document.getElementById('pause-btn');
  if (!logsWebSocket || !btn) return;

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

/**
 * Toggles auto-scroll behavior
 */
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

/**
 * Clears all logs from the display
 */
function clearLogs() {
  const logsOutput = document.getElementById('logs-output');
  if (logsOutput) {
    logsOutput.innerHTML = '';
  }
}

/**
 * Refreshes logs by clearing and reloading with current filters
 */
function refreshLogs() {
  clearLogs();
  loadHistoricalLogs();
}

/**
 * Scrolls log container to the bottom
 */
function scrollToBottom() {
  const logsOutput = document.getElementById('logs-output');
  if (logsOutput) {
    logsOutput.scrollTop = logsOutput.scrollHeight;
  }
}

/**
 * Updates connection status display
 * @param {string} status - Connection status: 'connected', 'reconnecting', 'error', 'failed'
 */
function handleConnectionStatus(status) {
  const statusIndicator = document.getElementById('connection-status');
  if (!statusIndicator) return;

  switch (status) {
    case 'connected':
      statusIndicator.innerHTML = '🟢 Connected';
      statusIndicator.className = 'status-indicator connected';
      break;
    case 'reconnecting':
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

// ============================================================================
// VIRTUAL SCROLLING
// ============================================================================

/**
 * Initializes virtual scrolling with scroll event listener
 */
function setupVirtualScrolling() {
  const logsOutput = document.getElementById('logs-output');
  if (!logsOutput) return;

  logsOutput.addEventListener('scroll', () => {
    updateVirtualScroll();
  });
}

/**
 * Updates visibility of log entries based on scroll position
 * Hides entries outside viewport to improve performance
 */
function updateVirtualScroll() {
  const logsOutput = document.getElementById('logs-output');
  if (!logsOutput) return;

  const scrollTop = logsOutput.scrollTop;
  const containerHeight = logsOutput.clientHeight;

  const visibleStart = Math.max(0, Math.floor(scrollTop / VIRTUAL_SCROLL_CONFIG.itemHeight) - VIRTUAL_SCROLL_CONFIG.bufferSize);
  const visibleEnd = Math.ceil((scrollTop + containerHeight) / VIRTUAL_SCROLL_CONFIG.itemHeight) + VIRTUAL_SCROLL_CONFIG.bufferSize;

  const entries = logsOutput.querySelectorAll('.log-entry');
  entries.forEach((entry, index) => {
    if (index >= visibleStart && index < visibleEnd) {
      entry.style.display = '';
    } else {
      entry.style.display = 'none';
    }
  });
}

// ============================================================================
// UTILITIES
// ============================================================================

/**
 * Formats timestamp to local time string
 * @param {string} ts - ISO timestamp string
 * @returns {string} Formatted time (e.g., "14:30:45")
 */
function formatTimestamp(ts) {
  try {
    return new Date(ts).toLocaleTimeString();
  } catch {
    return 'N/A';
  }
}

/**
 * Escapes HTML special characters to prevent XSS
 * @param {string} text - Raw text to escape
 * @returns {string} HTML-escaped text
 */
function escapeHtml(text) {
  if (!text) return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// ============================================================================
// TEST HELPERS (DEVELOPMENT ONLY)
// ============================================================================

/**
 * Simulates receiving a new log entry (for testing)
 * @param {Object} logData - Log data to simulate
 */
window.simulateNewLog = function(logData) {
  handleNewLogEntry({
    id: Math.random(),
    level: logData.level || 'INFO',
    message: logData.message || '',
    service: logData.service || 'unknown',
    context: logData.context,
    stackTrace: logData.stackTrace,
    created_at: new Date().toISOString(),
    ...logData
  });
};

/**
 * Updates a log entry (for testing)
 * @param {number} index - Log entry index
 * @param {Object} updates - Updates to apply
 */
window.updateLog = function(index, updates) {
  const entries = document.querySelectorAll('.log-entry');
  if (entries[index]) {
    const entry = entries[index];
    if (updates.stackTrace) {
      const stackTraceEl = entry.querySelector('.stack-trace');
      if (stackTraceEl) {
        stackTraceEl.textContent = updates.stackTrace;
      }
    }
  }
};
