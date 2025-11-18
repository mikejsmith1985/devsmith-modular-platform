// DevSmith Review service main entry point.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	app_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/review/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/config"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logging"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/middleware"
	review_circuit "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/circuit"
	review_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/github"
	review_handlers "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/handlers"
	review_health "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/health"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	review_tracing "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/tracing"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/session"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// nolint:gocyclo // Main initialization is inherently complex with multiple setup steps
func main() {
	// Create app-level context that will be cancelled on shutdown
	appCtx, cancelAppCtx := context.WithCancel(context.Background())
	defer cancelAppCtx()

	router := gin.Default()

	// Load and validate logs service configuration (allow configurable fallback)
	logURL, logsEnabled, err := config.LoadLogsConfigWithFallbackFor("review")
	if err != nil {
		log.Fatalf("Failed to load logging configuration: %v", err)
	}
	if !logsEnabled {
		log.Printf("Logging disabled at startup (LOGS_STRICT=false and config invalid)")
	}

	// Initialize structured logger for this service
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	reviewLogger, err := logger.NewLogger(&logger.Config{
		ServiceName:     "review",
		LogLevel:        logLevel,
		LogURL:          logURL,
		BatchSize:       100,
		BatchTimeoutSec: 5,
		LogToStdout:     true,
		EnableStdout:    true,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize OpenTelemetry tracing
	tracingEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if tracingEndpoint == "" {
		tracingEndpoint = "http://jaeger:4318" // Default to docker-compose service name
	}
	shutdownTracer, err := review_tracing.InitTracer("devsmith-review", tracingEndpoint)
	if err != nil {
		log.Printf("Warning: Failed to initialize tracing: %v", err)
	} else {
		defer shutdownTracer(context.Background())
		log.Printf("Tracing initialized (endpoint: %s)", tracingEndpoint)
	}

	// Middleware: Log all requests (async, non-blocking)
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path != "/health" {
			reviewLogger.Info("Incoming request", "method", c.Request.Method, "path", c.Request.URL.Path)
		}
		c.Next()
	})

	// Register debug routes (development only)
	debug.RegisterDebugRoutes(router, "review")

	// --- Database connection (PostgreSQL, pgx) ---
	dbURL := os.Getenv("REVIEW_DB_URL")
	if dbURL == "" {
		log.Fatal("REVIEW_DB_URL environment variable is required")
	}
	sqlDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	// Configure connection pool to prevent exhaustion
	sqlDB.SetMaxOpenConns(10)               // Max 10 connections per service
	sqlDB.SetMaxIdleConns(5)                // Keep 5 idle
	sqlDB.SetConnMaxLifetime(3600000000000) // 1 hour
	sqlDB.SetConnMaxIdleTime(600000000000)  // 10 minutes

	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing DB: %v", err)
		}
	}()
	if err := sqlDB.Ping(); err != nil {
		log.Printf("Failed to ping DB: %v", err)
		return
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
		if err := sessionStore.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}()
	reviewLogger.Info("Redis session store initialized", "addr", redisAddr, "ttl", "7 days")

	// Repository and service setup
	analysisRepo := review_db.NewAnalysisRepository(sqlDB)
	githubRepo := review_db.NewGitHubRepository(sqlDB)
	promptRepo := review_db.NewPromptTemplateRepository(sqlDB)

	// Start retention job for troubleshooting analysis captures (default 14 days)
	retentionDays := 14
	if v := os.Getenv("ANALYSIS_RETENTION_DAYS"); v != "" {
		if d, err := strconv.Atoi(v); err == nil {
			retentionDays = d
		}
	}
	retentionInterval := 24 * time.Hour
	if v := os.Getenv("ANALYSIS_RETENTION_INTERVAL_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			retentionInterval = time.Duration(h) * time.Hour
		}
	}
	// Start retention job (best-effort, uses analysisRepo.DeleteOlderThan)
	review_services.StartRetentionJob(appCtx, analysisRepo, retentionDays, retentionInterval, reviewLogger)

	// ==========================================
	// AI CLIENT INITIALIZATION
	// ==========================================
	// Initialize unified AI client that fetches configs from Portal's AI Factory
	// No environment variables needed - users configure models through AI Factory UI
	portalURL := os.Getenv("PORTAL_URL")
	if portalURL == "" {
		portalURL = "http://portal:3001" // Default to Docker Compose service name
	}
	reviewLogger.Info("Initializing AI client", "portal_url", portalURL, "config_source", "Portal AI Factory")

	unifiedAIClient := review_services.NewUnifiedAIClient(portalURL)

	// Wrap unified AI client with circuit breaker for resilience
	aiClientWithCircuitBreaker := review_circuit.NewOllamaCircuitBreaker(unifiedAIClient, reviewLogger)
	reviewLogger.Info("Circuit breaker initialized", "threshold", 5, "timeout", "60s")

	// NOTE: ModelService and MultiFileAnalyzer still use direct Ollama for model discovery
	// These will be refactored in future to use Portal AI Factory as well
	// For now, we keep a minimal Ollama client just for these legacy services
	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://host.docker.internal:11434"
	}
	ollamaDefaultModel := "mistral:7b-instruct" // Used only for multiFileAnalyzer fallback
	ollamaClient := providers.NewOllamaClient(ollamaEndpoint, ollamaDefaultModel)

	// Wire up services with circuit breaker wrapper (fail-fast when AI is unhealthy)
	previewService := review_services.NewPreviewService(aiClientWithCircuitBreaker, reviewLogger)
	skimService := review_services.NewSkimService(aiClientWithCircuitBreaker, analysisRepo, reviewLogger)
	scanService := review_services.NewScanService(aiClientWithCircuitBreaker, analysisRepo, reviewLogger)
	detailedService := review_services.NewDetailedService(aiClientWithCircuitBreaker, analysisRepo, reviewLogger)
	criticalService := review_services.NewCriticalService(aiClientWithCircuitBreaker, analysisRepo, reviewLogger)

	// Initialize health checker with all services
	healthChecker := review_health.NewServiceHealthChecker(
		previewService,
		skimService,
		scanService,
		detailedService,
		criticalService,
		unifiedAIClient, // Use unified client for health checks
		sqlDB,
		reviewLogger,
	)

	// Health and root endpoints (registered after healthChecker initialization)
	router.GET("/api/review/health", func(c *gin.Context) {
		// Perform comprehensive health check
		health, err := healthChecker.CheckHealth(c.Request.Context())
		if err != nil {
			reviewLogger.Error("Health check failed", "error", err)
			c.JSON(500, gin.H{
				"service": "review",
				"status":  "error",
				"error":   err.Error(),
			})
			return
		}

		// Map health status to HTTP status code
		statusCode := 200
		if health.Status == review_health.HealthStatusDegraded {
			statusCode = 200 // Still serving traffic
		} else if health.Status == review_health.HealthStatusUnhealthy {
			statusCode = 503 // Service unavailable
		}

		c.JSON(statusCode, health)
	})
	router.HEAD("/health", func(c *gin.Context) {
		reviewLogger.Info("HEAD /health endpoint hit")
		c.Status(200)
	})
	// Add GET handler for Traefik load balancer health check (uses GET by default)
	router.GET("/health", func(c *gin.Context) {
		reviewLogger.Info("GET /health endpoint hit")
		c.Status(200)
	})

	// Prepare logging client to send lightweight events to Logs service (optional)
	var logClient *logging.Client
	if logsEnabled && logURL != "" {
		logClient = logging.NewClient(logURL)
	} else {
		logClient = nil
	}

	// Create model service for dynamic model discovery (needs Ollama endpoint)
	modelService := review_services.NewModelService(reviewLogger, ollamaEndpoint)

	// Handler setup with services (UIHandler takes logger, logging client, and AI services)
	uiHandler := app_handlers.NewUIHandler(reviewLogger, logClient, previewService, skimService, scanService, detailedService, criticalService, modelService)

	// Initialize GitHub client for Phase 2 GitHub integration
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		reviewLogger.Warn("GITHUB_TOKEN not set - GitHub API rate limited to 60 requests/hour")
	}
	githubClient := github.NewDefaultClient()

	// Initialize multi-file analyzer service for GitHub session analysis
	// Note: MultiFileAnalyzer uses ai.Provider interface, uses Ollama client directly
	// TODO: Refactor to use Portal AI Factory
	multiFileAnalyzer := review_services.NewMultiFileAnalyzer(ollamaClient, ollamaDefaultModel)

	// Initialize GitHub session handler for repository integration
	githubSessionHandler := review_handlers.NewGitHubSessionHandler(githubRepo, githubClient, multiFileAnalyzer)

	// Initialize GitHub handler for Phase 1 GitHub integration (tree, file, quick-scan endpoints)
	// Pass previewService so Quick Scan can run AI analysis
	githubHandler := review_handlers.NewGitHubHandler(reviewLogger, previewService)

	// Initialize prompt template service and handler for prompt management
	promptService := review_services.NewPromptTemplateService(promptRepo)
	promptHandler := review_handlers.NewPromptHandler(promptService)

	// Serve static files (CSS, JS) from apps/review/static
	router.Static("/static", "./apps/review/static")
	reviewLogger.Info("Static files configured", "path", "/static", "dir", "./apps/review/static")

	// Public endpoints (no authentication required)
	router.GET("/api/review/models", uiHandler.GetAvailableModels) // Model list is public

	// Home/landing page - REQUIRES authentication via Redis session (SSO with Portal)
	// Handles both / (legacy direct access) and /review (Traefik gateway access)
	router.GET("/", middleware.RedisSessionAuthMiddleware(sessionStore), uiHandler.HomeHandler)
	router.HEAD("/", middleware.RedisSessionAuthMiddleware(sessionStore), uiHandler.HomeHandler)
	router.GET("/review", middleware.RedisSessionAuthMiddleware(sessionStore), uiHandler.HomeHandler)
	router.HEAD("/review", middleware.RedisSessionAuthMiddleware(sessionStore), uiHandler.HomeHandler)

	// Protected endpoints group (require JWT authentication with Redis session validation)
	protected := router.Group("/")
	protected.Use(middleware.RedisSessionAuthMiddleware(sessionStore))
	{
		// Workspace access (requires auth to track user sessions)
		// Browser accesses /review/workspace/123 and Traefik passes it as-is (no prefix stripping)
		protected.GET("/review/workspace/:session_id", uiHandler.ShowWorkspace)

		// Analysis endpoints (require auth for usage tracking and rate limiting)
		protected.GET("/analysis", uiHandler.AnalysisResultHandler)
		protected.POST("/api/review/sessions", uiHandler.CreateSessionHandler)
		protected.GET("/api/review/sessions/:id/progress", uiHandler.SessionProgressSSE)

		// Mode endpoints - all require authentication
		protected.POST("/api/review/modes/preview", uiHandler.HandlePreviewMode)
		protected.POST("/api/review/modes/skim", uiHandler.HandleSkimMode)
		protected.POST("/api/review/modes/scan", uiHandler.HandleScanMode)
		protected.POST("/api/review/modes/detailed", uiHandler.HandleDetailedMode)
		protected.POST("/api/review/modes/critical", uiHandler.HandleCriticalMode)

		// Session management endpoints (all require auth)
		protected.GET("/api/review/sessions/list", uiHandler.ListSessionsHTMX)
		protected.GET("/api/review/sessions/search", uiHandler.SearchSessionsHTMX)
		protected.GET("/api/review/sessions/:id", uiHandler.GetSessionDetailHTMX)
		protected.POST("/api/review/sessions/:id/resume", uiHandler.ResumeSessionHTMX)
		protected.POST("/api/review/sessions/:id/duplicate", uiHandler.DuplicateSessionHTMX)
		protected.POST("/api/review/sessions/:id/archive", uiHandler.ArchiveSessionHTMX)

		// GitHub session endpoints (Phase 2 - GitHub integration)
		protected.POST("/api/review/sessions/github", githubSessionHandler.CreateSession)
		protected.GET("/api/review/sessions/:id/github", githubSessionHandler.GetSession)
		protected.GET("/api/review/sessions/:id/tree", githubSessionHandler.GetTree)
		protected.POST("/api/review/sessions/:id/files", githubSessionHandler.OpenFile)
		protected.GET("/api/review/sessions/:id/files", githubSessionHandler.GetOpenFiles)
		protected.DELETE("/api/review/files/:tab_id", githubSessionHandler.CloseFile)
		protected.PATCH("/api/review/sessions/:id/files/activate", githubSessionHandler.SetActiveTab)
		protected.POST("/api/review/sessions/:id/analyze", githubSessionHandler.AnalyzeMultipleFiles)

		// GitHub Phase 1 endpoints (tree, file, quick-scan)
		protected.GET("/api/review/github/tree", githubHandler.GetRepoTree)
		protected.GET("/api/review/github/file", githubHandler.GetRepoFile)
		protected.GET("/api/review/github/quick-scan", githubHandler.QuickRepoScan)

		// Prompt template endpoints (Issue #2 - Details button)
		protected.GET("/api/review/prompts", promptHandler.GetPrompt)
		protected.PUT("/api/review/prompts", promptHandler.SavePrompt)
		protected.DELETE("/api/review/prompts", promptHandler.ResetPrompt)
		protected.GET("/api/review/prompts/history", promptHandler.GetHistory)
	}
	router.DELETE("/api/review/sessions/:id", uiHandler.DeleteSessionHTMX)            // Delete session (HTMX, replaces sessionHandler.DeleteSession)
	router.GET("/api/review/sessions/:id/stats", uiHandler.GetSessionStatsHTMX)       // Session statistics
	router.GET("/api/review/sessions/:id/metadata", uiHandler.GetSessionMetadataHTMX) // Session metadata
	router.GET("/api/review/sessions/:id/export", uiHandler.ExportSessionHTMX)        // Export session

	// Debug routes (TODO: remove in production or guard with env flag)
	app_handlers.RegisterDebugRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Create HTTP server with graceful shutdown support
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		reviewLogger.Info("Review service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			reviewLogger.Error("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	reviewLogger.Info("Shutting down gracefully...")

	// Cancel app context to signal retention job and other background tasks to stop
	cancelAppCtx()

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		reviewLogger.Error("Server forced to shutdown", "error", err)
		return
	}

	reviewLogger.Info("Server shutdown complete")
}
