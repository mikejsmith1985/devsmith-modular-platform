import React from 'react';
import PropTypes from 'prop-types';

/**
 * TagFilter Component
 * 
 * Displays available tags as clickable badges for filtering logs.
 * Supports multi-select with visual feedback.
 * 
 * Phase 3: Smart Tagging System
 */
const TagFilter = ({ availableTags, selectedTags, onTagToggle }) => {
  if (!availableTags || availableTags.length === 0) {
    return (
      <div className="alert alert-info">
        <i className="bi bi-info-circle me-2"></i>
        No tags available. Tags are automatically generated from log content.
      </div>
    );
  }

  return (
    <div className="tag-filter">
      <div className="d-flex align-items-center mb-2">
        <i className="bi bi-tags me-2"></i>
        <strong>Filter by Tags:</strong>
        <span className="ms-2 text-muted small">
          ({selectedTags.length} selected)
        </span>
      </div>
      
      <div className="d-flex flex-wrap gap-2">
        {availableTags.map((tag) => {
          const isSelected = selectedTags.includes(tag);
          return (
            <button
              key={tag}
              type="button"
              className={`btn btn-sm ${isSelected ? 'btn-primary' : 'btn-outline-secondary'}`}
              onClick={() => onTagToggle(tag)}
              aria-pressed={isSelected}
            >
              <i className={`bi ${isSelected ? 'bi-check-circle-fill' : 'bi-tag'} me-1`}></i>
              {tag}
            </button>
          );
        })}
      </div>

      {selectedTags.length > 0 && (
        <div className="mt-2">
          <button
            type="button"
            className="btn btn-sm btn-outline-danger"
            onClick={() => selectedTags.forEach(tag => onTagToggle(tag))}
          >
            <i className="bi bi-x-circle me-1"></i>
            Clear all filters
          </button>
        </div>
      )}
    </div>
  );
};

TagFilter.propTypes = {
  availableTags: PropTypes.arrayOf(PropTypes.string).isRequired,
  selectedTags: PropTypes.arrayOf(PropTypes.string).isRequired,
  onTagToggle: PropTypes.func.isRequired,
};

export default TagFilter;
