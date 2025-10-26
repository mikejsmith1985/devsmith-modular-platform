// Logs Dashboard JavaScript
const LOGS_SERVICE_URL = 'http://localhost:8082';

async function updateDashboard() {
	const timeRange = document.getElementById('time-range')?.value || 'last_24_hours';
	const service = document.getElementById('service-filter')?.value || '';
	
	try {
		// Fetch dashboard stats
		const statsResp = await fetch(`${LOGS_SERVICE_URL}/api/logs/dashboard/stats?time_range=${timeRange}&service=${service}`);
		if (statsResp.ok) {
			const stats = await statsResp.json();
			updateStatsCards(stats);
		}

		// Fetch top errors
		const errorsResp = await fetch(`${LOGS_SERVICE_URL}/api/logs/validations/top-errors?service=${service}&limit=10&days=1`);
		if (errorsResp.ok) {
			const errors = await errorsResp.json();
			updateTopErrors(errors);
		}

		// Fetch trends
		const trendsResp = await fetch(`${LOGS_SERVICE_URL}/api/logs/validations/trends?service=${service}&days=1&interval=hourly`);
		if (trendsResp.ok) {
			const trends = await trendsResp.json();
			updateTrends(trends);
		}

		// Fetch alerts
		await loadAlerts(service);
	} catch (error) {
		console.error('Dashboard update error:', error);
	}
}

function updateStatsCards(stats) {
	if (!stats) return;
	
	document.getElementById('stat-total-logs').textContent = formatNumber(stats.total_logs || 0);
	document.getElementById('stat-error-rate').textContent = (stats.error_rate || 0).toFixed(1) + '%';
	document.getElementById('stat-warning-count').textContent = formatNumber(stats.warning_count || 0);
	document.getElementById('stat-active-services').textContent = stats.service_count || 0;
}

function updateTopErrors(errors) {
	const listEl = document.getElementById('top-errors-list');
	if (!errors || errors.length === 0) {
		listEl.innerHTML = '<div class="empty-state">No errors found</div>';
		return;
	}

	listEl.innerHTML = errors.map((err, idx) => `
		<div class="error-item">
			<div class="error-rank">#${idx + 1}</div>
			<div class="error-details">
				<div class="error-type">${escapeHtml(err.error_type || 'Unknown')}</div>
				<div class="error-message">${escapeHtml((err.message || '').substring(0, 100))}</div>
			</div>
			<div class="error-count">${formatNumber(err.count || 0)}</div>
		</div>
	`).join('');
}

function updateTrends(trends) {
	const chartEl = document.getElementById('trends-chart');
	if (!trends || trends.length === 0) {
		chartEl.innerHTML = '<div class="empty-state">No trend data available</div>';
		return;
	}

	// Simple bar chart visualization
	const maxCount = Math.max(...trends.map(t => t.count || 0));
	chartEl.innerHTML = trends.map(trend => `
		<div class="trend-bar-item">
			<div class="trend-time">${trend.timestamp || 'N/A'}</div>
			<div class="trend-bar-container">
				<div class="trend-bar" style="width: ${(trend.count / maxCount * 100) || 0}%"></div>
			</div>
			<div class="trend-count">${formatNumber(trend.count || 0)}</div>
		</div>
	`).join('');
}

async function loadAlerts(service = '') {
	const alertsList = document.getElementById('alerts-list');
	try {
		const resp = await fetch(`${LOGS_SERVICE_URL}/api/logs/alert-config/${service || 'all'}`);
		if (resp.ok) {
			const alerts = await resp.json();
			renderAlerts(alerts);
		} else {
			alertsList.innerHTML = '<div class="empty-state">No alerts configured</div>';
		}
	} catch (error) {
		console.error('Failed to load alerts:', error);
		alertsList.innerHTML = '<div class="error-state">Failed to load alerts</div>';
	}
}

