import React from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';

export default function AnalyticsPage() {
  const { user, logout } = useAuth();

  return (
    <div className="container mt-4">
      <nav className="navbar navbar-expand-lg navbar-light bg-light rounded mb-4">
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
          <div className="card">
            <div className="card-body">
              <h2 className="card-title">
                <i className="bi bi-graph-up text-info me-2"></i>
                Analytics Dashboard
              </h2>
              <p className="card-text">
                Analyze trends and patterns in your application data
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-md-4 mb-4">
          <div className="card">
            <div className="card-body text-center">
              <i className="bi bi-bar-chart-line text-primary" style={{ fontSize: '3rem' }}></i>
              <h5 className="card-title mt-3">Trends</h5>
              <p className="card-text">
                Analyze data trends over time
              </p>
            </div>
          </div>
        </div>

        <div className="col-md-4 mb-4">
          <div className="card">
            <div className="card-body text-center">
              <i className="bi bi-exclamation-diamond text-warning" style={{ fontSize: '3rem' }}></i>
              <h5 className="card-title mt-3">Anomalies</h5>
              <p className="card-text">
                Detect unusual patterns and outliers
              </p>
            </div>
          </div>
        </div>

        <div className="col-md-4 mb-4">
          <div className="card">
            <div className="card-body text-center">
              <i className="bi bi-speedometer2 text-success" style={{ fontSize: '3rem' }}></i>
              <h5 className="card-title mt-3">Performance</h5>
              <p className="card-text">
                Monitor system performance metrics
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-12">
          <div className="card">
            <div className="card-body">
              <h5 className="card-title">Coming Soon</h5>
              <p className="card-text">
                Analytics features are currently in development. Check back soon for:
              </p>
              <ul>
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
    </div>
  );
}
