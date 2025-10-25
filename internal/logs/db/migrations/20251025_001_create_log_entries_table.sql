CREATE SCHEMA IF NOT EXISTS logs;

CREATE TABLE IF NOT EXISTS logs.log_entries (
    id BIGSERIAL PRIMARY KEY,
    user_id INT,
    service VARCHAR(50),
    level VARCHAR(20),
    message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_service_level ON logs.log_entries(service, level, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_user ON logs.log_entries(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_created ON logs.log_entries(created_at DESC);

COMMENT ON TABLE logs.log_entries IS 'Log entries from all services for audit and monitoring';
COMMENT ON COLUMN logs.log_entries.user_id IS 'User who triggered the log (optional, may be null for system logs)';
COMMENT ON COLUMN logs.log_entries.service IS 'Service name: portal, review, logging, analytics, etc.';
COMMENT ON COLUMN logs.log_entries.level IS 'Log level: debug, info, warn, error';
