// Package main starts the logs service for the DevSmith platform.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	apphandlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/logs/handlers"
	resthandlers "github.com/mikejsmith1985/devsmith-modular-platform/cmd/logs/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	internal_logs_handlers "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/handlers"
	logs_middleware "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/middleware"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/middleware"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/monitoring"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/session"
	"github.com/sirupsen/logrus"
)

var dbConn *sql.DB

// resolveLLMConfig handles the complex logic of finding an LLM configuration
// Returns the resolved config ID as string
func resolveLLMConfig(lookupErr error, configID string, db *sql.DB) string {
	switch {
	case errors.Is(lookupErr, sql.ErrNoRows):
		return handleMissingAppConfig(db)
	case lookupErr != nil:
		// Handle case where app_llm_preferences table doesn't exist yet (e.g., migrations not run)
		// This prevents service crash in environments where Phase 1 migrations haven't been applied
		if strings.Contains(lookupErr.Error(), "relation \"portal.app_llm_preferences\" does not exist") {
			log.Println("WARN: app_llm_preferences table not found (Phase 1 migrations may not be applied)")
			log.Println("Falling back to default LLM configuration")
			return handleMissingAppConfig(db)
		}
		log.Fatalf("FATAL: Failed to query LLM preference from database: %v", lookupErr)
	default:
		log.Println("Using logs app-specific LLM configuration")
	}
	return configID
}

// handleMissingAppConfig attempts to find default LLM configuration
func handleMissingAppConfig(db *sql.DB) string {
	var defaultConfigID string
	err := db.QueryRow(`
		SELECT id 
		FROM portal.llm_configs 
		WHERE is_default = true
		LIMIT 1
	`).Scan(&defaultConfigID)

	if errors.Is(err, sql.ErrNoRows) {
		log.Println("WARN: No default LLM configured in AI Factory")
		log.Println("INFO: Logs service will continue without AI analysis features")
		log.Println("INFO: Configure an LLM in AI Factory (http://localhost:3000/ai-factory) for AI-powered diagnostics")
		return "" // Return empty string to indicate no LLM available
	}
	if err != nil {
		// Check if table doesn't exist (Portal migrations not yet run)
		if strings.Contains(err.Error(), "relation \"portal.llm_configs\" does not exist") {
			log.Println("WARN: portal.llm_configs table not found (Portal migrations may not be applied)")
			log.Println("INFO: Logs service will continue without AI analysis features")
			return "" // Return empty string to indicate no LLM available
		}
		log.Fatalf("FATAL: Failed to query default LLM config: %v", err)
	}
	log.Println("Using default LLM configuration")
	return defaultConfigID
}

