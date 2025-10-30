-- Phase 3: Health Intelligence Tables
-- Stores health check history, auto-repair actions, security scans, and policies

-- Health check results (stores full reports)
CREATE TABLE IF NOT EXISTS logs.health_checks (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    overall_status VARCHAR(20) NOT NULL, -- pass/warn/fail
    duration_ms INTEGER NOT NULL,
    check_count INTEGER NOT NULL,
    passed_count INTEGER NOT NULL,
    warned_count INTEGER NOT NULL,
    failed_count INTEGER NOT NULL,
    report_json JSONB NOT NULL, -- Full HealthReport
    triggered_by VARCHAR(50) NOT NULL, -- 'manual', 'scheduled', 'api'
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_checks_timestamp ON logs.health_checks(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_health_checks_status ON logs.health_checks(overall_status);
CREATE INDEX IF NOT EXISTS idx_health_checks_triggered ON logs.health_checks(triggered_by);

-- Individual check details (for detailed analysis)
CREATE TABLE IF NOT EXISTS logs.health_check_details (
    id SERIAL PRIMARY KEY,
    health_check_id INTEGER NOT NULL REFERENCES logs.health_checks(id) ON DELETE CASCADE,
    check_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL, -- pass/warn/fail
    message TEXT,
    error TEXT,
    duration_ms INTEGER NOT NULL,
    details_json JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_check_details_check_id ON logs.health_check_details(health_check_id);
CREATE INDEX IF NOT EXISTS idx_health_check_details_name ON logs.health_check_details(check_name);
CREATE INDEX IF NOT EXISTS idx_health_check_details_status ON logs.health_check_details(status);

-- Security scan results (from Trivy)
CREATE TABLE IF NOT EXISTS logs.security_scans (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    scan_type VARCHAR(50) NOT NULL, -- 'image', 'config', 'filesystem'
    target VARCHAR(255) NOT NULL, -- image name or path
    critical_count INTEGER DEFAULT 0,
    high_count INTEGER DEFAULT 0,
    medium_count INTEGER DEFAULT 0,
    low_count INTEGER DEFAULT 0,
    scan_json JSONB NOT NULL, -- Full Trivy output
    health_check_id INTEGER REFERENCES logs.health_checks(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_scans_timestamp ON logs.security_scans(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_security_scans_critical ON logs.security_scans(critical_count DESC);
CREATE INDEX IF NOT EXISTS idx_security_scans_target ON logs.security_scans(target);
CREATE INDEX IF NOT EXISTS idx_security_scans_type ON logs.security_scans(scan_type);

-- Auto-repair actions and history
CREATE TABLE IF NOT EXISTS logs.auto_repairs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    health_check_id INTEGER REFERENCES logs.health_checks(id) ON DELETE SET NULL,
    service_name VARCHAR(100) NOT NULL,
    issue_type VARCHAR(100) NOT NULL, -- 'timeout', 'crash', 'dependency', 'security'
    repair_action VARCHAR(50) NOT NULL, -- 'restart', 'rebuild', 'rollback'
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'success', 'failed'
    error TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_auto_repairs_timestamp ON logs.auto_repairs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_auto_repairs_service ON logs.auto_repairs(service_name);
CREATE INDEX IF NOT EXISTS idx_auto_repairs_status ON logs.auto_repairs(status);
CREATE INDEX IF NOT EXISTS idx_auto_repairs_health_check ON logs.auto_repairs(health_check_id);

-- Custom health policies per service
CREATE TABLE IF NOT EXISTS logs.health_policies (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL UNIQUE,
    max_response_time_ms INTEGER DEFAULT 1000,
    auto_repair_enabled BOOLEAN DEFAULT true,
    repair_strategy VARCHAR(50) DEFAULT 'restart', -- 'restart', 'rebuild', 'none'
    alert_on_warn BOOLEAN DEFAULT false,
    alert_on_fail BOOLEAN DEFAULT true,
    policy_json JSONB, -- Additional custom rules
    updated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_policies_service ON logs.health_policies(service_name);

-- Add data retention policy comment
COMMENT ON TABLE logs.health_checks IS 'Stores health check results with 30-day retention policy';
COMMENT ON TABLE logs.security_scans IS 'Stores Trivy security scan results for trend analysis';
COMMENT ON TABLE logs.auto_repairs IS 'Tracks auto-repair actions and outcomes';
COMMENT ON TABLE logs.health_policies IS 'Per-service health check policies and repair strategies';
