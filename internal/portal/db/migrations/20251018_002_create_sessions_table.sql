CREATE TABLE portal.sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES portal.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON portal.sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON portal.sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON portal.sessions(expires_at);

COMMENT ON TABLE portal.sessions IS 'Active user sessions with JWT tokens';
COMMENT ON COLUMN portal.sessions.token_hash IS 'SHA-256 hash of JWT token';
COMMENT ON COLUMN portal.sessions.expires_at IS 'Expiration timestamp (default 24h)';
