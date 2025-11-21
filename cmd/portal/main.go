// Package main starts the portal service for the DevSmith platform.
// The portal service provides an interface for users to access
// various features of the DevSmith platform, including authentication,
// dashboard access, and more. It serves as the entry point for users
// and handles routing, middleware, and template rendering.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx PostgreSQL driver for DB connection
	handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/config"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/middleware"
	portal_handlers "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/handlers"
	portal_repositories "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/repositories"
	portal_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/session"
)

func main() {
	// Get port from environment or default to 3001
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	// Create Gin router
	router := gin.Default()

	// Initialize instrumentation logger for this service (use validated config)
	logsServiceURL, logsEnabled, err := config.LoadLogsConfigWithFallbackFor("portal")
	if err != nil {
		log.Fatalf("Failed to load logging configuration: %v", err)
	}
	if !logsEnabled {
		log.Printf("Instrumentation/logging disabled: continuing startup without external logs")
		logsServiceURL = "" // instrumentation will treat empty URL as disabled
	}
	instrLogger := instrumentation.NewServiceInstrumentationLogger("portal", logsServiceURL)

	// Middleware for logging requests (skip health checks to reduce noise)
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path != "/health" {
			log.Printf("Incoming request: %s %s", c.Request.Method, c.Request.URL.Path)
			// Log to instrumentation service asynchronously
			//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
			instrLogger.LogEvent(c.Request.Context(), "request_received", map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			})
		}
		c.Next()
	})

	// Health check endpoint - moved to /api/portal/health to avoid conflict with frontend /health route
	router.GET("/api/portal/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "portal",
			"version": "1.0.0",
		})
	})

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	dbConn, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}

	// Configure connection pool to prevent exhaustion
	dbConn.SetMaxOpenConns(10)               // Max 10 connections per service
	dbConn.SetMaxIdleConns(5)                // Keep 5 idle
	dbConn.SetConnMaxLifetime(3600000000000) // 1 hour
	dbConn.SetConnMaxIdleTime(600000000000)  // 10 minutes

	// Ping the database to verify connection
	if err := dbConn.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("Error closing DB connection: %v", closeErr)
		}
		return
	}

	// Run database migrations
	log.Println("Running database migrations...")
	if err := runMigrations(dbConn); err != nil {
		log.Printf("Failed to run migrations: %v", err)
		log.Printf("WARNING: Continuing startup, but some features may not work correctly")
	} else {
		log.Println("Database migrations completed successfully")
	}

	// Initialize Redis session store
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379" // Default for local development
	}
	sessionStore, err := session.NewRedisStore(redisURL, 7*24*time.Hour) // 7 day session TTL
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("Error closing DB connection: %v", closeErr)
		}
		return
	}
	defer func() {
		if closeErr := sessionStore.Close(); closeErr != nil {
			log.Printf("Error closing Redis connection: %v", closeErr)
		}
	}()
	log.Printf("Redis session store initialized at %s", redisURL)

	// Initialize LLM configuration services
	encryptionService := portal_services.NewEncryptionService()
	llmConfigRepo := portal_repositories.NewLLMConfigRepository(dbConn)
	llmConfigService := portal_services.NewLLMConfigService(llmConfigRepo, encryptionService)

	// Register authentication routes (pass session store)
	handlers.RegisterAuthRoutesWithSession(router, dbConn, sessionStore)

	// Register version endpoint (public - no auth required)
	router.GET("/api/portal/version", handlers.HandleVersion)
	router.GET("/version", handlers.HandleVersionShort)

	// Register debug routes (development only)
	debug.RegisterDebugRoutes(router, "portal")

	// Register dashboard routes (use RedisSessionAuth for SSO)
	authenticated := router.Group("/")
	authenticated.Use(middleware.RedisSessionAuthMiddleware(sessionStore))
	authenticated.GET("/api/v1/dashboard/user", handlers.GetUserInfoHandler)

	// Register LLM configuration routes (requires authentication)
	apiAuthenticated := router.Group("/api/portal")
	apiAuthenticated.Use(middleware.RedisSessionAuthMiddleware(sessionStore))
	portal_handlers.RegisterLLMConfigRoutes(apiAuthenticated, llmConfigService)

	// Serve static files (path works in both local dev and Docker)
	staticPath := "apps/portal/static"
	if _, err = os.Stat("./static"); err == nil {
		// Running in Docker container
		staticPath = "./static"
	}
	router.Static("/static", staticPath)
	router.Static("/assets", staticPath+"/assets") // Serve React app assets (JS, CSS, fonts)

	// Serve React SPA
	// Root route - serves index.html (portal dashboard / login page)
	router.GET("/", func(c *gin.Context) {
		indexPath := staticPath + "/index.html"
		c.File(indexPath)
	})

	// SPA fallback - catch all non-API routes for client-side routing
	// This allows React Router to handle routes like /dashboard, /projects, etc.
	router.NoRoute(func(c *gin.Context) {
		// If it's an API call, return 404 JSON
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			log.Printf("404 Not Found (API): %s", c.Request.URL.Path)
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		// For non-API routes, serve index.html (SPA will handle routing)
		log.Printf("SPA fallback route: %s", c.Request.URL.Path)
		indexPath := staticPath + "/index.html"
		c.File(indexPath)
	})

	// Validate required OAuth environment variables
	if err := validateOAuthEnvironment(); err != nil {
		log.Printf("FATAL: %v", err)
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("Error closing DB connection: %v", closeErr)
		}
		os.Exit(1)
	}
	log.Printf("OAuth configured: redirect_uri=%s", os.Getenv("REDIRECT_URI"))

	// Replace fmt.Printf with log.Printf for better logging consistency
	log.Printf("Portal service starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		// Replace os.Exit with proper error handling
		log.Printf("Failed to start server: %v", err)
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("Error closing DB connection: %v", closeErr)
		}
		os.Exit(1) // Ensure the application exits with a non-zero status
	}
}

