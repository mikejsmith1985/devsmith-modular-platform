-- Add AI analysis columns to logs.entries table
-- Migration: 009_add_ai_analysis_columns.sql

ALTER TABLE logs.entries 
ADD COLUMN IF NOT EXISTS issue_type VARCHAR(50),
ADD COLUMN IF NOT EXISTS ai_analysis JSONB,
ADD COLUMN IF NOT EXISTS severity_score INT;

-- Create index for efficient querying by issue type
CREATE INDEX IF NOT EXISTS idx_logs_entries_issue_type 
ON logs.entries(issue_type, created_at DESC);

-- Create index for severity queries
CREATE INDEX IF NOT EXISTS idx_logs_entries_severity 
ON logs.entries(severity_score DESC, created_at DESC);

-- Add comments for documentation
COMMENT ON COLUMN logs.entries.issue_type IS 'Categorized error type: db_connection, auth_failure, null_pointer, rate_limit, network_timeout, unknown';
COMMENT ON COLUMN logs.entries.ai_analysis IS 'Cached AI analysis result with root cause, suggested fix, and fix steps';
COMMENT ON COLUMN logs.entries.severity_score IS 'Severity rating from AI analysis: 1-5 (1=info, 5=critical)';
