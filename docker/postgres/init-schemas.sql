-- Ensure schemas are created before any table definitions
CREATE SCHEMA IF NOT EXISTS portal;
CREATE SCHEMA IF NOT EXISTS reviews;
CREATE SCHEMA IF NOT EXISTS logs;
CREATE SCHEMA IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS builds;

-- Create roles for authentication (check if they exist first)
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'devsmith') THEN
    CREATE ROLE devsmith WITH LOGIN PASSWORD 'test_password';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'root') THEN
    CREATE ROLE root WITH LOGIN PASSWORD 'test_password';
  END IF;
END
$$;

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

-- Create logs.entries table
CREATE TABLE IF NOT EXISTS logs.entries (
    id BIGSERIAL PRIMARY KEY,
    user_id INT,
    service VARCHAR(50),      -- 'portal', 'review', 'logging', etc.
    level VARCHAR(20),        -- 'debug', 'info', 'warn', 'error'
    message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_service_level ON logs.entries(service, level, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_user ON logs.entries(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_logs_created ON logs.entries(created_at DESC);

-- Grant CREATEDB privilege to devsmith user
GRANT CREATEDB TO devsmith;
