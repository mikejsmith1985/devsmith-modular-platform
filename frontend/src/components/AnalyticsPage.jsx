import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';

export default function AnalyticsPage() {
  const { user, logout } = useAuth();
  const [metrics, setMetrics] = useState(null);
  const [violations, setViolations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadMetrics();
    const interval = setInterval(loadMetrics, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const loadMetrics = async () => {
    try {
      const [dashboardRes, violationsRes] = await Promise.all([
        fetch('/api/analytics/metrics/dashboard'),
        fetch('/api/analytics/metrics/violations')
      ]);

      if (dashboardRes.ok) {
        const data = await dashboardRes.json();
        setMetrics(data.data);
      }

      if (violationsRes.ok) {
        const data = await violationsRes.json();
        setViolations(data.violations || []);
      }

      setLoading(false);
    } catch (err) {
      setError(err.message);
      setLoading(false);
    }
  };

  const formatValue = (value) => {
    if (typeof value === 'number' && !Number.isInteger(value)) {
      return value.toFixed(2);
    }
    return value;
  };

  const getTrendIcon = (trend) => {
    if (!trend) return '→';
    if (trend.direction === 'up') return '↑';
    if (trend.direction === 'down') return '↓';
    return '→';
  };

  const getTrendColor = (trend) => {
    if (!trend) return 'text-secondary';
    if (trend.direction === 'up') return 'text-success';
    if (trend.direction === 'down') return 'text-danger';
    return 'text-secondary';
  };

  if (loading) {
    return (
      <div className="container mt-4">
        <div className="text-center">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
          <p className="mt-3">Loading metrics...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mt-4">
        <div className="alert alert-danger">
          Error loading metrics: {error}
          <button className="btn btn-sm btn-outline-danger ms-3" onClick={loadMetrics}>
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mt-4">
      <nav className="navbar navbar-expand-lg navbar-light frosted-card mb-4">
        <div className="container-fluid">
          <Link to="/" className="navbar-brand">
            <i className="bi bi-arrow-left me-2"></i>
            Back to Dashboard
          </Link>
          <div className="d-flex align-items-center">
            <span className="me-3">Welcome, {user?.username || user?.name}!</span>
            <button
              className="btn btn-outline-danger btn-sm"
              onClick={() => logout()}
            >
              Logout
            </button>
          </div>
        </div>
      </nav>

      <div className="row">
        <div className="col-12 mb-4">
          <div className="frosted-card p-4">
            <h2 className="mb-3">
              <i className="bi bi-graph-up text-info me-2"></i>
              Analytics Dashboard
            </h2>
            <p className="mb-0">
              Platform metrics and trends
            </p>
          </div>
        </div>
      </div>

      {/* Metrics Cards */}
      {metrics && (
        <div className="row mb-4">
          <div className="col-md-3 mb-3">
            <div className="frosted-card p-4">
              <h6 className="text-uppercase text-muted mb-2" style={{ fontSize: '0.75rem' }}>
                Test Pass Rate
              </h6>
              <div className="d-flex justify-content-between align-items-end">
                <h3 className="mb-0 text-success">
                  {formatValue(metrics.test_pass_rate)}%
                </h3>
                {metrics.trends?.test_pass_rate && (
                  <span className={`${getTrendColor(metrics.trends.test_pass_rate)}`}>
                    {getTrendIcon(metrics.trends.test_pass_rate)}{' '}
                    {Math.abs(metrics.trends.test_pass_rate.change).toFixed(1)}%
                  </span>
                )}
              </div>
            </div>
          </div>

          <div className="col-md-3 mb-3">
            <div className="frosted-card p-4">
              <h6 className="text-uppercase text-muted mb-2" style={{ fontSize: '0.75rem' }}>
                Deployment Frequency
              </h6>
              <div className="d-flex justify-content-between align-items-end">
                <h3 className="mb-0 text-primary">
                  {formatValue(metrics.deployment_frequency)}/day
                </h3>
                {metrics.trends?.deployment_frequency && (
                  <span className={`${getTrendColor(metrics.trends.deployment_frequency)}`}>
                    {getTrendIcon(metrics.trends.deployment_frequency)}
                  </span>
                )}
              </div>
            </div>
          </div>

          <div className="col-md-3 mb-3">
            <div className="frosted-card p-4">
              <h6 className="text-uppercase text-muted mb-2" style={{ fontSize: '0.75rem' }}>
                Service Health
              </h6>
              <div className="d-flex justify-content-between align-items-end">
                <h3 className="mb-0 text-success">
                  {formatValue(metrics.avg_service_health)}%
                </h3>
              </div>
            </div>
          </div>

          <div className="col-md-3 mb-3">
            <div className="frosted-card p-4">
              <h6 className="text-uppercase text-muted mb-2" style={{ fontSize: '0.75rem' }}>
                Rule Violations
              </h6>
              <div className="d-flex justify-content-between align-items-end">
                <h3 className="mb-0 text-warning">
                  {metrics.rule_violations}
                </h3>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Violations Table */}
      {violations && violations.length > 0 && (
        <div className="row">
          <div className="col-12">
            <div className="frosted-card p-4">
              <h5 className="mb-3">Rule Violations</h5>
              <div className="table-responsive">
                <table className="table table-dark table-hover">
                  <thead>
                    <tr>
                      <th>Rule</th>
                      <th className="text-center">Count</th>
                      <th>Severity</th>
                      <th>Last Seen</th>
                    </tr>
                  </thead>
                  <tbody>
                    {violations.map((violation, index) => (
                      <tr key={index}>
                        <td><code>{violation.rule}</code></td>
                        <td className="text-center">
                          <span className="badge bg-danger">{violation.count}</span>
                        </td>
                        <td>
                          <span className={`badge bg-${violation.severity === 'critical' ? 'danger' : 'warning'}`}>
                            {violation.severity}
                          </span>
                        </td>
                        <td>{new Date(violation.last_seen).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Empty State */}
      {metrics && violations.length === 0 && (
        <div className="row">
          <div className="col-12">
            <div className="frosted-card p-4 text-center">
              <i className="bi bi-check-circle text-success" style={{ fontSize: '3rem' }}></i>
              <h5 className="mt-3">No Rule Violations</h5>
              <p className="text-muted">All quality gates are passing!</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
