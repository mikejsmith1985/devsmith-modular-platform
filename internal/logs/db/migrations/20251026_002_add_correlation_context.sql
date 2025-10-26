-- Add correlation tracking fields to log entries for Feature #38
-- Enables distributed tracing and request correlation across services

-- Add correlation_id column for grouping related logs
ALTER TABLE IF EXISTS logs.entries 
ADD COLUMN IF NOT EXISTS correlation_id TEXT;

-- Add context column for storing request context (JSONB)
ALTER TABLE IF EXISTS logs.entries 
ADD COLUMN IF NOT EXISTS context JSONB;

-- Update timestamp column for log entry timestamp (separate from created_at)
ALTER TABLE IF EXISTS logs.entries 
ADD COLUMN IF NOT EXISTS timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Update user_id if not exists (for tracking which user generated log)
ALTER TABLE IF EXISTS logs.entries 
ADD COLUMN IF NOT EXISTS user_id BIGINT;

-- Add indexes for correlation queries
CREATE INDEX IF NOT EXISTS idx_logs_entries_correlation_id 
  ON logs.entries(correlation_id) 
  WHERE correlation_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_logs_entries_context 
  ON logs.entries USING GIN(context);

-- Add index for context->>'correlation_id' for JSONB extraction
CREATE INDEX IF NOT EXISTS idx_logs_entries_context_correlation_id 
  ON logs.entries((context->>'correlation_id'))
  WHERE context IS NOT NULL;

-- Create view for viewing correlated logs with context
CREATE OR REPLACE VIEW logs.correlated_logs AS
SELECT 
  le.id,
  le.correlation_id,
  le.context,
  le.service,
  le.level,
  le.message,
  le.timestamp,
  le.created_at,
  le.user_id
FROM logs.entries le
WHERE le.correlation_id IS NOT NULL 
  OR le.context IS NOT NULL
ORDER BY le.timestamp DESC, le.id DESC;

-- Add comments for documentation
COMMENT ON COLUMN logs.entries.correlation_id IS 'Unique identifier for tracing related logs across services';
COMMENT ON COLUMN logs.entries.context IS 'JSONB object containing request context (user_id, session_id, trace_id, span_id, hostname, etc.)';
COMMENT ON COLUMN logs.entries.timestamp IS 'Log entry timestamp (may differ from created_at for retroactive logs)';
COMMENT ON COLUMN logs.entries.user_id IS 'User ID that generated the log entry (optional)';
COMMENT ON VIEW logs.correlated_logs IS 'View of log entries with correlation context for distributed tracing';
