-- Migration: Create analysis_results table for troubleshooting persistence
-- Date: 2025-11-03
-- Purpose: Store minimal analysis artifacts on failures for debugging

-- Create analysis_results table
CREATE TABLE IF NOT EXISTS reviews.analysis_results (
    id SERIAL PRIMARY KEY,
    review_id INTEGER,  -- Optional reference to reviews.sessions if linked
    mode VARCHAR(50) NOT NULL,  -- 'preview', 'skim', 'scan', 'detailed', 'critical'
    prompt TEXT,  -- The prompt sent to LLM
    summary TEXT,  -- Brief summary of what was attempted
    metadata JSONB DEFAULT '{}',  -- Additional context (filename, settings, etc.)
    model_used VARCHAR(100),  -- Which model was used (e.g., 'mistral:7b')
    raw_output TEXT,  -- Raw LLM output that failed to parse (excerpt)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Add indexes for efficient querying
CREATE INDEX idx_analysis_results_created_at ON reviews.analysis_results(created_at);
CREATE INDEX idx_analysis_results_mode ON reviews.analysis_results(mode);
CREATE INDEX idx_analysis_results_review_id_mode ON reviews.analysis_results(review_id, mode);

-- Add comment for documentation
COMMENT ON TABLE reviews.analysis_results IS 'Stores analysis artifacts on failures for troubleshooting. Retention policy: 14 days (configurable via ANALYSIS_RETENTION_DAYS).';
COMMENT ON COLUMN reviews.analysis_results.raw_output IS 'Excerpt of raw LLM output (truncated to prevent storage bloat)';
COMMENT ON COLUMN reviews.analysis_results.created_at IS 'Used by retention job for automatic cleanup';
