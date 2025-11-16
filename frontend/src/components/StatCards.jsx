import React from 'react';

const LEVEL_CONFIG = {
  debug: {
    icon: 'bi-bug-fill',
    color: 'success',
    bgColor: 'rgba(25, 135, 84, 0.1)',
    label: 'DEBUG'
  },
  info: {
    icon: 'bi-info-circle-fill',
    color: 'primary',
    bgColor: 'rgba(13, 110, 253, 0.1)',
    label: 'INFO'
  },
  warning: {
    icon: 'bi-exclamation-triangle-fill',
    color: 'warning',
    bgColor: 'rgba(255, 193, 7, 0.1)',
    label: 'WARNING'
  },
  error: {
    icon: 'bi-x-circle-fill',
    color: 'danger',
    bgColor: 'rgba(220, 53, 69, 0.1)',
    label: 'ERROR'
  },
  critical: {
    icon: 'bi-fire',
    color: 'danger',
    bgColor: 'rgba(220, 53, 69, 0.2)',
    label: 'CRITICAL'
  }
};

function StatCard({ level, count, isActive, onClick }) {
  const config = LEVEL_CONFIG[level.toLowerCase()] || LEVEL_CONFIG.info;

  return (
    <div className="col-md-6 col-lg mb-3">
      <div 
        className={`frosted-card h-100 ${isActive ? 'border border-2' : ''}`}
        style={{ 
          backgroundColor: config.bgColor,
          borderColor: isActive ? `var(--bs-${config.color})` : 'transparent',
          cursor: 'pointer',
          transition: 'all 0.2s ease',
          transform: isActive ? 'scale(1.05)' : 'scale(1)'
        }}
        onClick={onClick}
        role="button"
        tabIndex={0}
        onKeyPress={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            onClick();
          }
        }}
      >
        <div className="text-center p-4">
          <i className={`bi ${config.icon} text-${config.color}`} style={{ fontSize: '2.5rem' }}></i>
          <h3 className="mt-3 mb-0">{count.toLocaleString()}</h3>
          <p className={`text-${config.color} fw-bold mb-0`}>{config.label}</p>
          {isActive && (
            <small className="d-block mt-2 text-muted">
              <i className="bi bi-funnel-fill me-1"></i>
              Filtered
            </small>
          )}
        </div>
      </div>
    </div>
  );
}

export default function StatCards({ stats, selectedLevel, onLevelClick }) {
  const levels = ['debug', 'info', 'warning', 'error', 'critical'];

  return (
    <div className="row">
      {levels.map(level => (
        <StatCard
          key={level}
          level={level}
          count={stats[level] || 0}
          isActive={selectedLevel === level}
          onClick={() => onLevelClick(level)}
        />
      ))}
    </div>
  );
}
