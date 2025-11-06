import React from 'react';
import { useAuth } from '../context/AuthContext';
import { useNavigate, Link } from 'react-router-dom';

export default function Dashboard() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="container mt-4">
      <nav className="navbar navbar-expand-lg navbar-light bg-light rounded mb-4">
        <div className="container-fluid">
          <span className="navbar-brand">DevSmith Platform</span>
          <div className="d-flex align-items-center">
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
          <div className="card">
            <div className="card-body">
              <h2 className="card-title">Dashboard</h2>
              <p className="card-text">
                Welcome to DevSmith Platform! Choose an application below to get started.
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-md-6 col-lg-3 mb-4">
          <Link to="/logs" className="text-decoration-none">
            <div className="card h-100 hover-card">
              <div className="card-body text-center">
                <i className="bi bi-journal-text text-primary" style={{ fontSize: '3rem' }}></i>
                <h5 className="card-title mt-3">Logs</h5>
                <p className="card-text">
                  View and analyze application logs with real-time monitoring
                </p>
              </div>
            </div>
          </Link>
        </div>

        <div className="col-md-6 col-lg-3 mb-4">
          <Link to="/review" className="text-decoration-none">
            <div className="card h-100 hover-card">
              <div className="card-body text-center">
                <i className="bi bi-code-square text-success" style={{ fontSize: '3rem' }}></i>
                <h5 className="card-title mt-3">Code Review</h5>
                <p className="card-text">
                  AI-powered code review with five reading modes
                </p>
              </div>
            </div>
          </Link>
        </div>

        <div className="col-md-6 col-lg-3 mb-4">
          <Link to="/analytics" className="text-decoration-none">
            <div className="card h-100 hover-card">
              <div className="card-body text-center">
                <i className="bi bi-graph-up text-info" style={{ fontSize: '3rem' }}></i>
                <h5 className="card-title mt-3">Analytics</h5>
                <p className="card-text">
                  Analyze trends and patterns in your data
                </p>
              </div>
            </div>
          </Link>
        </div>

        <div className="col-md-6 col-lg-3 mb-4">
          <div className="card h-100 hover-card opacity-75">
            <div className="card-body text-center">
              <i className="bi bi-hammer text-warning" style={{ fontSize: '3rem' }}></i>
              <h5 className="card-title mt-3">Build</h5>
              <p className="card-text">
                Terminal interface and autonomous coding (Coming Soon)
              </p>
            </div>
          </div>
        </div>
      </div>

      <style jsx>{`
        .hover-card {
          transition: transform 0.2s, box-shadow 0.2s;
          cursor: pointer;
        }
        .hover-card:hover {
          transform: translateY(-5px);
          box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        }
        .text-decoration-none {
          color: inherit;
        }
      `}</style>
    </div>
  );
}
