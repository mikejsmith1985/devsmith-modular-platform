package logscontext

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// SessionTokenKey is the context key for storing the session token
// This allows the DynamicAIClient to authenticate with Portal API
const SessionTokenKey contextKey = "session_token"

// ModelContextKey is the context key for storing the model override
// This allows users to specify a different model than the default
const ModelContextKey contextKey = "model"
