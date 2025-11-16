import React, { useState, useEffect, useMemo } from 'react';
import { reviewApi } from '../utils/api';
import { useTheme } from '../context/ThemeContext';

/**
 * Phase 4, Task 4.1: Prompt Editor Modal Component
 * 
 * Allows users to view, edit, and customize AI prompts for each review mode.
 * 
 * @component
 * @param {Object} props
 * @param {boolean} props.isOpen - Controls modal visibility
 * @param {Function} props.onClose - Callback when modal closes
 * @param {string} props.mode - Analysis mode (preview|skim|scan|detailed|critical)
 * @param {string} props.userLevel - User expertise level (beginner|intermediate|advanced)
 * @param {string} props.outputMode - Output verbosity (quick|detailed|comprehensive)
 * 
 * @features
 * - Display current prompt (system default or user custom)
 * - Syntax highlighting for variables
 * - Variable reference panel
 * - Character count with 2000 char limit
 * - Save custom prompt
 * - Factory reset to system default
 * - Validation of required variables
 * 
 * @example
 * <PromptEditorModal 
 *   isOpen={showModal}
 *   onClose={handleClose}
 *   mode="critical"
 *   userLevel="intermediate"
 *   outputMode="detailed"
 * />
 */

// Constants
const MAX_PROMPT_LENGTH = 2000;

const ERROR_MESSAGES = {
  LOAD_FAILED: 'Failed to load prompt. Please try again.',
  SAVE_FAILED: 'Failed to save prompt. Please try again.',
  RESET_FAILED: 'Failed to reset prompt. Please try again.',
  VALIDATION_REQUIRED_VARS: 'Prompt must contain all required variables',
  VALIDATION_MAX_LENGTH: `Prompt cannot exceed ${MAX_PROMPT_LENGTH} characters`
};

const SUCCESS_MESSAGES = {
  SAVE_SUCCESS: 'Custom prompt saved successfully',
  RESET_SUCCESS: 'Prompt reset to system default'
};

// Variable definitions for each mode (extracted from component for clarity)
const MODE_VARIABLES = {
  preview: [
    { name: '{{code}}', description: 'Code to analyze', required: true }
  ],
  skim: [
    { name: '{{code}}', description: 'Code to analyze', required: true }
  ],
  scan: [
    { name: '{{code}}', description: 'Code to analyze', required: true },
    { name: '{{query}}', description: 'Search query', required: true }
  ],
  detailed: [
    { name: '{{code}}', description: 'Code to analyze', required: true }
  ],
  critical: [
    { name: '{{code}}', description: 'Code to analyze', required: true }
  ]
};

/**
 * PromptEditorModal Component
 */
