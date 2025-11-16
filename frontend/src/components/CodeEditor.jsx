import React, { useRef, useState } from 'react';
import Editor from '@monaco-editor/react';
import { useTheme } from '../context/ThemeContext';

// Monaco Editor component with VS Code-style editing capabilities
// Supports syntax highlighting, autocomplete, code folding, and more
export default function CodeEditor({ 
  value = '', 
  onChange, 
  language = 'javascript',
  placeholder = 'Enter your code here...',
  readOnly = false,
  className = '',
  height = '600px'
}) {
  const { isDarkMode } = useTheme();
  const editorRef = useRef(null);
  const [fontSize, setFontSize] = useState('medium');

  const fontSizes = {
    xsmall: 12,    // 12px
    small: 14,     // 14px
    medium: 16,    // 16px (default - middle size)
    large: 18,     // 18px
    xlarge: 20     // 20px
  };

  // Map common file extensions to Monaco language IDs
  const getMonacoLanguage = (lang) => {
    const languageMap = {
      'js': 'javascript',
      'jsx': 'javascript',
      'ts': 'typescript',
      'tsx': 'typescript',
      'py': 'python',
      'go': 'go',
      'sql': 'sql',
      'json': 'json',
      'yaml': 'yaml',
      'yml': 'yaml',
      'md': 'markdown',
      'html': 'html',
      'css': 'css',
      'sh': 'shell',
      'bash': 'shell'
    };
    return languageMap[lang.toLowerCase()] || lang.toLowerCase();
  };

  const handleEditorChange = (value) => {
    if (onChange) {
      onChange(value || '');
    }
  };

  const handleEditorDidMount = (editor, monaco) => {
    editorRef.current = editor;
    
    // Optional: Add custom keybindings or configurations here
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      // Prevent default save behavior (browser save dialog)
      // Could trigger a save action if implemented
    });
  };

  const monacoLanguage = getMonacoLanguage(language);

  return (
    <div className={`code-editor-container ${className}`}>
      {/* Font Size Controls */}
      <div className="d-flex align-items-center justify-content-between mb-2 px-2">
        <small style={{ color: 'var(--bs-gray-200)' }}>
          Language: <span className="badge" style={{ 
            backgroundColor: '#6366f1',
            color: 'white',
            fontWeight: '500'
          }}>{monacoLanguage}</span>
        </small>
        
        <div className="d-flex align-items-center gap-2">
          <span style={{ 
            fontSize: '0.875rem', 
            color: 'var(--bs-gray-200)',
            opacity: 0.9
          }}>
            <i className="bi bi-type me-1"></i>
            Editor Size:
          </span>
          <div className="btn-group btn-group-sm" role="group">
            <button
              type="button"
              className={`btn ${fontSize === 'xsmall' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('xsmall')}
              title="Extra Small (12px)"
              style={{ fontSize: '0.7rem', padding: '0.25rem 0.5rem' }}
            >
              A⁻⁻
            </button>
            <button
              type="button"
              className={`btn ${fontSize === 'small' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('small')}
              title="Small (14px)"
              style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}
            >
              A⁻
            </button>
            <button
              type="button"
              className={`btn ${fontSize === 'medium' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('medium')}
              title="Medium (16px) - Default"
              style={{ fontSize: '0.875rem', padding: '0.25rem 0.5rem' }}
            >
              A
            </button>
            <button
              type="button"
              className={`btn ${fontSize === 'large' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('large')}
              title="Large (18px)"
              style={{ fontSize: '1rem', padding: '0.25rem 0.5rem' }}
            >
              A⁺
            </button>
            <button
              type="button"
              className={`btn ${fontSize === 'xlarge' ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => setFontSize('xlarge')}
              title="Extra Large (20px)"
              style={{ fontSize: '1.125rem', padding: '0.25rem 0.5rem' }}
            >
              A⁺⁺
            </button>
          </div>
        </div>
      </div>
      
      <div className="editor-wrapper" style={{ 
        border: '1px solid var(--border-color, #dee2e6)',
        borderRadius: '0.375rem',
        overflow: 'hidden'
      }}>
        <Editor
          height={height}
          defaultLanguage={monacoLanguage}
          language={monacoLanguage}
          theme={isDarkMode ? "vs-dark" : "light"}
          value={value}
          onChange={handleEditorChange}
          onMount={handleEditorDidMount}
          options={{
            // Editor behavior
            readOnly: readOnly,
            automaticLayout: true,
            scrollBeyondLastLine: false,
            wordWrap: 'on',
            wrappingIndent: 'indent',
            
            // Line numbers and folding
            lineNumbers: 'on',
            folding: true,
            foldingStrategy: 'indentation',
            showFoldingControls: 'always',
            
            // Minimap
            minimap: {
              enabled: true,
              maxColumn: 120,
              renderCharacters: true,
              showSlider: 'mouseover',
              side: 'right'
            },
            
            // Font and rendering
            fontSize: fontSizes[fontSize],
            fontFamily: "'Fira Code', 'Cascadia Code', 'Courier New', monospace",
            fontLigatures: true,
            lineHeight: 1.6,
            letterSpacing: 0.5,
            renderWhitespace: 'selection',
            renderLineHighlight: 'all',
            
            // Indentation
            tabSize: 2,
            insertSpaces: true,
            detectIndentation: true,
            
            // Scrolling
            scrollbar: {
              vertical: 'visible',
              horizontal: 'visible',
              verticalScrollbarSize: 10,
              horizontalScrollbarSize: 10
            },
            
            // Suggestions and IntelliSense
            quickSuggestions: {
              other: true,
              comments: false,
              strings: false
            },
            suggestOnTriggerCharacters: true,
            acceptSuggestionOnCommitCharacter: true,
            acceptSuggestionOnEnter: 'on',
            
            // Brackets and matching
            bracketPairColorization: {
              enabled: true
            },
            matchBrackets: 'always',
            autoClosingBrackets: 'always',
            autoClosingQuotes: 'always',
            
            // Other useful features
            contextmenu: true,
            mouseWheelZoom: true,
            smoothScrolling: true,
            cursorBlinking: 'smooth',
            cursorSmoothCaretAnimation: 'on',
            formatOnPaste: true,
            formatOnType: false
          }}
          loading={
            <div className="d-flex justify-content-center align-items-center" style={{ height }}>
              <div className="spinner-border text-primary" role="status">
                <span className="visually-hidden">Loading editor...</span>
              </div>
            </div>
          }
        />
      </div>
    </div>
  );
}