//nolint:gocognit,gocyclo // main() initialization is necessarily complex with multiple service setups
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Initialize instrumentation logger for this service
	// Note: Logs service has circular dependency prevention built in
	logsServiceURL := os.Getenv("LOGS_SERVICE_URL")
	if logsServiceURL == "" {
		logsServiceURL = "http://localhost:8082" // Default for local development
	}
	instrLogger := instrumentation.NewServiceInstrumentationLogger("logs", logsServiceURL)

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Initialize database
	dbURL := os.Getenv("DATABASE_URL")
	var err error
	dbConn, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool to prevent exhaustion
	dbConn.SetMaxOpenConns(10)               // Max 10 connections per service
	dbConn.SetMaxIdleConns(5)                // Keep 5 idle
	dbConn.SetConnMaxLifetime(3600000000000) // 1 hour
	dbConn.SetConnMaxIdleTime(600000000000)  // 10 minutes

	// Verify connection
	if pingErr := dbConn.Ping(); pingErr != nil {
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("[ERROR] Failed to close database: %v", closeErr)
		}
		log.Fatal("Failed to ping database:", pingErr)
	}

	// --- Redis session store initialization ---
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default to local Redis
	}
	sessionStore, err := session.NewRedisStore(redisAddr, 7*24*time.Hour) // 7 day session TTL
	if err != nil {
		log.Fatalf("Failed to initialize Redis session store: %v", err)
	}
	defer func() {
		if closeErr := sessionStore.Close(); closeErr != nil {
			log.Printf("Error closing Redis: %v", closeErr)
		}
	}()
	log.Printf("Redis session store initialized: addr=%s, ttl=7 days", redisAddr)

	// Run database migrations
	if err = runMigrations(dbConn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// OAuth2 configuration (for GitHub)
	required := []string{"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "REDIRECT_URI"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			log.Printf("FATAL: %s environment variable not set", key)
			return
		}
	}
	log.Printf("OAuth configured: redirect_uri=%s", os.Getenv("REDIRECT_URI"))

	// Initialize Gin router
	router := gin.Default()

	// Middleware for logging requests (skip health checks in event log, but still track them)
	router.Use(func(c *gin.Context) {
		// Log all requests asynchronously (health checks too, for observability)
		if logErr := instrLogger.LogEvent(c.Request.Context(), "request_received", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}); logErr != nil {
			// Log error but don't fail the request
			log.Printf("Failed to log request event: %v", logErr)
		}
		c.Next()
	})

	// Middleware to inject DATABASE_URL into context for health checks
	router.Use(func(c *gin.Context) {
		c.Set("DATABASE_URL", dbURL)
		c.Next()
	})

	// Serve static files for logs dashboard
	router.Static("/static", "./apps/logs/static")

	// Register debug routes (development only)
	debug.RegisterDebugRoutes(router, "logs")

	// Initialize database repositories for REST API
	logRepo := logs_db.NewLogRepository(dbConn)
	restSvc := logs_services.NewRestLogService(logRepo, logger)

	// Issue #023: Production Enhancements - Initialize alert and aggregation services
	alertConfigRepo := logs_db.NewAlertConfigRepository(dbConn)
	alertViolationRepo := logs_db.NewAlertViolationRepository(dbConn)

	// Create alert service for threshold management (implements AlertThresholdService interface)
	alertSvc := logs_services.NewAlertService(alertViolationRepo, alertConfigRepo, logRepo, logger)

	// Create validation aggregation service for analytics
	validationAgg := logs_services.NewValidationAggregation(logRepo, logger)

	// Phase 1: AI-Driven Diagnostics - Initialize AI analysis services
	// AI Configuration - Query AI Factory database directly (no HTTP/auth overhead)
	// Uses same priority logic as Portal's GetEffectiveConfig:
	// 1. App-specific preference for 'logs' app
	// 2. User's default configuration (is_default = true)
	// 3. FATAL error if nothing configured
	log.Println("Fetching AI configuration from AI Factory database...")

	// Priority 1: Check for logs app-specific preference
	var logsConfigID string
	err = dbConn.QueryRow(`
		SELECT llm_config_id 
		FROM portal.app_llm_preferences 
		WHERE app_name = 'logs'
		LIMIT 1
	`).Scan(&logsConfigID)

	// Handle LLM configuration lookup results
	logsConfigID = resolveLLMConfig(err, logsConfigID, dbConn)

	// Initialize AI provider (if LLM configured)
	var rawAIClient ai.Provider
	var adaptedAIClient logs_services.AIProvider

	if logsConfigID == "" {
		// No LLM configured - continue without AI features
		log.Println("INFO: Logs service starting without AI analysis (no LLM configured)")
		rawAIClient = nil
		adaptedAIClient = nil
	} else {
		// Query for the specific LLM configuration
		var (
			configName  string
			provider    string
			modelName   string
			apiEndpoint sql.NullString
			apiKeyEnc   sql.NullString
		)
		err = dbConn.QueryRow(`
			SELECT id, provider, model_name, api_endpoint, api_key_encrypted
			FROM portal.llm_configs
			WHERE id = $1
			LIMIT 1
		`, logsConfigID).Scan(&configName, &provider, &modelName, &apiEndpoint, &apiKeyEnc)

		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("WARN: LLM config %s not found in AI Factory database - continuing without AI", logsConfigID)
			rawAIClient = nil
			adaptedAIClient = nil
		} else if err != nil {
			log.Printf("WARN: Failed to query LLM config from database: %v - continuing without AI", err)
			rawAIClient = nil
			adaptedAIClient = nil
		} else {
			log.Printf("AI Factory configuration loaded: %s (%s - %s)\n", configName, provider, modelName)

			// Decrypt API key (AI Factory stores encrypted keys)
			var apiKey string
			if apiKeyEnc.Valid && apiKeyEnc.String != "" {
				// TODO: Implement decryption using Portal's encryption service
				// For now, assume keys are stored as base64-encoded
				apiKey = apiKeyEnc.String
			}

			endpoint := apiEndpoint.String
			if endpoint == "" && provider == "ollama" {
				endpoint = "http://host.docker.internal:11434" // Default Ollama endpoint
			}

			// Initialize AI provider based on configuration
			switch provider {
			case "anthropic":
				if apiKey == "" {
					log.Printf("WARN: Anthropic API key not set in AI Factory for config '%s' - continuing without AI", configName)
					rawAIClient = nil
					adaptedAIClient = nil
				} else {
					anthropicClient := providers.NewAnthropicClient(apiKey, modelName)
					rawAIClient = anthropicClient
					adaptedAIClient = logs_services.NewAnthropicAdapter(anthropicClient)
					log.Printf("✓ AI provider ready: Anthropic (%s)\n", modelName)
				}

			case "ollama":
				if endpoint == "" {
					log.Printf("WARN: Ollama endpoint not configured in AI Factory - continuing without AI")
					rawAIClient = nil
					adaptedAIClient = nil
				} else {
					ollamaClient := providers.NewOllamaClient(endpoint, modelName)
					rawAIClient = ollamaClient
					adaptedAIClient = logs_services.NewOllamaAdapter(ollamaClient)
					log.Printf("✓ AI provider ready: Ollama (%s - %s)\n", endpoint, modelName)
				}

			case "openai":
				if apiKey == "" {
					log.Printf("WARN: OpenAI API key not set in AI Factory for config '%s' - continuing without AI", configName)
					rawAIClient = nil
					adaptedAIClient = nil
				} else {
					// OpenAI client implementation would go here
					log.Printf("WARN: OpenAI provider not yet implemented - continuing without AI")
					rawAIClient = nil
					adaptedAIClient = nil
				}

			default:
				log.Printf("WARN: Unsupported AI provider '%s' in AI Factory - continuing without AI", provider)
				rawAIClient = nil
				adaptedAIClient = nil
			}
		}
	}

	// Initialize AI analysis services (if AI available)
	var analysisHandler *internal_logs_handlers.AnalysisHandler
	if rawAIClient != nil {
		aiAnalyzer := logs_services.NewAIAnalyzer(rawAIClient)
		patternMatcher := logs_services.NewPatternMatcher()
		analysisService := logs_services.NewAnalysisService(aiAnalyzer, patternMatcher)
		analysisHandler = internal_logs_handlers.NewAnalysisHandler(analysisService, logger)
		log.Printf("✓ AI analysis services ready\n")
	} else {
		log.Printf("INFO: AI analysis services disabled (no LLM configured)\n")
	}

	// Week 1: Cross-Repository Logging - Initialize batch ingestion services
	projectRepo := logs_db.NewProjectRepository(dbConn)
	projectService := logs_services.NewProjectService(projectRepo)
	logEntryRepo := logs_db.NewLogEntryRepository(dbConn)
	batchHandler := internal_logs_handlers.NewBatchHandler(logEntryRepo, projectRepo, projectService)
	projectHandler := internal_logs_handlers.NewProjectHandler(projectService)

	log.Println("Batch ingestion service initialized for cross-repository logging")

	// Register REST API routes
	router.POST("/api/logs", func(c *gin.Context) {
		resthandlers.PostLogs(restSvc)(c)
	})

	// Week 1: Cross-Repository Logging - Batch ingestion endpoint
	// This endpoint allows external applications to send logs in batches (100x performance improvement)
	// Authentication: Simple API token validation (fast O(1) lookup)
	// Rate limit: 100 requests/minute per API key (TODO: implement rate limiting middleware)
	//
	// Standalone: Works for ANY external codebase (Node.js, Go, Java, Python, etc.)
	// No dependency on Portal service - projects can be unclaimed (user_id=NULL)
	router.POST("/api/logs/batch", logs_middleware.SimpleAPITokenAuth(projectRepo), batchHandler.IngestBatch)

	// Week 1: Cross-Repository Logging - Project management endpoints
	// Authentication: Redis session middleware (requires GitHub OAuth login)
	// These endpoints allow authenticated users to create projects and manage API keys
	projectRoutes := router.Group("/api/logs/projects")
	projectRoutes.Use(middleware.RedisSessionAuthMiddleware(sessionStore))
	projectRoutes.POST("", projectHandler.CreateProject)
	projectRoutes.GET("", projectHandler.ListProjects)
	projectRoutes.GET("/:id", projectHandler.GetProject)
	projectRoutes.POST("/:id/regenerate-key", projectHandler.RegenerateAPIKey)
	projectRoutes.DELETE("/:id", projectHandler.DeleteProject)

	router.GET("/api/logs", func(c *gin.Context) {
		resthandlers.GetLogs(restSvc)(c)
	})
	router.GET("/api/logs/:id", func(c *gin.Context) {
		resthandlers.GetLogByID(restSvc)(c)
	})
	router.GET("/api/logs/stats", func(c *gin.Context) {
		resthandlers.GetStats(restSvc)(c)
	})
	router.DELETE("/api/logs", func(c *gin.Context) {
		resthandlers.DeleteLogs(restSvc)(c)
	})

	// TODO: Add protected routes group when authentication is required
	// Example:
	// protected := router.Group("/")
	// protected.Use(middleware.RedisSessionAuthMiddleware(sessionStore))
	// {
	//     protected.POST("/api/logs/sensitive", ...) // Protected endpoint
	// }

	// Also register /api/v1/logs routes (for consistency and direct access)
	router.POST("/api/v1/logs", func(c *gin.Context) {
		resthandlers.PostLogs(restSvc)(c)
	})
	router.GET("/api/v1/logs", func(c *gin.Context) {
		resthandlers.GetLogs(restSvc)(c)
	})
	router.GET("/api/v1/logs/:id", func(c *gin.Context) {
		resthandlers.GetLogByID(restSvc)(c)
	})
	router.GET("/api/v1/logs/stats", func(c *gin.Context) {
		resthandlers.GetStats(restSvc)(c)
	})
	router.DELETE("/api/v1/logs", func(c *gin.Context) {
		resthandlers.DeleteLogs(restSvc)(c)
	})

	// Issue #023: Production Enhancements - Dashboard & Alert Endpoints
	// Dashboard statistics endpoint
	router.GET("/api/logs/dashboard/stats", func(c *gin.Context) {
		resthandlers.GetDashboardStats(validationAgg)(c)
	})

	// React Frontend Stats API - Log counts by level for StatCards
	router.GET("/api/logs/v1/stats", func(c *gin.Context) {
		// Add timeout to prevent hanging on DB deadlock/self-logging loop
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		stats, err := logRepo.GetLogStatsByLevel(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch log stats")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
			return
		}
		c.JSON(http.StatusOK, stats)
	})

	// Validation analytics endpoints
	router.GET("/api/logs/validations/top-errors", func(c *gin.Context) {
		resthandlers.GetTopErrors(validationAgg)(c)
	})
	router.GET("/api/logs/validations/trends", func(c *gin.Context) {
		resthandlers.GetErrorTrends(validationAgg)(c)
	})

	// Alert configuration management endpoints
	router.POST("/api/logs/alert-config", func(c *gin.Context) {
		resthandlers.CreateAlertConfig(alertSvc)(c)
	})
	router.GET("/api/logs/alert-config/:service", func(c *gin.Context) {
		resthandlers.GetAlertConfig(alertSvc)(c)
	})
	router.PUT("/api/logs/alert-config/:service", func(c *gin.Context) {
		resthandlers.UpdateAlertConfig(alertSvc)(c)
	})

	// Phase 1: AI-Driven Diagnostics - Analysis endpoints (if AI available)
	if analysisHandler != nil {
		router.POST("/api/logs/analyze", analysisHandler.AnalyzeLog)
		router.POST("/api/logs/classify", analysisHandler.ClassifyLog)
	} else {
		router.POST("/api/logs/analyze", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "AI analysis not available - no LLM configured"})
		})
		router.POST("/api/logs/classify", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "AI classification not available - no LLM configured"})
		})
	}

	// Phase 2: AI Insights - Initialize AI insights services (if AI available)
	var aiInsightsHandler *internal_logs_handlers.AIInsightsHandler
	if adaptedAIClient != nil {
		aiInsightsRepo := logs_db.NewAIInsightsRepository(dbConn)
		logRepoAdapter := logs_services.NewLogRepositoryAdapter(logRepo)
		aiInsightsService := logs_services.NewAIInsightsService(adaptedAIClient, logRepoAdapter, aiInsightsRepo)
		aiInsightsHandler = internal_logs_handlers.NewAIInsightsHandler(aiInsightsService, logger, logEntryRepo)
		log.Println("AI insights service initialized - ready for log analysis")
	}

	// AI insights endpoints (if AI available)
	if aiInsightsHandler != nil {
		router.POST("/api/logs/:id/insights", aiInsightsHandler.GenerateInsights)
		router.GET("/api/logs/:id/insights", aiInsightsHandler.GetInsights)
	} else {
		router.POST("/api/logs/:id/insights", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "AI insights not available - no LLM configured"})
		})
		router.GET("/api/logs/:id/insights", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "AI insights not available - no LLM configured"})
		})
	}

	// Phase 3: Smart Tagging System - Initialize tag management
	tagsHandler := internal_logs_handlers.NewTagsHandler(logRepo)

	// Tag management endpoints
	router.GET("/api/logs/tags", tagsHandler.GetAvailableTags)             // Get all unique tags with counts
	router.POST("/api/logs/:id/tags", tagsHandler.AddTagToLog)             // Add manual tag to log entry
	router.DELETE("/api/logs/:id/tags/:tag", tagsHandler.RemoveTagFromLog) // Remove tag from log entry

	log.Println("Tag management service initialized - 3 endpoints registered (auto-tagging + manual)")

	// Health Monitoring Dashboard - Real-time metrics and alerts
	metricsCollector := monitoring.NewSQLMetricsCollector(dbConn)
	monitoringHandler := internal_logs_handlers.NewMonitoringHandler(metricsCollector)

	router.GET("/api/logs/monitoring/metrics", monitoringHandler.GetMetrics)
	router.GET("/api/logs/monitoring/alerts", monitoringHandler.GetAlerts)
	router.GET("/api/logs/monitoring/stats", monitoringHandler.GetStats)

	// Start Alert Engine - Background monitoring and alerting
	alertThresholds := monitoring.DefaultAlertThresholds()
	alertEngine := monitoring.NewAlertEngine(dbConn, alertThresholds, 1*time.Minute, log.Default())
	alertEngine.Start()
	defer alertEngine.Stop()

	// Phase 3: WebSocket hub re-enabled with frontend connection
	hub := logs_services.NewWebSocketHub()
	go hub.Run()
	defer hub.Stop() // Ensure graceful shutdown of WebSocket hub

	// Register WebSocket routes
	logs_services.RegisterWebSocketRoutes(router, hub)

	// Health check endpoint (system-wide diagnostics)
	router.GET("/api/logs/healthcheck", resthandlers.GetHealthCheck)

	// Simple health endpoint for smoke tests
	router.GET("/api/logs/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "logs",
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	// Phase 3: Health Intelligence - Initialize services
	storageService := logs_services.NewHealthStorageService(dbConn)
	policyService := logs_services.NewHealthPolicyService(dbConn)
	repairService := logs_services.NewAutoRepairService(dbConn, policyService)

	// Initialize default policies on startup
	if err := policyService.InitializeDefaultPolicies(context.Background()); err != nil {
		log.Printf("Warning: Failed to initialize health policies: %v", err)
	}

	// Initialize UI handler with policy service
	uiHandler := apphandlers.NewUIHandler(logger, policyService)

	// Register UI routes for dashboard
	apphandlers.RegisterUIRoutes(router, uiHandler)

	// Register Phase 3 API endpoints
	router.GET("/api/health/history", resthandlers.GetHealthHistory(storageService))
	router.GET("/api/health/trends/:service", resthandlers.GetHealthTrends(storageService))
	router.GET("/api/health/policies", resthandlers.GetHealthPolicies(policyService))
	router.GET("/api/health/policies/:service", resthandlers.GetHealthPolicy(policyService))
	router.PUT("/api/health/policies/:service", resthandlers.UpdateHealthPolicy(policyService))
	router.GET("/api/health/repairs", resthandlers.GetRepairHistory(repairService))
	router.POST("/api/health/repair/:service", resthandlers.ManualRepair(repairService, storageService))

	// Start health scheduler (runs background checks every 5 minutes)
	scheduler := logs_services.NewHealthScheduler(5*time.Minute, storageService, repairService)
	scheduler.Start()
	defer scheduler.Stop() // Ensure graceful shutdown of health scheduler

	log.Println("Health intelligence system initialized - scheduler running every 5 minutes")

	log.Printf("Starting logs service on port %s", port)

	// Create an HTTP server with timeouts
	// WriteTimeout increased to 60s for AI generation endpoints (can take 10-20s for complex logs)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      60 * time.Second, // Increased from 10s for AI generation
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("[ERROR] Failed to close database: %v", closeErr)
		}
		log.Fatalf("[ERROR] Failed to start server: %v", err)
	}
}

