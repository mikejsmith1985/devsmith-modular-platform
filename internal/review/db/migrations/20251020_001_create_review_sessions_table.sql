-- Migration: Create reviews.sessions table for Review Service
CREATE SCHEMA IF NOT EXISTS reviews;

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

-- Index for fast lookup
CREATE INDEX IF NOT EXISTS idx_review_sessions_user_id ON reviews.sessions(user_id);
