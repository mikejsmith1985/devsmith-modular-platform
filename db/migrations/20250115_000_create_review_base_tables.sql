-- Migration: Create Review Base Tables
-- Date: 2025-01-15
-- Purpose: Initialize review schema with foundational tables

-- Create sessions table (primary review sessions)
CREATE TABLE IF NOT EXISTS reviews.sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(255),
    code_source VARCHAR(20) CHECK (code_source IN ('github', 'paste', 'upload')),
    github_repo VARCHAR(255),
    github_branch VARCHAR(100),
    pasted_code TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    last_accessed TIMESTAMP DEFAULT NOW()
);

-- Create reading_sessions table (one per mode analysis)
CREATE TABLE IF NOT EXISTS reviews.reading_sessions (
    id SERIAL PRIMARY KEY,
    session_id INT REFERENCES reviews.sessions(id) ON DELETE CASCADE,
    reading_mode VARCHAR(20) CHECK (reading_mode IN ('preview', 'skim', 'scan', 'detailed', 'critical')),
    target_path VARCHAR(500),
    scan_query TEXT,
    ai_response JSONB,
    user_annotations TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create critical_issues table (issues found in Critical mode)
CREATE TABLE IF NOT EXISTS reviews.critical_issues (
    id SERIAL PRIMARY KEY,
    reading_session_id INT REFERENCES reviews.reading_sessions(id) ON DELETE CASCADE,
    issue_type VARCHAR(50),
    severity VARCHAR(20),
    file_path VARCHAR(500),
    line_number INT,
    description TEXT,
    suggested_fix TEXT,
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indices for performance
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON reviews.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON reviews.sessions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_session_id ON reviews.reading_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_mode ON reviews.reading_sessions(reading_mode);
CREATE INDEX IF NOT EXISTS idx_critical_issues_reading_session_id ON reviews.critical_issues(reading_session_id);
CREATE INDEX IF NOT EXISTS idx_critical_issues_severity ON reviews.critical_issues(severity);

-- Add comment for clarity
COMMENT ON TABLE reviews.sessions IS 'Primary review sessions containing code to analyze';
COMMENT ON TABLE reviews.reading_sessions IS 'Analysis sessions for each of the 5 reading modes';
COMMENT ON TABLE reviews.critical_issues IS 'Issues identified during Critical mode analysis';
