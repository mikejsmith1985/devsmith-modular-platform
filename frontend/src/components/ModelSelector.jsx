import React, { useState, useEffect } from 'react';
import { apiRequest } from '../utils/api';
import { useTheme } from '../context/ThemeContext';

export default function ModelSelector({ selectedModel, onModelSelect, disabled = false }) {
  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { isDarkMode } = useTheme();

  useEffect(() => {
    const loadModels = async () => {
      try {
        setLoading(true);
        // Use AI Factory endpoint instead of ollama list
        const response = await apiRequest('/api/portal/llm-configs');
        console.log('AI Factory response:', response);
        
        let modelList = [];
        if (Array.isArray(response)) {
          // Transform AI Factory configs to model selector format
          modelList = response.map(config => ({
            name: config.model_name,
            displayName: config.display_name || config.model_name,
            provider: config.provider,
            isDefault: config.is_default
          }));
          
          // Sort models: default first, then alphabetically by display name
          modelList.sort((a, b) => {
            if (a.isDefault && !b.isDefault) return -1;
            if (!a.isDefault && b.isDefault) return 1;
            return a.displayName.localeCompare(b.displayName);
          });
        } else {
          console.warn('Unexpected AI Factory response format:', response);
        }

        setModels(modelList);
        console.log('Models loaded from AI Factory:', modelList);
        
        // Auto-select default model from AI Factory
        if (!selectedModel && modelList.length > 0) {
          const defaultModel = modelList.find(m => m.isDefault);
          if (defaultModel) {
            console.log('Setting default model:', defaultModel.name);
            onModelSelect(defaultModel.name);
          } else {
            // Fallback to first model if no default
            console.log('No default model found, using first:', modelList[0].name);
            onModelSelect(modelList[0].name);
          }
        }
      } catch (err) {
        console.error('Failed to load AI Factory models:', err);
        setError(err.message);
        // Fallback to default models with recommended first
        const defaultModels = [
          { name: 'qwen2.5-coder:7b-instruct-q4_K_M', description: 'Qwen 2.5 Coder 7B (Recommended for 8GB VRAM)', provider: 'Ollama', isDefault: true },
          { name: 'mistral:7b-instruct', description: 'Mistral 7B', provider: 'Ollama' },
          { name: 'deepseek-coder-v2:16b-lite-instruct-q4_K_M', description: 'DeepSeek Coder V2 16B (Requires 16GB+ VRAM)', provider: 'Ollama' }
        ];
        setModels(defaultModels);
        if (!selectedModel) {
          onModelSelect('qwen2.5-coder:7b-instruct-q4_K_M');
        }
      } finally {
        setLoading(false);
      }
    };

    loadModels();
  }, [selectedModel, onModelSelect]);

  const handleModelChange = (e) => {
    console.log('Model selected:', e.target.value);
    if (onModelSelect) {
      onModelSelect(e.target.value);
    }
  };

  if (loading) {
    return (
      <div className="model-selector mb-3">
        <label className="form-label">AI Model:</label>
        <div className="d-flex align-items-center">
          <div className="spinner-border spinner-border-sm me-2" role="status">
            <span className="visually-hidden">Loading models...</span>
          </div>
          <small className="text-muted">Loading available models...</small>
        </div>
      </div>
    );
  }

  return (
    <div className="model-selector mb-3">
      <label className="form-label" htmlFor="model-select">
        AI Model:
      </label>
      
      {error && (
        <div className="alert alert-warning py-1 mb-2" role="alert">
          <small>
            <strong>Warning:</strong> {error}. Using fallback models.
          </small>
        </div>
      )}
      
      <select 
        id="model-select"
        className={`form-select ${isDarkMode ? 'bg-dark text-light border-secondary' : ''}`}
        style={isDarkMode ? { 
          backgroundColor: '#1a1d2e',
          color: '#e0e7ff',
          borderColor: '#4a5568'
        } : {}}
        value={selectedModel || ''}
        onChange={handleModelChange}
        disabled={disabled || models.length === 0}
      >
        {models.length === 0 ? (
          <option value="">No models available</option>
        ) : (
          models.map((model) => {
            const modelName = typeof model === 'string' ? model : model.name;
            const displayName = typeof model === 'object' && model.displayName ? model.displayName : modelName;
            const provider = typeof model === 'object' && model.provider ? model.provider : '';
            
            return (
              <option key={modelName} value={modelName}>
                {provider} - {displayName}
              </option>
            );
          })
        )}
      </select>
      
      <small className="form-text text-muted mt-1">
        Choose the AI model for code analysis. Larger models provide more detailed analysis.
      </small>
    </div>
  );
}