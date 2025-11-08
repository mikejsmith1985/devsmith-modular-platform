import React, { useState, useEffect } from 'react';
import api from '../utils/api';

const PROVIDERS = [
  { value: 'anthropic', label: 'Anthropic (Claude)' },
  { value: 'openai', label: 'OpenAI (GPT)' },
  { value: 'ollama', label: 'Ollama (Local)' },
  { value: 'deepseek', label: 'DeepSeek' },
  { value: 'mistral', label: 'Mistral AI' }
];

const MODELS_BY_PROVIDER = {
  anthropic: [
    'claude-3-5-sonnet-20241022',
    'claude-3-5-haiku-20241022',
    'claude-3-opus-20240229',
    'claude-3-sonnet-20240229',
    'claude-3-haiku-20240307'
  ],
  openai: [
    'gpt-4-turbo-preview',
    'gpt-4',
    'gpt-3.5-turbo',
    'gpt-4-32k'
  ],
  ollama: [
    'llama3.1:70b',
    'llama3.1:8b',
    'deepseek-coder-v2:16b',
    'deepseek-coder:6.7b',
    'qwen2.5-coder:7b',
    'codestral:22b'
  ],
  deepseek: [
    'deepseek-chat',
    'deepseek-coder'
  ],
  mistral: [
    'mistral-large',
    'mistral-medium',
    'mistral-small',
    'codestral-latest'
  ]
};

