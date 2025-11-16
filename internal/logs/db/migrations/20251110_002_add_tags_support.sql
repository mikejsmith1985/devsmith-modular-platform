-- Phase 3: Smart Tagging System
-- Add tags support to logs.entries

-- Add tags column
ALTER TABLE logs.entries 
ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';

-- Create GIN index for efficient tag queries
CREATE INDEX IF NOT EXISTS idx_logs_entries_tags ON logs.entries USING GIN(tags);

-- Function to auto-generate tags based on content
CREATE OR REPLACE FUNCTION logs.auto_generate_tags()
RETURNS TRIGGER AS $$
BEGIN
    -- Initialize empty tags array if NULL
    IF NEW.tags IS NULL THEN
        NEW.tags := '{}';
    END IF;
    
    -- Service tag (always add)
    NEW.tags := array_append(NEW.tags, NEW.service);
    
    -- Level tag (always add, lowercase)
    NEW.tags := array_append(NEW.tags, lower(NEW.level));
    
    -- Content-based tags (keyword matching)
    -- Network/Routing
    IF NEW.message ~* 'traefik|gateway|routing|proxy|nginx|load.?balancer' THEN
        NEW.tags := array_append(NEW.tags, 'network');
    END IF;
    
    -- Containers
    IF NEW.message ~* 'docker|container|image|build|compose' THEN
        NEW.tags := array_append(NEW.tags, 'docker');
    END IF;
    
    -- Frontend
    IF NEW.message ~* 'react|vite|npm|javascript|jsx|webpack|frontend' THEN
        NEW.tags := array_append(NEW.tags, 'frontend');
    END IF;
    
    -- Backend
    IF NEW.message ~* 'golang|gin|api|handler|endpoint|backend' THEN
        NEW.tags := array_append(NEW.tags, 'backend');
    END IF;
    
    -- Database
    IF NEW.message ~* 'postgres|sql|migration|query|database|connection.?pool' THEN
        NEW.tags := array_append(NEW.tags, 'database');
    END IF;
    
    -- Authentication
    IF NEW.message ~* 'oauth|jwt|token|login|authentication|session|unauthorized' THEN
        NEW.tags := array_append(NEW.tags, 'auth');
    END IF;
    
    -- AI/LLM
    IF NEW.message ~* 'ollama|anthropic|openai|claude|model|ai|llm|inference' THEN
        NEW.tags := array_append(NEW.tags, 'ai');
    END IF;
    
    -- Performance
    IF NEW.message ~* 'timeout|slow|performance|latency|bottleneck|cache' THEN
        NEW.tags := array_append(NEW.tags, 'performance');
    END IF;
    
    -- Security
    IF NEW.message ~* 'security|vulnerability|attack|unauthorized|forbidden|csrf|xss' THEN
        NEW.tags := array_append(NEW.tags, 'security');
    END IF;
    
    -- Deployment
    IF NEW.message ~* 'deploy|release|version|rollback|upgrade' THEN
        NEW.tags := array_append(NEW.tags, 'deployment');
    END IF;
    
    -- Remove duplicates and sort
    NEW.tags := array(SELECT DISTINCT unnest(NEW.tags) ORDER BY 1);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for auto-tagging on INSERT and UPDATE
DROP TRIGGER IF EXISTS trigger_auto_generate_tags ON logs.entries;
CREATE TRIGGER trigger_auto_generate_tags
    BEFORE INSERT OR UPDATE ON logs.entries
    FOR EACH ROW
    EXECUTE FUNCTION logs.auto_generate_tags();

-- Backfill tags for existing logs (if any)
UPDATE logs.entries SET tags = tags WHERE id > 0;

COMMENT ON COLUMN logs.entries.tags IS 'Auto-generated and manual tags for categorization';
COMMENT ON INDEX idx_logs_entries_tags IS 'GIN index for fast tag-based queries';
COMMENT ON FUNCTION logs.auto_generate_tags() IS 'Automatically generates tags based on log content (service, level, keywords)';
