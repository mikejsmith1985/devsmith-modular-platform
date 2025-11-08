import React from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';

export default function AnalyticsPage() {
  const { user, logout } = useAuth();

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
              Analyze trends and patterns in your application data
            </p>
          </div>
        </div>
      </div>

            <div className="row">
        <div className="col-md-4 mb-4">
          <div className="frosted-card p-4 text-center h-100">
            <i className="bi bi-bar-chart-line text-primary" style={{ fontSize: '3rem' }}></i>
            <h5 className="mt-3 mb-3">Trends</h5>
            <p className="mb-0" style={{ color: 'var(--bs-gray-200)' }}>
              Identify patterns and trends in your data
            </p>
          </div>
        </div>

        <div className="col-md-4 mb-4">
          <div className="frosted-card p-4 text-center h-100">
            <i className="bi bi-exclamation-diamond text-warning" style={{ fontSize: '3rem' }}></i>
            <h5 className="mt-3 mb-3">Anomalies</h5>
            <p className="mb-0" style={{ color: 'var(--bs-gray-200)' }}>
              Detect unusual patterns and outliers
            </p>
          </div>
        </div>

        <div className="col-md-4 mb-4">
          <div className="frosted-card p-4 text-center h-100">
            <i className="bi bi-speedometer2 text-success" style={{ fontSize: '3rem' }}></i>
            <h5 className="mt-3 mb-3">Performance</h5>
            <p className="mb-0" style={{ color: 'var(--bs-gray-200)' }}>
              Monitor system performance metrics
            </p>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-12">
          <div className="frosted-card p-4">
            <h5 className="mb-3">Coming Soon</h5>
            <p className="mb-3" style={{ color: 'var(--bs-gray-200)' }}>
              Analytics features are currently in development. Check back soon for:
            </p>
            <ul style={{ color: 'var(--bs-gray-200)' }}>
              <li>Real-time trend analysis</li>
              <li>Anomaly detection and alerts</li>
              <li>Performance monitoring dashboards</li>
              <li>Custom report generation</li>
              <li>Data visualization and charts</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
