// Enable debug logging only in development mode
const DEBUG_ENABLED = window.location.hostname === 'localhost' || 
                      window.location.hostname === '127.0.0.1' ||
                      window.DEBUG_ENABLED === true;

// Internal debug logger - only logs if DEBUG_ENABLED
function _debug(message, ...args) {
  if (DEBUG_ENABLED) {
    console.log(`[Analytics] ${message}`, ...args);
  }
}

function _error(message, ...args) {
  if (DEBUG_ENABLED) {
    console.error(`[Analytics] ${message}`, ...args);
  }
}

let currentTimeRange = '24h';
let trendsChart = null;

// Initialize dashboard
document.addEventListener('DOMContentLoaded', () => {
  loadTrends();
  loadAnomalies();
  loadTopIssues();

  // Set up event listeners
  setupEventListeners();
});

// Setup all event listeners
function setupEventListeners() {
  // Time range selector
  const timeRangeSelect = document.getElementById('time-range');
  if (timeRangeSelect) {
    timeRangeSelect.addEventListener('change', (e) => {
      currentTimeRange = e.target.value;
      loadTrends();
      loadAnomalies();
      loadTopIssues();
    });
  }

  // Issues level filter
  const issuesLevelSelect = document.getElementById('issues-level');
  if (issuesLevelSelect) {
    issuesLevelSelect.addEventListener('change', () => {
      loadTopIssues();
    });
  }

  // Export buttons
  const exportCsvBtn = document.getElementById('export-csv');
  if (exportCsvBtn) {
    exportCsvBtn.addEventListener('click', () => {
      exportData('csv');
    });
  }

  const exportJsonBtn = document.getElementById('export-json');
  if (exportJsonBtn) {
    exportJsonBtn.addEventListener('click', () => {
      exportData('json');
    });
  }
}

// Load trends data and render chart
async function loadTrends() {
  try {
    const trendSection = document.querySelector('.trends-section');
    if (!trendSection) return;

    const response = await fetch(`/api/analytics/trends?time_range=${currentTimeRange}`);
    if (!response.ok) {
      throw new Error('Failed to fetch trends');
    }

    const data = await response.json();
    renderTrendsChart(data);

    const trendsLoading = document.getElementById('trends-loading');
    if (trendsLoading) {
      trendsLoading.style.display = 'none';
    }
  } catch (error) {
    _error('Failed to load trends:', error);
    showError('trends-chart', 'Failed to load trend data');
  }
}

// Load anomalies
async function loadAnomalies() {
  try {
    const anomaliesContainer = document.getElementById('anomalies-container');
    if (!anomaliesContainer) return;

    const response = await fetch(`/api/analytics/anomalies?time_range=${currentTimeRange}`);
    if (!response.ok) {
      throw new Error('Failed to fetch anomalies');
    }

    const anomalies = await response.json();
    renderAnomalies(anomalies || []);
  } catch (error) {
    _error('Failed to load anomalies:', error);
    showError('anomalies-container', 'Failed to load anomalies');
  }
}

// Load top issues
async function loadTopIssues() {
  try {
    const issuesContainer = document.getElementById('issues-container');
    if (!issuesContainer) return;

    const levelSelect = document.getElementById('issues-level');
    const level = levelSelect ? levelSelect.value : 'all';

    const response = await fetch(
      `/api/analytics/top-issues?time_range=${currentTimeRange}&level=${level}&limit=10`
    );
    if (!response.ok) {
      throw new Error('Failed to fetch top issues');
    }

    const issues = await response.json();
    renderTopIssues(issues || []);
  } catch (error) {
    _error('Failed to load top issues:', error);
    showError('issues-container', 'Failed to load top issues');
  }
}