function renderAlerts(alerts) {
	const alertsList = document.getElementById('alerts-list');
	if (!alerts || alerts.length === 0) {
		alertsList.innerHTML = '<div class="empty-state">No alerts configured</div>';
		return;
	}

	alertsList.innerHTML = (Array.isArray(alerts) ? alerts : [alerts]).map(alert => `
		<div class="alert-item">
			<div class="alert-header">
				<div class="alert-service">${escapeHtml(alert.service || 'Unknown')}</div>
				<div class="alert-status ${alert.enabled ? 'enabled' : 'disabled'}">
					${alert.enabled ? '✓ Enabled' : '✕ Disabled'}
				</div>
			</div>
			<div class="alert-details">
				<div class="alert-detail">
					<span class="label">Error Threshold:</span>
					<span class="value">${alert.error_threshold_per_min || 0}/min</span>
				</div>
				<div class="alert-detail">
					<span class="label">Warning Threshold:</span>
					<span class="value">${alert.warning_threshold_per_min || 0}/min</span>
				</div>
				${alert.alert_email ? `<div class="alert-detail"><span class="label">Email:</span><span class="value">${escapeHtml(alert.alert_email)}</span></div>` : ''}
			</div>
			<div class="alert-actions">
				<button class="btn-secondary btn-small" onclick="editAlert('${alert.service}')">Edit</button>
				<button class="btn-danger btn-small" onclick="deleteAlert('${alert.service}')">Delete</button>
			</div>
		</div>
	`).join('');
}

function showAddAlertModal() {
	document.getElementById('add-alert-modal').style.display = 'flex';
}

function closeAddAlertModal() {
	document.getElementById('add-alert-modal').style.display = 'none';
	document.getElementById('add-alert-form').reset();
}

async function saveAlert(event) {
	event.preventDefault();
	const form = event.target;
	const formData = new FormData(form);
	
	const alert = {
		service: formData.get('service'),
		error_threshold_per_min: parseInt(formData.get('error_threshold_per_min')),
		warning_threshold_per_min: parseInt(formData.get('warning_threshold_per_min')),
		alert_email: formData.get('alert_email'),
		enabled: true
	};

	try {
		const resp = await fetch(`${LOGS_SERVICE_URL}/api/logs/alert-config`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(alert)
		});

		if (resp.ok) {
			closeAddAlertModal();
			await loadAlerts();
			showNotification('Alert created successfully', 'success');
		} else {
			showNotification('Failed to create alert', 'error');
		}
	} catch (error) {
		console.error('Error saving alert:', error);
		showNotification('Error creating alert: ' + error.message, 'error');
	}
}

async function deleteAlert(service) {
	if (!confirm('Delete alert for ' + service + '?')) return;
	try {
		const resp = await fetch(`${LOGS_SERVICE_URL}/api/logs/alert-config/${service}`, {
			method: 'DELETE'
		});
		if (resp.ok) {
			await loadAlerts();
			showNotification('Alert deleted', 'success');
		}
	} catch (error) {
		console.error('Error deleting alert:', error);
		showNotification('Error deleting alert', 'error');
	}
}

function editAlert(service) {
	showNotification('Edit feature coming soon', 'info');
}

function refreshDashboard() {
	updateDashboard();
}

function formatNumber(num) {
	if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
	if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
	return num.toString();
}

function escapeHtml(text) {
	const div = document.createElement('div');
	div.textContent = text;
	return div.innerHTML;
}

function showNotification(message, type = 'info') {
	const notification = document.createElement('div');
	notification.className = `notification ${type}`;
	notification.textContent = message;
	notification.style.cssText = `
		position: fixed; top: 20px; right: 20px; padding: 15px 20px;
		background: ${type === 'success' ? '#4CAF50' : type === 'error' ? '#f44336' : '#2196F3'};
		color: white; border-radius: 4px; z-index: 10000; box-shadow: 0 2px 5px rgba(0,0,0,0.2);
	`;
	document.body.appendChild(notification);
	setTimeout(() => notification.remove(), 3000);
}

// Load dashboard on page load
window.addEventListener('load', updateDashboard);
