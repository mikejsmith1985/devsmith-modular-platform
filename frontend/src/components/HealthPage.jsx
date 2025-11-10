import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import { Link } from 'react-router-dom';
import { Modal } from 'react-bootstrap';
import StatCards from './StatCards';
import ModelSelector from './ModelSelector';

export default function HealthPage() {
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
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [selectedLog, setSelectedLog] = useState(null);
  const [aiInsights, setAiInsights] = useState(null);
  const [loadingInsights, setLoadingInsights] = useState(false);

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

  const getLevelColor = (level) => {
    const levelLower = level.toLowerCase();
    switch (levelLower) {
      case 'debug': return 'secondary';
      case 'info': return 'info';
      case 'warning': return 'warning';
      case 'error': return 'danger';
      case 'critical': return 'danger';
      default: return 'secondary';
    }
  };

  const openDetailModal = async (log) => {
    setSelectedLog(log);
    setAiInsights(null); // Reset insights when opening new log
    setShowDetailModal(true);
    
    // Check if insights already exist for this log
    await fetchExistingInsights(log.id);
  };

  const fetchExistingInsights = async (logId) => {
    try {
      const response = await fetch(`/api/logs/${logId}/insights`);
      if (response.ok) {
        const data = await response.json();
        setAiInsights(data);
      }
      // If 404, no insights exist yet (that's okay)
    } catch (error) {
      console.error('Error fetching existing insights:', error);
    }
  };

  const generateAIInsights = async (logId) => {
    setLoadingInsights(true);
    try {
      // Call backend to generate AI insights
      const response = await fetch(`/api/logs/${logId}/insights`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          model: selectedModel
        })
      });

      if (!response.ok) {
        throw new Error(`Failed to generate insights: ${response.statusText}`);
      }

      const data = await response.json();
      setAiInsights(data);
    } catch (error) {
      console.error('Error generating AI insights:', error);
      setAiInsights({
        analysis: `Error: ${error.message}`,
        root_cause: null,
        suggestions: []
      });
    } finally {
      setLoadingInsights(false);
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
          <Link to="/health" className="navbar-brand">
            Health
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
                <i className="bi bi-heart-pulse text-primary me-2"></i>
                Health
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
              <li className="nav-item">
                <button
                  className={`nav-link ${activeTab === 'analytics' ? 'active' : ''}`}
                  onClick={() => setActiveTab('analytics')}
                  style={{
                    backgroundColor: activeTab === 'analytics' ? 'rgba(99, 102, 241, 0.2)' : 'transparent',
                    color: activeTab === 'analytics' ? 'var(--bs-primary)' : 'var(--bs-gray-300)',
                    border: 'none',
                    borderBottom: activeTab === 'analytics' ? '2px solid var(--bs-primary)' : 'none'
                  }}
                >
                  <i className="bi bi-graph-up me-2"></i>
                  Analytics
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

                    {/* Logs Cards */}
                    {filteredLogs.length === 0 ? (
                      <div className="text-center py-5" style={{ color: 'var(--bs-gray-400)' }}>
                        <i className="bi bi-inbox display-1 d-block mb-3"></i>
                        <p className="mb-0">No logs found matching your filters</p>
                      </div>
                    ) : (
                      <div className="log-cards-container" style={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '0.5rem'
                      }}>
                        {filteredLogs.map((log) => (
                          <div
                            key={log.id}
                            className="log-card"
                            onClick={() => openDetailModal(log)}
                            style={{
                              background: theme === 'dark' ? 'rgba(255, 255, 255, 0.05)' : 'var(--bs-card-bg)',
                              border: `1px solid ${theme === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'var(--bs-border-color)'}`,
                              borderRadius: '0.5rem',
                              padding: '1rem',
                              cursor: 'pointer',
                              transition: 'all 0.2s'
                            }}
                            onMouseEnter={(e) => {
                              e.currentTarget.style.borderColor = 'var(--bs-primary)';
                              e.currentTarget.style.boxShadow = '0 2px 8px rgba(99, 102, 241, 0.15)';
                              e.currentTarget.style.transform = 'translateY(-1px)';
                            }}
                            onMouseLeave={(e) => {
                              e.currentTarget.style.borderColor = theme === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'var(--bs-border-color)';
                              e.currentTarget.style.boxShadow = 'none';
                              e.currentTarget.style.transform = 'translateY(0)';
                            }}
                          >
                            <div style={{
                              display: 'grid',
                              gridTemplateColumns: '100px 150px 180px 1fr',
                              gap: '1rem',
                              alignItems: 'center'
                            }}>
                              <div>
                                <span className={`badge bg-${getLevelColor(log.level)}`}>
                                  {log.level.toUpperCase()}
                                </span>
                              </div>
                              <div style={{
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                whiteSpace: 'nowrap'
                              }}>
                                <code className="text-primary">{log.service}</code>
                              </div>
                              <div style={{
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                whiteSpace: 'nowrap'
                              }}>
                                <small style={{ color: 'var(--bs-gray-400)' }}>
                                  {formatTimestamp(log.created_at)}
                                </small>
                              </div>
                              <div style={{
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                whiteSpace: 'nowrap'
                              }}>
                                {log.message}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </>
          )}
        </>
      )}

      {activeTab === 'monitoring' && (
        <div className="monitoring-tab text-center py-5">
          <i className="bi bi-activity" style={{ fontSize: '4rem', opacity: 0.3 }}></i>
          <h3 className="mt-3 text-muted">Monitoring Dashboard</h3>
          <p className="text-muted">Coming Soon</p>
          <p className="small text-muted">
            Real-time service health, request rates, and error tracking.
          </p>
        </div>
      )}

      {activeTab === 'analytics' && (
        <div className="analytics-tab text-center py-5">
          <i className="bi bi-graph-up" style={{ fontSize: '4rem', opacity: 0.3 }}></i>
          <h3 className="mt-3 text-muted">Analytics Dashboard</h3>
          <p className="text-muted">Coming Soon</p>
          <p className="small text-muted">
            Historical trends, error patterns, and AI-powered insights.
          </p>
        </div>
      )}

      {/* Log Detail Modal */}
      {selectedLog && (
        <Modal show={showDetailModal} onHide={() => setShowDetailModal(false)} size="lg">
          <Modal.Header closeButton className={theme === 'dark' ? 'bg-dark text-light border-secondary' : ''}>
            <Modal.Title>
              <span className={`badge bg-${getLevelColor(selectedLog.level)} me-2`}>
                {selectedLog.level.toUpperCase()}
              </span>
              Log Details
            </Modal.Title>
          </Modal.Header>
          <Modal.Body className={theme === 'dark' ? 'bg-dark text-light' : ''}>
            {/* Key Info Section */}
            <div className="row mb-3">
              <div className="col-md-6">
                <strong>Service:</strong> <code className="text-primary">{selectedLog.service}</code>
              </div>
              <div className="col-md-6">
                <strong>Timestamp:</strong> {formatTimestamp(selectedLog.created_at)}
              </div>
            </div>

            {/* Message Section */}
            <div className="mb-3">
              <strong>Message:</strong>
              <div className="mt-2 p-3 rounded" style={{
                backgroundColor: theme === 'dark' ? 'rgba(0,0,0,0.3)' : 'rgba(0,0,0,0.05)',
                border: `1px solid ${theme === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)'}`
              }}>
                {selectedLog.message}
              </div>
            </div>

            {/* Metadata Section */}
            {selectedLog.metadata && Object.keys(selectedLog.metadata).length > 0 && (
              <div className="mb-3">
                <strong>Metadata:</strong>
                <pre className="mt-2 p-3 rounded" style={{
                  backgroundColor: theme === 'dark' ? 'rgba(0,0,0,0.3)' : 'rgba(0,0,0,0.05)',
                  color: theme === 'dark' ? 'var(--bs-gray-300)' : 'var(--bs-gray-800)',
                  border: `1px solid ${theme === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)'}`
                }}>
                  {JSON.stringify(selectedLog.metadata, null, 2)}
                </pre>
              </div>
            )}

            {/* Tags Section */}
            {selectedLog.tags && selectedLog.tags.length > 0 && (
              <div className="mb-3">
                <strong>Tags:</strong>
                <div className="mt-2">
                  {selectedLog.tags.map((tag, idx) => (
                    <span key={idx} className="badge bg-secondary me-2">{tag}</span>
                  ))}
                </div>
              </div>
            )}

            {/* AI Insights Section */}
            <div className="mb-3">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <strong>AI Insights:</strong>
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => generateAIInsights(selectedLog.id)}
                  disabled={loadingInsights}
                >
                  {loadingInsights ? (
                    <>
                      <span className="spinner-border spinner-border-sm me-2"></span>
                      Analyzing...
                    </>
                  ) : aiInsights ? (
                    <>
                      <i className="bi bi-arrow-clockwise me-2"></i>
                      Regenerate
                    </>
                  ) : (
                    <>
                      <i className="bi bi-stars me-2"></i>
                      Generate Insights
                    </>
                  )}
                </button>
              </div>
              {aiInsights && (
                <div className="p-3 rounded" style={{
                  backgroundColor: theme === 'dark' ? 'rgba(99,102,241,0.1)' : 'rgba(99,102,241,0.05)',
                  border: '1px solid rgba(99,102,241,0.3)'
                }}>
                  <div className="mb-2">
                    <strong>Analysis:</strong>
                    <p className="mb-2">{aiInsights.analysis}</p>
                  </div>
                  {aiInsights.root_cause && (
                    <div className="mb-2">
                      <strong>Root Cause:</strong>
                      <p className="mb-2">{aiInsights.root_cause}</p>
                    </div>
                  )}
                  {aiInsights.suggestions && aiInsights.suggestions.length > 0 && (
                    <div>
                      <strong>Suggestions:</strong>
                      <ul className="mb-0">
                        {aiInsights.suggestions.map((suggestion, idx) => (
                          <li key={idx}>{suggestion}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}
            </div>
          </Modal.Body>
          <Modal.Footer className={theme === 'dark' ? 'bg-dark border-secondary' : ''}>
            <button className="btn btn-secondary" onClick={() => setShowDetailModal(false)}>
              Close
            </button>
          </Modal.Footer>
        </Modal>
      )}
    </div>
  );
}
