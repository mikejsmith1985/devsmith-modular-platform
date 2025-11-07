import React from 'react';
import { analysisModesConfig } from './AnalysisModeSelector';

export default function AnalysisOutput({ 
  result, 
  loading = false, 
  error = null, 
  mode = 'preview',
  onRetry = null 
}) {
  if (loading) {
    return (
      <div className="analysis-output">
        <div className="d-flex justify-content-center align-items-center" style={{ minHeight: '200px' }}>
          <div className="text-center">
            <div className="spinner-border text-primary mb-3" role="status">
              <span className="visually-hidden">Analyzing code...</span>
            </div>
            <h6 className="text-muted">
              Running {analysisModesConfig[mode]?.name || mode} analysis...
            </h6>
            <small className="text-muted">
              This may take a few moments depending on code complexity and selected model.
            </small>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="analysis-output">
        <div className="alert alert-danger" role="alert">
          <h6 className="alert-heading">Analysis Failed</h6>
          <p className="mb-2">{error}</p>
          {onRetry && (
            <button className="btn btn-outline-danger btn-sm" onClick={onRetry}>
              Try Again
            </button>
          )}
        </div>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="analysis-output">
        <div className="d-flex justify-content-center align-items-center text-center" style={{ minHeight: '200px' }}>
          <div>
            <div className="fs-1 text-muted mb-3">📝</div>
            <h6 className="text-muted">No Analysis Yet</h6>
            <p className="text-muted mb-0">
              Add some code to the left pane, select an analysis mode, and click "Analyze" to get started.
            </p>
          </div>
        </div>
      </div>
    );
  }

  // Handle different types of analysis results
  const renderResult = () => {
    if (typeof result === 'string') {
      return (
        <div className="analysis-result">
          <pre className="bg-light p-3 rounded" style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
            {result}
          </pre>
        </div>
      );
    }

    if (typeof result === 'object') {
      // Try to render structured result
      if (result.summary && result.details) {
        return (
          <div className="analysis-result">
            <div className="mb-3">
              <h6>Summary</h6>
              <div className="bg-light p-3 rounded">
                {result.summary}
              </div>
            </div>
            <div>
              <h6>Details</h6>
              <div className="bg-light p-3 rounded">
                <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                  {typeof result.details === 'string' ? result.details : JSON.stringify(result.details, null, 2)}
                </pre>
              </div>
            </div>
          </div>
        );
      }

      // Fallback to JSON display
      return (
        <div className="analysis-result">
          <h6>Analysis Result</h6>
          <pre className="bg-light p-3 rounded" style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
            {JSON.stringify(result, null, 2)}
          </pre>
        </div>
      );
    }

    return (
      <div className="analysis-result">
        <div className="bg-light p-3 rounded">
          {String(result)}
        </div>
      </div>
    );
  };

  return (
    <div className="analysis-output">
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h6 className="mb-0">
          <span className={`badge bg-${analysisModesConfig[mode]?.color || 'primary'} me-2`}>
            {analysisModesConfig[mode]?.icon || '📋'}
          </span>
          {analysisModesConfig[mode]?.name || mode} Analysis
        </h6>
        <small className="text-muted">
          {new Date().toLocaleTimeString()}
        </small>
      </div>
      {renderResult()}
    </div>
  );
}