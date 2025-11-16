-- Migration: Enhanced Session Management
-- Date: 2025-01-15
-- Purpose: Add cross-mode state tracking and session history

-- Add new columns to sessions table for enhanced tracking
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS code_content TEXT;
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS github_path VARCHAR(500);
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS language VARCHAR(50);
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active';
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS current_mode VARCHAR(20);
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP;
ALTER TABLE review.sessions ADD COLUMN IF NOT EXISTS session_duration_seconds BIGINT DEFAULT 0;

-- Create mode_states table for tracking per-mode analysis state
CREATE TABLE IF NOT EXISTS review.mode_states (
    id SERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES review.sessions(id) ON DELETE CASCADE,
    mode VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    is_completed BOOLEAN DEFAULT FALSE,
    analysis_started_at TIMESTAMP,
    analysis_completed_at TIMESTAMP,
    analysis_duration_ms BIGINT DEFAULT 0,
    result_id BIGINT,
    user_notes TEXT,
    issues_found INT DEFAULT 0,
    quality_score INT DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(session_id, mode)
);

-- Create session_history table for audit trail
CREATE TABLE IF NOT EXISTS review.session_history (
    id SERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES review.sessions(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    mode VARCHAR(20),
    old_value TEXT,
    new_value TEXT,
    changes JSONB,
    acted_by BIGINT,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_session_history_session_id (session_id),
    INDEX idx_session_history_created_at (created_at)
);

-- Create indices for performance
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON review.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON review.sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON review.sessions(created_at);
CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON review.sessions(updated_at);
CREATE INDEX IF NOT EXISTS idx_mode_states_session_id ON review.mode_states(session_id);
CREATE INDEX IF NOT EXISTS idx_mode_states_mode ON review.mode_states(mode);

-- Create view for session summary
CREATE OR REPLACE VIEW review.session_summaries AS
SELECT
    s.id,
    s.title,
    s.code_source,
    s.language,
    s.status,
    s.current_mode,
    ROUND(
        100.0 * (SELECT COUNT(*) FROM review.mode_states WHERE session_id = s.id AND is_completed = TRUE) /
        NULLIF((SELECT COUNT(*) FROM review.mode_states WHERE session_id = s.id), 0),
        0
    )::INT AS mode_progress,
    s.created_at,
    s.last_accessed,
    EXTRACT(EPOCH FROM (COALESCE(s.completed_at, NOW()) - s.created_at))::BIGINT AS duration_seconds
FROM review.sessions s;

-- Add update trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION review.update_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF NOT EXISTS sessions_updated_at_trigger ON review.sessions;
CREATE TRIGGER sessions_updated_at_trigger
BEFORE UPDATE ON review.sessions
FOR EACH ROW
EXECUTE FUNCTION review.update_sessions_updated_at();

-- Add update trigger to auto-update mode_states.updated_at
CREATE OR REPLACE FUNCTION review.update_mode_states_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF NOT EXISTS mode_states_updated_at_trigger ON review.mode_states;
CREATE TRIGGER mode_states_updated_at_trigger
BEFORE UPDATE ON review.mode_states
FOR EACH ROW
EXECUTE FUNCTION review.update_mode_states_updated_at();
