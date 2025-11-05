-- Migration: GitHub Session Management for Phase 2
-- Date: 2025-11-04
-- Purpose: Add GitHub repository session tracking with file tree and multi-tab support

-- Create github_sessions table for GitHub repository metadata
CREATE TABLE IF NOT EXISTS reviews.github_sessions (
    id SERIAL PRIMARY KEY,
    session_id INT NOT NULL REFERENCES reviews.sessions(id) ON DELETE CASCADE,
    github_url VARCHAR(500) NOT NULL,
    owner VARCHAR(255) NOT NULL,
    repo VARCHAR(255) NOT NULL,
    branch VARCHAR(100) DEFAULT 'main',
    commit_sha VARCHAR(40),
    file_tree JSONB,  -- Cached repository tree structure
    total_files INT DEFAULT 0,
    total_directories INT DEFAULT 0,
    tree_last_synced TIMESTAMP,
    is_private BOOLEAN DEFAULT FALSE,
    stars_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(session_id)  -- One GitHub session per review session
);

-- Create open_files table for multi-tab system
CREATE TABLE IF NOT EXISTS reviews.open_files (
    id SERIAL PRIMARY KEY,
    github_session_id INT NOT NULL REFERENCES reviews.github_sessions(id) ON DELETE CASCADE,
    tab_id UUID NOT NULL,  -- Client-generated UUID for tab tracking
    file_path VARCHAR(500) NOT NULL,
    file_sha VARCHAR(40),
    file_content TEXT,
    file_size BIGINT,
    language VARCHAR(50),
    is_active BOOLEAN DEFAULT FALSE,
    tab_order INT DEFAULT 0,
    opened_at TIMESTAMP DEFAULT NOW(),
    last_accessed TIMESTAMP DEFAULT NOW(),
    analysis_count INT DEFAULT 0  -- How many times analyzed in this tab
);

-- Create multi_file_analysis table for tracking cross-file analysis
CREATE TABLE IF NOT EXISTS reviews.multi_file_analysis (
    id SERIAL PRIMARY KEY,
    github_session_id INT NOT NULL REFERENCES reviews.github_sessions(id) ON DELETE CASCADE,
    file_paths TEXT[],  -- Array of file paths analyzed together
    reading_mode VARCHAR(20) CHECK (reading_mode IN ('preview', 'skim', 'scan', 'detailed', 'critical')),
    combined_content TEXT,  -- Concatenated file contents
    ai_response JSONB,  -- Analysis result from AI
    cross_file_dependencies JSONB,  -- Detected dependencies between files
    shared_abstractions JSONB,  -- Shared patterns/abstractions found
    architecture_patterns JSONB,  -- Architecture patterns detected
    analysis_duration_ms BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indices for performance
CREATE INDEX IF NOT EXISTS idx_github_sessions_session_id ON reviews.github_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_github_sessions_owner_repo ON reviews.github_sessions(owner, repo);
CREATE INDEX IF NOT EXISTS idx_github_sessions_updated_at ON reviews.github_sessions(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_open_files_github_session_id ON reviews.open_files(github_session_id);
CREATE INDEX IF NOT EXISTS idx_open_files_tab_id ON reviews.open_files(tab_id);
CREATE INDEX IF NOT EXISTS idx_open_files_is_active ON reviews.open_files(github_session_id, is_active);
CREATE INDEX IF NOT EXISTS idx_open_files_tab_order ON reviews.open_files(github_session_id, tab_order);
CREATE INDEX IF NOT EXISTS idx_multi_file_analysis_github_session_id ON reviews.multi_file_analysis(github_session_id);
CREATE INDEX IF NOT EXISTS idx_multi_file_analysis_created_at ON reviews.multi_file_analysis(created_at DESC);

-- Create GIN index for JSONB columns for efficient querying
CREATE INDEX IF NOT EXISTS idx_github_sessions_file_tree ON reviews.github_sessions USING GIN (file_tree);
CREATE INDEX IF NOT EXISTS idx_multi_file_analysis_dependencies ON reviews.multi_file_analysis USING GIN (cross_file_dependencies);
CREATE INDEX IF NOT EXISTS idx_multi_file_analysis_abstractions ON reviews.multi_file_analysis USING GIN (shared_abstractions);

-- Add comments for documentation
COMMENT ON TABLE reviews.github_sessions IS 'GitHub repository session metadata with cached file tree';
COMMENT ON TABLE reviews.open_files IS 'Tracks files opened in multi-tab UI for each GitHub session';
COMMENT ON TABLE reviews.multi_file_analysis IS 'Cross-file analysis results for multiple files analyzed together';

COMMENT ON COLUMN reviews.github_sessions.file_tree IS 'JSON structure of repository tree: {rootNodes: [{path, type, sha, size, children}]}';
COMMENT ON COLUMN reviews.open_files.tab_id IS 'Client-side generated UUID for tab identification and ordering';
COMMENT ON COLUMN reviews.open_files.tab_order IS 'Display order in tab bar (0-indexed, left to right)';
COMMENT ON COLUMN reviews.multi_file_analysis.file_paths IS 'Array of file paths that were analyzed together';
COMMENT ON COLUMN reviews.multi_file_analysis.cross_file_dependencies IS 'Detected import relationships and dependencies between files';
COMMENT ON COLUMN reviews.multi_file_analysis.shared_abstractions IS 'Common interfaces, patterns, and abstractions shared across files';

-- Add trigger to update github_sessions.updated_at on changes
CREATE OR REPLACE FUNCTION update_github_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_github_sessions_updated_at
    BEFORE UPDATE ON reviews.github_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_github_sessions_updated_at();

-- Add trigger to update open_files.last_accessed on update
CREATE OR REPLACE FUNCTION update_open_files_last_accessed()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_active = TRUE AND OLD.is_active = FALSE THEN
        NEW.last_accessed = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_open_files_last_accessed
    BEFORE UPDATE ON reviews.open_files
    FOR EACH ROW
    EXECUTE FUNCTION update_open_files_last_accessed();

-- Add helper function to count files in tree (recursive)
CREATE OR REPLACE FUNCTION count_files_in_tree(tree_json JSONB)
RETURNS TABLE(file_count INT, dir_count INT) AS $$
DECLARE
    file_cnt INT := 0;
    dir_cnt INT := 0;
    node JSONB;
BEGIN
    -- Iterate through root nodes
    FOR node IN SELECT jsonb_array_elements(tree_json->'rootNodes')
    LOOP
        IF node->>'type' = 'file' THEN
            file_cnt := file_cnt + 1;
        ELSIF node->>'type' = 'dir' THEN
            dir_cnt := dir_cnt + 1;
            -- Recursively count children
            IF node->'children' IS NOT NULL THEN
                SELECT file_cnt + cf.file_count, dir_cnt + cf.dir_count
                INTO file_cnt, dir_cnt
                FROM count_files_in_tree(jsonb_build_object('rootNodes', node->'children')) cf;
            END IF;
        END IF;
    END LOOP;
    
    RETURN QUERY SELECT file_cnt, dir_cnt;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION count_files_in_tree IS 'Recursively counts files and directories in a tree JSON structure';
