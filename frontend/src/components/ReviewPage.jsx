import React, { useState, useRef } from 'react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import { Link } from 'react-router-dom';
import CodeEditor from './CodeEditor';
import AnalysisModeSelector from './AnalysisModeSelector';
import ModelSelector from './ModelSelector';
import AnalysisOutput from './AnalysisOutput';
import { reviewApi } from '../utils/api';

// Default code for demonstration
const defaultCode = `// Example JavaScript function to analyze
function fibonacci(n) {
  if (n <= 1) {
    return n;
  }
  return fibonacci(n - 1) + fibonacci(n - 2);
}

// TODO: Optimize for large values of n
// Consider using memoization or iterative approach
console.log(fibonacci(10));
`;

export default function ReviewPage() {
  const { user, logout } = useAuth();
  const { isDarkMode, toggleTheme } = useTheme();
  
  // State management
  const [code, setCode] = useState(defaultCode);
  const [selectedMode, setSelectedMode] = useState('preview');
  const [selectedModel, setSelectedModel] = useState('');
  const [scanQuery, setScanQuery] = useState(''); // Add scan query state
  const [analysisResult, setAnalysisResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [sessionId] = useState(() => `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`);

  // Refs for managing focus
  const codeEditorRef = useRef(null);

  const handleAnalyze = async () => {
    if (!code.trim()) {
      setError('Please enter some code to analyze');
      return;
    }

    if (!selectedModel) {
      setError('Please select an AI model');
      return;
    }

    // Validate scan query for Scan mode
    if (selectedMode === 'scan' && !scanQuery.trim()) {
      setError('Please enter a search query for Scan mode');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      setAnalysisResult(null);

      let result;
      switch (selectedMode) {
        case 'preview':
          result = await reviewApi.runPreview(sessionId, code, selectedModel);
          break;
        case 'skim':
          result = await reviewApi.runSkim(sessionId, code, selectedModel);
          break;
        case 'scan':
          // Pass scan query to API for context-aware search
          result = await reviewApi.runScan(sessionId, code, selectedModel, scanQuery);
          break;
        case 'detailed':
          result = await reviewApi.runDetailed(sessionId, code, selectedModel);
          break;
        case 'critical':
          result = await reviewApi.runCritical(sessionId, code, selectedModel);
          break;
        default:
          throw new Error(`Unknown analysis mode: ${selectedMode}`);
      }

      setAnalysisResult(result);
    } catch (err) {
      console.error('Analysis failed:', err);
      setError(err.message || 'Analysis failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleRetry = () => {
    setError(null);
    handleAnalyze();
  };

  const clearCode = () => {
    setCode('');
    setAnalysisResult(null);
    setError(null);
  };

  const resetToDefault = () => {
    setCode(defaultCode);
    setAnalysisResult(null);
    setError(null);
  };

  return (
    <div className="container-fluid py-3">
      {/* Navigation Header */}
      <nav className="frosted-card mb-4 p-3">
        <div className="d-flex justify-content-between align-items-center">
          <Link to="/" className="btn btn-outline-primary btn-sm">
            <i className="bi bi-arrow-left me-2"></i>
            Back to Dashboard
          </Link>
          <div className="d-flex align-items-center gap-3">
            <button onClick={toggleTheme} className="theme-toggle">
              <i className={`bi ${isDarkMode ? 'bi-sun-fill' : 'bi-moon-fill'}`}></i>
            </button>
            <span className="me-2">Welcome, {user?.username || user?.name}!</span>
            <button
              className="btn btn-outline-danger btn-sm"
              onClick={() => logout()}
            >
              Logout
            </button>
          </div>
        </div>
      </nav>

      {/* Header */}
      <div className="row mb-4">
        <div className="col-12">
          <div className="d-flex justify-content-between align-items-center">
            <div>
              <h2 className="mb-1">
                <i className="bi bi-code-square text-primary me-2"></i>
                Code Review
              </h2>
              <p className="text-muted mb-0">AI-powered code analysis with five distinct reading modes</p>
            </div>
            <div>
              <small className="text-muted">Session: {sessionId}</small>
            </div>
          </div>
        </div>
      </div>

      {/* Analysis Mode Selection */}
      <div className="row mb-3">
        <div className="col-12">
          <AnalysisModeSelector 
            selectedMode={selectedMode}
            onModeSelect={setSelectedMode}
            disabled={loading}
          />
        </div>
      </div>

      {/* Scan Mode Search Bar */}
      {selectedMode === 'scan' && (
        <div className="row mb-3">
          <div className="col-12">
            <div className="frosted-card p-3">
              <label htmlFor="scanQuery" className="form-label mb-2">
                <i className="bi bi-search me-2"></i>
                <strong>What are you looking for?</strong>
              </label>
              <input
                type="text"
                id="scanQuery"
                className="form-control"
                placeholder='Try "functions", "error handling", "database queries", "authentication logic", etc.'
                value={scanQuery}
                onChange={(e) => setScanQuery(e.target.value)}
                disabled={loading}
                style={{
                  backgroundColor: 'rgba(255, 255, 255, 0.95)',
                  border: '2px solid rgba(99, 102, 241, 0.2)',
                  borderRadius: '8px',
                  padding: '0.75rem'
                }}
              />
              <small className="text-muted mt-2 d-block">
                <i className="bi bi-info-circle me-1"></i>
                This is a <strong>context-aware search</strong> - ask for concepts like "functions" or "error handling" 
                rather than exact text matches. The AI will find and analyze relevant code patterns.
              </small>
            </div>
          </div>
        </div>
      )}

      {/* Model Selection and Controls */}
      <div className="row mb-3">
        <div className="col-md-4">
          <ModelSelector 
            selectedModel={selectedModel}
            onModelSelect={setSelectedModel}
            disabled={loading}
          />
        </div>
        <div className="col-md-8">
          <div className="d-flex gap-2 align-items-end h-100">
            <button 
              className="btn btn-primary"
              onClick={handleAnalyze}
              disabled={loading || !code.trim() || !selectedModel}
            >
              {loading ? (
                <>
                  <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                  Analyzing...
                </>
              ) : (
                'Analyze Code'
              )}
            </button>
            <button 
              className="btn btn-outline-secondary"
              onClick={resetToDefault}
              disabled={loading}
            >
              Reset to Example
            </button>
            <button 
              className="btn btn-outline-danger"
              onClick={clearCode}
              disabled={loading}
            >
              Clear
            </button>
          </div>
        </div>
      </div>

      {/* Main 2-Pane Layout */}
      <div className="row g-3">
        {/* Left Pane - Code Editor */}
        <div className="col-md-6">
          <div className="frosted-card h-100" style={{ display: 'flex', flexDirection: 'column' }}>
            <div className="p-3 border-bottom d-flex justify-content-between align-items-center">
              <h6 className="mb-0">
                <i className="bi bi-file-code me-2"></i>
                Code Input
              </h6>
              <div className="d-flex gap-3">
                <small style={{ color: 'var(--bs-gray-200)' }}>
                  <i className="bi bi-type me-1"></i>
                  {code.length} chars
                </small>
                <small style={{ color: 'var(--bs-gray-200)' }}>
                  <i className="bi bi-list-ol me-1"></i>
                  {code.split('\n').length} lines
                </small>
              </div>
            </div>
            <div className="p-0 flex-grow-1">
              <CodeEditor 
                ref={codeEditorRef}
                value={code}
                onChange={setCode}
                language="javascript"
                placeholder="Enter your code here for analysis..."
                className="h-100"
              />
            </div>
          </div>
        </div>

        {/* Right Pane - Analysis Output */}
        <div className="col-md-6">
          <AnalysisOutput 
            result={analysisResult}
            loading={loading}
            error={error}
            mode={selectedMode}
            onRetry={handleRetry}
          />
        </div>
      </div>

      {/* Footer with tips */}
      <div className="row mt-4">
        <div className="col-12">
          <div className="frosted-card p-3">
            <small style={{ 
              color: 'var(--bs-gray-200)',
              opacity: 0.95
            }}>
              <i className="bi bi-lightbulb me-1"></i>
              <strong>Tips:</strong> Try different analysis modes to understand code from various perspectives. 
              Preview for structure, Skim for abstractions, Scan for specific elements, Detailed for algorithms, 
              and Critical for quality assessment.
            </small>
          </div>
        </div>
      </div>
    </div>
  );
}
