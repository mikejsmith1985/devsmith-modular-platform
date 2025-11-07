import React, { useState } from 'react';

/**
 * FileTreeBrowser Component
 * 
 * Displays hierarchical file tree for repository navigation
 * Features:
 * - Collapsible folders
 * - File selection (single and multiple)
 * - Search/filter files
 * - File type icons
 * - Batch actions (analyze selected files)
 */
export default function FileTreeBrowser({
  treeData = [],
  selectedFiles = [],
  onFileSelect,
  onFilesAnalyze,
  loading = false
}) {
  const [expandedFolders, setExpandedFolders] = useState(new Set());
  const [searchQuery, setSearchQuery] = useState('');

  const toggleFolder = (path) => {
    setExpandedFolders(prev => {
      const newSet = new Set(prev);
      if (newSet.has(path)) {
        newSet.delete(path);
      } else {
        newSet.add(path);
      }
      return newSet;
    });
  };

  const handleFileClick = (file, event) => {
    if (onFileSelect) {
      // Support multi-select with Ctrl/Cmd
      const isMultiSelect = event.ctrlKey || event.metaKey;
      onFileSelect(file, isMultiSelect);
    }
  };

  const handleAnalyzeSelected = () => {
    if (onFilesAnalyze && selectedFiles.length > 0) {
      onFilesAnalyze(selectedFiles);
    }
  };

  // Filter tree based on search query
  const filterTree = (nodes, query) => {
    if (!query) return nodes;
    
    const lowerQuery = query.toLowerCase();
    return nodes.filter(node => {
      if (node.type === 'file') {
        return node.name.toLowerCase().includes(lowerQuery) ||
               node.path.toLowerCase().includes(lowerQuery);
      } else {
        // For directories, check if any children match
        const filteredChildren = filterTree(node.children || [], query);
        return filteredChildren.length > 0 || node.name.toLowerCase().includes(lowerQuery);
      }
    }).map(node => {
      if (node.type === 'directory') {
        return {
          ...node,
          children: filterTree(node.children || [], query)
        };
      }
      return node;
    });
  };

  const filteredTree = filterTree(treeData, searchQuery);

  return (
    <div className="file-tree-browser frosted-card" style={{
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      overflow: 'hidden'
    }}>
      {/* Header with search */}
      <div className="p-3 border-bottom" style={{
        borderBottomColor: 'rgba(99, 102, 241, 0.2)'
      }}>
        <div className="d-flex align-items-center justify-content-between mb-2">
          <h6 className="mb-0" style={{ color: '#6366f1' }}>
            <i className="bi bi-folder2-open me-2"></i>
            File Explorer
          </h6>
          {selectedFiles.length > 0 && (
            <span className="badge" style={{
              backgroundColor: 'rgba(99, 102, 241, 0.2)',
              color: '#6366f1'
            }}>
              {selectedFiles.length} selected
            </span>
          )}
        </div>

        {/* Search bar */}
        <div className="input-group input-group-sm">
          <span className="input-group-text">
            <i className="bi bi-search"></i>
          </span>
          <input
            type="text"
            className="form-control"
            placeholder="Search files..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchQuery && (
            <button
              className="btn btn-outline-secondary"
              onClick={() => setSearchQuery('')}
              title="Clear search"
            >
              <i className="bi bi-x"></i>
            </button>
          )}
        </div>
      </div>

      {/* File tree */}
      <div className="flex-grow-1" style={{
        overflowY: 'auto',
        overflowX: 'hidden',
        padding: '0.5rem'
      }}>
        {filteredTree.length === 0 ? (
          <div className="text-center text-muted p-4">
            <i className="bi bi-folder-x" style={{ fontSize: '2rem' }}></i>
            <p className="mt-2 mb-0">
              {searchQuery ? 'No files match your search' : 'No files found'}
            </p>
          </div>
        ) : (
          <TreeNodeRenderer
            nodes={filteredTree}
            expandedFolders={expandedFolders}
            selectedFiles={selectedFiles}
            onToggleFolder={toggleFolder}
            onFileClick={handleFileClick}
            level={0}
          />
        )}
      </div>

      {/* Actions footer */}
      {selectedFiles.length > 0 && (
        <div className="p-3 border-top" style={{
          borderTopColor: 'rgba(99, 102, 241, 0.2)',
          backgroundColor: 'rgba(99, 102, 241, 0.05)'
        }}>
          <button
            className="btn btn-primary w-100"
            onClick={handleAnalyzeSelected}
            disabled={loading}
          >
            {loading ? (
              <>
                <span className="spinner-border spinner-border-sm me-2"></span>
                Analyzing...
              </>
            ) : (
              <>
                <i className="bi bi-lightning-charge me-2"></i>
                Analyze {selectedFiles.length} File{selectedFiles.length !== 1 ? 's' : ''}
              </>
            )}
          </button>
        </div>
      )}
    </div>
  );
}

