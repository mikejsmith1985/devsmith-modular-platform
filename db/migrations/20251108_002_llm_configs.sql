-- Migration: 20251108_002_llm_configs
-- Description: Create tables for multi-LLM configuration and usage tracking
-- Author: DevSmith Platform
-- Date: 2025-11-08

-- LLM Configurations Table
-- Stores user's AI model configurations with encrypted API keys
CREATE TABLE portal.llm_configs (
    id VARCHAR(64) PRIMARY KEY,
    user_id INT NOT NULL REFERENCES portal.users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('openai', 'anthropic', 'ollama', 'deepseek', 'mistral', 'google')),
    model_name VARCHAR(100) NOT NULL,
    api_key_encrypted TEXT,
    api_endpoint VARCHAR(255),
    is_default BOOLEAN DEFAULT false,
    max_tokens INT DEFAULT 4096 CHECK (max_tokens > 0),
    temperature DECIMAL(3,2) DEFAULT 0.7 CHECK (temperature >= 0.0 AND temperature <= 2.0),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Ensure one config per user/provider/model combination
    UNIQUE(user_id, provider, model_name)
);

-- Indexes for llm_configs
CREATE INDEX idx_llm_configs_user ON portal.llm_configs(user_id);
CREATE INDEX idx_llm_configs_provider ON portal.llm_configs(provider);
CREATE INDEX idx_llm_configs_default ON portal.llm_configs(user_id, is_default) WHERE is_default = true;

-- App LLM Preferences Table
-- Maps each app to a specific LLM config for a user
CREATE TABLE portal.app_llm_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES portal.users(id) ON DELETE CASCADE,
    app_name VARCHAR(50) NOT NULL CHECK (app_name IN ('review', 'logs', 'analytics', 'build')),
    llm_config_id VARCHAR(64) REFERENCES portal.llm_configs(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- One preference per user/app
    UNIQUE(user_id, app_name)
);

-- Indexes for app_llm_preferences
CREATE INDEX idx_app_llm_prefs_user ON portal.app_llm_preferences(user_id, app_name);
CREATE INDEX idx_app_llm_prefs_config ON portal.app_llm_preferences(llm_config_id);

-- LLM Usage Logs Table
-- Tracks token usage, latency, and costs for billing and analytics
CREATE TABLE portal.llm_usage_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    app_name VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    tokens_used INT NOT NULL DEFAULT 0,
    latency_ms INT NOT NULL DEFAULT 0,
    cost_usd DECIMAL(10,6) DEFAULT 0.000000 CHECK (cost_usd >= 0),
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for llm_usage_logs (optimized for analytics queries)
CREATE INDEX idx_llm_usage_user_date ON portal.llm_usage_logs(user_id, created_at DESC);
CREATE INDEX idx_llm_usage_app ON portal.llm_usage_logs(app_name, created_at DESC);
CREATE INDEX idx_llm_usage_provider ON portal.llm_usage_logs(provider, created_at DESC);
CREATE INDEX idx_llm_usage_cost ON portal.llm_usage_logs(cost_usd DESC, created_at DESC);

-- Trigger to update updated_at timestamp for llm_configs
CREATE OR REPLACE FUNCTION portal.update_llm_config_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_llm_config_timestamp
    BEFORE UPDATE ON portal.llm_configs
    FOR EACH ROW
    EXECUTE FUNCTION portal.update_llm_config_timestamp();

-- Trigger to update updated_at timestamp for app_llm_preferences
CREATE TRIGGER trigger_update_app_llm_pref_timestamp
    BEFORE UPDATE ON portal.app_llm_preferences
    FOR EACH ROW
    EXECUTE FUNCTION portal.update_llm_config_timestamp();

-- Trigger to ensure only one default config per user
CREATE OR REPLACE FUNCTION portal.ensure_single_default_llm()
RETURNS TRIGGER AS $$
BEGIN
    -- If setting this config as default, unset all other defaults for this user
    IF NEW.is_default = true THEN
        UPDATE portal.llm_configs
        SET is_default = false
        WHERE user_id = NEW.user_id
          AND id != NEW.id
          AND is_default = true;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_single_default_llm
    BEFORE INSERT OR UPDATE ON portal.llm_configs
    FOR EACH ROW
    WHEN (NEW.is_default = true)
    EXECUTE FUNCTION portal.ensure_single_default_llm();

-- Comments for documentation
COMMENT ON TABLE portal.llm_configs IS 'User AI model configurations with encrypted API keys';
COMMENT ON COLUMN portal.llm_configs.api_key_encrypted IS 'AES-256-GCM encrypted API key using user-specific derived key';
COMMENT ON COLUMN portal.llm_configs.is_default IS 'Default LLM config used when no app-specific preference set';
COMMENT ON COLUMN portal.llm_configs.temperature IS 'Sampling temperature (0.0-2.0). Lower = more deterministic, higher = more creative';

COMMENT ON TABLE portal.app_llm_preferences IS 'Maps each app to specific LLM config per user';
COMMENT ON COLUMN portal.app_llm_preferences.llm_config_id IS 'NULL means use user default config';

COMMENT ON TABLE portal.llm_usage_logs IS 'Audit log of LLM API calls for billing and analytics';
COMMENT ON COLUMN portal.llm_usage_logs.cost_usd IS 'Calculated cost based on provider pricing (input + output tokens)';
COMMENT ON COLUMN portal.llm_usage_logs.success IS 'False if API call failed';
