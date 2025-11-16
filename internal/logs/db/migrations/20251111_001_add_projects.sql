-- Migration: Add projects table for cross-repo logging
-- Date: 2025-11-11
-- Purpose: Enable DevSmith to monitor external applications/repositories

-- Create projects table
CREATE TABLE IF NOT EXISTS logs.projects (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,  -- Owner of this project (references portal.users)
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,  -- URL-safe identifier (e.g., 'my-ecommerce-app')
    description TEXT,
    repository_url VARCHAR(500),  -- Optional GitHub/GitLab URL
    api_key_hash VARCHAR(255) NOT NULL,  -- Bcrypt-hashed API key
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    
    -- Constraints
    CONSTRAINT unique_user_slug UNIQUE(user_id, slug),
    CONSTRAINT slug_format CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
);

-- Index for fast API key lookups (authentication)
CREATE INDEX idx_projects_api_key ON logs.projects(api_key_hash);

-- Index for user's projects list
CREATE INDEX idx_projects_user ON logs.projects(user_id, created_at DESC);

-- Index for active projects
CREATE INDEX idx_projects_active ON logs.projects(is_active, created_at DESC);

-- Add project_id to log_entries (nullable for backward compatibility)
ALTER TABLE logs.entries ADD COLUMN IF NOT EXISTS project_id INT;

-- Add service_name to log_entries (for microservices identification)
ALTER TABLE logs.entries ADD COLUMN IF NOT EXISTS service_name VARCHAR(100);

-- Foreign key (optional - allows logs to exist without project)
-- ON DELETE SET NULL means logs remain if project is deleted
ALTER TABLE logs.entries 
    ADD CONSTRAINT fk_project 
    FOREIGN KEY (project_id) 
    REFERENCES logs.projects(id) 
    ON DELETE SET NULL;

-- Index for filtering by project
CREATE INDEX idx_entries_project ON logs.entries(project_id, created_at DESC);

-- Index for filtering by project + service
CREATE INDEX idx_entries_project_service ON logs.entries(project_id, service_name, created_at DESC);

-- Index for filtering by project + level
CREATE INDEX idx_entries_project_level ON logs.entries(project_id, level, created_at DESC);

-- Create view for project statistics
CREATE OR REPLACE VIEW logs.project_stats AS
SELECT 
    p.id,
    p.name,
    p.slug,
    COUNT(e.id) AS total_logs,
    COUNT(CASE WHEN e.level = 'ERROR' THEN 1 END) AS error_count,
    COUNT(CASE WHEN e.level = 'WARN' THEN 1 END) AS warn_count,
    COUNT(CASE WHEN e.level = 'INFO' THEN 1 END) AS info_count,
    COUNT(CASE WHEN e.level = 'DEBUG' THEN 1 END) AS debug_count,
    MAX(e.created_at) AS last_log_at,
    COUNT(DISTINCT e.service_name) AS service_count
FROM logs.projects p
LEFT JOIN logs.entries e ON p.id = e.project_id
WHERE p.is_active = true
GROUP BY p.id, p.name, p.slug;

-- Add comment for documentation
COMMENT ON TABLE logs.projects IS 'External projects/applications that send logs to DevSmith';
COMMENT ON COLUMN logs.projects.api_key_hash IS 'Bcrypt-hashed API key for authentication (never store plain keys)';
COMMENT ON COLUMN logs.projects.slug IS 'URL-safe identifier, used in API requests and URLs';
COMMENT ON COLUMN logs.entries.project_id IS 'Reference to external project (NULL for DevSmith internal logs)';
COMMENT ON COLUMN logs.entries.service_name IS 'Microservice name within project (e.g., api-gateway, user-service)';

-- Create internal "devsmith-platform" project for existing logs
-- This ensures backward compatibility - existing logs get assigned to DevSmith project
INSERT INTO logs.projects (user_id, name, slug, description, api_key_hash, is_active)
VALUES (
    1,  -- Admin user (adjust if needed)
    'DevSmith Platform',
    'devsmith-platform',
    'Internal DevSmith platform logs',
    '$2a$10$dummyhashfordevsmithplatform',  -- Dummy hash, not used for auth
    true
)
ON CONFLICT (user_id, slug) DO NOTHING;

-- Update existing logs to belong to DevSmith project
UPDATE logs.entries 
SET project_id = (SELECT id FROM logs.projects WHERE slug = 'devsmith-platform')
WHERE project_id IS NULL;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION logs.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_projects_updated_at
    BEFORE UPDATE ON logs.projects
    FOR EACH ROW
    EXECUTE FUNCTION logs.update_updated_at_column();
