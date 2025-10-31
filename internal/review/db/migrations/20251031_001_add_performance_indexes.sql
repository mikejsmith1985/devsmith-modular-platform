-- Migration: Add performance indexes to review service tables
-- Purpose: Optimize query performance for Issue #26
-- Date: 2025-10-31

-- Index on analysis_results for cache lookups (review_id, mode)
CREATE INDEX IF NOT EXISTS idx_analysis_results_review_mode 
ON reviews.analysis_results(review_id, mode);

-- Index on analysis_results for ordering by created_at
CREATE INDEX IF NOT EXISTS idx_analysis_results_created_at 
ON reviews.analysis_results(created_at DESC);

-- Index on sessions for user listing (most common query)
CREATE INDEX IF NOT EXISTS idx_sessions_user_created 
ON reviews.sessions(user_id, created_at DESC);

-- Index on sessions for expiration queries
CREATE INDEX IF NOT EXISTS idx_sessions_created_at 
ON reviews.sessions(created_at DESC);

-- Composite index for session access patterns
CREATE INDEX IF NOT EXISTS idx_sessions_user_accessed 
ON reviews.sessions(user_id, last_accessed DESC);
