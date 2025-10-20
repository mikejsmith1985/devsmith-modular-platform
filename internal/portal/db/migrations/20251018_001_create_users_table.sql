CREATE TABLE portal.users (
    id SERIAL PRIMARY KEY,
    github_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    avatar_url TEXT,
    github_access_token TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_github_id ON portal.users(github_id);
CREATE INDEX idx_users_username ON portal.users(username);

COMMENT ON TABLE portal.users IS 'Authenticated users from GitHub OAuth';
COMMENT ON COLUMN portal.users.github_id IS 'Unique GitHub user ID (immutable)';
COMMENT ON COLUMN portal.users.github_access_token IS 'Encrypted OAuth token for GitHub API';