// Render trends chart using Chart.js
function renderTrendsChart(data) {
  const chartElement = document.getElementById('trends-chart');
  if (!chartElement) return;

  const ctx = chartElement.getContext('2d');

  if (trendsChart) {
    trendsChart.destroy();
  }

  trendsChart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: data.timestamps || [],
      datasets: [
        {
          label: 'INFO',
          data: data.info_counts || [],
          borderColor: '#0366d6',
          backgroundColor: 'rgba(3, 102, 214, 0.1)',
          tension: 0.3,
          borderWidth: 2,
          pointRadius: 4,
          pointHoverRadius: 6,
        },
        {
          label: 'WARN',
          data: data.warn_counts || [],
          borderColor: '#ffc107',
          backgroundColor: 'rgba(255, 193, 7, 0.1)',
          tension: 0.3,
          borderWidth: 2,
          pointRadius: 4,
          pointHoverRadius: 6,
        },
        {
          label: 'ERROR',
          data: data.error_counts || [],
          borderColor: '#dc3545',
          backgroundColor: 'rgba(220, 53, 69, 0.1)',
          tension: 0.3,
          borderWidth: 2,
          pointRadius: 4,
          pointHoverRadius: 6,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'top',
          labels: {
            usePointStyle: true,
            padding: 20,
          },
        },
        title: {
          display: true,
          text: `Log Trends (${currentTimeRange})`,
          font: {
            size: 14,
            weight: 'bold',
          },
        },
      },
      scales: {
        y: {
          beginAtZero: true,
          title: {
            display: true,
            text: 'Log Count',
          },
          grid: {
            drawBorder: true,
          },
        },
        x: {
          title: {
            display: true,
            text: 'Time',
          },
          grid: {
            display: false,
          },
        },
      },
      interaction: {
        mode: 'index',
        intersect: false,
      },
    },
  });
}

// Render anomalies
function renderAnomalies(anomalies) {
  const container = document.getElementById('anomalies-container');
  if (!container) return;

  if (anomalies.length === 0) {
    container.innerHTML = '<p class="no-data">No anomalies detected in this time range.</p>';
    return;
  }

  container.innerHTML = anomalies.map(anomaly => `
    <div class="anomaly-card ${anomaly.severity || 'low'}">
      <div class="anomaly-header">
        <span class="severity-badge ${anomaly.severity || 'low'}">${anomaly.severity || 'low'}</span>
        <span class="anomaly-time">${formatTimestamp(anomaly.detected_at || new Date())}</span>
      </div>
      <div class="anomaly-content">
        <h4>${escapeHtml(anomaly.metric || 'Unknown Metric')}</h4>
        <p class="anomaly-description">${escapeHtml(anomaly.description || 'No description available')}</p>
        <div class="anomaly-stats">
          <span>Expected: ${(anomaly.expected_value || 0).toFixed(2)}</span>
          <span>Actual: ${(anomaly.actual_value || 0).toFixed(2)}</span>
          <span>Deviation: ${(anomaly.deviation || 0).toFixed(1)}σ</span>
        </div>
      </div>
    </div>
  `).join('');
}

// Render top issues table
function renderTopIssues(issues) {
  const container = document.getElementById('issues-container');
  if (!container) return;

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
            <td><span class="level-badge ${issue.level || 'info'}">${issue.level || 'info'}</span></td>
            <td class="message" title="${escapeHtml(issue.message_pattern || 'N/A')}">${escapeHtml(issue.message_pattern || 'N/A')}</td>
            <td class="count">${issue.count || 0}</td>
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
      `/api/analytics/export?format=${format}&time_range=${currentTimeRange}`
    );

    if (!response.ok) {
      throw new Error('Export failed');
    }

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `analytics-${currentTimeRange}.${format === 'csv' ? 'csv' : 'json'}`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(url);
  } catch (error) {
    _error('Export failed:', error);
    alert('Failed to export data. Please try again.');
  }
}

// Utility: Format timestamp
function formatTimestamp(ts) {
  if (!ts) return 'N/A';
  try {
    const date = new Date(ts);
    return date.toLocaleString();
  } catch {
    return 'N/A';
  }
}

// Utility: Escape HTML
function escapeHtml(text) {
  if (!text) return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Utility: Show error
function showError(containerId, message) {
  const container = document.getElementById(containerId);
  if (container) {
    container.innerHTML = `<p class="error-message">${escapeHtml(message)}</p>`;
  }
}
