-- Migration: Make logs service truly standalone
-- Date: 2025-11-12
-- Purpose: Remove Portal dependency, enable universal cross-repo monitoring

-- Step 1: Drop foreign key constraint if it exists
-- This allows projects to exist without Portal users
DO $$
BEGIN
    ALTER TABLE logs.projects DROP CONSTRAINT IF EXISTS fk_projects_user;
    ALTER TABLE logs.projects DROP CONSTRAINT IF EXISTS projects_user_id_fkey;
EXCEPTION
    WHEN undefined_object THEN NULL;
END $$;

-- Step 2: Make user_id nullable (projects can be unclaimed)
-- This allows external projects to auto-create on first batch without user association
ALTER TABLE logs.projects ALTER COLUMN user_id DROP NOT NULL;

-- Step 3: Update constraint to allow NULL user_id
-- Projects can have same slug if one is unclaimed (user_id = NULL)
ALTER TABLE logs.projects DROP CONSTRAINT IF EXISTS unique_user_slug;

-- Create new constraint that allows NULL user_id
CREATE UNIQUE INDEX unique_user_slug ON logs.projects(user_id, slug) 
WHERE user_id IS NOT NULL;

-- Also ensure unclaimed projects have unique slugs
CREATE UNIQUE INDEX unique_unclaimed_slug ON logs.projects(slug) 
WHERE user_id IS NULL;

-- Step 4: Add claimed_at timestamp for tracking when users claim projects
ALTER TABLE logs.projects ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMP;

-- Step 5: Update indexes for better performance with nullable user_id
DROP INDEX IF EXISTS idx_projects_user;
CREATE INDEX idx_projects_user ON logs.projects(user_id, created_at DESC) 
WHERE user_id IS NOT NULL;

-- Index for unclaimed projects
CREATE INDEX idx_projects_unclaimed ON logs.projects(created_at DESC) 
WHERE user_id IS NULL;

-- Step 6: Update project_stats view to handle nullable user_id
CREATE OR REPLACE VIEW logs.project_stats AS
SELECT 
    p.id,
    p.user_id,
    p.name,
    p.slug,
    p.claimed_at,
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
GROUP BY p.id, p.user_id, p.name, p.slug, p.claimed_at;

-- Step 7: Update comments
COMMENT ON COLUMN logs.projects.user_id IS 'Optional Portal user who claimed this project (NULL for unclaimed projects from external codebases)';
COMMENT ON COLUMN logs.projects.claimed_at IS 'Timestamp when user claimed this project via Portal (NULL if unclaimed)';

-- Step 8: Allow DevSmith platform project to be unclaimed initially
UPDATE logs.projects 
SET user_id = NULL, claimed_at = NULL
WHERE slug = 'devsmith-platform' AND user_id = 1;
