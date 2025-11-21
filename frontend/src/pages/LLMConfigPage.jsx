import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useTheme } from '../context/ThemeContext';
import { useAuth } from '../context/AuthContext';
import { apiRequest } from '../utils/api';
import AddLLMConfigModal from '../components/AddLLMConfigModal';

/**
 * LLMConfigPage Component
 * 
 * Phase 5, Task 5.2 - Full Implementation (GREEN phase)
 * 
 * Allows users to manage AI model configurations including:
 * - Add/Edit/Delete LLM configs (Anthropic, OpenAI, Ollama, etc.)
 * - Set app-specific preferences
 * - View usage statistics
 * - Test API connections
 * 
 * @component
 */
export default function LLMConfigPage() {
  const { isDarkMode, toggleTheme } = useTheme();
  const { user } = useAuth();
  
  // State
  const [configs, setConfigs] = useState([]);
  const [appPreferences, setAppPreferences] = useState({});
  const [usageSummary, setUsageSummary] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [editingConfig, setEditingConfig] = useState(null);
  const [deletingConfigId, setDeletingConfigId] = useState(null);
  
  // Load data on mount
  useEffect(() => {
    loadConfigs();
    loadAppPreferences();
    loadUsageSummary();
  }, []);
  
  const loadConfigs = async () => {
    try {
      setLoading(true);
      const data = await apiRequest('/api/portal/llm-configs');
      setConfigs(data || []);
      setError(null);
    } catch (err) {
      console.error('Failed to load configs:', err);
      setError('Failed to load AI model configurations');
      setConfigs([]);
    } finally {
      setLoading(false);
    }
  };
  
  const loadAppPreferences = async () => {
    try {
      const data = await apiRequest('/api/portal/app-llm-preferences');
      setAppPreferences(data || {});
    } catch (err) {
      console.error('Failed to load app preferences:', err);
      setAppPreferences({});
    }
  };
  
  const loadUsageSummary = async () => {
    try {
      const data = await apiRequest('/api/portal/llm-usage/summary?period=30d');
      setUsageSummary(data);
    } catch (err) {
      console.error('Failed to load usage summary:', err);
      setUsageSummary(null);
    }
  };
  
  const handleDeleteConfig = async (configId) => {
    try {
      await apiRequest(`/api/portal/llm-configs/${configId}`, { method: 'DELETE' });
      await loadConfigs();
      setDeletingConfigId(null);
    } catch (err) {
      console.error('Failed to delete config:', err);
      alert('Failed to delete configuration: ' + (err.message || 'Unknown error'));
    }
  };
  
  const handleSetAppPreference = async (appName, configId) => {
    try {
      await apiRequest(`/api/portal/app-llm-preferences/${appName}`, {
        method: 'PUT',
        body: JSON.stringify({ llm_config_id: configId })
      });
      await loadAppPreferences();
    } catch (err) {
      console.error('Failed to set app preference:', err);
      alert('Failed to set app preference');
    }
  };

  const handleSetDefault = async (configId) => {
    try {
      setSettingDefault(configId);
      await apiRequest(`/api/portal/llm-configs/${configId}/set-default`, {
        method: 'PUT',
        body: JSON.stringify({ is_default: true })
      });
      await loadConfigs();
    } catch (err) {
      console.error('Failed to set default:', err);
      alert('Failed to set default configuration: ' + (err.message || 'Unknown error'));
    } finally {
      setSettingDefault(null);
    }
  };  const handleToggleDefault = async (configId, currentDefault) => {
    try {
      await apiRequest(`/api/portal/llm-configs/${configId}/set-default`, {
        method: 'PUT',
        body: JSON.stringify({ is_default: !currentDefault })
      });
      await loadConfigs();
    } catch (err) {
      console.error('Failed to set default:', err);
      alert('Failed to set default configuration: ' + (err.message || 'Unknown error'));
    }
  };

  const handleSaveConfig = async (configData) => {
    try {
      // Save the configuration to backend
      await apiClient.post('/api/portal/llm-configs', configData);
      
      // Refresh the config list
      await loadConfigs();
      setShowAddModal(false);
      setEditingConfig(null);
    } catch (err) {
      console.error('Error saving config:', err);
      setError(err.message || 'Failed to save configuration');
    }
  };

  return (
    <div className={`container mt-4 ${isDarkMode ? 'text-light' : ''}`}>
      {/* Navigation Bar */}
      <nav className={`navbar navbar-expand-lg mb-4 frosted-card ${isDarkMode ? 'navbar-dark bg-dark border-secondary' : 'navbar-light'}`}>
        <div className="container-fluid">
          <div className="d-flex align-items-center gap-3">
            <Link to="/portal" className={`btn btn-sm ${isDarkMode ? 'btn-outline-light' : 'btn-outline-secondary'}`}>
              <i className="bi bi-arrow-left me-1"></i>
              Back to Dashboard
            </Link>
            <span className="navbar-brand fw-bold mb-0" style={{ fontSize: '1.5rem', color: isDarkMode ? '#e0e7ff' : '#1e293b' }}>
              <i className="bi bi-code-square me-2"></i>
              DevSmith Platform
            </span>
          </div>
          <div className="d-flex align-items-center gap-3">
            <button
              onClick={toggleTheme}
              className={`btn btn-sm ${isDarkMode ? 'btn-outline-light' : 'btn-outline-dark'}`}
              title={isDarkMode ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
            >
              <i className={`bi bi-${isDarkMode ? 'sun' : 'moon'}-fill`}></i>
            </button>
          </div>
        </div>
      </nav>

      {/* Page Header */}
      <div className="row">
        <div className="col-12 mb-4">
          <div className={`card shadow-sm ${isDarkMode ? 'bg-dark text-light border-secondary' : ''}`}>
            <div className="card-body">
              <h2 className={`card-title mb-3 ${isDarkMode ? 'text-light' : ''}`}>
                <i className="bi bi-robot me-2" style={{ fontSize: '1.5rem', verticalAlign: 'middle' }}></i>
                AI Factory
              </h2>
              <p className={`card-text mb-0 ${isDarkMode ? 'text-light' : 'text-muted'}`}>
                Manage your AI model configurations, API keys, and app-specific preferences.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="row">
          <div className="col-12 mb-4">
            <div className="alert alert-danger">
              <i className="bi bi-exclamation-triangle me-2"></i>
              {error}
            </div>
          </div>
        </div>
      )}

      {/* Your AI Models Section */}
      <div className="row">
        <div className="col-12 mb-4">
          <div className={`frosted-card p-4 ${isDarkMode ? 'bg-dark text-light border-secondary' : ''}`}>
            <div className="d-flex justify-content-between align-items-center mb-3">
              <h4 className={`mb-0 ${isDarkMode ? 'text-light' : ''}`}>Your AI Models</h4>
              <button 
                className="btn btn-primary btn-sm"
                onClick={() => setShowAddModal(true)}
              >
                <i className="bi bi-plus-circle me-1"></i>
                Add Model
              </button>
            </div>
            
            {loading ? (
              <div className="text-center py-4">
                <div className="spinner-border" role="status">
                  <span className="visually-hidden">Loading...</span>
                </div>
              </div>
            ) : configs.length === 0 ? (
              <div className={`alert ${isDarkMode ? 'alert-secondary' : 'alert-info'} mb-0`}>
                <i className="bi bi-info-circle me-2"></i>
                No AI models configured yet. Click "Add Model" to get started.
              </div>
            ) : (
              <div className="table-responsive">
                <table className={`table table-hover ${isDarkMode ? 'table-dark' : ''}`}>
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Provider</th>
                      <th>Model</th>
                      <th>API Key</th>
                      <th>Default</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {configs.map((config) => (
                      <tr key={config.id}>
                        <td>{config.name}</td>
                        <td>
                          <span className="badge bg-secondary">
                            {config.provider}
                          </span>
                        </td>
                        <td>{config.model}</td>
                        <td>
                          {config.has_api_key ? (
                            <span className="badge bg-success">
                              <i className="bi bi-lock-fill me-1"></i>
                              Set
                            </span>
                          ) : (
                            <span className="badge bg-secondary">None</span>
                          )}
                        </td>
                                          <td>
                    <div className="form-check form-switch">
                      <input
                        className="form-check-input"
                        type="checkbox"
                        role="switch"
                        id={`default-toggle-${config.id}`}
                        checked={config.is_default}
                        onChange={() => handleToggleDefault(config.id, config.is_default)}
                        style={{
                          cursor: 'pointer',
                          width: '3em',
                          height: '1.5em'
                        }}
                      />
                    </div>
                  </td>
                        <td>
                          <div className="btn-group btn-group-sm">
                            <button 
                              className="btn btn-outline-primary"
                              title="Edit"
                              onClick={() => {
                                setEditingConfig(config);
                                setShowAddModal(true);
                              }}
                            >
                              <i className="bi bi-pencil"></i>
                            </button>
                            <button 
                              className="btn btn-outline-danger"
                              title="Delete"
                              onClick={() => setDeletingConfigId(config.id)}
                            >
                              <i className="bi bi-trash"></i>
                            </button>
                          </div>
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

      {/* App Preferences Section */}
      <div className="row">
        <div className="col-12 mb-4">
          <div className={`frosted-card p-4 ${isDarkMode ? 'bg-dark text-light border-secondary' : ''}`}>
            <h4 className={`mb-3 ${isDarkMode ? 'text-light' : ''}`}>App-Specific Preferences</h4>
            <p className={`mb-3 ${isDarkMode ? 'text-light' : 'text-muted'}`}>
              Choose which AI model each app should use by default.
            </p>
            
            {['review', 'logs', 'analytics'].map((appName) => (
              <div key={appName} className="row mb-3">
                <div className="col-md-3">
                  <label className={`form-label text-capitalize ${isDarkMode ? 'text-light' : ''}`}>{appName} App:</label>
                </div>
                <div className="col-md-6">
                  <select 
                    className={`form-select ${isDarkMode ? 'bg-dark text-light border-secondary' : ''}`}
                    name={`${appName}-preference`}
                    value={appPreferences[appName]?.config_id || ''}
                    onChange={(e) => handleSetAppPreference(appName, e.target.value)}
                  >
                    <option value="">Use Default</option>
                    {configs.map((config) => (
                      <option key={config.id} value={config.id}>
                        {config.name} ({config.provider_type})
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Usage Summary Section */}
      <div className="row">
        <div className="col-12 mb-4">
          <div className={`frosted-card p-4 ${isDarkMode ? 'bg-dark text-light border-secondary' : ''}`}>
            <h4 className={`mb-3 ${isDarkMode ? 'text-light' : ''}`}>Usage Summary (Last 30 Days)</h4>
            
            {usageSummary ? (
              <div className="row">
                <div className="col-md-4">
                  <div className={`text-center p-3 rounded ${isDarkMode ? 'bg-secondary text-light' : 'bg-light'}`}>
                    <h5 className={isDarkMode ? 'text-light' : 'text-muted'}>Total Tokens</h5>
                    <h3 className={isDarkMode ? 'text-light' : ''}>{usageSummary.total_tokens?.toLocaleString() || 0}</h3>
                  </div>
                </div>
                <div className="col-md-4">
                  <div className={`text-center p-3 rounded ${isDarkMode ? 'bg-secondary text-light' : 'bg-light'}`}>
                    <h5 className={isDarkMode ? 'text-light' : 'text-muted'}>Requests</h5>
                    <h3 className={isDarkMode ? 'text-light' : ''}>{usageSummary.total_requests?.toLocaleString() || 0}</h3>
                  </div>
                </div>
                <div className="col-md-4">
                  <div className={`text-center p-3 rounded ${isDarkMode ? 'bg-secondary text-light' : 'bg-light'}`}>
                    <h5 className={isDarkMode ? 'text-light' : 'text-muted'}>Estimated Cost</h5>
                    <h3 className={isDarkMode ? 'text-light' : ''}>${usageSummary.estimated_cost?.toFixed(2) || '0.00'}</h3>
                  </div>
                </div>
              </div>
            ) : (
              <div className={`alert ${isDarkMode ? 'alert-secondary' : 'alert-info'} mb-0`}>
                <i className="bi bi-info-circle me-2"></i>
                No usage data yet. Start using AI features to see statistics here.
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      {deletingConfigId && (
        <div className="modal show d-block" style={{ backgroundColor: 'rgba(0,0,0,0.5)' }}>
          <div className="modal-dialog">
            <div className={`modal-content ${isDarkMode ? 'bg-dark text-light' : ''}`}>
              <div className={`modal-header ${isDarkMode ? 'border-secondary' : ''}`}>
                <h5 className="modal-title">Confirm Deletion</h5>
                <button 
                  type="button" 
                  className={`btn-close ${isDarkMode ? 'btn-close-white' : ''}`}
                  onClick={() => setDeletingConfigId(null)}
                ></button>
              </div>
              <div className={`modal-body ${isDarkMode ? 'bg-dark text-light' : ''}`}>
                <p>Are you sure you want to delete this AI model configuration?</p>
                <p className="text-danger mb-0">
                  <i className="bi bi-exclamation-triangle me-1"></i>
                  This action cannot be undone.
                </p>
              </div>
              <div className={`modal-footer ${isDarkMode ? 'border-secondary' : ''}`}>
                <button 
                  type="button" 
                  className="btn btn-secondary"
                  onClick={() => setDeletingConfigId(null)}
                >
                  Cancel
                </button>
                <button 
                  type="button" 
                  className="btn btn-danger"
                  onClick={() => handleDeleteConfig(deletingConfigId)}
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Add/Edit Modal */}
      <AddLLMConfigModal
        isOpen={showAddModal}
        onClose={() => {
          setShowAddModal(false);
          setEditingConfig(null);
        }}
        onSave={handleSaveConfig}
        editingConfig={editingConfig}
      />
    </div>
  );
}