// runMigrations executes the database migration SQL file
func runMigrations(db *sql.DB) error {
	migrationSQL := `-- Phase 3: Health Intelligence & Automation
-- Creates tables for health check history, security scans, auto-repairs, and policies

-- Store health check results over time
CREATE TABLE IF NOT EXISTS logs.health_checks (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    overall_status VARCHAR(20) NOT NULL,
    duration_ms INTEGER NOT NULL,
    check_count INTEGER NOT NULL,
    passed_count INTEGER NOT NULL,
    warned_count INTEGER NOT NULL,
    failed_count INTEGER NOT NULL,
    report_json JSONB NOT NULL,
    triggered_by VARCHAR(50) DEFAULT 'manual'
);

CREATE INDEX IF NOT EXISTS idx_health_checks_timestamp ON logs.health_checks(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_health_checks_status ON logs.health_checks(overall_status);

-- Store individual check results for detailed analysis
CREATE TABLE IF NOT EXISTS logs.health_check_details (
    id SERIAL PRIMARY KEY,
    health_check_id INTEGER NOT NULL REFERENCES logs.health_checks(id) ON DELETE CASCADE,
    check_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    error TEXT,
    duration_ms INTEGER NOT NULL,
    details_json JSONB
);

CREATE INDEX IF NOT EXISTS idx_health_check_details_check_id ON logs.health_check_details(health_check_id);
CREATE INDEX IF NOT EXISTS idx_health_check_details_name ON logs.health_check_details(check_name);
CREATE INDEX IF NOT EXISTS idx_health_check_details_status ON logs.health_check_details(status);

-- Store Trivy security scan results
CREATE TABLE IF NOT EXISTS logs.security_scans (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    scan_type VARCHAR(50) NOT NULL,
    target VARCHAR(255) NOT NULL,
    critical_count INTEGER DEFAULT 0,
    high_count INTEGER DEFAULT 0,
    medium_count INTEGER DEFAULT 0,
    low_count INTEGER DEFAULT 0,
    scan_json JSONB NOT NULL,
    health_check_id INTEGER REFERENCES logs.health_checks(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_security_scans_timestamp ON logs.security_scans(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_security_scans_critical ON logs.security_scans(critical_count DESC);
CREATE INDEX IF NOT EXISTS idx_security_scans_type ON logs.security_scans(scan_type);

-- Store auto-repair actions
CREATE TABLE IF NOT EXISTS logs.auto_repairs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    health_check_id INTEGER REFERENCES logs.health_checks(id) ON DELETE SET NULL,
    service_name VARCHAR(100) NOT NULL,
    issue_type VARCHAR(100) NOT NULL,
    repair_action VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error TEXT,
    duration_ms INTEGER
);

CREATE INDEX IF NOT EXISTS idx_auto_repairs_service ON logs.auto_repairs(service_name);
CREATE INDEX IF NOT EXISTS idx_auto_repairs_status ON logs.auto_repairs(status);
CREATE INDEX IF NOT EXISTS idx_auto_repairs_timestamp ON logs.auto_repairs(timestamp DESC);

-- Store custom health policies per service
CREATE TABLE IF NOT EXISTS logs.health_policies (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL UNIQUE,
    max_response_time_ms INTEGER DEFAULT 1000,
    auto_repair_enabled BOOLEAN DEFAULT true,
    repair_strategy VARCHAR(50) DEFAULT 'restart',
    alert_on_warn BOOLEAN DEFAULT false,
    alert_on_fail BOOLEAN DEFAULT true,
    policy_json JSONB,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_policies_service ON logs.health_policies(service_name);

-- Add comments for documentation
COMMENT ON TABLE logs.health_checks IS 'Stores health check results over time for trend analysis';
COMMENT ON TABLE logs.health_check_details IS 'Stores individual check results (e.g., HTTP, database, container)';
COMMENT ON TABLE logs.security_scans IS 'Stores Trivy security scan results for vulnerability tracking';
COMMENT ON TABLE logs.auto_repairs IS 'Stores auto-repair action history and outcomes';
COMMENT ON TABLE logs.health_policies IS 'Stores custom health policies for each service';

-- Phase 1: AI-Driven Diagnostics
-- Add AI analysis columns to logs.entries table
ALTER TABLE logs.entries 
ADD COLUMN IF NOT EXISTS issue_type VARCHAR(50),
ADD COLUMN IF NOT EXISTS ai_analysis JSONB,
ADD COLUMN IF NOT EXISTS severity_score INT;

-- Create index for efficient querying by issue type
CREATE INDEX IF NOT EXISTS idx_logs_entries_issue_type 
ON logs.entries(issue_type, created_at DESC);

-- Create index for severity queries
CREATE INDEX IF NOT EXISTS idx_logs_entries_severity 
ON logs.entries(severity_score DESC, created_at DESC);

-- Add comments for documentation
COMMENT ON COLUMN logs.entries.issue_type IS 'Categorized error type: db_connection, auth_failure, null_pointer, rate_limit, network_timeout, unknown';
COMMENT ON COLUMN logs.entries.ai_analysis IS 'Cached AI analysis result with root cause, suggested fix, and fix steps';
COMMENT ON COLUMN logs.entries.severity_score IS 'Severity rating from AI analysis: 1-5 (1=info, 5=critical)';

-- Phase 4: Health Monitoring Dashboard & Alert Engine
-- Create monitoring schema for health metrics and alerts
CREATE SCHEMA IF NOT EXISTS monitoring;

-- Store API call metrics for error rate and response time analysis
CREATE TABLE IF NOT EXISTS monitoring.api_metrics (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER NOT NULL,
    payload_size_bytes INTEGER DEFAULT 0,
    user_id INTEGER,
    error_type VARCHAR(100),
    error_message TEXT,
    service_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_metrics_timestamp ON monitoring.api_metrics(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_api_metrics_errors ON monitoring.api_metrics(timestamp, status_code) WHERE status_code >= 400;
CREATE INDEX IF NOT EXISTS idx_api_metrics_service ON monitoring.api_metrics(service_name, timestamp DESC);

-- Store detected alerts from alert engine
CREATE TABLE IF NOT EXISTS monitoring.alerts (
    id BIGSERIAL PRIMARY KEY,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    value FLOAT,
    threshold FLOAT,
    service_name VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_alerts_active ON monitoring.alerts(created_at DESC) WHERE resolved_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_service ON monitoring.alerts(service_name, created_at DESC);

-- Add service column to health_checks table for service health monitoring
ALTER TABLE logs.health_checks 
ADD COLUMN IF NOT EXISTS service VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_health_checks_service ON logs.health_checks(service, timestamp DESC);

COMMENT ON SCHEMA monitoring IS 'Health monitoring metrics and alerts for real-time dashboard';
COMMENT ON TABLE monitoring.api_metrics IS 'API call metrics for error rate and response time analysis';
COMMENT ON TABLE monitoring.alerts IS 'Detected alerts from alert engine evaluation';
`

	if _, err := db.Exec(migrationSQL); err != nil {
		return fmt.Errorf("migration execution failed: %w", err)
	}
	log.Println("Database migrations completed successfully")
	return nil
}
