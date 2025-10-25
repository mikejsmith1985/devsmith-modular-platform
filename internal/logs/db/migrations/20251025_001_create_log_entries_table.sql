CREATE SCHEMA IF NOT EXISTS logs;

CREATE TABLE IF NOT EXISTS logs.entries (
    id BIGSERIAL PRIMARY KEY,
    service TEXT NOT NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_entries_created_at ON logs.entries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_entries_service ON logs.entries(service);
CREATE INDEX IF NOT EXISTS idx_logs_entries_level ON logs.entries(level);
CREATE INDEX IF NOT EXISTS idx_logs_entries_metadata ON logs.entries USING GIN(metadata);

COMMENT ON TABLE logs.entries IS 'Log entries from all services for audit and monitoring';
COMMENT ON COLUMN logs.entries.service IS 'Service name: portal, review, logging, analytics, etc.';
COMMENT ON COLUMN logs.entries.level IS 'Log level: debug, info, warn, error';
COMMENT ON COLUMN logs.entries.metadata IS 'Flexible JSONB metadata for contextual information';
