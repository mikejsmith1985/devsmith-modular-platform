-- Ensure schemas are created before any table definitions
CREATE SCHEMA IF NOT EXISTS portal;
CREATE SCHEMA IF NOT EXISTS reviews;
CREATE SCHEMA IF NOT EXISTS logs;
CREATE SCHEMA IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS builds;

-- Create roles for authentication
CREATE ROLE devsmith WITH LOGIN PASSWORD 'test_password';
CREATE ROLE root WITH LOGIN PASSWORD 'test_password';

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

-- Create portal.users table
CREATE TABLE IF NOT EXISTS portal.users (
    id SERIAL PRIMARY KEY,
    github_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    avatar_url TEXT,
    github_access_token TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
