import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import { Link } from 'react-router-dom';
import { Modal } from 'react-bootstrap';
import StatCards from './StatCards';
import ModelSelector from './ModelSelector';
import TagFilter from './TagFilter';
import { logError, logWarning, logInfo, logDebug } from '../utils/logger';
import { apiRequest } from '../utils/api';

// Helper function to classify log severity and detect critical events
const classifyLogSeverity = (log) => {
  const level = (log.level || 'info').toLowerCase();
  const message = (log.message || '').toLowerCase();
  const metadata = log.metadata || {};
  const metadataStr = JSON.stringify(metadata).toLowerCase();
  
  // Critical keywords that indicate high severity
  const criticalKeywords = ['critical', 'fatal', 'panic', 'emergency', 'down', 'crash', 'failed to start'];
  const isCritical = criticalKeywords.some(keyword => 
    message.includes(keyword) || metadataStr.includes(keyword)
  );
  
  // Return appropriate severity level
  if (isCritical || level === 'critical') return 'critical';
  if (level === 'error') return 'error';
  if (level === 'warn') return 'warning';
  return level;
};

export default function HealthPage() {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const [activeTab, setActiveTab] = useState('logs');
  // Unfiltered stats - always shows total counts regardless of active filters
  const [unfilteredStats, setUnfilteredStats] = useState({
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
    project: 'all',  // Week 3: Add project filter
    search: ''
  });
  const [autoRefresh, setAutoRefresh] = useState(false);  // OFF by default - Phase 1 fix
  const [selectedModel, setSelectedModel] = useState('');
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [selectedLog, setSelectedLog] = useState(null);
  const [aiInsights, setAiInsights] = useState(null);
  const [loadingInsights, setLoadingInsights] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);  // Phase 4: Prevent concurrent AI requests
  
  // Phase 3: WebSocket connection state
  const [wsConnected, setWsConnected] = useState(false);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null); // Track reconnect timeout for cleanup
  
  // Phase 3: Smart Tagging System
  const [availableTags, setAvailableTags] = useState([]);
  const [selectedTags, setSelectedTags] = useState([]);
  
  // Phase 3: Manual Tag Management
  const [newTagInput, setNewTagInput] = useState('');
  
  // Week 3: Cross-repo logging - project management
  const [projects, setProjects] = useState([]);
  const [loadingProjects, setLoadingProjects] = useState(false);
  const [addingTag, setAddingTag] = useState(false);

  useEffect(() => {
    // Phase 5: Batch all initial API calls in parallel for faster page load
    const loadInitialData = async () => {
      try {
        setLoading(true);
        const [statsData, logsData, tagsData] = await Promise.all([
          apiRequest('/api/logs/v1/stats'),
          apiRequest('/api/logs?limit=100'),
          apiRequest('/api/logs/tags')
        ]);
      
        const entries = logsData.entries || [];
        
        // Store unfiltered stats from API (always shows total database counts)
        setUnfilteredStats(statsData);
        
        setLogs(entries);
        setAvailableTags(tagsData.tags || []);
        setError(null);
    } catch (err) {
      logError(err, { context: 'Health page initial data load failed' });
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };
      loadInitialData();
  }, [activeTab]); // Remove autoRefresh from dependencies - WebSocket handles updates

  // Phase 3: WebSocket connection management
  useEffect(() => {
    if (!autoRefresh) {
    // Disconnect WebSocket when auto-refresh OFF
    if (wsRef.current) {
      logDebug('WebSocket disconnecting (auto-refresh disabled)');
      wsRef.current.close();
      wsRef.current = null;
      setWsConnected(false);
    }
    return;
  }

  // Connect WebSocket when auto-refresh ON
  const connectWebSocket = () => {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws/logs`;
    
    logDebug('WebSocket connecting', { url: wsUrl });
    const ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      logInfo('WebSocket connection established', { autoRefresh });
      setWsConnected(true);
    };
    
    ws.onmessage = (event) => {
      try {
        const newLog = JSON.parse(event.data);
        logDebug('WebSocket received log', { logId: newLog.id, level: newLog.level });
        
        // Add new log to the top of the list (limit to 100)
        setLogs(prev => [newLog, ...prev].slice(0, 100));
        
        // Update unfiltered stats incrementally
        setUnfilteredStats(prev => ({
          ...prev,
          [newLog.level.toLowerCase()]: (prev[newLog.level.toLowerCase()] || 0) + 1
        }));
      } catch (error) {
        logError(error, { context: 'WebSocket message parsing failed' });
      }
    };
    
    ws.onerror = (error) => {
      logError(new Error('WebSocket error'), { errorEvent: error.toString() });
      setWsConnected(false);
    };
    
    ws.onclose = () => {
      logInfo('WebSocket connection closed');
      setWsConnected(false);
      
      // Reconnect after 5 seconds if auto-refresh still enabled
      if (autoRefresh) {
        logDebug('WebSocket reconnecting in 5 seconds');
        reconnectTimeoutRef.current = setTimeout(connectWebSocket, 5000);
      }
    };
    
    wsRef.current = ws;
  };

  connectWebSocket();

  return () => {
    // Clear any pending reconnect timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    // Close WebSocket connection
    if (wsRef.current) {
      logDebug('WebSocket cleanup - closing connection');
      wsRef.current.close();
    }
  };
}, [autoRefresh]);  // Define fetchData with useCallback to prevent infinite loops
  const fetchData = useCallback(async (isBackgroundRefresh = false) => {
    try {
      // Only show loading spinner on initial load, not during background refresh
      if (!isBackgroundRefresh) {
        setLoading(true);
      }
      
      // Build query string with level filter if set
      let logsQuery = '/api/logs?limit=100';
      if (filters.level !== 'all') {
        logsQuery += `&level=${filters.level}`;
      }
      if (filters.service !== 'all') {
        logsQuery += `&service=${filters.service}`;
      }
      // Week 3: Add project filter for cross-repo logging
      if (filters.project !== 'all') {
        logsQuery += `&project_id=${filters.project}`;
      }
      
      // Fetch stats and logs in parallel
      const [statsData, logsData] = await Promise.all([
        apiRequest('/api/logs/v1/stats'),
        apiRequest(logsQuery)
      ]);
      
      const entries = logsData.entries || [];
      
      // Store unfiltered stats from API (always shows total database counts)
      setUnfilteredStats(statsData);
      
      setLogs(entries);
      setError(null);
    setError(null);
  } catch (err) {
    logError(err, { context: 'Health page data fetch failed' });
    setError(err.message);
  } finally {
    setLoading(false);
  }
}, [filters.level, filters.service]); // Only depend on filter values we use

// Refetch data when level or service filters change
useEffect(() => {
  fetchData();
}, [fetchData]);

useEffect(() => {
  applyFilters();
}, [logs, filters, selectedTags]); // Depend on actual values, not the callback

// Phase 3: Fetch available tags
const fetchAvailableTags = async () => {
  try {
    const data = await apiRequest('/api/logs/tags');
    setAvailableTags(data.tags || []);
  } catch (error) {
    logWarning('Failed to fetch log tags', { error: error.message });
  }
};

// Week 3: Fetch projects for cross-repo logging
const fetchProjects = async () => {
  try {
    setLoadingProjects(true);
    const data = await apiRequest('/api/logs/projects');
    setProjects(Array.isArray(data) ? data : data.projects || []);
  } catch (error) {
    logWarning('Failed to fetch projects', { error: error.message });
    setProjects([]);
  } finally {
    setLoadingProjects(false);
  }
};

useEffect(() => {
  fetchAvailableTags();
  fetchProjects();  // Week 3: Fetch projects on mount
}, []);  // Phase 3: Toggle tag selection
  const handleTagToggle = (tag) => {
    setSelectedTags(prev => {
      if (prev.includes(tag)) {
        return prev.filter(t => t !== tag);
      } else {
        return [...prev, tag];
      }
    });
  };


  const applyFilters = useCallback(() => {
    let filtered = [...logs];
    
    // NOTE: Level and service filtering already handled by backend API
    // Backend filters: /api/logs?level=${filters.level}&service=${filters.service}
    // Frontend only needs to filter by search terms and tags (not handled by backend)
    
    // Filter by search term
    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      filtered = filtered.filter(log => 
        log.message.toLowerCase().includes(searchLower) ||
        log.service.toLowerCase().includes(searchLower)
      );
    }
    
    // Phase 3: Filter by tags (AND logic - log must have ALL selected tags)
    if (selectedTags.length > 0) {
      filtered = filtered.filter(log => {
        if (!log.tags || log.tags.length === 0) return false;
        return selectedTags.every(tag => log.tags.includes(tag));
      });
    }
    
    setFilteredLogs(filtered);
  }, [logs, filters, selectedTags]);

  // Week 3: Filter services by selected project
  const getUniqueServices = () => {
    // If a project is selected, only show services from that project's logs
    const logsToFilter = filters.project !== 'all' 
      ? logs.filter(log => log.project_id === parseInt(filters.project))
      : logs;
    
    const services = new Set(logsToFilter.map(log => log.service));
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
      // Phase 4 Fix: Use apiRequest() for connection pooling
      const data = await apiRequest(`/api/logs/${logId}/insights`);
      logDebug('Fetched existing AI insights', { logId, insightsCount: data ? 1 : 0 });
      setAiInsights(data);
      
      logInfo('Fetched existing AI insights', {
        log_id: logId,
        action: 'fetch_insights_success'
      });
    } catch (error) {
      // If 404, no insights exist yet (that's okay, no need to log)
      if (error.status === 404) {
        logDebug('No existing insights found for log', { logId });
        return;
      }
      
      logWarning('Failed to fetch existing insights', {
        log_id: logId,
        error: error.message,
        action: 'fetch_insights_error'
      });
    }
  };

  const generateAIInsights = async (logId) => {
    // Phase 4 Fix: Prevent concurrent AI requests (debouncing)
    if (isGenerating) {
      logDebug('Already generating insights, ignoring duplicate request');
      return;
    }
    
    setLoadingInsights(true);
    setIsGenerating(true);
    
    try {
      logInfo('Generating AI insights', {
        log_id: logId,
        model: selectedModel,
        action: 'generate_insights_start'
      });

      // Phase 4 Fix: Use apiRequest() for connection pooling and timeout handling
      const data = await apiRequest(`/api/logs/${logId}/insights`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          model: selectedModel
        }),
        timeout: 60000  // 60 second timeout
      });

      logDebug('AI insights response received', { logId, hasAnalysis: !!data?.analysis });
      setAiInsights(data);
      
      logInfo('AI insights generated successfully', {
        log_id: logId,
        model: selectedModel,
        action: 'generate_insights_success'
      });
    } catch (error) {
      
      // Check if it was a timeout
      if (error.name === 'AbortError' || error.message?.includes('timeout')) {
        const timeoutMsg = 'AI analysis timed out after 60 seconds. Try a smaller/faster model or retry when server is less busy.';
        
        logError(new Error(timeoutMsg), {
          log_id: logId,
          model: selectedModel,
          action: 'generate_insights_timeout'
        });
        
        setAiInsights({
          analysis: `⏱️ ${timeoutMsg}`,
          root_cause: 'Request exceeded 60 second timeout limit',
          suggestions: [
            'Try a smaller model like qwen2.5-coder:7b-instruct-q4_K_M',
            'Check server logs for model loading issues',
            'Retry when server is less busy'
          ]
        });
      } else {
        // Log the actual error with full context
        logError(error, {
          log_id: logId,
          model: selectedModel,
          status_code: error.status,
          error_message: error.message,
          action: 'generate_insights_failed'
        });
        
        setAiInsights({
          analysis: `❌ Failed to generate insights: ${error.message}`,
          root_cause: 'AI service error',
          suggestions: [
            'Check that the AI model is running',
            'Verify model name is correct',
            'Check server logs for details'
          ]
        });
      }
    } finally {
      setLoadingInsights(false);
      setIsGenerating(false);  // Phase 4: Re-enable after completion
    }
  };

  // Phase 3: Add tag to log entry
  const handleAddTag = async (logId, tag) => {
    if (!tag || tag.trim() === '') return;
    
    setAddingTag(true);
    try {
      await apiRequest(`/api/logs/${logId}/tags`, {
        method: 'POST',
        body: { tag: tag.trim() }
      });

      // Update the selected log with new tag
      const updatedLog = {
        ...selectedLog,
        tags: [...(selectedLog.tags || []), tag.trim()]
      };
      setSelectedLog(updatedLog);

      // Update the log in the main list
      setLogs(prevLogs =>
        prevLogs.map(log =>
          log.id === logId ? updatedLog : log
        )
      );

      // Refresh available tags
      await fetchAvailableTags();

      // Clear input
      setNewTagInput('');
    } catch (error) {
      logError(error, {
        log_id: logId,
        tag,
        action: 'add_tag_failed'
      });
      alert(`Failed to add tag: ${error.message}`);
    } finally {
      setAddingTag(false);
    }
  };

  // Phase 3: Remove tag from log entry
  const handleRemoveTag = async (logId, tag) => {
    try {
      await apiRequest(`/api/logs/${logId}/tags/${encodeURIComponent(tag)}`, {
        method: 'DELETE'
      });

      // Update the selected log without the removed tag
      const updatedLog = {
        ...selectedLog,
        tags: (selectedLog.tags || []).filter(t => t !== tag)
      };
      setSelectedLog(updatedLog);

      // Update the log in the main list
      setLogs(prevLogs =>
        prevLogs.map(log =>
          log.id === logId ? updatedLog : log
        )
      );

      // Refresh available tags
      await fetchAvailableTags();
    } catch (error) {
      logError(error, {
        log_id: logId,
        tag,
        action: 'remove_tag_failed'
      });
      alert(`Failed to remove tag: ${error.message}`);
    }
  };

  // Phase 5: Show loading spinner during initial data fetch
  if (loading && logs.length === 0) {
    return (
      <div className="container mt-4">
        <div className="text-center py-5">
          <div className="spinner-border text-primary" role="status" style={{ width: '3rem', height: '3rem' }}>
            <span className="visually-hidden">Loading...</span>
          </div>
          <p className="mt-3 text-muted">Loading health data...</p>
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
              {/* Phase 1: Card-Based Dashboard Layout */}
              <div className="row g-3">
                {/* Left Column: Main Logs Feed (8 columns) */}
                <div className="col-lg-8">
                  {/* Stats Cards - Horizontal on main column */}
                  <div className="mb-3">
                    <StatCards 
                      stats={unfilteredStats} 
                      selectedLevel={filters.level === 'all' ? null : filters.level}
                      onLevelClick={(level) => {
                        setFilters({ 
                          ...filters, 
                          level: filters.level === level ? 'all' : level 
                        });
                      }}
                    />
                  </div>

                  {/* Logs Feed Card */}
                  <div className="frosted-card p-4">
                    <div className="d-flex justify-content-between align-items-center mb-3">
                      <h5 className="mb-0">
                        <i className="bi bi-list-ul me-2"></i>
                        Logs Feed ({filteredLogs.length})
                      </h5>
                      <div className="d-flex gap-2 align-items-center">
                        {/* Phase 3: WebSocket connection status indicator */}
                        {autoRefresh && (
                          <span className={`badge ${wsConnected ? 'bg-success' : 'bg-secondary'}`}>
                            <i className={`bi bi-${wsConnected ? 'check-circle' : 'x-circle'} me-1`}></i>
                            {wsConnected ? 'Connected' : 'Disconnected'}
                          </span>
                        )}
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
                      {/* Week 3: Project filter for cross-repo logging */}
                      <div className="col-md-3">
                        <select
                          className="form-select form-select-sm bg-dark text-light border-secondary"
                          value={filters.project}
                          onChange={(e) => setFilters({ ...filters, project: e.target.value })}
                          disabled={loadingProjects}
                        >
                          <option value="all">All Projects</option>
                          {projects.map(project => (
                            <option key={project.id} value={project.id}>
                              {project.name}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div className="col-md-3">
                        <select
                          className="form-select form-select-sm bg-dark text-light border-secondary"
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
                          className="form-control form-control-sm bg-dark text-light border-secondary"
                          placeholder="Search logs..."
                          value={filters.search}
                          onChange={(e) => setFilters({ ...filters, search: e.target.value })}
                        />
                      </div>
                    </div>

                    {/* Phase 3: Tag Filter */}
                    <div className="mb-3">
                      <TagFilter
                        availableTags={availableTags}
                        selectedTags={selectedTags}
                        onTagToggle={handleTagToggle}
                      />
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

                {/* Right Column: Quick Info Sidebar (4 columns) */}
                <div className="col-lg-4">
                  {/* Quick Stats Card */}
                  <div className="frosted-card p-3 mb-3">
                    <h6 className="mb-3">
                      <i className="bi bi-speedometer2 me-2"></i>
                      Quick Stats
                    </h6>
                    <div className="d-flex flex-column gap-2">
                      <div className="d-flex justify-content-between align-items-center p-2 rounded" style={{ backgroundColor: 'rgba(99, 102, 241, 0.05)' }}>
                        <span className="small">Total Logs</span>
                        <strong>{logs.length}</strong>
                      </div>
                      <div className="d-flex justify-content-between align-items-center p-2 rounded" style={{ backgroundColor: 'rgba(220, 38, 38, 0.05)' }}>
                        <span className="small">Errors</span>
                        <strong className="text-danger">{stats.error + stats.critical}</strong>
                      </div>
                      <div className="d-flex justify-content-between align-items-center p-2 rounded" style={{ backgroundColor: 'rgba(234, 179, 8, 0.05)' }}>
                        <span className="small">Warnings</span>
                        <strong className="text-warning">{stats.warning}</strong>
                      </div>
                      <div className="d-flex justify-content-between align-items-center p-2 rounded" style={{ backgroundColor: 'rgba(59, 130, 246, 0.05)' }}>
                        <span className="small">Info</span>
                        <strong className="text-info">{stats.info}</strong>
                      </div>
                      <div className="d-flex justify-content-between align-items-center p-2 rounded" style={{ backgroundColor: 'rgba(34, 197, 94, 0.05)' }}>
                        <span className="small">Debug</span>
                        <strong className="text-success">{stats.debug}</strong>
                      </div>
                    </div>
                  </div>

                  {/* Active Filters Card */}
                  {(filters.level !== 'all' || filters.service !== 'all' || filters.search || selectedTags.length > 0) && (
                    <div className="frosted-card p-3 mb-3">
                      <h6 className="mb-3">
                        <i className="bi bi-funnel me-2"></i>
                        Active Filters
                      </h6>
                      <div className="d-flex flex-column gap-2">
                        {filters.level !== 'all' && (
                          <div className="d-flex justify-content-between align-items-center">
                            <span className="small">Level:</span>
                            <span className={`badge bg-${getLevelColor(filters.level)}`}>
                              {filters.level.toUpperCase()}
                            </span>
                          </div>
                        )}
                        {filters.service !== 'all' && (
                          <div className="d-flex justify-content-between align-items-center">
                            <span className="small">Service:</span>
                            <code className="small text-primary">{filters.service}</code>
                          </div>
                        )}
                        {filters.search && (
                          <div className="d-flex justify-content-between align-items-center">
                            <span className="small">Search:</span>
                            <code className="small">{filters.search}</code>
                          </div>
                        )}
                        {selectedTags.length > 0 && (
                          <div className="d-flex flex-column gap-1">
                            <span className="small">Tags:</span>
                            <div className="d-flex flex-wrap gap-1">
                              {selectedTags.map(tag => (
                                <span key={tag} className="badge bg-secondary small">
                                  {tag}
                                </span>
                              ))}
                            </div>
                          </div>
                        )}
                        <button 
                          className="btn btn-sm btn-outline-secondary mt-2"
                          onClick={() => {
                            setFilters({ level: 'all', service: 'all', search: '' });
                            setSelectedTags([]);
                          }}
                        >
                          <i className="bi bi-x-circle me-1"></i>
                          Clear All
                        </button>
                      </div>
                    </div>
                  )}

                  {/* Recent Critical Events Card */}
                  <div className="frosted-card p-3">
                    <h6 className="mb-3">
                      <i className="bi bi-exclamation-triangle-fill text-danger me-2"></i>
                      Critical Events
                    </h6>
                    <div className="d-flex flex-column gap-2">
                      {logs.filter(log => log.level === 'error' || log.level === 'critical')
                        .slice(0, 5)
                        .map(log => (
                          <div 
                            key={log.id}
                            className="p-2 rounded"
                            style={{ 
                              backgroundColor: 'rgba(220, 38, 38, 0.05)',
                              borderLeft: '3px solid var(--bs-danger)',
                              cursor: 'pointer'
                            }}
                            onClick={() => handleViewDetails(log)}
                          >
                            <div className="d-flex justify-content-between align-items-start mb-1">
                              <small className="text-muted">{formatTimestamp(log.created_at)}</small>
                              <span className={`badge badge-sm bg-${getLevelColor(log.level)}`}>
                                {log.level}
                              </span>
                            </div>
                            <div className="small" style={{ 
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              display: '-webkit-box',
                              WebkitLineClamp: 2,
                              WebkitBoxOrient: 'vertical'
                            }}>
                              {log.message}
                            </div>
                          </div>
                        ))}
                      {logs.filter(log => log.level === 'error' || log.level === 'critical').length === 0 && (
                        <div className="text-center py-3 text-muted small">
                          <i className="bi bi-check-circle me-1"></i>
                          No critical events
                        </div>
                      )}
                    </div>
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

            {/* Tags Section - Phase 3: Enhanced with Manual Management */}
            <div className="mb-3">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <strong>Tags:</strong>
              </div>
              
              {/* Display existing tags with remove button */}
              <div className="mb-2">
                {selectedLog.tags && selectedLog.tags.length > 0 ? (
                  <div className="d-flex flex-wrap gap-2">
                    {selectedLog.tags.map((tag, idx) => (
                      <span 
                        key={idx} 
                        className="badge bg-secondary d-flex align-items-center"
                        style={{ fontSize: '0.9rem' }}
                      >
                        {tag}
                        <button
                          className="btn-close btn-close-white ms-2"
                          style={{ fontSize: '0.6rem' }}
                          onClick={() => handleRemoveTag(selectedLog.id, tag)}
                          title="Remove tag"
                          aria-label={`Remove ${tag} tag`}
                        ></button>
                      </span>
                    ))}
                  </div>
                ) : (
                  <div className="text-muted small">No tags yet</div>
                )}
              </div>

              {/* Add new tag input */}
              <div className="input-group input-group-sm">
                <input
                  type="text"
                  className={`form-control ${theme === 'dark' ? 'bg-dark text-light border-secondary' : ''}`}
                  placeholder="Add a tag (e.g., investigated, resolved)"
                  value={newTagInput}
                  onChange={(e) => setNewTagInput(e.target.value)}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      handleAddTag(selectedLog.id, newTagInput);
                    }
                  }}
                  disabled={addingTag}
                />
                <button
                  className="btn btn-outline-primary btn-sm"
                  onClick={() => handleAddTag(selectedLog.id, newTagInput)}
                  disabled={addingTag || !newTagInput.trim()}
                >
                  {addingTag ? (
                    <>
                      <span className="spinner-border spinner-border-sm me-1"></span>
                      Adding...
                    </>
                  ) : (
                    <>
                      <i className="bi bi-plus-circle me-1"></i>
                      Add Tag
                    </>
                  )}
                </button>
              </div>
              <small className="text-muted">
                Tags help categorize and filter logs. Press Enter or click Add Tag.
              </small>
            </div>

            {/* AI Insights Section */}
            <div className="mb-3">
              <div className="d-flex justify-content-between align-items-center mb-2">
                <strong>AI Insights:</strong>
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => generateAIInsights(selectedLog.id)}
                  disabled={loadingInsights || isGenerating}
                >
                  {(loadingInsights || isGenerating) ? (
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
// Force rebuild 1763072492
