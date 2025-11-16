package reviewcontext

// Context keys for review service
type contextKey string

// ModelContextKey is used to pass the selected LLM model through the request context
const ModelContextKey contextKey = "model"

// SessionTokenKey is used to pass the user's session token through the request context
// This is set by RedisSessionAuthMiddleware and used to query Portal's AI Factory
const SessionTokenKey contextKey = "session_token"
