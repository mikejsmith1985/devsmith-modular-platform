-- Migration: Add timestamp column to logs.entries for cross-repo logging
-- Date: 2025-11-11
-- Purpose: Store original timestamp from external applications (separate from created_at)

-- Add timestamp column (nullable for backward compatibility with existing logs)
ALTER TABLE logs.entries 
    ADD COLUMN IF NOT EXISTS timestamp TIMESTAMP;

-- Set default timestamp to created_at for existing rows
UPDATE logs.entries 
SET timestamp = created_at 
WHERE timestamp IS NULL;

-- Index for timestamp-based queries (cross-repo time-series analysis)
CREATE INDEX IF NOT EXISTS idx_entries_timestamp ON logs.entries(timestamp DESC);

-- Index for project + timestamp queries
CREATE INDEX IF NOT EXISTS idx_entries_project_timestamp ON logs.entries(project_id, timestamp DESC);

-- Comment explaining the difference
COMMENT ON COLUMN logs.entries.timestamp IS 'Original log timestamp from source application (may differ from created_at due to network latency)';
COMMENT ON COLUMN logs.entries.created_at IS 'Timestamp when log was received by DevSmith platform';
