-- Create analytics schema (isolated from logs and other schemas)
CREATE SCHEMA IF NOT EXISTS analytics;

-- Aggregations table: Pre-computed hourly metrics
CREATE TABLE IF NOT EXISTS analytics.aggregations (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    service VARCHAR(50) NOT NULL,
    value NUMERIC NOT NULL,
    time_bucket TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT uq_aggregation UNIQUE (metric_type, service, time_bucket)
);

-- Indexes for common queries
CREATE INDEX idx_aggregations_metric_service ON analytics.aggregations(metric_type, service, time_bucket DESC);
CREATE INDEX idx_aggregations_time_bucket ON analytics.aggregations(time_bucket DESC);
CREATE INDEX idx_aggregations_service ON analytics.aggregations(service, time_bucket DESC);