function AddLLMConfigModal({ isOpen, onClose, onSave, editingConfig }) {
  const [formData, setFormData] = useState({
    name: '',
    provider_type: 'anthropic',
    model_name: '',
    api_key: '',
    custom_endpoint: '',
    temperature: 0.7,
    max_tokens: 4096,
    is_default: false
  });

  const [availableModels, setAvailableModels] = useState(MODELS_BY_PROVIDER.anthropic);
  const [testingConnection, setTestingConnection] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [errors, setErrors] = useState({});

  // Populate form when editing
  useEffect(() => {
    if (editingConfig) {
      setFormData({
        name: editingConfig.name || '',
        provider_type: editingConfig.provider_type || 'anthropic',
        model_name: editingConfig.model_name || '',
        api_key: '', // Never pre-fill API key for security
        custom_endpoint: editingConfig.custom_endpoint || '',
        temperature: editingConfig.temperature || 0.7,
        max_tokens: editingConfig.max_tokens || 4096,
        is_default: editingConfig.is_default || false
      });
      setAvailableModels(MODELS_BY_PROVIDER[editingConfig.provider_type] || []);
    } else {
      // Reset form for new config
      setFormData({
        name: '',
        provider_type: 'anthropic',
        model_name: '',
        api_key: '',
        custom_endpoint: '',
        temperature: 0.7,
        max_tokens: 4096,
        is_default: false
      });
      setAvailableModels(MODELS_BY_PROVIDER.anthropic);
    }
    setTestResult(null);
    setErrors({});
  }, [editingConfig, isOpen]);

  // Update available models when provider changes
  const handleProviderChange = (provider) => {
    setFormData(prev => ({
      ...prev,
      provider_type: provider,
      model_name: '' // Reset model when provider changes
    }));
    setAvailableModels(MODELS_BY_PROVIDER[provider] || []);
  };

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
    // Clear error for this field
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: null }));
    }
  };

  const validateForm = () => {
    const newErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }

    if (!formData.model_name) {
      newErrors.model_name = 'Model is required';
    }

    // API key required for non-Ollama providers
    if (formData.provider_type !== 'ollama' && !formData.api_key && !editingConfig) {
      newErrors.api_key = 'API key is required';
    }

    if (formData.temperature < 0 || formData.temperature > 2) {
      newErrors.temperature = 'Temperature must be between 0 and 2';
    }

    if (formData.max_tokens < 1 || formData.max_tokens > 100000) {
      newErrors.max_tokens = 'Max tokens must be between 1 and 100000';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleTestConnection = async () => {
    if (!validateForm()) {
      return;
    }

    setTestingConnection(true);
    setTestResult(null);

    try {
      const response = await api.post('/api/portal/llm-configs/test', {
        provider_type: formData.provider_type,
        model_name: formData.model_name,
        api_key: formData.api_key || undefined,
        custom_endpoint: formData.custom_endpoint || undefined
      });

      setTestResult({
        success: true,
        message: response.data?.message || 'Connection successful!'
      });
    } catch (err) {
      setTestResult({
        success: false,
        message: err.response?.data?.error || 'Connection failed'
      });
    } finally {
      setTestingConnection(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      let response;
      if (editingConfig) {
        // Update existing config
        const updateData = { ...formData };
        // Don't send api_key if empty (keeping existing key)
        if (!updateData.api_key) {
          delete updateData.api_key;
        }
        response = await api.put(`/api/portal/llm-configs/${editingConfig.id}`, updateData);
      } else {
        // Create new config
        response = await api.post('/api/portal/llm-configs', formData);
      }

      onSave(response.data);
      onClose();
    } catch (err) {
      console.error('Failed to save config:', err);
      alert('Failed to save configuration: ' + (err.response?.data?.error || err.message));
    }
  };

  if (!isOpen) {
    return null;
  }

  const isFormValid = formData.name.trim() && formData.model_name && 
    (formData.provider_type === 'ollama' || formData.api_key || editingConfig);

  return (
    <div className="modal show d-block" style={{ backgroundColor: 'rgba(0,0,0,0.5)' }} role="dialog">
      <div className="modal-dialog modal-lg">
        <div className="modal-content">
          <div className="modal-header">
            <h5 className="modal-title">
              {editingConfig ? 'Edit' : 'Add'} AI Model Configuration
            </h5>
            <button 
              type="button" 
              className="btn-close" 
              onClick={onClose}
              aria-label="Close"
            ></button>
          </div>

          <form onSubmit={handleSubmit}>
            <div className="modal-body">
              {/* Name Field */}
              <div className="mb-3">
                <label htmlFor="config-name" className="form-label">
                  Configuration Name <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  className={`form-control ${errors.name ? 'is-invalid' : ''}`}
                  id="config-name"
                  name="name"
                  value={formData.name}
                  onChange={handleInputChange}
                  placeholder="e.g., Claude for Review"
                  required
                />
                {errors.name && <div className="invalid-feedback">{errors.name}</div>}
              </div>

              {/* Provider Dropdown */}
              <div className="mb-3">
                <label htmlFor="provider" className="form-label">
                  Provider <span className="text-danger">*</span>
                </label>
                <select
                  className="form-select"
                  id="provider"
                  name="provider_type"
                  value={formData.provider_type}
                  onChange={(e) => handleProviderChange(e.target.value)}
                  required
                >
                  {PROVIDERS.map(provider => (
                    <option key={provider.value} value={provider.value}>
                      {provider.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Model Dropdown */}
              <div className="mb-3">
                <label htmlFor="model" className="form-label">
                  Model <span className="text-danger">*</span>
                </label>
                <select
                  className={`form-select ${errors.model_name ? 'is-invalid' : ''}`}
                  id="model"
                  name="model_name"
                  value={formData.model_name}
                  onChange={handleInputChange}
                  required
                >
                  <option value="">Select a model</option>
                  {availableModels.map(model => (
                    <option key={model} value={model}>
                      {model}
                    </option>
                  ))}
                </select>
                {errors.model_name && <div className="invalid-feedback">{errors.model_name}</div>}
              </div>

              {/* API Key Field (hidden for Ollama) */}
              {formData.provider_type !== 'ollama' && (
                <div className="mb-3">
                  <label htmlFor="api-key" className="form-label">
                    API Key {!editingConfig && <span className="text-danger">*</span>}
                  </label>
                  <input
                    type="password"
                    className={`form-control ${errors.api_key ? 'is-invalid' : ''}`}
                    id="api-key"
                    name="api_key"
                    value={formData.api_key}
                    onChange={handleInputChange}
                    placeholder={editingConfig ? 'Leave blank to keep existing key' : 'Enter API key'}
                    required={!editingConfig}
                  />
                  {errors.api_key && <div className="invalid-feedback">{errors.api_key}</div>}
                  {editingConfig && (
                    <small className="form-text text-muted">
                      Leave blank to keep the existing API key
                    </small>
                  )}
                </div>
              )}

              {/* Custom Endpoint (Optional) */}
              <div className="mb-3">
                <label htmlFor="endpoint" className="form-label">
                  Custom Endpoint (Optional)
                </label>
                <input
                  type="url"
                  className="form-control"
                  id="endpoint"
                  name="custom_endpoint"
                  value={formData.custom_endpoint}
                  onChange={handleInputChange}
                  placeholder="https://api.example.com/v1"
                />
                <small className="form-text text-muted">
                  Override the default API endpoint
                </small>
              </div>

              {/* Advanced Settings */}
              <div className="accordion mb-3" id="advancedSettings">
                <div className="accordion-item">
                  <h2 className="accordion-header">
                    <button 
                      className="accordion-button collapsed" 
                      type="button" 
                      data-bs-toggle="collapse" 
                      data-bs-target="#advancedCollapse"
                      aria-expanded="false"
                    >
                      Advanced Settings
                    </button>
                  </h2>
                  <div id="advancedCollapse" className="accordion-collapse collapse">
                    <div className="accordion-body">
                      {/* Temperature */}
                      <div className="mb-3">
                        <label htmlFor="temperature" className="form-label">
                          Temperature: {formData.temperature}
                        </label>
                        <input
                          type="range"
                          className="form-range"
                          id="temperature"
                          name="temperature"
                          min="0"
                          max="2"
                          step="0.1"
                          value={formData.temperature}
                          onChange={handleInputChange}
                        />
                        <small className="form-text text-muted">
                          Controls randomness (0 = deterministic, 2 = very creative)
                        </small>
                      </div>

                      {/* Max Tokens */}
                      <div className="mb-3">
                        <label htmlFor="max-tokens" className="form-label">
                          Max Tokens
                        </label>
                        <input
                          type="number"
                          className={`form-control ${errors.max_tokens ? 'is-invalid' : ''}`}
                          id="max-tokens"
                          name="max_tokens"
                          value={formData.max_tokens}
                          onChange={handleInputChange}
                          min="1"
                          max="100000"
                        />
                        {errors.max_tokens && <div className="invalid-feedback">{errors.max_tokens}</div>}
                      </div>

                      {/* Set as Default */}
                      <div className="form-check">
                        <input
                          type="checkbox"
                          className="form-check-input"
                          id="is-default"
                          name="is_default"
                          checked={formData.is_default}
                          onChange={handleInputChange}
                        />
                        <label className="form-check-label" htmlFor="is-default">
                          Set as default configuration
                        </label>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Test Connection Result */}
              {testResult && (
                <div className={`alert ${testResult.success ? 'alert-success' : 'alert-danger'}`}>
                  <i className={`bi ${testResult.success ? 'bi-check-circle' : 'bi-x-circle'} me-2`}></i>
                  {testResult.message}
                </div>
              )}
            </div>

            <div className="modal-footer">
              <button
                type="button"
                className="btn btn-secondary"
                onClick={handleTestConnection}
                disabled={testingConnection || !isFormValid}
              >
                {testingConnection ? (
                  <>
                    <span className="spinner-border spinner-border-sm me-2" role="status"></span>
                    Testing...
                  </>
                ) : (
                  <>
                    <i className="bi bi-plug me-1"></i>
                    Test Connection
                  </>
                )}
              </button>
              <button
                type="button"
                className="btn btn-secondary"
                onClick={onClose}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn btn-primary"
                disabled={!isFormValid}
              >
                <i className="bi bi-save me-1"></i>
                {editingConfig ? 'Update' : 'Save'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export default AddLLMConfigModal;
