-- Add realistic test log entries for Health app development and testing
-- This migration creates diverse log entries across all services and levels
-- to properly demonstrate filtering, searching, and tag management features

INSERT INTO logs.entries (service, level, message, tags, metadata, created_at) VALUES
-- Portal service logs (authentication, navigation, user actions)
('portal', 'INFO', 'User logged in successfully', ARRAY['auth', 'login'], '{"user_id": 1, "method": "github_oauth", "ip": "192.168.1.100"}', NOW() - INTERVAL '5 minutes'),
('portal', 'INFO', 'User navigated to review service', ARRAY['navigation', 'review'], '{"user_id": 1, "from": "/portal", "to": "/review"}', NOW() - INTERVAL '4 minutes'),
('portal', 'INFO', 'User navigated to health app', ARRAY['navigation', 'health'], '{"user_id": 1, "from": "/portal", "to": "/logs"}', NOW() - INTERVAL '3 minutes'),
('portal', 'DEBUG', 'Session token refreshed', ARRAY['auth', 'session'], '{"user_id": 1, "session_id": "sess_abc123"}', NOW() - INTERVAL '2 minutes'),
('portal', 'WARNING', 'Slow API response detected', ARRAY['performance', 'api'], '{"endpoint": "/api/portal/dashboard", "duration_ms": 1250, "threshold_ms": 1000}', NOW() - INTERVAL '10 minutes'),
('portal', 'ERROR', 'Failed to load user preferences', ARRAY['database', 'error'], '{"user_id": 1, "error": "connection timeout", "retry_count": 3}', NOW() - INTERVAL '15 minutes'),

-- Review service logs (file operations, analysis, AI interactions)
('review', 'INFO', 'File tree loaded successfully', ARRAY['file-system', 'github'], '{"repo": "devsmith-platform", "file_count": 247, "duration_ms": 342}', NOW() - INTERVAL '6 minutes'),
('review', 'INFO', 'Code analysis started', ARRAY['analysis', 'ai'], '{"file": "HealthPage.jsx", "lines": 831, "mode": "detailed"}', NOW() - INTERVAL '5 minutes'),
('review', 'INFO', 'Code analysis completed', ARRAY['analysis', 'ai', 'success'], '{"file": "HealthPage.jsx", "duration_ms": 2840, "insights": 12}', NOW() - INTERVAL '4 minutes 30 seconds'),
('review', 'DEBUG', 'File content fetched from cache', ARRAY['cache', 'performance'], '{"file": "StatCards.jsx", "cache_hit": true, "size_bytes": 2456}', NOW() - INTERVAL '3 minutes'),
('review', 'WARNING', 'AI rate limit approaching', ARRAY['ai', 'rate-limit'], '{"requests_remaining": 15, "reset_in_seconds": 180}', NOW() - INTERVAL '8 minutes'),
('review', 'ERROR', 'Failed to fetch file from GitHub', ARRAY['github', 'api-error'], '{"repo": "test-repo", "file": "main.go", "status": 404, "message": "Not Found"}', NOW() - INTERVAL '12 minutes'),

-- Logs service logs (ingestion, health checks, database operations)
('logs', 'INFO', 'Log entry ingested', ARRAY['ingestion'], '{"source": "portal", "level": "INFO", "processing_ms": 12}', NOW() - INTERVAL '7 minutes'),
('logs', 'INFO', 'Health check performed', ARRAY['monitoring', 'health'], '{"services_checked": 4, "all_healthy": true, "duration_ms": 234}', NOW() - INTERVAL '5 minutes'),
('logs', 'INFO', 'Database cleanup completed', ARRAY['database', 'maintenance'], '{"deleted_entries": 0, "retention_days": 90, "duration_ms": 156}', NOW() - INTERVAL '20 minutes'),
('logs', 'DEBUG', 'WebSocket connection established', ARRAY['websocket', 'realtime'], '{"client_id": "ws_xyz789", "subscribed_to": ["logs"]}', NOW() - INTERVAL '1 minute'),
('logs', 'WARNING', 'High log ingestion rate detected', ARRAY['performance', 'monitoring'], '{"rate_per_second": 45, "threshold": 50, "source": "automated_tests"}', NOW() - INTERVAL '9 minutes'),
('logs', 'CRITICAL', 'Database connection pool exhausted', ARRAY['database', 'critical'], '{"active_connections": 100, "max_connections": 100, "waiting_queries": 23}', NOW() - INTERVAL '25 minutes'),

