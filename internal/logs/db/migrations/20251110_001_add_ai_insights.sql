-- AI Insights table for storing AI-generated log analysis
CREATE TABLE IF NOT EXISTS logs.ai_insights (
    id SERIAL PRIMARY KEY,
    log_id BIGINT NOT NULL REFERENCES logs.entries(id) ON DELETE CASCADE,
    analysis TEXT NOT NULL,
    root_cause TEXT,
    suggestions JSONB,
    model_used VARCHAR(255) NOT NULL,
    generated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(log_id) -- One insight per log (regenerate overwrites)
);

CREATE INDEX idx_ai_insights_log_id ON logs.ai_insights(log_id);
CREATE INDEX idx_ai_insights_generated_at ON logs.ai_insights(generated_at DESC);

COMMENT ON TABLE logs.ai_insights IS 'AI-generated insights and analysis for log entries';
COMMENT ON COLUMN logs.ai_insights.log_id IS 'Foreign key to logs.entries';
COMMENT ON COLUMN logs.ai_insights.analysis IS 'General analysis of what the log indicates';
COMMENT ON COLUMN logs.ai_insights.root_cause IS 'Identified root cause (for errors/warnings)';
COMMENT ON COLUMN logs.ai_insights.suggestions IS 'JSON array of actionable suggestions';
COMMENT ON COLUMN logs.ai_insights.model_used IS 'AI model used for generation (e.g., ollama/deepseek-coder-v2:16b)';
