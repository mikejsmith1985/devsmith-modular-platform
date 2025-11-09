import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import { Link } from 'react-router-dom';
import StatCards from './StatCards';
import MonitoringDashboard from './MonitoringDashboard';
import ModelSelector from './ModelSelector';

export default function LogsPage() {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const [activeTab, setActiveTab] = useState('logs');
  const [stats, setStats] = useState({
    debug: 0,
    info: 0,
    warning: 0,
    error: 0,
    critical: 0
  });
  const [logs, setLogs] = useState([]);
  const [filteredLogs, setFilteredLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filters, setFilters] = useState({
    level: 'all',
    service: 'all',
    search: ''
  });
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [selectedModel, setSelectedModel] = useState('');

  useEffect(() => {
    fetchData();
    
    // Auto-refresh every 5 seconds if enabled
    if (autoRefresh && activeTab === 'logs') {
      const interval = setInterval(fetchData, 5000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, activeTab]);

  useEffect(() => {
    applyFilters();
  }, [logs, filters]);

  const fetchData = async () => {
    try {
      setLoading(true);
      
      // Fetch both stats and logs in parallel
      const [statsResponse, logsResponse] = await Promise.all([
        fetch('http://localhost:3000/api/logs/v1/stats', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('devsmith_token')}`
          }
        }),
        fetch('http://localhost:3000/api/logs?limit=100', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('devsmith_token')}`
          }
        })
      ]);

      if (!statsResponse.ok) {
        throw new Error('Failed to fetch stats');
      }
      
      if (!logsResponse.ok) {
        throw new Error('Failed to fetch logs');
      }

      const statsData = await statsResponse.json();
      const logsData = await logsResponse.json();
      
      setStats(statsData);
      setLogs(logsData.entries || []);
      setError(null);
    } catch (err) {
      console.error('Error fetching data:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const applyFilters = () => {
    let filtered = [...logs];
    
    // Filter by level
    if (filters.level !== 'all') {
      filtered = filtered.filter(log => log.level.toLowerCase() === filters.level.toLowerCase());
    }
    
    // Filter by service
    if (filters.service !== 'all') {
      filtered = filtered.filter(log => log.service === filters.service);
    }
    
    // Filter by search term
    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      filtered = filtered.filter(log => 
        log.message.toLowerCase().includes(searchLower) ||
        log.service.toLowerCase().includes(searchLower)
      );
    }
    
    setFilteredLogs(filtered);
  };

  const getUniqueServices = () => {
    const services = new Set(logs.map(log => log.service));
    return Array.from(services).sort();
  };

  const getLevelBadgeClass = (level) => {
    const levelLower = level.toLowerCase();
    switch (levelLower) {
      case 'debug': return 'badge bg-secondary';
      case 'info': return 'badge bg-info';
      case 'warning': return 'badge bg-warning text-dark';
      case 'error': return 'badge bg-danger';
      case 'critical': return 'badge bg-danger';
      default: return 'badge bg-secondary';
    }
  };

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
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
            <div className="me-3">
              <ModelSelector
                selectedModel={selectedModel}
                onModelSelect={setSelectedModel}
              />
            </div>
            <button
              className="btn btn-link p-2 me-3"
              onClick={toggleTheme}
              title="Toggle Dark Mode"
              style={{ fontSize: '1.25rem' }}
            >
              <i className={`bi bi-${theme === 'dark' ? 'sun' : 'moon'}-fill`}></i>
            </button>
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
                    <div className="d-flex justify-content-between align-items-center mb-3">
                      <h5 className="mb-0">Recent Logs ({filteredLogs.length})</h5>
                      <div className="d-flex gap-2 align-items-center">
                        <div className="form-check form-switch">
                          <input
                            className="form-check-input"
                            type="checkbox"
                            id="autoRefresh"
                            checked={autoRefresh}
                            onChange={(e) => setAutoRefresh(e.target.checked)}
                          />
                          <label className="form-check-label" htmlFor="autoRefresh">
                            Auto-refresh
                          </label>
                        </div>
                        <button 
                          className="btn btn-sm btn-outline-primary"
                          onClick={() => fetchData()}
                          disabled={loading}
                        >
                          <i className="bi bi-arrow-clockwise me-1"></i>
                          Refresh
                        </button>
                      </div>
                    </div>

                    {/* Filters */}
                    <div className="row mb-3">
                      <div className="col-md-3">
                        <select
                          className="form-select form-select-sm"
                          value={filters.level}
                          onChange={(e) => setFilters({ ...filters, level: e.target.value })}
                        >
                          <option value="all">All Levels</option>
                          <option value="debug">Debug</option>
                          <option value="info">Info</option>
                          <option value="warning">Warning</option>
                          <option value="error">Error</option>
                          <option value="critical">Critical</option>
                        </select>
                      </div>
                      <div className="col-md-3">
                        <select
                          className="form-select form-select-sm"
                          value={filters.service}
                          onChange={(e) => setFilters({ ...filters, service: e.target.value })}
                        >
                          <option value="all">All Services</option>
                          {getUniqueServices().map(service => (
                            <option key={service} value={service}>{service}</option>
                          ))}
                        </select>
                      </div>
                      <div className="col-md-6">
                        <input
                          type="text"
                          className="form-control form-control-sm"
                          placeholder="Search logs..."
                          value={filters.search}
                          onChange={(e) => setFilters({ ...filters, search: e.target.value })}
                        />
                      </div>
                    </div>

                    {/* Logs Table */}
                    {filteredLogs.length === 0 ? (
                      <div className="text-center py-5" style={{ color: 'var(--bs-gray-400)' }}>
                        <i className="bi bi-inbox display-1 d-block mb-3"></i>
                        <p className="mb-0">No logs found matching your filters</p>
                      </div>
                    ) : (
                      <div className="table-responsive">
                        <table className="table table-hover">
                          <thead>
                            <tr>
                              <th style={{ width: '10%' }}>Level</th>
                              <th style={{ width: '15%' }}>Service</th>
                              <th style={{ width: '15%' }}>Timestamp</th>
                              <th style={{ width: '60%' }}>Message</th>
                            </tr>
                          </thead>
                          <tbody>
                            {filteredLogs.map((log) => (
                              <tr key={log.id}>
                                <td>
                                  <span className={getLevelBadgeClass(log.level)}>
                                    {log.level.toUpperCase()}
                                  </span>
                                </td>
                                <td>
                                  <code className="text-primary">{log.service}</code>
                                </td>
                                <td>
                                  <small style={{ color: 'var(--bs-gray-400)' }}>
                                    {formatTimestamp(log.created_at)}
                                  </small>
                                </td>
                                <td>
                                  {log.message}
                                  {log.metadata && Object.keys(log.metadata).length > 0 && (
                                    <details className="mt-1">
                                      <summary style={{ cursor: 'pointer', color: 'var(--bs-gray-400)' }}>
                                        <small>Metadata</small>
                                      </summary>
                                      <pre 
                                        className="p-2 rounded mt-2" 
                                        style={{ 
                                          fontSize: '0.75rem',
                                          backgroundColor: theme === 'dark' ? 'rgba(0, 0, 0, 0.3)' : 'rgba(0, 0, 0, 0.05)',
                                          color: theme === 'dark' ? 'var(--bs-gray-300)' : 'var(--bs-gray-800)',
                                          border: theme === 'dark' ? '1px solid rgba(255, 255, 255, 0.1)' : '1px solid rgba(0, 0, 0, 0.1)'
                                        }}
                                      >
                                        {JSON.stringify(log.metadata, null, 2)}
                                      </pre>
                                    </details>
                                  )}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
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
