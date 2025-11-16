-- Migration: 20251108_001_prompt_templates
-- Description: Create tables for AI prompt templates and execution tracking
-- Author: DevSmith Platform
-- Date: 2025-11-08

-- Create review schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS review;

-- Prompt Templates Table
-- Stores both system default prompts (user_id = NULL) and user customizations
CREATE TABLE review.prompt_templates (
    id VARCHAR(64) PRIMARY KEY,
    user_id INT REFERENCES portal.users(id) ON DELETE CASCADE,
    mode VARCHAR(20) NOT NULL CHECK (mode IN ('preview', 'skim', 'scan', 'detailed', 'critical')),
    user_level VARCHAR(20) NOT NULL CHECK (user_level IN ('beginner', 'intermediate', 'expert')),
    output_mode VARCHAR(20) NOT NULL CHECK (output_mode IN ('quick', 'detailed', 'comprehensive')),
    prompt_text TEXT NOT NULL,
    variables JSONB DEFAULT '[]'::jsonb,
    is_default BOOLEAN DEFAULT false,
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Ensure one custom prompt per user/mode/level/output combination
    UNIQUE(user_id, mode, user_level, output_mode)
);

-- Indexes for prompt templates
CREATE INDEX idx_prompt_templates_user ON review.prompt_templates(user_id);
CREATE INDEX idx_prompt_templates_mode ON review.prompt_templates(mode, user_level, output_mode);
CREATE INDEX idx_prompt_templates_default ON review.prompt_templates(is_default) WHERE is_default = true;

-- Prompt Executions Table
-- Logs every time a prompt is used, for analytics and feedback
CREATE TABLE review.prompt_executions (
    id SERIAL PRIMARY KEY,
    template_id VARCHAR(64) REFERENCES review.prompt_templates(id) ON DELETE SET NULL,
    user_id INT NOT NULL,
    rendered_prompt TEXT NOT NULL,
    response TEXT,
    model_used VARCHAR(100) NOT NULL,
    latency_ms INT,
    tokens_used INT,
    user_rating INT CHECK (user_rating BETWEEN 1 AND 5),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for prompt executions
CREATE INDEX idx_prompt_executions_user ON review.prompt_executions(user_id, created_at DESC);
CREATE INDEX idx_prompt_executions_template ON review.prompt_executions(template_id, created_at DESC);
CREATE INDEX idx_prompt_executions_model ON review.prompt_executions(model_used, created_at DESC);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION review.update_prompt_template_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_prompt_template_timestamp
    BEFORE UPDATE ON review.prompt_templates
    FOR EACH ROW
    EXECUTE FUNCTION review.update_prompt_template_timestamp();

-- Comments for documentation
COMMENT ON TABLE review.prompt_templates IS 'Stores AI prompt templates - system defaults (user_id=NULL) and user customizations';
COMMENT ON COLUMN review.prompt_templates.user_id IS 'NULL for system defaults, user ID for custom prompts';
COMMENT ON COLUMN review.prompt_templates.mode IS 'Review mode: preview, skim, scan, detailed, or critical';
COMMENT ON COLUMN review.prompt_templates.user_level IS 'Target user expertise: beginner, intermediate, or expert';
COMMENT ON COLUMN review.prompt_templates.output_mode IS 'Output verbosity: quick, detailed, or comprehensive';
COMMENT ON COLUMN review.prompt_templates.variables IS 'JSON array of variable names used in template (e.g., ["{{code}}", "{{query}}"])';
COMMENT ON COLUMN review.prompt_templates.is_default IS 'True for system defaults that cannot be deleted';

COMMENT ON TABLE review.prompt_executions IS 'Audit log of prompt usage for analytics and user feedback';
COMMENT ON COLUMN review.prompt_executions.rendered_prompt IS 'Final prompt sent to AI with variables substituted';
COMMENT ON COLUMN review.prompt_executions.user_rating IS 'User feedback rating (1-5 stars), NULL if not rated';
