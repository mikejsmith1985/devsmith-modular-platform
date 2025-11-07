import React from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';

export default function ReviewPage() {
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
                <i className="bi bi-code-square text-success me-2"></i>
                Code Review
              </h2>
              <p className="card-text">
                AI-powered code review with five reading modes: Preview, Skim, Scan, Detailed, and Critical
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-md-6 mb-4">
          <div className="card h-100">
            <div className="card-body">
              <h5 className="card-title">
                <i className="bi bi-eye text-info me-2"></i>
                Preview Mode
              </h5>
              <p className="card-text">
                Quick structural assessment of code organization
              </p>
            </div>
          </div>
        </div>

        <div className="col-md-6 mb-4">
          <div className="card h-100">
            <div className="card-body">
              <h5 className="card-title">
                <i className="bi bi-list-ul text-primary me-2"></i>
                Skim Mode
              </h5>
              <p className="card-text">
                Understand abstractions and function signatures
              </p>
            </div>
          </div>
        </div>

        <div className="col-md-6 mb-4">
          <div className="card h-100">
            <div className="card-body">
              <h5 className="card-title">
                <i className="bi bi-search text-warning me-2"></i>
                Scan Mode
              </h5>
              <p className="card-text">
                Find specific information with semantic search
              </p>
            </div>
          </div>
        </div>

        <div className="col-md-6 mb-4">
          <div className="card h-100">
            <div className="card-body">
              <h5 className="card-title">
                <i className="bi bi-zoom-in text-success me-2"></i>
                Detailed Mode
              </h5>
              <p className="card-text">
                Line-by-line analysis of complex algorithms
              </p>
            </div>
          </div>
        </div>

        <div className="col-12 mb-4">
          <div className="card border-danger">
            <div className="card-body">
              <h5 className="card-title">
                <i className="bi bi-shield-exclamation text-danger me-2"></i>
                Critical Mode
              </h5>
              <p className="card-text">
                Quality evaluation: architecture, security, performance, and testing analysis
              </p>
              <button className="btn btn-danger mt-2">
                Start Critical Review
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-12">
          <div className="alert alert-info">
            <i className="bi bi-info-circle me-2"></i>
            Full review interface coming soon. The existing review functionality is still available at the legacy endpoint.
          </div>
        </div>
      </div>
    </div>
  );
}
