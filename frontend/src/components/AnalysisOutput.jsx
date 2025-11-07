import React, { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { analysisModesConfig } from './AnalysisModeSelector';
import DetailedAnalysisView from './DetailedAnalysisView';

export default function AnalysisOutput({ 
  result, 
  loading = false, 
  error = null, 
  mode = 'preview',
  onRetry = null 
}) {
  const [viewMode, setViewMode] = useState('markdown'); // 'markdown' or 'raw'
  const [fontSize, setFontSize] = useState('medium'); // 'xsmall', 'small', 'medium', 'large', 'xlarge'

  const fontSizes = {
    xsmall: '0.875rem',    // 14px
    small: '1.0rem',       // 16px
    medium: '1.125rem',    // 18px (default - middle size)
    large: '1.25rem',      // 20px
    xlarge: '1.375rem'     // 22px
  };

  if (loading) {
    return (
      <div className="analysis-output frosted-card h-100" style={{ display: 'flex', flexDirection: 'column' }}>
        <div className="d-flex justify-content-center align-items-center flex-grow-1">
          <div className="text-center">
            <div className="spinner-border text-primary mb-3" role="status" style={{ width: '3rem', height: '3rem' }}>
              <span className="visually-hidden">Analyzing code...</span>
            </div>
            <h6 className="text-muted mb-2">
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
      <div className="analysis-output frosted-card h-100" style={{ display: 'flex', flexDirection: 'column' }}>
        <div className="alert alert-danger m-3" role="alert">
          <h6 className="alert-heading">
            <i className="bi bi-exclamation-triangle-fill me-2"></i>
            Analysis Failed
          </h6>
          <p className="mb-2">{error}</p>
          {onRetry && (
            <button className="btn btn-outline-danger btn-sm" onClick={onRetry}>
              <i className="bi bi-arrow-clockwise me-1"></i>
              Try Again
            </button>
          )}
        </div>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="analysis-output frosted-card h-100" style={{ display: 'flex', flexDirection: 'column' }}>
        <div className="d-flex justify-content-center align-items-center text-center flex-grow-1">
          <div>
            <div className="fs-1 mb-3">📝</div>
            <h6 style={{ color: 'var(--bs-gray-200)' }} className="mb-2">No Analysis Yet</h6>
            <p style={{ color: 'var(--bs-gray-300)' }} className="mb-0">
              Add some code to the left pane, select an analysis mode, and click "Analyze" to get started.
            </p>
          </div>
        </div>
      </div>
    );
  }

  // Convert result to markdown-friendly format (for Formatted view)
  const formatAsMarkdown = (text) => {
    if (!text) return '';
    
    if (typeof text === 'string') {
      // Clean up the text
      let cleaned = text.replace(/<[^>]*>/g, ''); // Remove HTML tags
      
      // Convert common patterns to markdown
      cleaned = cleaned.replace(/\*\*Note:\*\*/g, '> **📌 Note:**');
      cleaned = cleaned.replace(/\*\*Important:\*\*/g, '> **⚠️ Important:**');
      cleaned = cleaned.replace(/\*\*Tip:\*\*/g, '> **💡 Tip:**');
      
      // Ensure proper line breaks
      cleaned = cleaned.replace(/\n\n/g, '\n\n');
      
      return cleaned;
    }
    
    if (typeof text === 'object') {
      return JSON.stringify(text, null, 2);
    }
    
    return String(text);
  };

  // Get the raw, unformatted result (for Raw view)
  const getRawText = () => {
    // Extract actual markdown from HTML if present
    if (typeof result === 'string') {
      // If it contains HTML tags, use DOMParser for proper extraction
      if (result.includes('<div') || result.includes('<span') || result.includes('<p')) {
        try {
          // Use DOMParser for robust HTML parsing
          const parser = new DOMParser();
          const doc = parser.parseFromString(result, 'text/html');
          
          // Extract text with preserved structure
          let extractedText = '';
          
          // Process each node to preserve formatting
          const processNode = (node) => {
            if (node.nodeType === Node.TEXT_NODE) {
              // Add text content, trimming excessive whitespace
              const text = node.textContent;
              if (text.trim()) {
                extractedText += text;
              }
            } else if (node.nodeType === Node.ELEMENT_NODE) {
              const tagName = node.tagName.toLowerCase();
              
              // Add line breaks for block elements
              if (['p', 'div', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'li', 'br'].includes(tagName)) {
                if (extractedText && !extractedText.endsWith('\n')) {
                  extractedText += '\n';
                }
              }
              
              // Special handling for code blocks - preserve whitespace
              if (tagName === 'pre' || tagName === 'code') {
                extractedText += node.textContent;
                if (tagName === 'pre') extractedText += '\n';
              } else {
                // Recursively process child nodes
                node.childNodes.forEach(processNode);
              }
              
              // Add line breaks after block elements
              if (['p', 'div', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'li'].includes(tagName)) {
                if (!extractedText.endsWith('\n')) {
                  extractedText += '\n';
                }
              }
            }
          };
          
          doc.body.childNodes.forEach(processNode);
          
          // Clean up excessive line breaks (max 2 consecutive)
          extractedText = extractedText.replace(/\n{3,}/g, '\n\n');
          
          return extractedText.trim() || result;
        } catch (e) {
          console.warn('Failed to parse HTML, returning raw result:', e);
          return result;
        }
      }
      // Return as-is if no HTML detected
      return result;
    }
    if (typeof result === 'object') {
      // For objects, check if there's a 'result' or 'text' property with the actual content
      if (result.result) return String(result.result);
      if (result.text) return String(result.text);
      if (result.content) return String(result.content);
      return JSON.stringify(result, null, 2); // Pretty JSON as fallback
    }
    return String(result);
  };

  const markdownText = formatAsMarkdown(result);
  const rawText = getRawText();

  return (
    <div className="analysis-output frosted-card h-100" style={{ display: 'flex', flexDirection: 'column' }}>
      {/* Tab Navigation */}
      <div className="border-bottom d-flex align-items-center justify-content-between px-3">
        <ul className="nav nav-tabs border-0" role="tablist">
          <li className="nav-item" role="presentation">
            <button
              className={`nav-link ${viewMode === 'markdown' ? 'active' : ''}`}
              onClick={() => setViewMode('markdown')}
              type="button"
              style={{
                borderTopLeftRadius: '16px',
                borderBottom: viewMode === 'markdown' ? '3px solid #6366f1' : 'none'
              }}
            >
              <i className="bi bi-markdown me-2"></i>
              Formatted
            </button>
          </li>
          <li className="nav-item" role="presentation">
            <button
              className={`nav-link ${viewMode === 'raw' ? 'active' : ''}`}
              onClick={() => setViewMode('raw')}
              type="button"
              style={{
                borderBottom: viewMode === 'raw' ? '3px solid #6366f1' : 'none'
              }}
            >
              <i className="bi bi-code-slash me-2"></i>
              Raw
            </button>
          </li>
        </ul>
        
        {/* Font Size Selector */}
        <div className="d-flex align-items-center gap-2">
          <span style={{ 
            fontSize: '0.875rem', 
            color: 'var(--bs-gray-200)',
            opacity: 0.9
          }}>
            <i className="bi bi-type me-1"></i>
            Text Size:
          </span>
          <div className="btn-group btn-group-sm" role="group">
            <button
              type="button"
              className={`btn ${fontSize === 'xsmall' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('xsmall')}
              title="Extra Small (14px)"
              style={{ fontSize: '0.7rem', padding: '0.25rem 0.5rem' }}
            >
              A⁻⁻
            </button>
            <button
              type="button"
              className={`btn ${fontSize === 'small' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('small')}
              title="Small (16px)"
              style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}
            >
              A⁻
            </button>
            <button
              type="button"
              className={`btn ${fontSize === 'medium' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('medium')}
              title="Medium (18px) - Default"
              style={{ fontSize: '0.875rem', padding: '0.25rem 0.5rem' }}
            >
              A
            </button>
            <button
              type="button"
              className={`btn ${fontSize === 'large' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('large')}
              title="Large (20px)"
              style={{ fontSize: '1rem', padding: '0.25rem 0.5rem' }}
            >
              A⁺
            </button>
            <button
              type="button"
              className={`btn ${fontSize === 'xlarge' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('xlarge')}
              title="Extra Large (22px)"
              style={{ fontSize: '1.125rem', padding: '0.25rem 0.5rem' }}
            >
              A⁺⁺
            </button>
          </div>
        </div>
      </div>

      {/* Content Area - with flex-grow and overflow scroll */}
      <div className="p-4 flex-grow-1" style={{ overflowY: 'auto', overflowX: 'hidden' }}>
        {/* Use special view for Detailed mode */}
        {mode === 'detailed' && viewMode === 'markdown' ? (
          <DetailedAnalysisView 
            analysis={getRawText()} 
            fontSize={fontSize}
          />
        ) : viewMode === 'markdown' ? (
          <div className="markdown-content" style={{ 
            lineHeight: '1.8',
            fontSize: fontSizes[fontSize],
            wordWrap: 'break-word',
            overflowWrap: 'break-word'
          }}>
            <ReactMarkdown
              components={{
                // Custom rendering for better readability
                h1: ({node, ...props}) => <h1 className="mb-3 mt-4" {...props} />,
                h2: ({node, ...props}) => <h2 className="mb-3 mt-4" {...props} />,
                h3: ({node, ...props}) => <h3 className="mb-2 mt-3" {...props} />,
                h4: ({node, ...props}) => <h4 className="mb-2 mt-3" {...props} />,
                p: ({node, ...props}) => <p className="mb-3" style={{ wordWrap: 'break-word' }} {...props} />,
                ul: ({node, ...props}) => <ul className="mb-3" style={{ paddingLeft: '1.5rem' }} {...props} />,
                ol: ({node, ...props}) => <ol className="mb-3" style={{ paddingLeft: '1.5rem' }} {...props} />,
                li: ({node, ...props}) => <li className="mb-1" {...props} />,
                code: ({node, inline, ...props}) => 
                  inline ? (
                    <code className="px-2 py-1 rounded" style={{ 
                      background: 'rgba(99, 102, 241, 0.1)',
                      color: '#6366f1',
                      fontSize: '0.9em',
                      wordWrap: 'break-word'
                    }} {...props} />
                  ) : (
                    <pre className="p-3 rounded mb-3" style={{ 
                      background: 'rgba(0, 0, 0, 0.05)',
                      overflowX: 'auto',
                      whiteSpace: 'pre-wrap',
                      wordWrap: 'break-word'
                    }}>
                      <code {...props} />
                    </pre>
                  ),
                blockquote: ({node, ...props}) => (
                  <blockquote className="border-start border-4 border-primary ps-3 mb-3" 
                    style={{ 
                      background: 'rgba(99, 102, 241, 0.05)',
                      padding: '0.5rem 1rem',
                      borderRadius: '0.25rem'
                    }} 
                    {...props} 
                  />
                ),
              }}
            >
              {markdownText}
            </ReactMarkdown>
          </div>
        ) : (
          <pre className="mb-0" style={{ 
            whiteSpace: 'pre-wrap', 
            wordBreak: 'break-word',
            lineHeight: '1.6',
            fontSize: fontSizes[fontSize],
            fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace',
            overflowX: 'hidden'
          }}>
            {rawText}
          </pre>
        )}
      </div>
    </div>
  );
}
