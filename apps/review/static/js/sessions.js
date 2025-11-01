/**
 * Session Management JavaScript
 * Handles session list, detail view, and mode tab switching
 */

// Global state
let currentView = 'list'; // 'list', 'detail', 'history'
let currentSessionId = null;
let currentFilters = {
  status: '',
  language: ''
};

/**
 * Initialize session management on page load
 */
function initSessions() {
  loadSessionsList();
  setupEventListeners();
}

/**
 * Setup event listeners for session interactions
 */
function setupEventListeners() {
  // Filter changes
  const statusFilter = document.getElementById('status-filter');
  const languageFilter = document.getElementById('language-filter');
  
  if (statusFilter) {
    statusFilter.addEventListener('change', applyFilters);
  }
  if (languageFilter) {
    languageFilter.addEventListener('change', applyFilters);
  }

  // New session button
  const newSessionBtn = document.getElementById('new-session-btn');
  if (newSessionBtn) {
    newSessionBtn.addEventListener('click', showNewSessionDialog);
  }
}

/**
 * Load and display sessions list
 */
async function loadSessionsList() {
  try {
    const params = new URLSearchParams({
      status: currentFilters.status,
      language: currentFilters.language,
      limit: 10,
      offset: 0
    });

    const response = await fetch(`/api/review/sessions?${params}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      console.error('Failed to load sessions:', response.statusText);
      showNotification('Failed to load sessions', 'error');
      return;
    }

    const data = await response.json();
    displaySessionsList(data.sessions);
    currentView = 'list';
  } catch (error) {
    console.error('Error loading sessions:', error);
    showNotification('Error loading sessions', 'error');
  }
}

/**
 * Display sessions in the list view
 */
function displaySessionsList(sessions) {
  const container = document.querySelector('.sessions-grid');
  if (!container) return;

  if (!sessions || sessions.length === 0) {
    container.innerHTML = '<div class="empty-state"><p>📭 No sessions yet</p></div>';
    return;
  }

  container.innerHTML = sessions.map(session => `
    <div class="session-card" onclick="viewSession(${session.id})">
      <div class="session-header">
        <h3 class="session-title">${escapeHtml(session.title)}</h3>
        <span class="status-badge ${session.status}">${session.status}</span>
      </div>
      <div class="session-details">
        <div class="detail-row">
          <span class="label">Language:</span>
          <span class="value">${session.language || 'Unknown'}</span>
        </div>
        <div class="detail-row">
          <span class="label">Source:</span>
          <span class="value">${session.code_source}</span>
        </div>
        <div class="detail-row">
          <span class="label">Mode:</span>
          <span class="value">${session.current_mode || 'None'}</span>
        </div>
      </div>
      <div class="progress-section">
        <div class="progress-label">
          <span>Progress</span>
          <span>${Math.round(session.mode_progress || 0)}%</span>
        </div>
        <div class="progress-bar">
          <div class="progress-fill" style="width: ${session.mode_progress || 0}%"></div>
        </div>
      </div>
      <div class="session-footer">
        <div class="time-info">
          <span class="text-sm text-gray-600 dark:text-gray-400">
            Created: ${formatDate(session.created_at)}
          </span>
        </div>
        <div class="session-actions">
          <button class="btn-icon" title="View details" onclick="event.stopPropagation(); viewSession(${session.id})">
            👁️
          </button>
          <button class="btn-icon" title="Delete session" onclick="event.stopPropagation(); deleteSession(${session.id})">
            🗑️
          </button>
        </div>
      </div>
    </div>
  `).join('');
}

/**
 * View detailed session information
 */
async function viewSession(sessionId) {
  try {
    const response = await fetch(`/api/review/sessions/${sessionId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      console.error('Failed to load session:', response.statusText);
      showNotification('Failed to load session', 'error');
      return;
    }

    const session = await response.json();
    currentSessionId = sessionId;
    displaySessionDetail(session);
    currentView = 'detail';
  } catch (error) {
    console.error('Error loading session:', error);
    showNotification('Error loading session', 'error');
  }
}

/**
 * Display session detail view
 */
function displaySessionDetail(session) {
  const container = document.querySelector('main');
  if (!container) return;

  // Switch to detail view (in a real app, would navigate to new page or overlay)
  console.log('Session Detail:', session);
  showNotification(`Viewing session: ${session.title}`, 'info');
}

/**
 * Switch between mode tabs
 */
function switchMode(mode) {
  // Update active tab
  document.querySelectorAll('.tab-button').forEach(btn => {
    btn.classList.remove('active');
  });
  document.querySelector(`[data-mode="${mode}"]`)?.classList.add('active');

  // Show mode content
  document.querySelectorAll('.mode-content').forEach(content => {
    content.style.display = 'none';
  });
  document.querySelector(`.mode-content[data-mode="${mode}"]`).style.display = 'block';
}

/**
 * Apply filter selections
 */
function applyFilters() {
  const statusFilter = document.getElementById('status-filter');
  const languageFilter = document.getElementById('language-filter');

  currentFilters.status = statusFilter?.value || '';
  currentFilters.language = languageFilter?.value || '';

  loadSessionsList();
}

/**
 * Create new session dialog
 */
function showNewSessionDialog() {
  const title = prompt('Enter session title:');
  if (!title) return;

  const codeSource = prompt('Code source (paste/github/file):');
  if (!codeSource) return;

  createSession(title, codeSource);
}

/**
 * Create a new session
 */
async function createSession(title, codeSource) {
  try {
    const response = await fetch('/api/review/sessions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        title,
        code_source: codeSource,
        description: `New session for code review`
      })
    });

    if (!response.ok) {
      console.error('Failed to create session:', response.statusText);
      showNotification('Failed to create session', 'error');
      return;
    }

    showNotification('Session created successfully', 'success');
    loadSessionsList();
  } catch (error) {
    console.error('Error creating session:', error);
    showNotification('Error creating session', 'error');
  }
}

