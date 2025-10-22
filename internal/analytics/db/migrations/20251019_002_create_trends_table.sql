-- Trends table: Detected patterns over time
CREATE TABLE IF NOT EXISTS analytics.trends (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    service VARCHAR(50) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    confidence NUMERIC NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_direction CHECK (direction IN ('increasing', 'decreasing', 'stable'))
);

CREATE INDEX idx_trends_service ON analytics.trends(service, created_at DESC);
CREATE INDEX idx_trends_time_range ON analytics.trends(start_time, end_time);