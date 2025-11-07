import React from 'react';

const analysisModesConfig = {
  preview: {
    name: 'Preview',
    description: 'Quick structural assessment',
    icon: '👁️',
    color: 'primary',
    purpose: 'Get a high-level overview of code structure and organization'
  },
  skim: {
    name: 'Skim',
    description: 'Understand abstractions',
    icon: '📋',
    color: 'info',
    purpose: 'Focus on interfaces, function signatures, and key workflows'
  },
  scan: {
    name: 'Scan',
    description: 'Find specific information',
    icon: '🔍',
    color: 'success',
    purpose: 'Search for specific patterns, functions, or code elements'
  },
  detailed: {
    name: 'Detailed',
    description: 'Deep algorithm understanding',
    icon: '🔬',
    color: 'warning',
    purpose: 'Line-by-line analysis of complex logic and algorithms'
  },
  critical: {
    name: 'Critical',
    description: 'Quality evaluation',
    icon: '⚠️',
    color: 'danger',
    purpose: 'Identify issues, security concerns, and improvement opportunities'
  }
};

export default function AnalysisModeSelector({ selectedMode, onModeSelect, disabled = false }) {
  return (
    <div className="analysis-mode-selector mb-3">
      <h6 className="mb-2">Analysis Mode:</h6>
      <div className="row g-2">
        {Object.entries(analysisModesConfig).map(([mode, config]) => (
          <div key={mode} className="col-md-2 col-sm-4 col-6">
            <div 
              className={`frosted-card h-100 cursor-pointer ${
                selectedMode === mode ? 'border-primary border-3' : ''
              } ${disabled ? 'opacity-50' : ''}`}
              onClick={() => !disabled && onModeSelect(mode)}
              style={{ 
                cursor: disabled ? 'not-allowed' : 'pointer',
                transition: 'all 0.2s',
                padding: '0.75rem'
              }}
            >
              <div className="text-center">
                <div className="fs-4 mb-1">{config.icon}</div>
                <h6 className={`mb-1 ${
                  selectedMode === mode ? 'text-primary' : ''
                }`}>
                  {config.name}
                </h6>
                <small style={{ 
                  color: 'var(--bs-gray-300)',
                  opacity: 0.9
                }}>{config.description}</small>
              </div>
            </div>
          </div>
        ))}
      </div>
      
      {/* Show selected mode purpose */}
      {selectedMode && (
        <div className="mt-2">
          <small style={{ 
            color: 'var(--bs-gray-200)',
            opacity: 0.95
          }}>
            <strong>{analysisModesConfig[selectedMode].name}:</strong>{' '}
            {analysisModesConfig[selectedMode].purpose}
          </small>
        </div>
      )}
    </div>
  );
}

export { analysisModesConfig };