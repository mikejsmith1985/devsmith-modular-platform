import React, { useState, useEffect, useRef } from 'react';

// Simple code editor component using textarea
// In the future, this could be replaced with Monaco Editor for better functionality
export default function CodeEditor({ 
  value = '', 
  onChange, 
  language = 'javascript',
  placeholder = 'Enter your code here...',
  readOnly = false,
  className = ''
}) {
  const textareaRef = useRef(null);

  // Auto-resize textarea based on content
  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = 'auto';
      textarea.style.height = Math.max(textarea.scrollHeight, 200) + 'px';
    }
  }, [value]);

  const handleChange = (e) => {
    if (onChange) {
      onChange(e.target.value);
    }
  };

  // Add line numbers and syntax highlighting classes for basic visual enhancement
  return (
    <div className={`code-editor-container ${className}`}>
      <textarea
        ref={textareaRef}
        value={value}
        onChange={handleChange}
        placeholder={placeholder}
        readOnly={readOnly}
        className="form-control font-monospace"
        style={{
          minHeight: '200px',
          resize: 'vertical',
          fontSize: '14px',
          lineHeight: '1.5',
          border: '1px solid #dee2e6',
          borderRadius: '0.375rem'
        }}
        spellCheck={false}
      />
      {/* Language indicator */}
      <small className="text-muted mt-1 d-block">
        Language: {language}
      </small>
    </div>
  );
}