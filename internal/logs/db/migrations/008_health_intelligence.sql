-- Phase 3: Health Intelligence & Automation
-- Creates tables for health check history, security scans, auto-repairs, and policies

-- Store health check results over time
CREATE TABLE IF NOT EXISTS logs.health_checks (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    overall_status VARCHAR(20) NOT NULL,
    duration_ms INTEGER NOT NULL,
    check_count INTEGER NOT NULL,
    passed_count INTEGER NOT NULL,
    warned_count INTEGER NOT NULL,
    failed_count INTEGER NOT NULL,
    report_json JSONB NOT NULL,
    triggered_by VARCHAR(50) DEFAULT 'manual'
);

CREATE INDEX IF NOT EXISTS idx_health_checks_timestamp ON logs.health_checks(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_health_checks_status ON logs.health_checks(overall_status);

-- Store individual check results for detailed analysis
CREATE TABLE IF NOT EXISTS logs.health_check_details (
    id SERIAL PRIMARY KEY,
    health_check_id INTEGER NOT NULL REFERENCES logs.health_checks(id) ON DELETE CASCADE,
    check_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    error TEXT,
    duration_ms INTEGER NOT NULL,
    details_json JSONB
);

CREATE INDEX IF NOT EXISTS idx_health_check_details_check_id ON logs.health_check_details(health_check_id);
CREATE INDEX IF NOT EXISTS idx_health_check_details_name ON logs.health_check_details(check_name);
CREATE INDEX IF NOT EXISTS idx_health_check_details_status ON logs.health_check_details(status);

-- Store Trivy security scan results
CREATE TABLE IF NOT EXISTS logs.security_scans (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    scan_type VARCHAR(50) NOT NULL,
    target VARCHAR(255) NOT NULL,
    critical_count INTEGER DEFAULT 0,
    high_count INTEGER DEFAULT 0,
    medium_count INTEGER DEFAULT 0,
    low_count INTEGER DEFAULT 0,
    scan_json JSONB NOT NULL,
    health_check_id INTEGER REFERENCES logs.health_checks(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_security_scans_timestamp ON logs.security_scans(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_security_scans_critical ON logs.security_scans(critical_count DESC);
CREATE INDEX IF NOT EXISTS idx_security_scans_type ON logs.security_scans(scan_type);

-- Store auto-repair actions
CREATE TABLE IF NOT EXISTS logs.auto_repairs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    health_check_id INTEGER REFERENCES logs.health_checks(id) ON DELETE SET NULL,
    service_name VARCHAR(100) NOT NULL,
    issue_type VARCHAR(100) NOT NULL,
    repair_action VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error TEXT,
    duration_ms INTEGER
);

CREATE INDEX IF NOT EXISTS idx_auto_repairs_service ON logs.auto_repairs(service_name);
CREATE INDEX IF NOT EXISTS idx_auto_repairs_status ON logs.auto_repairs(status);
CREATE INDEX IF NOT EXISTS idx_auto_repairs_timestamp ON logs.auto_repairs(timestamp DESC);

-- Store custom health policies per service
CREATE TABLE IF NOT EXISTS logs.health_policies (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL UNIQUE,
    max_response_time_ms INTEGER DEFAULT 1000,
    auto_repair_enabled BOOLEAN DEFAULT true,
    repair_strategy VARCHAR(50) DEFAULT 'restart',
    alert_on_warn BOOLEAN DEFAULT false,
    alert_on_fail BOOLEAN DEFAULT true,
    policy_json JSONB,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_policies_service ON logs.health_policies(service_name);

-- Add comments for documentation
COMMENT ON TABLE logs.health_checks IS 'Stores health check results over time for trend analysis';
COMMENT ON TABLE logs.health_check_details IS 'Stores individual check results (e.g., HTTP, database, container)';
COMMENT ON TABLE logs.security_scans IS 'Stores Trivy security scan results for vulnerability tracking';
COMMENT ON TABLE logs.auto_repairs IS 'Stores auto-repair action history and outcomes';
COMMENT ON TABLE logs.health_policies IS 'Stores custom health policies for each service';
