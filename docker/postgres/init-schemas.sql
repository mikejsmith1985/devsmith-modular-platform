
-- Create reviews.sessions table for Review Service
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
-- Create schemas for modular platform
CREATE SCHEMA IF NOT EXISTS portal;
CREATE SCHEMA IF NOT EXISTS reviews;
CREATE SCHEMA IF NOT EXISTS logs;
CREATE SCHEMA IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS builds;
