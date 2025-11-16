import React from 'react';

/**
 * FileTabs Component
 * 
 * VS Code-style tabs for managing multiple open files
 * Features:
 * - Click to switch between files
 * - Close button with confirmation for unsaved changes
 * - Visual indicator for active tab
 * - Unsaved changes indicator (dot)
 * - Add new file button
 * - Drag and drop reordering (future enhancement)
 */
export default function FileTabs({ 
  files = [],
  activeFileId,
  onFileSelect,
  onFileClose,
  onFileAdd,
  onFileRename
}) {
  const handleCloseTab = (e, fileId) => {
    e.stopPropagation(); // Prevent tab selection when clicking close button
    
    const file = files.find(f => f.id === fileId);
    
    // Show confirmation if file has unsaved changes
    if (file?.hasUnsavedChanges) {
      const confirmed = window.confirm(
        `"${file.name}" has unsaved changes. Do you want to close it anyway?`
      );
      if (!confirmed) return;
    }
    
    if (onFileClose) {
      onFileClose(fileId);
    }
  };

  const handleTabClick = (fileId) => {
    if (onFileSelect) {
      onFileSelect(fileId);
    }
  };

  const handleTabDoubleClick = (fileId) => {
    if (onFileRename) {
      const file = files.find(f => f.id === fileId);
      const newName = prompt('Enter new file name:', file.name);
      if (newName && newName.trim() && newName !== file.name) {
        onFileRename(fileId, newName.trim());
      }
    }
  };

  return (
    <div className="file-tabs-container" style={{
      display: 'flex',
      alignItems: 'center',
      backgroundColor: isDarkMode => isDarkMode ? 'rgba(30, 33, 48, 0.95)' : 'rgba(250, 250, 255, 0.95)',
      borderBottom: '2px solid rgba(99, 102, 241, 0.2)',
      padding: '0.25rem 0.5rem',
      gap: '0.25rem',
      overflowX: 'auto',
      overflowY: 'hidden',
      whiteSpace: 'nowrap',
      maxWidth: '100%'
    }}>
      {/* File Tabs */}
      {files.map(file => (
        <div
          key={file.id}
          onClick={() => handleTabClick(file.id)}
          onDoubleClick={() => handleTabDoubleClick(file.id)}
          className={`file-tab ${file.id === activeFileId ? 'active' : ''}`}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.5rem',
            padding: '0.5rem 0.75rem',
            borderRadius: '8px 8px 0 0',
            cursor: 'pointer',
            transition: 'all 0.2s ease',
            position: 'relative',
            backgroundColor: file.id === activeFileId 
              ? 'rgba(99, 102, 241, 0.15)' 
              : 'transparent',
            border: file.id === activeFileId
              ? '2px solid rgba(99, 102, 241, 0.4)'
              : '2px solid transparent',
            borderBottom: 'none',
            minWidth: 'fit-content'
          }}
          title={file.path || file.name}
        >
          {/* Language/File Type Icon */}
          <span style={{
            fontSize: '1rem',
            opacity: 0.8
          }}>
            {getFileIcon(file.language)}
          </span>

          {/* File Name */}
          <span style={{
            fontSize: '0.875rem',
            fontWeight: file.id === activeFileId ? '600' : '400',
            color: file.id === activeFileId ? '#6366f1' : 'inherit',
            maxWidth: '150px',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap'
          }}>
            {file.name}
          </span>

          {/* Unsaved Changes Indicator */}
          {file.hasUnsavedChanges && (
            <span style={{
              width: '6px',
              height: '6px',
              borderRadius: '50%',
              backgroundColor: '#ec4899',
              flexShrink: 0
            }} title="Unsaved changes" />
          )}

          {/* Close Button */}
          <button
            onClick={(e) => handleCloseTab(e, file.id)}
            className="btn-close-tab"
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '20px',
              height: '20px',
              border: 'none',
              background: 'transparent',
              cursor: 'pointer',
              borderRadius: '4px',
              padding: 0,
              opacity: 0.6,
              transition: 'all 0.2s ease',
              fontSize: '1rem',
              color: 'inherit'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.opacity = '1';
              e.currentTarget.style.backgroundColor = 'rgba(239, 68, 68, 0.2)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.opacity = '0.6';
              e.currentTarget.style.backgroundColor = 'transparent';
            }}
            title="Close (Ctrl+W)"
          >
            ×
          </button>
        </div>
      ))}

      {/* Add New File Button */}
      <button
        onClick={onFileAdd}
        className="btn-add-file"
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '0.5rem',
          border: 'none',
          background: 'transparent',
          cursor: 'pointer',
          borderRadius: '4px',
          transition: 'all 0.2s ease',
          fontSize: '1.25rem',
          color: '#6366f1',
          opacity: 0.7,
          minWidth: 'fit-content'
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.opacity = '1';
          e.currentTarget.style.backgroundColor = 'rgba(99, 102, 241, 0.1)';
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.opacity = '0.7';
          e.currentTarget.style.backgroundColor = 'transparent';
        }}
        title="Add new file (Ctrl+N)"
      >
        <i className="bi bi-plus-circle"></i>
      </button>

      {/* Tab Count Indicator */}
      {files.length > 1 && (
        <span style={{
          fontSize: '0.75rem',
          color: 'rgba(99, 102, 241, 0.6)',
          marginLeft: 'auto',
          padding: '0.25rem 0.5rem',
          borderRadius: '12px',
          backgroundColor: 'rgba(99, 102, 241, 0.1)',
          fontWeight: '500',
          minWidth: 'fit-content'
        }}>
          {files.length} {files.length === 1 ? 'file' : 'files'}
        </span>
      )}
    </div>
  );
}

/**
 * Get appropriate icon for file type/language
 */
function getFileIcon(language) {
  const iconMap = {
    'javascript': '📜',
    'typescript': '📘',
    'python': '🐍',
    'go': '🔷',
    'java': '☕',
    'rust': '🦀',
    'c': '⚙️',
    'cpp': '⚙️',
    'csharp': '#️⃣',
    'sql': '🗄️',
    'html': '🌐',
    'css': '🎨',
    'json': '📋',
    'yaml': '📝',
    'markdown': '📖',
    'shell': '💻',
    'bash': '💻',
    'php': '🐘',
    'ruby': '💎',
    'swift': '🦅',
    'kotlin': '🟣'
  };

  return iconMap[language?.toLowerCase()] || '📄';
}
