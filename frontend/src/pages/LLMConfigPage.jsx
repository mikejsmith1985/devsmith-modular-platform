import React from 'react';
import { Link } from 'react-router-dom';
import { useTheme } from '../context/ThemeContext';

/**
 * LLMConfigPage Component
 * 
 * Phase 5, Task 5.2 - Placeholder (to be fully implemented)
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
  const { isDarkMode } = useTheme();

  return (
    <div className="container mt-4">
      {/* Navigation Bar */}
      <nav className="navbar navbar-expand-lg navbar-light frosted-card mb-4">
        <div className="container-fluid">
          <span className="navbar-brand fw-bold" style={{ fontSize: '1.5rem', color: isDarkMode ? '#e0e7ff' : '#1e293b' }}>
            <i className="bi bi-robot me-2"></i>
            AI Model Management
          </span>
          <div className="d-flex align-items-center gap-3">
            <Link to="/portal" className="btn btn-outline-secondary btn-sm">
              <i className="bi bi-arrow-left me-1"></i>
              Back to Dashboard
            </Link>
          </div>
        </div>
      </nav>

      {/* Page Header */}
      <div className="row">
        <div className="col-12 mb-4">
          <div className="frosted-card p-4">
            <h2 className="mb-3">
              <i className="bi bi-robot me-2"></i>
              AI Model Configuration
            </h2>
            <p className="mb-0">
              Manage your AI model configurations, API keys, and app-specific preferences.
            </p>
          </div>
        </div>
      </div>

      {/* Placeholder Content */}
      <div className="row">
        <div className="col-12 mb-4">
          <div className="frosted-card p-4">
            <h4 className="mb-3">Your AI Models</h4>
            <div className="alert alert-info">
              <i className="bi bi-info-circle me-2"></i>
              <strong>Coming Soon:</strong> Full LLM configuration management will be implemented in Task 5.2.
            </div>
            <p className="text-muted mb-0">
              Features in development:
            </p>
            <ul className="text-muted">
              <li>Add/Edit/Delete LLM configurations</li>
              <li>Configure Anthropic Claude, OpenAI, Ollama, DeepSeek, and Mistral models</li>
              <li>Secure API key storage with encryption</li>
              <li>Test connection before saving</li>
              <li>Set default model per application</li>
              <li>View usage statistics and costs</li>
            </ul>
          </div>
        </div>
      </div>

      {/* App Preferences Section */}
      <div className="row">
        <div className="col-12 mb-4">
          <div className="frosted-card p-4">
            <h4 className="mb-3">App Preferences</h4>
            <div className="alert alert-info">
              <i className="bi bi-info-circle me-2"></i>
              <strong>Coming Soon:</strong> Set which AI model each app should use.
            </div>
          </div>
        </div>
      </div>

      {/* Usage Summary Section */}
      <div className="row">
        <div className="col-12 mb-4">
          <div className="frosted-card p-4">
            <h4 className="mb-3">Usage Summary</h4>
            <div className="alert alert-info">
              <i className="bi bi-info-circle me-2"></i>
              <strong>Coming Soon:</strong> Track token usage and costs across all models.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
