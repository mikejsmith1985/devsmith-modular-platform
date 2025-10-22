-- Anomalies table: Detected unusual spikes/dips
CREATE TABLE IF NOT EXISTS analytics.anomalies (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    service VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    detected_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_severity CHECK (severity IN ('low', 'medium', 'high'))
);

CREATE INDEX idx_anomalies_service ON analytics.anomalies(service, detected_at DESC);
CREATE INDEX idx_anomalies_severity ON analytics.anomalies(severity, detected_at DESC);
CREATE INDEX idx_anomalies_detected ON analytics.anomalies(detected_at DESC);