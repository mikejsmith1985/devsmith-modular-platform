let logsWebSocket = null;
let autoScroll = true;
let currentFilters = {
  level: 'all',
  service: 'all',
  search: '',
};

// Initialize dashboard on page load
document.addEventListener('DOMContentLoaded', () => {
  loadHistoricalLogs();
  connectWebSocket();
  setupEventListeners();
});

// Connect to WebSocket for real-time logs
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
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const logs = await response.json();
    if (Array.isArray(logs)) {
      logs.forEach(log => renderLogEntry(log));
    }

    const loadingElement = document.getElementById('logs-loading');
    if (loadingElement) {
      loadingElement.style.display = 'none';
    }
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
  if (!logsOutput) return;

  const logDiv = document.createElement('div');
  logDiv.className = `log-entry log-${log.level}`;
  logDiv.innerHTML = `
    <span class="log-timestamp">${formatTimestamp(log.timestamp)}</span>
    <span class="log-level ${log.level}">${log.level.toUpperCase()}</span>
    <span class="log-service">[${escapeHtml(log.service)}]</span>
    <span class="log-message">${escapeHtml(log.message)}</span>
  `;

  logsOutput.appendChild(logDiv);

  // Limit to 1000 entries in UI (performance)
  const entries = logsOutput.children;
  if (entries.length > 1000) {
    logsOutput.removeChild(entries[0]);
  }
}

// Setup event listeners
function setupEventListeners() {
  const levelFilter = document.getElementById('level-filter');
  const serviceFilter = document.getElementById('service-filter');
  const searchInput = document.getElementById('search-input');
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
      currentFilters.search = e.target.value;
      refreshLogs();
    });
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
  const logsOutput = document.getElementById('logs-output');
  if (logsOutput) {
    logsOutput.innerHTML = '';
  }
}

function refreshLogs() {
  clearLogs();
  loadHistoricalLogs();
}

function scrollToBottom() {
  const logsOutput = document.getElementById('logs-output');
  if (logsOutput) {
    logsOutput.scrollTop = logsOutput.scrollHeight;
  }
}

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

function formatTimestamp(ts) {
  try {
    return new Date(ts).toLocaleTimeString();
  } catch {
    return 'N/A';
  }
}

function escapeHtml(text) {
  if (!text) return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function showError(message) {
  const logsOutput = document.getElementById('logs-output');
  if (logsOutput) {
    logsOutput.innerHTML = `<div class="error-message">${escapeHtml(message)}</div>`;
  }
}