-- Analytics service logs (aggregations, reports, queries)
('analytics', 'INFO', 'Daily aggregation started', ARRAY['aggregation', 'scheduled'], '{"date": "2025-11-10", "log_count": 1234, "services": 4}', NOW() - INTERVAL '30 minutes'),
('analytics', 'INFO', 'Trend analysis completed', ARRAY['analysis', 'trends'], '{"metric": "error_rate", "period": "7d", "direction": "decreasing", "change_percent": -15.3}', NOW() - INTERVAL '18 minutes'),
('analytics', 'DEBUG', 'Query cache hit', ARRAY['cache', 'performance'], '{"query": "top_errors_last_hour", "cached_result": true, "age_seconds": 45}', NOW() - INTERVAL '2 minutes 30 seconds'),
('analytics', 'WARNING', 'Anomaly detected in log patterns', ARRAY['monitoring', 'anomaly'], '{"pattern": "unusual_spike", "service": "review", "increase_percent": 234}', NOW() - INTERVAL '11 minutes'),
('analytics', 'ERROR', 'Failed to generate report', ARRAY['report', 'error'], '{"report_type": "weekly_summary", "error": "insufficient_data", "min_required": 100}', NOW() - INTERVAL '14 minutes'),

-- Frontend logs (client-side events via backend proxy)
('frontend', 'INFO', 'Page loaded successfully', ARRAY['performance', 'ux'], '{"page": "/logs", "load_time_ms": 342, "user_agent": "Chrome/120.0"}', NOW() - INTERVAL '3 minutes 15 seconds'),
('frontend', 'INFO', 'Tag filter applied', ARRAY['ui', 'filter'], '{"tag": "database", "matching_logs": 5}', NOW() - INTERVAL '2 minutes 45 seconds'),
('frontend', 'DEBUG', 'Component re-rendered', ARRAY['react', 'performance'], '{"component": "HealthPage", "render_time_ms": 18, "cause": "state_update"}', NOW() - INTERVAL '1 minute 30 seconds'),
('frontend', 'WARNING', 'Slow API response', ARRAY['api', 'performance'], '{"endpoint": "/api/logs", "duration_ms": 2150, "expected_max_ms": 1000}', NOW() - INTERVAL '6 minutes'),
('frontend', 'ERROR', 'Failed to add tag', ARRAY['ui', 'error'], '{"log_id": 999, "tag": "test", "status": 404, "message": "Log not found"}', NOW() - INTERVAL '16 minutes'),

-- Additional mixed service logs for variety
('portal', 'INFO', 'Dashboard widgets loaded', ARRAY['ui', 'performance'], '{"widget_count": 6, "load_time_ms": 245}', NOW() - INTERVAL '22 minutes'),
('review', 'INFO', 'GitHub repository cloned', ARRAY['github', 'clone'], '{"repo": "devsmith-platform", "branch": "main", "size_mb": 12.4, "duration_ms": 3421}', NOW() - INTERVAL '19 minutes'),
('logs', 'INFO', 'Tag added to log entry', ARRAY['tags', 'manual'], '{"log_id": 2, "tag": "investigated", "added_by": "user_1"}', NOW() - INTERVAL '1 minute'),
('analytics', 'INFO', 'Export completed', ARRAY['export', 'csv'], '{"rows": 500, "format": "csv", "file_size_kb": 234, "duration_ms": 456}', NOW() - INTERVAL '13 minutes'),
('frontend', 'INFO', 'Dark mode toggled', ARRAY['ui', 'theme'], '{"theme": "dark", "previous": "light", "user_id": 1}', NOW() - INTERVAL '4 minutes'),

-- Critical/urgent logs for attention
('portal', 'CRITICAL', 'Authentication service unavailable', ARRAY['auth', 'critical', 'outage'], '{"service": "github_oauth", "status": "down", "last_success": "2 minutes ago"}', NOW() - INTERVAL '1 minute 45 seconds'),
('review', 'CRITICAL', 'AI service quota exceeded', ARRAY['ai', 'critical', 'quota'], '{"service": "claude_api", "quota_used": 100, "quota_limit": 100, "reset_time": "2025-11-11T00:00:00Z"}', NOW() - INTERVAL '35 minutes'),
('logs', 'CRITICAL', 'Disk space critically low', ARRAY['storage', 'critical'], '{"mount": "/var/lib/postgresql", "free_gb": 0.8, "total_gb": 100, "threshold_gb": 5}', NOW() - INTERVAL '40 minutes');

-- Update the existing 2 error logs with more realistic data
UPDATE logs.entries 
SET 
  message = 'Failed to connect to PostgreSQL database: connection timeout after 5s',
  tags = ARRAY['ai', 'database', 'error', 'performance', 'portal'],
  metadata = '{"host": "postgres", "port": 5432, "database": "devsmith", "retry_count": 3, "last_error": "connection timeout"}'::jsonb,
  created_at = NOW() - INTERVAL '17 minutes'
WHERE id = 2;

UPDATE logs.entries
SET
  message = 'Failed to connect to AI service: connection timeout after 30s',
  tags = ARRAY['ai', 'review', 'error', 'timeout'],
  metadata = '{"service": "ai-factory", "host": "ai-factory:8083", "timeout_ms": 30000, "error": "dial tcp: lookup ai-factory: no such host"}'::jsonb,
  created_at = NOW() - INTERVAL '8 minutes'
WHERE id = 3;
