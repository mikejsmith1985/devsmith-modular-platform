import React, { useState } from 'react';
import { reviewApi } from '../utils/api';

/**
 * RepoImportModal - GitHub Repository Import Modal
 * 
 * Allows users to import code from GitHub repositories with two modes:
 * 1. Quick Repo Scan - Fetches 5-8 core files for instant analysis (~2 seconds)
 * 2. Full Repository Browser - Fetches complete file tree for exploration
 * 
 * Features:
 * - GitHub URL validation (github.com/owner/repo format)
 * - Branch selection (defaults to 'main')
 * - Mode selection via radio buttons
 * - Loading states with progress indicators
 * - Error handling with user-friendly messages
 * - Success callbacks for parent component integration
 */
export default function RepoImportModal({ show, onClose, onSuccess }) {
  // Form state
  const [githubUrl, setGithubUrl] = useState('');
  const [branch, setBranch] = useState('main');
  const [importMode, setImportMode] = useState('quick'); // 'quick' or 'full'
  
  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [validationError, setValidationError] = useState(null);

  // Validate GitHub URL format
  const validateGithubUrl = (url) => {
    // Accept formats:
    // - github.com/owner/repo
    // - https://github.com/owner/repo
    // - http://github.com/owner/repo
    const githubPattern = /^(https?:\/\/)?(www\.)?github\.com\/[\w-]+\/[\w.-]+\/?$/;
    
    if (!url.trim()) {
      return 'GitHub URL is required';
    }
    
    if (!githubPattern.test(url.trim())) {
      return 'Invalid GitHub URL format. Expected: github.com/owner/repo';
    }
    
    return null;
  };

  // Extract owner and repo from URL
  const parseGithubUrl = (url) => {
    const cleaned = url.trim().replace(/^(https?:\/\/)?(www\.)?/, '').replace(/\/$/, '');
    const parts = cleaned.split('/');
    if (parts.length >= 3 && parts[0] === 'github.com') {
      return {
        owner: parts[1],
        repo: parts[2],
        fullUrl: cleaned
      };
    }
    return null;
  };

  // Handle URL input change with validation
  const handleUrlChange = (e) => {
    const url = e.target.value;
    setGithubUrl(url);
    setValidationError(null);
    setError(null);
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validate URL
    const urlError = validateGithubUrl(githubUrl);
    if (urlError) {
      setValidationError(urlError);
      return;
    }
    
    const parsed = parseGithubUrl(githubUrl);
    if (!parsed) {
      setValidationError('Could not parse GitHub URL');
      return;
    }

    setLoading(true);
    setError(null);
    setValidationError(null);

    try {
      let result;
      
      if (importMode === 'quick') {
        // Quick Repo Scan Mode - Fetch 5-8 core files
        result = await reviewApi.githubQuickScan(parsed.fullUrl, branch);
      } else {
        // Full Browser Mode - Fetch complete tree structure
        result = await reviewApi.githubGetTree(parsed.fullUrl, branch);
      }

      // Success - call parent callback with results
      onSuccess({
        mode: importMode,
        data: result,
        repoInfo: {
          owner: parsed.owner,
          repo: parsed.repo,
          branch: branch,
          url: parsed.fullUrl
        }
      });
      
      // Reset form and close modal
      handleClose();
      
    } catch (err) {
      console.error('GitHub import failed:', err);
      
      // User-friendly error messages
      if (err.message.includes('404')) {
        setError('Repository not found. Check the URL and branch name.');
      } else if (err.message.includes('403')) {
        setError('Access denied. You may need to authenticate for private repositories.');
      } else if (err.message.includes('Authentication required')) {
        setError('Please log in to access GitHub repositories.');
      } else {
        setError(err.message || 'Failed to import repository. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  // Handle modal close
  const handleClose = () => {
    if (!loading) {
      setGithubUrl('');
      setBranch('main');
      setImportMode('quick');
      setError(null);
      setValidationError(null);
      onClose();
    }
  };

  // Don't render if not shown
  if (!show) return null;

  return (
    <>
      {/* Bootstrap Modal Backdrop */}
      <div 
        className="modal-backdrop fade show" 
        onClick={handleClose}
        style={{ zIndex: 1040 }}
      />
      
      {/* Bootstrap Modal */}
      <div 
        className="modal fade show d-block" 
        tabIndex="-1" 
        role="dialog"
        style={{ zIndex: 1050 }}
      >
        <div className="modal-dialog modal-dialog-centered modal-lg" role="document">
          <div className="modal-content" style={{ 
            backgroundColor: 'var(--bs-body-bg)', 
            border: '2px solid rgba(99, 102, 241, 0.3)',
            borderRadius: '12px',
            boxShadow: '0 10px 40px rgba(0, 0, 0, 0.3)'
          }}>
            
            {/* Modal Header */}
            <div className="modal-header" style={{ 
              borderBottom: '1px solid rgba(99, 102, 241, 0.2)',
              padding: '1.5rem'
            }}>
              <h5 className="modal-title">
                <i className="bi bi-github me-2 text-primary"></i>
                Import from GitHub
              </h5>
              <button 
                type="button" 
                className="btn-close" 
                onClick={handleClose}
                disabled={loading}
                aria-label="Close"
              />
            </div>

            {/* Modal Body */}
            <div className="modal-body" style={{ padding: '1.5rem' }}>
              <form onSubmit={handleSubmit}>
                
                {/* GitHub URL Input */}
                <div className="mb-3">
                  <label htmlFor="githubUrl" className="form-label">
                    <strong>GitHub Repository URL</strong>
                  </label>
                  <input
                    type="text"
                    id="githubUrl"
                    className={`form-control ${validationError ? 'is-invalid' : ''}`}
                    placeholder="github.com/owner/repo or https://github.com/owner/repo"
                    value={githubUrl}
                    onChange={handleUrlChange}
                    disabled={loading}
                    autoFocus
                    style={{
                      padding: '0.75rem',
                      fontSize: '1rem',
                      borderRadius: '8px'
                    }}
                  />
                  {validationError && (
                    <div className="invalid-feedback d-block">
                      <i className="bi bi-exclamation-circle me-1"></i>
                      {validationError}
                    </div>
                  )}
                  <small className="form-text text-muted">
                    <i className="bi bi-info-circle me-1"></i>
                    Enter the repository URL (e.g., github.com/golang/go)
                  </small>
                </div>

                {/* Branch Input */}
                <div className="mb-3">
                  <label htmlFor="branch" className="form-label">
                    <strong>Branch</strong>
                  </label>
                  <input
                    type="text"
                    id="branch"
                    className="form-control"
                    placeholder="main"
                    value={branch}
                    onChange={(e) => setBranch(e.target.value)}
                    disabled={loading}
                    style={{
                      padding: '0.75rem',
                      fontSize: '1rem',
                      borderRadius: '8px'
                    }}
                  />
                  <small className="form-text text-muted">
                    <i className="bi bi-info-circle me-1"></i>
                    Default: main (or master for older repos)
                  </small>
                </div>

                {/* Import Mode Selection */}
                <div className="mb-4">
                  <label className="form-label d-block">
                    <strong>Import Mode</strong>
                  </label>
                  
                  {/* Quick Scan Option */}
                  <div className="form-check mb-2 p-3" style={{
                    backgroundColor: importMode === 'quick' ? 'rgba(99, 102, 241, 0.1)' : 'transparent',
                    border: `2px solid ${importMode === 'quick' ? 'rgba(99, 102, 241, 0.4)' : 'rgba(99, 102, 241, 0.2)'}`,
                    borderRadius: '8px',
                    cursor: 'pointer',
                    transition: 'all 0.2s'
                  }} onClick={() => !loading && setImportMode('quick')}>
                    <input
                      className="form-check-input"
                      type="radio"
                      name="importMode"
                      id="quickScan"
                      value="quick"
                      checked={importMode === 'quick'}
                      onChange={(e) => setImportMode(e.target.value)}
                      disabled={loading}
                      style={{ cursor: 'pointer' }}
                    />
                    <label className="form-check-label ms-2" htmlFor="quickScan" style={{ cursor: 'pointer' }}>
                      <strong>Quick Repo Scan</strong>
                      <span className="badge bg-success ms-2">Fast</span>
                      <div className="text-muted mt-1" style={{ fontSize: '0.9rem' }}>
                        <i className="bi bi-lightning-fill me-1"></i>
                        Fetches 5-8 core files (README, package files, entry points) in ~2 seconds.
                        Best for quick project assessment.
                      </div>
                    </label>
                  </div>

                  {/* Full Browser Option */}
                  <div className="form-check p-3" style={{
                    backgroundColor: importMode === 'full' ? 'rgba(99, 102, 241, 0.1)' : 'transparent',
                    border: `2px solid ${importMode === 'full' ? 'rgba(99, 102, 241, 0.4)' : 'rgba(99, 102, 241, 0.2)'}`,
                    borderRadius: '8px',
                    cursor: 'pointer',
                    transition: 'all 0.2s'
                  }} onClick={() => !loading && setImportMode('full')}>
                    <input
                      className="form-check-input"
                      type="radio"
                      name="importMode"
                      id="fullBrowser"
                      value="full"
                      checked={importMode === 'full'}
                      onChange={(e) => setImportMode(e.target.value)}
                      disabled={loading}
                      style={{ cursor: 'pointer' }}
                    />
                    <label className="form-check-label ms-2" htmlFor="fullBrowser" style={{ cursor: 'pointer' }}>
                      <strong>Full Repository Browser</strong>
                      <span className="badge bg-primary ms-2">Complete</span>
                      <div className="text-muted mt-1" style={{ fontSize: '0.9rem' }}>
                        <i className="bi bi-folder-fill me-1"></i>
                        Fetches complete file tree structure. Explore all files and folders.
                        Files loaded on-demand when selected.
                      </div>
                    </label>
                  </div>
                </div>

                {/* Error Display */}
                {error && (
                  <div className="alert alert-danger d-flex align-items-center" role="alert">
                    <i className="bi bi-exclamation-triangle-fill me-2"></i>
                    <div>{error}</div>
                  </div>
                )}

                {/* Loading State */}
                {loading && (
                  <div className="alert alert-info d-flex align-items-center" role="alert">
                    <div className="spinner-border spinner-border-sm me-2" role="status">
                      <span className="visually-hidden">Loading...</span>
                    </div>
                    <div>
                      {importMode === 'quick' 
                        ? 'Fetching core files from repository...' 
                        : 'Fetching repository structure...'}
                    </div>
                  </div>
                )}
              </form>
            </div>

            {/* Modal Footer */}
            <div className="modal-footer" style={{ 
              borderTop: '1px solid rgba(99, 102, 241, 0.2)',
              padding: '1rem 1.5rem'
            }}>
              <button 
                type="button" 
                className="btn btn-outline-secondary" 
                onClick={handleClose}
                disabled={loading}
              >
                Cancel
              </button>
              <button 
                type="button" 
                className="btn btn-primary"
                onClick={handleSubmit}
                disabled={loading || !githubUrl.trim()}
              >
                {loading ? (
                  <>
                    <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                    Importing...
                  </>
                ) : (
                  <>
                    <i className="bi bi-download me-2"></i>
                    Import Repository
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
