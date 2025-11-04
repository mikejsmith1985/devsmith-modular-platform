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
	review_circuit "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/circuit"
	review_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	review_health "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/health"
	review_middleware "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/middleware"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	review_tracing "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/tracing"
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

	// Repository and service setup
	analysisRepo := review_db.NewAnalysisRepository(sqlDB)

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

	// Initialize Ollama client with configuration from environment
	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://localhost:11434" // Default to local Ollama
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "mistral:7b-instruct" // Default to mistral
	}

	// Allow a fast deterministic mock of Ollama for Playwright/CI runs.
	// Set OLLAMA_MOCK=true in environment (via docker-compose override) to enable.
	if os.Getenv("OLLAMA_MOCK") == "true" {
		// Start a local mock Ollama HTTP server listening on 127.0.0.1:11434
		mockAddr := "127.0.0.1:11434"
		reviewLogger.Info("Starting mock Ollama server for tests", "addr", mockAddr, "model", ollamaModel)
		shutdownMock, err := providers.StartMockOllamaServer(mockAddr, ollamaModel)
		if err != nil {
			reviewLogger.Error("Failed to start mock Ollama server", "error", err)
		} else {
			// Ensure shutdown during process exit
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = shutdownMock(ctx)
			}()

			// Point client at mock server
			ollamaEndpoint = "http://127.0.0.1:11434"
		}
	}

	reviewLogger.Info("Initializing Ollama client", "endpoint", ollamaEndpoint, "model", ollamaModel)
	ollamaClient := providers.NewOllamaClient(ollamaEndpoint, ollamaModel)

	// Verify Ollama is reachable
	if err := ollamaClient.HealthCheck(context.Background()); err != nil {
		reviewLogger.Warn("Ollama health check failed (will retry on first request)", "error", err.Error())
	} else {
		reviewLogger.Info("Ollama health check passed", "model", ollamaModel)
	}

	// Wrap OllamaClient with adapter to match review services interface
	ollamaAdapter := review_services.NewOllamaClientAdapter(ollamaClient)

	// Wrap Ollama adapter with circuit breaker for resilience
	ollamaWithCircuitBreaker := review_circuit.NewOllamaCircuitBreaker(ollamaAdapter, reviewLogger)
	reviewLogger.Info("Circuit breaker initialized", "threshold", 5, "timeout", "60s")

	// Wire up services with circuit breaker wrapper (fail-fast when Ollama is unhealthy)
	previewService := review_services.NewPreviewService(ollamaWithCircuitBreaker, reviewLogger)
	skimService := review_services.NewSkimService(ollamaWithCircuitBreaker, analysisRepo, reviewLogger)
	scanService := review_services.NewScanService(ollamaWithCircuitBreaker, analysisRepo, reviewLogger)
	detailedService := review_services.NewDetailedService(ollamaWithCircuitBreaker, analysisRepo, reviewLogger)
	criticalService := review_services.NewCriticalService(ollamaWithCircuitBreaker, analysisRepo, reviewLogger)

	// Initialize health checker with all services
	healthChecker := review_health.NewServiceHealthChecker(
		previewService,
		skimService,
		scanService,
		detailedService,
		criticalService,
		ollamaAdapter,
		sqlDB,
		reviewLogger,
	)

	// Health and root endpoints (registered after healthChecker initialization)
	router.GET("/health", func(c *gin.Context) {
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

	// Serve static files (CSS, JS) from apps/review/static
	router.Static("/static", "./apps/review/static")
	reviewLogger.Info("Static files configured", "path", "/static", "dir", "./apps/review/static")

	// Public endpoints (no authentication required)
	router.GET("/api/review/models", uiHandler.GetAvailableModels) // Model list is public

	// Home/landing page with optional auth (validates JWT if present, allows unauthenticated)
	router.GET("/", review_middleware.OptionalAuthMiddleware(reviewLogger), uiHandler.HomeHandler)
	router.GET("/review", review_middleware.OptionalAuthMiddleware(reviewLogger), uiHandler.HomeHandler) // Serve UI at /review for E2E tests

	// Protected endpoints group (require JWT authentication)
	protected := router.Group("/")
	protected.Use(review_middleware.JWTAuthMiddleware(reviewLogger))
	{
		// Workspace access (requires auth to track user sessions)
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