/**
 * TreeNodeRenderer - Recursive component for rendering tree nodes
 */
function TreeNodeRenderer({
  nodes,
  expandedFolders,
  selectedFiles,
  onToggleFolder,
  onFileClick,
  level
}) {
  return (
    <>
      {nodes.map((node, index) => {
        const isExpanded = expandedFolders.has(node.path);
        const isSelected = selectedFiles.some(f => f.path === node.path);
        const indent = level * 1.25; // rem

        if (node.type === 'directory') {
          return (
            <div key={node.path || index}>
              {/* Directory item */}
              <div
                onClick={() => onToggleFolder(node.path)}
                style={{
                  paddingLeft: `${indent}rem`,
                  paddingTop: '0.375rem',
                  paddingBottom: '0.375rem',
                  cursor: 'pointer',
                  borderRadius: '0.25rem',
                  transition: 'all 0.15s ease',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.backgroundColor = 'rgba(99, 102, 241, 0.08)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = 'transparent';
                }}
              >
                <i className={`bi bi-chevron-${isExpanded ? 'down' : 'right'}`}
                   style={{ fontSize: '0.75rem', color: '#6366f1' }}></i>
                <i className={`bi bi-folder${isExpanded ? '-open' : ''}`}
                   style={{ color: '#ec4899' }}></i>
                <span style={{ fontSize: '0.875rem', fontWeight: '500' }}>
                  {node.name}
                </span>
                {node.children && (
                  <span className="badge badge-sm" style={{
                    backgroundColor: 'rgba(99, 102, 241, 0.15)',
                    color: '#6366f1',
                    fontSize: '0.7rem',
                    marginLeft: 'auto'
                  }}>
                    {node.children.length}
                  </span>
                )}
              </div>

              {/* Children (if expanded) */}
              {isExpanded && node.children && (
                <TreeNodeRenderer
                  nodes={node.children}
                  expandedFolders={expandedFolders}
                  selectedFiles={selectedFiles}
                  onToggleFolder={onToggleFolder}
                  onFileClick={onFileClick}
                  level={level + 1}
                />
              )}
            </div>
          );
        } else {
          // File item
          return (
            <div
              key={node.path || index}
              onClick={(e) => onFileClick(node, e)}
              style={{
                paddingLeft: `${indent + 0.5}rem`,
                paddingTop: '0.375rem',
                paddingBottom: '0.375rem',
                cursor: 'pointer',
                borderRadius: '0.25rem',
                transition: 'all 0.15s ease',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                backgroundColor: isSelected ? 'rgba(99, 102, 241, 0.15)' : 'transparent'
              }}
              onMouseEnter={(e) => {
                if (!isSelected) {
                  e.currentTarget.style.backgroundColor = 'rgba(99, 102, 241, 0.08)';
                }
              }}
              onMouseLeave={(e) => {
                if (!isSelected) {
                  e.currentTarget.style.backgroundColor = 'transparent';
                }
              }}
            >
              {isSelected && (
                <i className="bi bi-check-circle-fill" style={{ color: '#6366f1', fontSize: '0.875rem' }}></i>
              )}
              <span style={{ fontSize: '1rem' }}>{getFileIcon(node.name)}</span>
              <span style={{
                fontSize: '0.875rem',
                color: isSelected ? '#6366f1' : 'inherit',
                fontWeight: isSelected ? '500' : '400'
              }}>
                {node.name}
              </span>
            </div>
          );
        }
      })}
    </>
  );
}

/**
 * Get file icon based on extension
 */
function getFileIcon(filename) {
  const ext = filename.split('.').pop().toLowerCase();
  const iconMap = {
    // Programming languages
    'js': '📜',
    'jsx': '⚛️',
    'ts': '📘',
    'tsx': '⚛️',
    'py': '🐍',
    'go': '🔷',
    'java': '☕',
    'rs': '🦀',
    'c': '⚙️',
    'cpp': '⚙️',
    'cs': '#️⃣',
    'php': '🐘',
    'rb': '💎',
    'swift': '🦅',
    'kt': '🟣',
    
    // Web
    'html': '🌐',
    'css': '🎨',
    'scss': '🎨',
    'vue': '💚',
    
    // Config/Data
    'json': '📋',
    'yaml': '📝',
    'yml': '📝',
    'xml': '📄',
    'toml': '⚙️',
    
    // Documentation
    'md': '📖',
    'txt': '📃',
    
    // Database
    'sql': '🗄️',
    
    // Shell
    'sh': '💻',
    'bash': '💻',
    
    // Other
    'gitignore': '🚫',
    'dockerignore': '🚫',
    'dockerfile': '🐳'
  };

  return iconMap[ext] || '📄';
}