func validateOAuthEnvironment() error {
	required := []string{"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "REDIRECT_URI"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			return fmt.Errorf("%s environment variable not set", key)
		}
	}
	return nil
}

// runMigrations executes the database migration SQL for Portal service
// This ensures all required tables exist before the service starts
func runMigrations(db *sql.DB) error {
	migrationSQL := `-- Portal Service Migrations: LLM Configs and User Management

-- LLM Configurations Table
-- Stores user's AI model configurations with encrypted API keys
CREATE TABLE IF NOT EXISTS portal.llm_configs (
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
    UNIQUE(user_id, provider, model_name)
);

-- Indexes for llm_configs
CREATE INDEX IF NOT EXISTS idx_llm_configs_user ON portal.llm_configs(user_id);
CREATE INDEX IF NOT EXISTS idx_llm_configs_provider ON portal.llm_configs(provider);
CREATE INDEX IF NOT EXISTS idx_llm_configs_default ON portal.llm_configs(user_id, is_default) WHERE is_default = true;

-- App LLM Preferences Table
-- Maps each app to a specific LLM config for a user
CREATE TABLE IF NOT EXISTS portal.app_llm_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES portal.users(id) ON DELETE CASCADE,
    app_name VARCHAR(50) NOT NULL CHECK (app_name IN ('review', 'logs', 'analytics', 'build')),
    llm_config_id VARCHAR(64) REFERENCES portal.llm_configs(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, app_name)
);

-- Indexes for app_llm_preferences
CREATE INDEX IF NOT EXISTS idx_app_llm_prefs_user ON portal.app_llm_preferences(user_id, app_name);
CREATE INDEX IF NOT EXISTS idx_app_llm_prefs_config ON portal.app_llm_preferences(llm_config_id);

-- LLM Usage Logs Table
-- Tracks token usage, latency, and costs for billing and analytics
CREATE TABLE IF NOT EXISTS portal.llm_usage_logs (
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
CREATE INDEX IF NOT EXISTS idx_llm_usage_user_date ON portal.llm_usage_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_llm_usage_app ON portal.llm_usage_logs(app_name, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_llm_usage_provider ON portal.llm_usage_logs(provider, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_llm_usage_cost ON portal.llm_usage_logs(cost_usd DESC, created_at DESC);

-- Trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION portal.update_llm_config_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to llm_configs
DROP TRIGGER IF EXISTS trigger_update_llm_config_timestamp ON portal.llm_configs;
CREATE TRIGGER trigger_update_llm_config_timestamp
    BEFORE UPDATE ON portal.llm_configs
    FOR EACH ROW
    EXECUTE FUNCTION portal.update_llm_config_timestamp();

-- Apply trigger to app_llm_preferences
DROP TRIGGER IF EXISTS trigger_update_app_llm_pref_timestamp ON portal.app_llm_preferences;
CREATE TRIGGER trigger_update_app_llm_pref_timestamp
    BEFORE UPDATE ON portal.app_llm_preferences
    FOR EACH ROW
    EXECUTE FUNCTION portal.update_llm_config_timestamp();

-- Trigger to ensure only one default config per user
CREATE OR REPLACE FUNCTION portal.ensure_single_default_llm()
RETURNS TRIGGER AS $$
BEGIN
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

DROP TRIGGER IF EXISTS trigger_ensure_single_default_llm ON portal.llm_configs;
CREATE TRIGGER trigger_ensure_single_default_llm
    BEFORE INSERT OR UPDATE ON portal.llm_configs
    FOR EACH ROW
    WHEN (NEW.is_default = true)
    EXECUTE FUNCTION portal.ensure_single_default_llm();
`

	if _, err := db.Exec(migrationSQL); err != nil {
		return fmt.Errorf("migration execution failed: %w", err)
	}

	return nil
}
