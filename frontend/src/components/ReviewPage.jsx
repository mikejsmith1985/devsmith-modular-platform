import React, { useState, useRef } from 'react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import { Link } from 'react-router-dom';
import CodeEditor from './CodeEditor';
import AnalysisModeSelector from './AnalysisModeSelector';
import ModelSelector from './ModelSelector';
import AnalysisOutput from './AnalysisOutput';
import FileTabs from './FileTabs';
import FileTreeBrowser from './FileTreeBrowser';
import RepoImportModal from './RepoImportModal';
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
  
  // Multi-file state management
  const [files, setFiles] = useState([
    {
      id: 'file_1',
      name: 'example.js',
      language: 'javascript',
      content: defaultCode,
      hasUnsavedChanges: false,
      path: null
    }
  ]);
  const [activeFileId, setActiveFileId] = useState('file_1');
  
  // Get current active file
  const activeFile = files.find(f => f.id === activeFileId);
  const code = activeFile?.content || '';
  
  // State management
  const [selectedMode, setSelectedMode] = useState('preview');
  const [selectedModel, setSelectedModel] = useState('');
  const [scanQuery, setScanQuery] = useState('');
  const [analysisResult, setAnalysisResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [sessionId] = useState(() => `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`);
  
  // GitHub import modal state
  const [showImportModal, setShowImportModal] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null); // Stores current repo info

  // Refs for managing focus
  const codeEditorRef = useRef(null);

  // File management functions
  const handleCodeChange = (newCode) => {
    setFiles(prevFiles => prevFiles.map(file => 
      file.id === activeFileId 
        ? { ...file, content: newCode, hasUnsavedChanges: true }
        : file
    ));
  };

  const handleFileSelect = (fileId) => {
    setActiveFileId(fileId);
    // Clear analysis when switching files
    setAnalysisResult(null);
    setError(null);
  };

  const handleFileClose = (fileId) => {
    const fileIndex = files.findIndex(f => f.id === fileId);
    const newFiles = files.filter(f => f.id !== fileId);
    
    // If closing the last file, create a new default file
    if (newFiles.length === 0) {
      const newFileId = `file_${Date.now()}`;
      setFiles([{
        id: newFileId,
        name: 'untitled.js',
        language: 'javascript',
        content: '',
        hasUnsavedChanges: false,
        path: null
      }]);
      setActiveFileId(newFileId);
      return;
    }
    
    setFiles(newFiles);
    
    // If closing active file, switch to another file
    if (fileId === activeFileId) {
      // Select previous file, or first file if closing first
      const newActiveIndex = fileIndex > 0 ? fileIndex - 1 : 0;
      setActiveFileId(newFiles[newActiveIndex].id);
    }
  };

  const handleFileAdd = () => {
    const newFileId = `file_${Date.now()}`;
    const newFile = {
      id: newFileId,
      name: `untitled-${files.length + 1}.js`,
      language: 'javascript',
      content: '// New file\n',
      hasUnsavedChanges: false,
      path: null
    };
    
    setFiles(prevFiles => [...prevFiles, newFile]);
    setActiveFileId(newFileId);
    setAnalysisResult(null);
    setError(null);
  };

  const handleFileRename = (fileId, newName) => {
    // Detect language from file extension
    const extension = newName.split('.').pop().toLowerCase();
    const languageMap = {
      'js': 'javascript',
      'jsx': 'javascript',
      'ts': 'typescript',
      'tsx': 'typescript',
      'py': 'python',
      'go': 'go',
      'rs': 'rust',
      'java': 'java',
      'c': 'c',
      'cpp': 'cpp',
      'cs': 'csharp',
      'sql': 'sql',
      'html': 'html',
      'css': 'css',
      'json': 'json',
      'yaml': 'yaml',
      'yml': 'yaml',
      'md': 'markdown',
      'sh': 'shell',
      'bash': 'shell'
    };
    
    const detectedLanguage = languageMap[extension] || 'javascript';
    
    setFiles(prevFiles => prevFiles.map(file =>
      file.id === fileId
        ? { ...file, name: newName, language: detectedLanguage, hasUnsavedChanges: true }
        : file
    ));
  };

  // Handle GitHub import success
  const handleGitHubImportSuccess = (importData) => {
    const { mode, data, repoInfo: repo } = importData;
    
    // Store repo info for reference
    setRepoInfo(repo);
    
    if (mode === 'quick') {
      // Quick Scan Mode - Open core files in tabs
      console.log('Quick scan data:', data);
      
      // Clear existing files
      setFiles([]);
      
      // Create tabs for each fetched file
      const newFiles = [];
      
      // Add README if present
      if (data.readme) {
        newFiles.push({
          id: `file_readme_${Date.now()}`,
          name: 'README.md',
          language: 'markdown',
          content: data.readme,
          hasUnsavedChanges: false,
          path: 'README.md',
          repoInfo: repo
        });
      }
      
      // Add entry point files
      if (data.entry_points && Array.isArray(data.entry_points)) {
        data.entry_points.forEach((entry, idx) => {
          if (entry.content) {
            const fileName = entry.path.split('/').pop();
            const extension = fileName.split('.').pop().toLowerCase();
            
            // Detect language from extension
            const languageMap = {
              'js': 'javascript', 'jsx': 'javascript',
              'ts': 'typescript', 'tsx': 'typescript',
              'py': 'python', 'go': 'go', 'rs': 'rust',
              'java': 'java', 'c': 'c', 'cpp': 'cpp',
              'json': 'json', 'yaml': 'yaml', 'yml': 'yaml'
            };
            
            newFiles.push({
              id: `file_entry_${idx}_${Date.now()}`,
              name: fileName,
              language: languageMap[extension] || 'plaintext',
              content: entry.content,
              hasUnsavedChanges: false,
              path: entry.path,
              repoInfo: repo
            });
          }
        });
      }
      
      // Add config files
      if (data.config_files && Array.isArray(data.config_files)) {
        data.config_files.forEach((config, idx) => {
          if (config.content) {
            const fileName = config.path.split('/').pop();
            const extension = fileName.split('.').pop().toLowerCase();
            
            newFiles.push({
              id: `file_config_${idx}_${Date.now()}`,
              name: fileName,
              language: extension === 'json' ? 'json' : 'yaml',
              content: config.content,
              hasUnsavedChanges: false,
              path: config.path,
              repoInfo: repo
            });
          }
        });
      }
      
      // If no files were added, create a placeholder
      if (newFiles.length === 0) {
        newFiles.push({
          id: `file_${Date.now()}`,
          name: 'info.txt',
          language: 'plaintext',
          content: `Repository: ${repo.owner}/${repo.repo}\nBranch: ${repo.branch}\n\nNo core files found.`,
          hasUnsavedChanges: false,
          path: null,
          repoInfo: repo
        });
      }
      
      setFiles(newFiles);
      setActiveFileId(newFiles[0].id);
      
    } else {
      // Full Browser Mode - Show tree in FileTreeBrowser
      console.log('Full tree data:', data);
      // TODO: Wire FileTreeBrowser component to display tree
      // FileTreeBrowser will handle file selection → file fetch → open in FileTabs
    }
    
    // Clear any errors
    setError(null);
  };

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
            <div className="d-flex gap-2 align-items-center">
              {repoInfo && (
                <div className="text-end me-3">
                  <small className="text-muted d-block">
                    <i className="bi bi-github me-1"></i>
                    {repoInfo.owner}/{repoInfo.repo}
                  </small>
                  <small className="text-muted">
                    <i className="bi bi-git me-1"></i>
                    {repoInfo.branch}
                  </small>
                </div>
              )}
              <button 
                className="btn btn-primary btn-sm"
                onClick={() => setShowImportModal(true)}
                disabled={loading}
              >
                <i className="bi bi-github me-2"></i>
                Import from GitHub
              </button>
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
            
            {/* File Tabs */}
            <FileTabs
              files={files}
              activeFileId={activeFileId}
              onFileSelect={handleFileSelect}
              onFileClose={handleFileClose}
              onFileAdd={handleFileAdd}
              onFileRename={handleFileRename}
            />
            
            <div className="p-0 flex-grow-1">
              <CodeEditor 
                ref={codeEditorRef}
                value={code}
                onChange={handleCodeChange}
                language={activeFile?.language || 'javascript'}
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

      {/* GitHub Import Modal */}
      <RepoImportModal 
        show={showImportModal}
        onClose={() => setShowImportModal(false)}
        onSuccess={handleGitHubImportSuccess}
      />
    </div>
  );
}
