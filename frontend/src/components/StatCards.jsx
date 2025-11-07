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

function StatCard({ level, count }) {
  const config = LEVEL_CONFIG[level.toLowerCase()] || LEVEL_CONFIG.info;

  return (
    <div className="col-md-6 col-lg mb-3">
      <div className="card h-100" style={{ backgroundColor: config.bgColor }}>
        <div className="card-body text-center">
          <i className={`bi ${config.icon} text-${config.color}`} style={{ fontSize: '2.5rem' }}></i>
          <h3 className="mt-3 mb-0">{count.toLocaleString()}</h3>
          <p className={`text-${config.color} fw-bold mb-0`}>{config.label}</p>
        </div>
      </div>
    </div>
  );
}

export default function StatCards({ stats }) {
  const levels = ['debug', 'info', 'warning', 'error', 'critical'];

  return (
    <div className="row">
      {levels.map(level => (
        <StatCard
          key={level}
          level={level}
          count={stats[level] || 0}
        />
      ))}
    </div>
  );
}
