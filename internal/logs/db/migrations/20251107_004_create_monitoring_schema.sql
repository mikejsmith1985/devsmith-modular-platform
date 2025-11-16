-- Migration: Create monitoring schema and tables for health metrics
-- Date: 2025-11-07
-- Purpose: Support alert engine and real-time health monitoring dashboard

-- Create monitoring schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS monitoring;

-- Table: API Metrics
-- Purpose: Store all API call metrics for error rate and response time analysis
CREATE TABLE IF NOT EXISTS monitoring.api_metrics (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER NOT NULL,
    payload_size_bytes INTEGER DEFAULT 0,
    user_id INTEGER,
    error_type VARCHAR(100),
    error_message TEXT,
    service_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for time-based queries (used by GetErrorRate, GetResponseTimes)
CREATE INDEX IF NOT EXISTS idx_api_metrics_timestamp ON monitoring.api_metrics(timestamp DESC);

-- Index for error rate queries (status_code >= 400)
CREATE INDEX IF NOT EXISTS idx_api_metrics_errors ON monitoring.api_metrics(timestamp, status_code) WHERE status_code >= 400;

-- Index for service-specific queries
CREATE INDEX IF NOT EXISTS idx_api_metrics_service ON monitoring.api_metrics(service_name, timestamp DESC);

-- Table: Alerts
-- Purpose: Store detected alerts from alert engine
CREATE TABLE IF NOT EXISTS monitoring.alerts (
    id BIGSERIAL PRIMARY KEY,
    alert_type VARCHAR(50) NOT NULL, -- 'error_rate', 'response_time', 'service_health'
    severity VARCHAR(20) NOT NULL, -- 'warning', 'critical'
    message TEXT NOT NULL,
    value FLOAT,
    threshold FLOAT,
    service_name VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- Index for active alerts queries (resolved_at IS NULL)
CREATE INDEX IF NOT EXISTS idx_alerts_active ON monitoring.alerts(created_at DESC) WHERE resolved_at IS NULL;

-- Index for service-specific alert queries
CREATE INDEX IF NOT EXISTS idx_alerts_service ON monitoring.alerts(service_name, created_at DESC);

-- Add service column to health_checks table if it doesn't exist
-- Purpose: Support service health monitoring in alert engine
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_schema = 'logs' 
        AND table_name = 'health_checks' 
        AND column_name = 'service'
    ) THEN
        ALTER TABLE logs.health_checks ADD COLUMN service VARCHAR(50);
    END IF;
END $$;

-- Create index on service column for health check queries
CREATE INDEX IF NOT EXISTS idx_health_checks_service ON logs.health_checks(service, timestamp DESC);

-- Grant permissions to logs user (if it exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'logs') THEN
        GRANT USAGE ON SCHEMA monitoring TO logs;
        GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA monitoring TO logs;
        GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA monitoring TO logs;
    END IF;
END $$;