export default function PromptEditorModal({ 
  isOpen, 
  onClose, 
  mode, 
  userLevel = 'intermediate', 
  outputMode = 'quick' 
}) {
  // Theme
  const { isDarkMode } = useTheme();
  
  // State management
  const [promptText, setPromptText] = useState('');
  const [originalPrompt, setOriginalPrompt] = useState('');
  const [isCustom, setIsCustom] = useState(false);
  const [canReset, setCanReset] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [validationError, setValidationError] = useState(null);
  const [showResetConfirm, setShowResetConfirm] = useState(false);
  const [variablesPanelExpanded, setVariablesPanelExpanded] = useState(false);

  // Get variables for current mode
  const variables = useMemo(() => MODE_VARIABLES[mode] || MODE_VARIABLES.preview, [mode]);

  // Load prompt when modal opens
  useEffect(() => {
    if (isOpen) {
      loadPrompt();
    }
  }, [isOpen, mode, userLevel, outputMode]);

  /**
   * Load prompt from API
   * Fetches the effective prompt (system default or custom) for the current mode/level/output combo
   */
  const loadPrompt = async () => {
    setLoading(true);
    setError(null);
    setValidationError(null);

    try {
      const response = await reviewApi.getPrompt(mode, userLevel, outputMode);
      setPromptText(response.prompt_text);
      setOriginalPrompt(response.prompt_text);
      setIsCustom(response.is_custom || false);
      setCanReset(response.can_reset || false);
    } catch (err) {
      setError(err.message || ERROR_MESSAGES.LOAD_FAILED);
    } finally {
      setLoading(false);
    }
  };

  /**
   * Validate prompt contains all required variables
   * @param {string} text - Prompt text to validate
   * @returns {string|null} Error message or null if valid
   */
  const validatePrompt = (text) => {
    // Check length
    if (text.length > MAX_PROMPT_LENGTH) {
      return ERROR_MESSAGES.VALIDATION_MAX_LENGTH;
    }

    // Check required variables
    const requiredVars = variables.filter(v => v.required);
    const missingVars = [];

    for (const variable of requiredVars) {
      if (!text.includes(variable.name)) {
        missingVars.push(variable.name);
      }
    }

    if (missingVars.length > 0) {
      return `${ERROR_MESSAGES.VALIDATION_REQUIRED_VARS}: ${missingVars.join(', ')}`;
    }

    return null;
  };

  /**
   * Save custom prompt
   * Validates prompt and sends to API
   */
  const handleSave = async () => {
    setValidationError(null);

    // Validate required variables
    const validationErr = validatePrompt(promptText);
    if (validationErr) {
      setValidationError(validationErr);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await reviewApi.savePrompt({
        mode,
        user_level: userLevel,
        output_mode: outputMode,
        prompt_text: promptText,
        variables: variables.map(v => v.name)
      });

      // Reload prompt to get updated metadata
      await loadPrompt();
      
      // Close modal after successful save
      onClose();
    } catch (err) {
      setError(err.message || ERROR_MESSAGES.SAVE_FAILED);
    } finally {
      setLoading(false);
    }
  };

  /**
   * Factory reset prompt to system default
   * Deletes custom prompt and reloads system default
   */
  const handleFactoryReset = async () => {
    setLoading(true);
    setError(null);
    setValidationError(null);

    try {
      await reviewApi.resetPrompt(mode, userLevel, outputMode);
      
      // Reload prompt to get system default
      await loadPrompt();
      
      // Close confirmation modal
      setShowResetConfirm(false);
    } catch (err) {
      setError(err.message || ERROR_MESSAGES.RESET_FAILED);
    } finally {
      setLoading(false);
    }
  };

  /**
   * Cancel editing and close modal
   * Resets prompt to original value without saving
   */
  const handleCancel = () => {
    // Reset to original prompt
    setPromptText(originalPrompt);
    setValidationError(null);
    setError(null);
    onClose();
  };

  /**
   * Handle prompt text changes
   * Clears validation errors when user edits
   */
  const handlePromptChange = (e) => {
    setPromptText(e.target.value);
    setValidationError(null); // Clear validation error on edit
  };

  /**
   * Highlight variables in prompt text for display
   * Converts {{variable}} syntax to styled spans for visual preview
   * 
   * @param {string} text - Prompt text containing variables
   * @returns {Array} Array of {type, content} objects for rendering
   */
  const highlightVariables = (text) => {
    const parts = [];
    let lastIndex = 0;
    const regex = /\{\{[\w_]+\}\}/g;
    let match;

    while ((match = regex.exec(text)) !== null) {
      // Add text before variable
      if (match.index > lastIndex) {
        parts.push({
          type: 'text',
          content: text.substring(lastIndex, match.index)
        });
      }
      
      // Add variable
      parts.push({
        type: 'variable',
        content: match[0]
      });
      
      lastIndex = match.index + match[0].length;
    }

    // Add remaining text
    if (lastIndex < text.length) {
      parts.push({
        type: 'text',
        content: text.substring(lastIndex)
      });
    }

    return parts;
  };

  if (!isOpen) return null;

  return (
    <>
      {/* Main Modal */}
      <div className="modal show d-block prompt-editor-modal" tabIndex="-1" role="dialog">
        <div className="modal-dialog modal-lg" role="document">
          <div className={`modal-content ${isDarkMode ? 'bg-dark text-light' : ''}`}>
            {/* Header */}
            <div className={`modal-header ${isDarkMode ? 'border-secondary' : ''}`}>
              <h5 className="modal-title">
                Edit Prompt - {mode.charAt(0).toUpperCase() + mode.slice(1)} Mode
              </h5>
              <div className="ms-auto d-flex align-items-center gap-2">
                {isCustom ? (
                  <span className="badge bg-primary badge-custom">Custom</span>
                ) : (
                  <span className="badge bg-secondary badge-default">System Default</span>
                )}
                <button 
                  type="button" 
                  className={`btn-close ${isDarkMode ? 'btn-close-white' : ''}`}
                  onClick={handleCancel}
                  disabled={loading}
                />
              </div>
            </div>

            {/* Body */}
            <div className={`modal-body ${isDarkMode ? 'bg-dark' : ''}`}>
              {error && (
                <div className="alert alert-danger alert-dismissible fade show" role="alert">
                  {error}
                  <button 
                    type="button" 
                    className="btn-close" 
                    onClick={() => setError(null)}
                  />
                </div>
              )}

              {validationError && (
                <div className="alert alert-warning validation-error" role="alert">
                  <i className="bi bi-exclamation-triangle me-2"></i>
                  {validationError}
                </div>
              )}

              {loading ? (
                <div className="text-center py-5">
                  <div className="spinner-border text-primary" role="status">
                    <span className="visually-hidden">Loading...</span>
                  </div>
                </div>
              ) : (
                <>
                  {/* Variable Reference Panel */}
                  <div className={`card mb-3 variable-reference-panel ${isDarkMode ? 'bg-dark border-secondary' : ''}`}>
                    <div 
                      className={`card-header d-flex justify-content-between align-items-center ${isDarkMode ? 'bg-dark border-secondary text-light' : ''}`}
                      style={{ cursor: 'pointer' }}
                      onClick={() => setVariablesPanelExpanded(!variablesPanelExpanded)}
                    >
                      <span>
                        <i className="bi bi-code-square me-2"></i>
                        Available Variables
                      </span>
                      <button 
                        className={`btn btn-sm btn-link expand-btn ${isDarkMode ? 'text-light' : ''}`}
                        type="button"
                      >
                        <i className={`bi bi-chevron-${variablesPanelExpanded ? 'up' : 'down'}`}></i>
                      </button>
                    </div>
                    {variablesPanelExpanded && (
                      <div className={`card-body ${isDarkMode ? 'bg-dark text-light' : ''}`}>
                        <div className="row">
                          {variables.map((variable, index) => (
                            <div key={index} className="col-md-6 mb-2">
                              <div className="d-flex align-items-start">
                                <code className="text-primary me-2">{variable.name}</code>
                                {variable.required && (
                                  <span className="badge bg-danger badge-sm">Required</span>
                                )}
                              </div>
                              <small className="text-muted">{variable.description}</small>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Prompt Editor */}
                  <div className="mb-3">
                    <label htmlFor="prompt-text" className="form-label">
                      Prompt Template
                    </label>
                    <textarea
                      id="prompt-text"
                      name="prompt"
                      className={`form-control font-monospace ${isDarkMode ? 'bg-dark text-light border-secondary' : ''}`}
                      rows="12"
                      value={promptText}
                      onChange={handlePromptChange}
                      placeholder="Enter your prompt template..."
                      style={{ fontSize: '14px' }}
                    />
                    <div className="d-flex justify-content-between mt-1">
                      <small className="text-muted">
                        Use variables like <code>{'{{code}}'}</code> in your prompt
                      </small>
                      <small className="text-muted character-count">
                        {promptText.length} characters
                      </small>
                    </div>
                  </div>

                  {/* Prompt Preview with Syntax Highlighting */}
                  <div className={`card ${isDarkMode ? 'bg-dark border-secondary' : 'bg-light'}`}>
                    <div className={`card-body ${isDarkMode ? 'text-light' : ''}`}>
                      <h6 className="card-title">Preview</h6>
                      <div className="font-monospace" style={{ fontSize: '13px', whiteSpace: 'pre-wrap' }}>
                        {highlightVariables(promptText).map((part, index) => (
                          part.type === 'variable' ? (
                            <span key={index} className="text-primary fw-bold">{part.content}</span>
                          ) : (
                            <span key={index}>{part.content}</span>
                          )
                        ))}
                      </div>
                    </div>
                  </div>
                </>
              )}
            </div>

            {/* Footer */}
            <div className={`modal-footer d-flex justify-content-between ${isDarkMode ? 'border-secondary' : ''}`}>
              <div>
                {canReset && (
                  <button
                    type="button"
                    className="btn btn-outline-danger btn-factory-reset"
                    onClick={() => setShowResetConfirm(true)}
                    disabled={loading}
                  >
                    <i className="bi bi-arrow-counterclockwise me-1"></i>
                    Factory Reset
                  </button>
                )}
              </div>
              <div className="d-flex gap-2">
                <button
                  type="button"
                  className="btn btn-secondary btn-cancel"
                  onClick={handleCancel}
                  disabled={loading}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="btn btn-primary btn-save"
                  onClick={handleSave}
                  disabled={loading || !promptText.trim()}
                >
                  {loading ? (
                    <>
                      <span className="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
                      Saving...
                    </>
                  ) : (
                    <>
                      <i className="bi bi-save me-1"></i>
                      Save Custom Prompt
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Backdrop */}
      <div className="modal-backdrop show"></div>

      {/* Reset Confirmation Modal */}
      {showResetConfirm && (
        <>
          <div className="modal show d-block" tabIndex="-1" role="dialog" style={{ zIndex: 1060 }}>
            <div className="modal-dialog modal-dialog-centered" role="document">
              <div className={`modal-content ${isDarkMode ? 'bg-dark text-light' : ''}`}>
                <div className={`modal-header ${isDarkMode ? 'border-secondary' : ''}`}>
                  <h5 className="modal-title">Confirm Factory Reset</h5>
                  <button 
                    type="button" 
                    className={`btn-close ${isDarkMode ? 'btn-close-white' : ''}`}
                    onClick={() => setShowResetConfirm(false)}
                    disabled={loading}
                  />
                </div>
                <div className={`modal-body ${isDarkMode ? 'bg-dark' : ''}`}>
                  <p>
                    Are you sure you want to reset this prompt to the system default?
                  </p>
                  <p className="text-muted mb-0">
                    <i className="bi bi-info-circle me-1"></i>
                    This will permanently delete your custom prompt for this mode.
                  </p>
                </div>
                <div className={`modal-footer ${isDarkMode ? 'border-secondary' : ''}`}>
                  <button
                    type="button"
                    className="btn btn-secondary"
                    onClick={() => setShowResetConfirm(false)}
                    disabled={loading}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    className="btn btn-danger confirm-reset-btn"
                    onClick={handleFactoryReset}
                    disabled={loading}
                  >
                    {loading ? 'Resetting...' : 'Yes, Reset to Default'}
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div className="modal-backdrop show" style={{ zIndex: 1055 }}></div>
        </>
      )}
    </>
  );
}
