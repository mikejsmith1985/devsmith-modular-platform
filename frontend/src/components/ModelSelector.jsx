import React, { useState, useEffect } from 'react';
import { reviewApi } from '../utils/api';

export default function ModelSelector({ selectedModel, onModelSelect, disabled = false }) {
  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const loadModels = async () => {
      try {
        setLoading(true);
        const response = await reviewApi.getModels();
        
        if (Array.isArray(response)) {
          setModels(response);
          // Auto-select first model if none selected
          if (!selectedModel && response.length > 0) {
            onModelSelect(response[0].name || response[0]);
          }
        } else if (response.models && Array.isArray(response.models)) {
          setModels(response.models);
          if (!selectedModel && response.models.length > 0) {
            onModelSelect(response.models[0].name || response.models[0]);
          }
        } else {
          console.warn('Unexpected models response format:', response);
          setModels([]);
        }
      } catch (err) {
        console.error('Failed to load models:', err);
        setError(err.message);
        // Fallback to default models
        const defaultModels = [
          { name: 'deepseek-coder:6.7b', description: 'Recommended for most users' },
          { name: 'deepseek-coder:1.5b', description: 'Faster, lower resource usage' }
        ];
        setModels(defaultModels);
        if (!selectedModel) {
          onModelSelect(defaultModels[0].name);
        }
      } finally {
        setLoading(false);
      }
    };

    loadModels();
  }, [selectedModel, onModelSelect]);

  const handleModelChange = (e) => {
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
        className="form-select"
        value={selectedModel || ''}
        onChange={handleModelChange}
        disabled={disabled || models.length === 0}
      >
        {models.length === 0 ? (
          <option value="">No models available</option>
        ) : (
          models.map((model) => {
            const modelName = typeof model === 'string' ? model : model.name;
            const modelDesc = typeof model === 'object' ? model.description : '';
            
            return (
              <option key={modelName} value={modelName}>
                {modelName} {modelDesc && `- ${modelDesc}`}
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