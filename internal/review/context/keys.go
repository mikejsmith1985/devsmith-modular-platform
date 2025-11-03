package reviewcontext

// Context keys for review service
type contextKey string

// ModelContextKey is used to pass the selected LLM model through the request context
const ModelContextKey contextKey = "model"