/**
 * Delete a session
 */
async function deleteSession(sessionId) {
  if (!confirm('Are you sure you want to delete this session?')) {
    return;
  }

  try {
    const response = await fetch(`/api/review/sessions/${sessionId}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      console.error('Failed to delete session:', response.statusText);
      showNotification('Failed to delete session', 'error');
      return;
    }

    showNotification('Session deleted successfully', 'success');
    loadSessionsList();
  } catch (error) {
    console.error('Error deleting session:', error);
    showNotification('Error deleting session', 'error');
  }
}

/**
 * View session history
 */
async function viewHistory(sessionId) {
  try {
    const response = await fetch(`/api/review/sessions/${sessionId}/history?limit=50`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      console.error('Failed to load history:', response.statusText);
      showNotification('Failed to load history', 'error');
      return;
    }

    const data = await response.json();
    displayHistory(data.history);
    currentView = 'history';
  } catch (error) {
    console.error('Error loading history:', error);
    showNotification('Error loading history', 'error');
  }
}

/**
 * Display session history
 */
function displayHistory(history) {
  console.log('History:', history);
  showNotification(`Loaded ${history ? history.length : 0} history entries`, 'info');
}

/**
 * Run a specific mode on the session
 */
async function runMode(mode, sessionId) {
  try {
    const response = await fetch(`/api/review/sessions/${currentSessionId || sessionId}/modes/${mode}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        status: 'in_progress'
      })
    });

    if (!response.ok) {
      console.error('Failed to run mode:', response.statusText);
      showNotification('Failed to run mode', 'error');
      return;
    }

    showNotification(`Running ${mode} mode...`, 'info');
  } catch (error) {
    console.error('Error running mode:', error);
    showNotification('Error running mode', 'error');
  }
}

/**
 * Add note to a mode
 */
async function addNote(mode, sessionId) {
  const note = prompt(`Add note for ${mode} mode:`);
  if (!note) return;

  try {
    const response = await fetch(`/api/review/sessions/${currentSessionId || sessionId}/notes`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        mode,
        note
      })
    });

    if (!response.ok) {
      console.error('Failed to add note:', response.statusText);
      showNotification('Failed to add note', 'error');
      return;
    }

    showNotification('Note added successfully', 'success');
  } catch (error) {
    console.error('Error adding note:', error);
    showNotification('Error adding note', 'error');
  }
}

/**
 * Navigate back to list view
 */
function backToList() {
  currentView = 'list';
  loadSessionsList();
}

/**
 * Navigate back to detail view
 */
function backToDetail() {
  if (currentSessionId) {
    viewSession(currentSessionId);
  } else {
    backToList();
  }
}

/**
 * Show notification toast
 */
function showNotification(message, type = 'info') {
  console.log(`[${type.toUpperCase()}] ${message}`);
  // In a real app, would display a toast notification
}

/**
 * Format date string
 */
function formatDate(dateStr) {
  if (!dateStr) return 'Unknown';
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

/**
 * Escape HTML to prevent XSS
 */
function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Navigate to specific page
 */
function goToPage(offset) {
  // In a real app, would load sessions at the new offset
  console.log('Go to page offset:', offset);
  loadSessionsList();
}

// Initialize on DOM content loaded
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initSessions);
} else {
  initSessions();
}
