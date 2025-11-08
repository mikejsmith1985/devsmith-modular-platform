import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';
import StatCards from './StatCards';
import MonitoringDashboard from './MonitoringDashboard';

export default function LogsPage() {
  const { user, logout } = useAuth();
  const [activeTab, setActiveTab] = useState('logs');
  const [stats, setStats] = useState({
    debug: 0,
    info: 0,
    warning: 0,
    error: 0,
    critical: 0
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchStats();
  }, []);

  const fetchStats = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:3000/api/logs/v1/stats', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('devsmith_token')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch stats');
      }

      const data = await response.json();
      setStats(data);
    } catch (err) {
      console.error('Error fetching stats:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

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

      <div className="row mb-4">
        <div className="col-12">
          <div className="frosted-card p-4">
            <div className="d-flex justify-content-between align-items-center mb-3">
              <h2 className="mb-0">
                <i className="bi bi-journal-text text-primary me-2"></i>
                Logs Dashboard
              </h2>
            </div>
            
            {/* Tab Navigation */}
            <ul className="nav nav-tabs mb-3">
              <li className="nav-item">
                <button
                  className={`nav-link ${activeTab === 'logs' ? 'active' : ''}`}
                  onClick={() => setActiveTab('logs')}
                  style={{
                    backgroundColor: activeTab === 'logs' ? 'rgba(99, 102, 241, 0.2)' : 'transparent',
                    color: activeTab === 'logs' ? 'var(--bs-primary)' : 'var(--bs-gray-300)',
                    border: 'none',
                    borderBottom: activeTab === 'logs' ? '2px solid var(--bs-primary)' : 'none'
                  }}
                >
                  <i className="bi bi-list-ul me-2"></i>
                  Logs
                </button>
              </li>
              <li className="nav-item">
                <button
                  className={`nav-link ${activeTab === 'monitoring' ? 'active' : ''}`}
                  onClick={() => setActiveTab('monitoring')}
                  style={{
                    backgroundColor: activeTab === 'monitoring' ? 'rgba(99, 102, 241, 0.2)' : 'transparent',
                    color: activeTab === 'monitoring' ? 'var(--bs-primary)' : 'var(--bs-gray-300)',
                    border: 'none',
                    borderBottom: activeTab === 'monitoring' ? '2px solid var(--bs-primary)' : 'none'
                  }}
                >
                  <i className="bi bi-activity me-2"></i>
                  Monitoring
                </button>
              </li>
            </ul>

            {activeTab === 'logs' && (
              <p className="mb-0" style={{ color: 'var(--bs-gray-200)' }}>
                Monitor your application logs in real-time
              </p>
            )}
          </div>
        </div>
      </div>

      {activeTab === 'logs' && (
        <>
          {loading && (
            <div className="text-center my-5">
              <div className="spinner-border text-primary" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            </div>
          )}

          {error && (
            <div className="alert alert-danger" role="alert">
              <i className="bi bi-exclamation-triangle-fill me-2"></i>
              {error}
            </div>
          )}

          {!loading && !error && (
            <>
              <StatCards stats={stats} />

              <div className="row mt-4">
                <div className="col-12">
                  <div className="frosted-card p-4">
                    <h5 className="mb-3">Recent Logs</h5>
                    <p className="mb-0" style={{ color: 'var(--bs-gray-200)' }}>
                      Log streaming and filtering features coming soon
                    </p>
                  </div>
                </div>
              </div>
            </>
          )}
        </>
      )}

      {activeTab === 'monitoring' && <MonitoringDashboard />}
    </div>
  );
}
