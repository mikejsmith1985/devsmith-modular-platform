import React from 'react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import { useNavigate, Link } from 'react-router-dom';

export default function Dashboard() {
  const { user, logout } = useAuth();
  const { isDarkMode, toggleTheme } = useTheme();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="container mt-4">
      <nav className="navbar navbar-expand-lg navbar-light frosted-card mb-4">
        <div className="container-fluid">
          <span className="navbar-brand fw-bold" style={{ fontSize: '1.5rem', color: isDarkMode ? '#e0e7ff' : '#1e293b' }}>
            DevSmith Platform
          </span>
          <div className="d-flex align-items-center gap-3">
            <button
              onClick={toggleTheme}
              className="theme-toggle"
              title="Toggle dark/light mode"
            >
              <i className={`bi ${isDarkMode ? 'bi-sun-fill' : 'bi-moon-fill'}`}></i>
            </button>
            <span className="me-3">Welcome, {user?.username || user?.name}!</span>
            <button
              className="btn btn-outline-danger btn-sm"
              onClick={handleLogout}
            >
              Logout
            </button>
          </div>
        </div>
      </nav>

      <div className="row">
        <div className="col-12 mb-4">
          <div className="frosted-card p-4">
            <h2 className="mb-3">Dashboard</h2>
            <p className="mb-0">
              Welcome to DevSmith Platform! Choose an application below to get started.
            </p>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-md-6 col-lg-3 mb-4">
          <Link to="/health" className="text-decoration-none">
            <div className="frosted-card p-4 text-center h-100">
              <i className="bi bi-heart-pulse mb-3" style={{fontSize: '3.3rem', color: '#06b6d4'}}></i>
              <h5 className="mb-3">Health</h5>
              <p className="mb-0" style={{fontSize: '0.9rem'}}>
                System Health monitoring, service logs, and diagnostics
              </p>
            </div>
          </Link>
        </div>

        <div className="col-md-6 col-lg-3 mb-4">
          <Link to="/review" className="text-decoration-none">
            <div className="frosted-card p-4 text-center h-100">
              <i className="bi bi-code-square mb-3" style={{fontSize: '3.3rem', color: '#8b5cf6'}}></i>
              <h5 className="mb-3">Code Review</h5>
              <p className="mb-0" style={{fontSize: '0.9rem'}}>
                AI-powered code review with five reading modes
              </p>
            </div>
          </Link>
        </div>

        <div className="col-md-6 col-lg-3 mb-4">
          <Link to="/analytics" className="text-decoration-none">
            <div className="frosted-card p-4 text-center h-100">
              <i className="bi bi-graph-up mb-3" style={{fontSize: '3.3rem', color: '#06b6d4'}}></i>
              <h5 className="mb-3">Analytics</h5>
              <p className="mb-0" style={{fontSize: '0.9rem'}}>
                Analyze trends and patterns in your data
              </p>
            </div>
          </Link>
        </div>

        <div className="col-md-6 col-lg-3 mb-4">
          <Link to="/llm-config" className="text-decoration-none">
            <div className="frosted-card p-4 text-center h-100">
              <i className="bi bi-robot mb-3" style={{fontSize: '3.3rem', color: '#10b981'}}></i>
              <h5 className="mb-3">AI Factory</h5>
              <p className="mb-0" style={{fontSize: '0.9rem'}}>
                Configure AI models and API keys for each app
              </p>
            </div>
          </Link>
        </div>
      </div>
    </div>
  );
}
