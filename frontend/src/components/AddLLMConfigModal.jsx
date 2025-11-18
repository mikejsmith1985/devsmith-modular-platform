import React, { useState, useEffect } from 'react';
import { apiRequest } from '../utils/api';
import { fetchOllamaModels } from '../utils/ollama';

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
    'claude-3-5-sonnet-20240620',
    'claude-3-opus-20240229',
    'claude-3-5-haiku-20241022',
    'claude-3-sonnet-20240229',
    'claude-3-haiku-20240307'
  ],
  openai: [
    'gpt-4-turbo-preview',
    'gpt-4',
    'gpt-3.5-turbo',
    'gpt-4-32k'
  ],
  // Ollama models will be fetched dynamically
  ollama: [],
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

// Max token limits by model
const MODEL_MAX_TOKENS = {
  'claude-3-5-sonnet-20241022': 8192,
  'claude-3-5-sonnet-20240620': 8192,
  'claude-3-opus-20240229': 4096,
  'claude-3-5-haiku-20241022': 8192,
  'claude-3-sonnet-20240229': 4096,
  'claude-3-haiku-20240307': 4096,
  'gpt-4-turbo-preview': 4096,
  'gpt-4': 8192,
  'gpt-3.5-turbo': 4096,
  'gpt-4-32k': 32768,
  'llama3.1:70b': 128000,
  'llama3.1:8b': 128000,
  'deepseek-coder-v2:16b': 128000,
  'qwen2.5-coder:7b': 32768,
  'codellama:13b': 16384,
  'deepseek-chat': 64000,
  'deepseek-coder': 64000,
  'mistral-large': 32768,
  'mistral-medium': 32768,
  'mistral-small': 32768,
  'codestral-latest': 32768
};

function AddLLMConfigModal({ isOpen, onClose, onSave, editingConfig }) {
  const [formData, setFormData] = useState({
    name: '',
    provider: 'anthropic',
    model: '',
    api_key: '',
    endpoint: '',
    is_default: false
  });

  const [availableModels, setAvailableModels] = useState(MODELS_BY_PROVIDER.anthropic);
  const [ollamaModels, setOllamaModels] = useState([]);
  const [testingConnection, setTestingConnection] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [errors, setErrors] = useState({});
  const [showAdvanced, setShowAdvanced] = useState(false);

  // Populate form when editing
  useEffect(() => {
    async function loadOllamaModels() {
      const models = await fetchOllamaModels();
      setOllamaModels(models);
      if (formData.provider === 'ollama') {
        setAvailableModels(models);
      }
    }
    if (editingConfig) {
      setFormData({
        name: editingConfig.name || '',
        provider: editingConfig.provider || 'anthropic',
        model: editingConfig.model || '',
        api_key: '', // Never pre-fill API key for security
        endpoint: editingConfig.endpoint || '',
        is_default: editingConfig.is_default || false
      });
      if (editingConfig.provider === 'ollama') {
        loadOllamaModels();
      } else {
        setAvailableModels(MODELS_BY_PROVIDER[editingConfig.provider] || []);
      }
    } else {
      // Reset form for new config
      setFormData({
        name: '',
        provider: 'anthropic',
        model: '',
        api_key: '',
        endpoint: '',
        is_default: false
      });
      setAvailableModels(MODELS_BY_PROVIDER.anthropic);
    }
    setTestResult(null);
    setErrors({});
  }, [editingConfig, isOpen]);

  // Update available models when provider changes
  const handleProviderChange = async (provider) => {
    setFormData(prev => ({
      ...prev,
      provider: provider,
      model: '' // Reset model when provider changes
    }));
    if (provider === 'ollama') {
      const models = ollamaModels.length ? ollamaModels : await fetchOllamaModels();
      setAvailableModels(models);
    } else {
      setAvailableModels(MODELS_BY_PROVIDER[provider] || []);
    }
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

    if (!formData.model) {
      newErrors.model = 'Model is required';
    }

    // API key required for non-Ollama providers
    if (formData.provider !== 'ollama' && !formData.api_key && !editingConfig) {
      newErrors.api_key = 'API key is required';
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
      const response = await apiRequest('/api/portal/llm-configs/test', {
        method: 'POST',
        body: JSON.stringify({
          provider: formData.provider,
          model: formData.model,
          api_key: formData.api_key || "",
          endpoint: formData.endpoint || ""
        })
      });

      setTestResult({
        success: true,
        message: response?.message || 'Connection successful!'
      });
    } catch (err) {
      setTestResult({
        success: false,
        message: err.message || 'Connection failed'
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
        response = await apiRequest(`/api/portal/llm-configs/${editingConfig.id}`, {
          method: 'PUT',
          body: JSON.stringify(updateData)
        });
      } else {
        // Create new config
        response = await apiRequest('/api/portal/llm-configs', {
          method: 'POST',
          body: JSON.stringify(formData)
        });
      }

      onSave(response);
      onClose();
    } catch (err) {
      console.error('Failed to save config:', err);
      alert('Failed to save configuration: ' + (err.message || 'Unknown error'));
    }
  };

  if (!isOpen) {
    return null;
  }

  const isFormValid = formData.name.trim() && formData.model && 
    (formData.provider === 'ollama' || formData.api_key || editingConfig);

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
                  name="provider"
                  value={formData.provider}
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
                  className={`form-select ${errors.model ? 'is-invalid' : ''}`}
                  id="model"
                  name="model"
                  value={formData.model}
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
                {errors.model && <div className="invalid-feedback">{errors.model}</div>}
              </div>

              {/* API Key Field (hidden for Ollama) */}
              {formData.provider !== 'ollama' && (
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
                  name="endpoint"
                  value={formData.endpoint}
                  onChange={handleInputChange}
                  placeholder={formData.provider === 'ollama' ? 'http://localhost:11434' : 'https://api.example.com/v1'}
                />
                <small className="form-text text-muted">
                  {formData.provider === 'ollama' 
                    ? 'Defaults to http://localhost:11434 if not specified'
                    : 'Override the default API endpoint'}
                </small>
              </div>

              {/* Advanced Settings */}
              <div className="mb-3">
                <button 
                  type="button"
                  className="btn btn-link p-0 text-decoration-none d-flex align-items-center"
                  onClick={() => setShowAdvanced(!showAdvanced)}
                >
                  <i className={`bi bi-chevron-${showAdvanced ? 'down' : 'right'} me-2`}></i>
                  Advanced Settings
                </button>
                
                {showAdvanced && (
                  <div className="mt-3 ps-4 border-start">
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
                )}
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
