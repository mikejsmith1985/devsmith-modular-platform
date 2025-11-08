import { useState, useEffect } from 'react';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

function MonitoringDashboard() {
  const [stats, setStats] = useState(null);
  const [metrics, setMetrics] = useState(null);
  const [alerts, setAlerts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Fetch monitoring data
  const fetchMonitoringData = async () => {
    try {
      setError(null);
      
      // Fetch all monitoring endpoints in parallel
      const [statsRes, metricsRes, alertsRes] = await Promise.all([
        fetch('/api/logs/monitoring/stats'),
        fetch('/api/logs/monitoring/metrics?window=1h'),
        fetch('/api/logs/monitoring/alerts?active=true')
      ]);

      if (!statsRes.ok || !metricsRes.ok || !alertsRes.ok) {
        throw new Error('Failed to fetch monitoring data');
      }

      const [statsData, metricsData, alertsData] = await Promise.all([
        statsRes.json(),
        metricsRes.json(),
        alertsRes.json()
      ]);

      setStats(statsData);
      setMetrics(metricsData);
      setAlerts(alertsData.alerts || []);
      setLoading(false);
    } catch (err) {
      console.error('Failed to fetch monitoring data:', err);
      setError(err.message);
      setLoading(false);
    }
  };

  // Initial load and auto-refresh
  useEffect(() => {
    fetchMonitoringData();

    if (autoRefresh) {
      const interval = setInterval(fetchMonitoringData, 30000); // Refresh every 30s
      return () => clearInterval(interval);
    }
  }, [autoRefresh]);

  // Response time chart data
  const responseTimeChartData = metrics ? {
    labels: ['P50', 'P95', 'P99', 'Max'],
    datasets: [
      {
        label: 'Response Time (ms)',
        data: [
          metrics.response_times?.p50 || 0,
          metrics.response_times?.p95 || 0,
          metrics.response_times?.p99 || 0,
          metrics.response_times?.max || 0
        ],
        backgroundColor: 'rgba(99, 102, 241, 0.2)',
        borderColor: 'rgba(99, 102, 241, 1)',
        borderWidth: 2,
        fill: true,
        tension: 0.4
      }
    ]
  } : null;

  const responseTimeChartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: true,
        labels: {
          color: 'var(--bs-gray-200)'
        }
      },
      title: {
        display: true,
        text: 'Response Time Distribution',
        color: 'var(--bs-gray-100)',
        font: { size: 16 }
      }
    },
    scales: {
      y: {
        beginAtZero: true,
        ticks: { color: 'var(--bs-gray-300)' },
        grid: { color: 'rgba(255, 255, 255, 0.1)' }
      },
      x: {
        ticks: { color: 'var(--bs-gray-300)' },
        grid: { color: 'rgba(255, 255, 255, 0.1)' }
      }
    }
  };

  if (loading) {
    return (
      <div className="container mt-4">
        <div className="frosted-card p-4 text-center">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
          <p className="mt-3" style={{ color: 'var(--bs-gray-200)' }}>Loading monitoring data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mt-4">
        <div className="frosted-card p-4">
          <div className="alert alert-danger" role="alert">
            <i className="bi bi-exclamation-triangle-fill me-2"></i>
            Failed to load monitoring data: {error}
          </div>
          <button className="btn btn-primary" onClick={fetchMonitoringData}>
            <i className="bi bi-arrow-clockwise me-2"></i>
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="container-fluid mt-4">
      {/* Header with auto-refresh toggle */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2 style={{ color: 'var(--bs-gray-100)' }}>
          <i className="bi bi-activity me-2"></i>
          Health Monitoring Dashboard
        </h2>
        <div className="d-flex gap-2">
          <div className="form-check form-switch">
            <input
              className="form-check-input"
              type="checkbox"
              id="autoRefreshToggle"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
            />
            <label className="form-check-label" htmlFor="autoRefreshToggle" style={{ color: 'var(--bs-gray-200)' }}>
              Auto-refresh (30s)
            </label>
          </div>
          <button className="btn btn-sm btn-outline-primary" onClick={fetchMonitoringData}>
            <i className="bi bi-arrow-clockwise me-1"></i>
            Refresh
          </button>
        </div>
      </div>

      {/* Summary Stats Cards */}
      <div className="row g-3 mb-4">
        <div className="col-md-3">
          <div className="frosted-card p-3 h-100">
            <div className="d-flex justify-content-between align-items-center">
              <div>
                <h6 style={{ color: 'var(--bs-gray-300)' }}>Services Up</h6>
                <h2 className="mb-0 text-success">{stats?.services_up || 0}</h2>
              </div>
              <i className="bi bi-check-circle-fill text-success" style={{ fontSize: '2rem' }}></i>
            </div>
          </div>
        </div>

        <div className="col-md-3">
          <div className="frosted-card p-3 h-100">
            <div className="d-flex justify-content-between align-items-center">
              <div>
                <h6 style={{ color: 'var(--bs-gray-300)' }}>Services Down</h6>
                <h2 className="mb-0 text-danger">{stats?.services_down || 0}</h2>
              </div>
              <i className="bi bi-x-circle-fill text-danger" style={{ fontSize: '2rem' }}></i>
            </div>
          </div>
        </div>

        <div className="col-md-3">
          <div className="frosted-card p-3 h-100">
            <div className="d-flex justify-content-between align-items-center">
              <div>
                <h6 style={{ color: 'var(--bs-gray-300)' }}>Error Rate</h6>
                <h2 className="mb-0 text-warning">
                  {stats?.error_rate?.toFixed(2) || '0.00'}/min
                </h2>
              </div>
              <i className="bi bi-exclamation-triangle-fill text-warning" style={{ fontSize: '2rem' }}></i>
            </div>
          </div>
        </div>

        <div className="col-md-3">
          <div className="frosted-card p-3 h-100">
            <div className="d-flex justify-content-between align-items-center">
              <div>
                <h6 style={{ color: 'var(--bs-gray-300)' }}>Active Alerts</h6>
                <h2 className="mb-0 text-info">{stats?.active_alerts || 0}</h2>
              </div>
              <i className="bi bi-bell-fill text-info" style={{ fontSize: '2rem' }}></i>
            </div>
          </div>
        </div>
      </div>

      {/* Charts Row */}
      <div className="row g-3 mb-4">
        <div className="col-lg-8">
          <div className="frosted-card p-4" style={{ height: '400px' }}>
            {responseTimeChartData && (
              <Line data={responseTimeChartData} options={responseTimeChartOptions} />
            )}
          </div>
        </div>

        <div className="col-lg-4">
          <div className="frosted-card p-4" style={{ height: '400px', overflowY: 'auto' }}>
            <h5 style={{ color: 'var(--bs-gray-100)' }} className="mb-3">
              <i className="bi bi-gear-fill me-2"></i>
              Service Health
            </h5>
            {stats?.service_health && Object.entries(stats.service_health).map(([service, health]) => (
              <div key={service} className="mb-3">
                <div className="d-flex justify-content-between align-items-center">
                  <span style={{ color: 'var(--bs-gray-200)', textTransform: 'capitalize' }}>
                    {service}
                  </span>
                  <span className={`badge ${
                    health === 'healthy' ? 'bg-success' :
                    health === 'degraded' ? 'bg-warning' :
                    'bg-danger'
                  }`}>
                    {health}
                  </span>
                </div>
                <div className="progress mt-2" style={{ height: '4px' }}>
                  <div
                    className={`progress-bar ${
                      health === 'healthy' ? 'bg-success' :
                      health === 'degraded' ? 'bg-warning' :
                      'bg-danger'
                    }`}
                    role="progressbar"
                    style={{ width: health === 'healthy' ? '100%' : health === 'degraded' ? '60%' : '20%' }}
                  ></div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Metrics Summary */}
      <div className="row g-3 mb-4">
        <div className="col-lg-6">
          <div className="frosted-card p-4">
            <h5 style={{ color: 'var(--bs-gray-100)' }} className="mb-3">
              <i className="bi bi-speedometer2 me-2"></i>
              Response Time Metrics
            </h5>
            <div className="row g-3">
              <div className="col-6">
                <div className="text-center p-3" style={{ backgroundColor: 'rgba(99, 102, 241, 0.1)', borderRadius: '8px' }}>
                  <small style={{ color: 'var(--bs-gray-300)' }}>Average</small>
                  <h4 className="mb-0 text-primary">{metrics?.response_times?.avg?.toFixed(0) || '0'}ms</h4>
                </div>
              </div>
              <div className="col-6">
                <div className="text-center p-3" style={{ backgroundColor: 'rgba(99, 102, 241, 0.1)', borderRadius: '8px' }}>
                  <small style={{ color: 'var(--bs-gray-300)' }}>P50</small>
                  <h4 className="mb-0 text-primary">{metrics?.response_times?.p50?.toFixed(0) || '0'}ms</h4>
                </div>
              </div>
              <div className="col-6">
                <div className="text-center p-3" style={{ backgroundColor: 'rgba(99, 102, 241, 0.1)', borderRadius: '8px' }}>
                  <small style={{ color: 'var(--bs-gray-300)' }}>P95</small>
                  <h4 className="mb-0 text-primary">{metrics?.response_times?.p95?.toFixed(0) || '0'}ms</h4>
                </div>
              </div>
              <div className="col-6">
                <div className="text-center p-3" style={{ backgroundColor: 'rgba(99, 102, 241, 0.1)', borderRadius: '8px' }}>
                  <small style={{ color: 'var(--bs-gray-300)' }}>P99</small>
                  <h4 className="mb-0 text-primary">{metrics?.response_times?.p99?.toFixed(0) || '0'}ms</h4>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="col-lg-6">
          <div className="frosted-card p-4">
            <h5 style={{ color: 'var(--bs-gray-100)' }} className="mb-3">
              <i className="bi bi-graph-up me-2"></i>
              Request Statistics
            </h5>
            <div className="row g-3">
              <div className="col-6">
                <div className="text-center p-3" style={{ backgroundColor: 'rgba(16, 185, 129, 0.1)', borderRadius: '8px' }}>
                  <small style={{ color: 'var(--bs-gray-300)' }}>Total Requests</small>
                  <h4 className="mb-0 text-success">{metrics?.request_count || 0}</h4>
                </div>
              </div>
              <div className="col-6">
                <div className="text-center p-3" style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)', borderRadius: '8px' }}>
                  <small style={{ color: 'var(--bs-gray-300)' }}>Errors</small>
                  <h4 className="mb-0 text-danger">{metrics?.error_count || 0}</h4>
                </div>
              </div>
              <div className="col-12">
                <div className="text-center p-3" style={{ backgroundColor: 'rgba(245, 158, 11, 0.1)', borderRadius: '8px' }}>
                  <small style={{ color: 'var(--bs-gray-300)' }}>Error Rate</small>
                  <h4 className="mb-0 text-warning">{metrics?.error_rate?.toFixed(2) || '0.00'} errors/min</h4>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Active Alerts */}
      {alerts.length > 0 && (
        <div className="frosted-card p-4">
          <h5 style={{ color: 'var(--bs-gray-100)' }} className="mb-3">
            <i className="bi bi-bell-fill me-2"></i>
            Active Alerts
          </h5>
          <div className="list-group">
            {alerts.map((alert, index) => (
              <div key={index} className="list-group-item bg-transparent border-0 border-bottom border-secondary">
                <div className="d-flex w-100 justify-content-between">
                  <h6 className="mb-1" style={{ color: 'var(--bs-gray-100)' }}>
                    <span className={`badge ${
                      alert.severity === 'critical' ? 'bg-danger' :
                      alert.severity === 'warning' ? 'bg-warning' :
                      'bg-info'
                    } me-2`}>
                      {alert.severity}
                    </span>
                    {alert.alert_type}
                  </h6>
                  <small style={{ color: 'var(--bs-gray-300)' }}>
                    {new Date(alert.triggered).toLocaleTimeString()}
                  </small>
                </div>
                <p className="mb-1" style={{ color: 'var(--bs-gray-200)' }}>{alert.message}</p>
                {alert.service_name && (
                  <small style={{ color: 'var(--bs-gray-300)' }}>
                    Service: {alert.service_name}
                  </small>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {alerts.length === 0 && (
        <div className="frosted-card p-4 text-center">
          <i className="bi bi-check-circle-fill text-success" style={{ fontSize: '3rem' }}></i>
          <h5 className="mt-3" style={{ color: 'var(--bs-gray-100)' }}>No Active Alerts</h5>
          <p style={{ color: 'var(--bs-gray-300)' }}>All systems are operating normally</p>
        </div>
      )}
    </div>
  );
}

export default MonitoringDashboard;
