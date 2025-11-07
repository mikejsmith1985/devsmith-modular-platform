// Package main starts the logs service for the DevSmith platform.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	apphandlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/logs/handlers"
	resthandlers "github.com/mikejsmith1985/devsmith-modular-platform/cmd/logs/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	internal_logs_handlers "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/handlers"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/monitoring"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/session"
	"github.com/sirupsen/logrus"
)

var dbConn *sql.DB

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
	if err := dbConn.Ping(); err != nil {
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("[ERROR] Failed to close database: %v", closeErr)
		}
		log.Fatal("Failed to ping database:", err)
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
	log.Printf("Redis session store initialized: addr=%s, ttl=7 days", redisAddr)

	// Run database migrations
	if err := runMigrations(dbConn); err != nil {
		log.Fatal("Failed to run migrations:", err)
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
		//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
		instrLogger.LogEvent(c.Request.Context(), "request_received", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		})
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
	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://host.docker.internal:11434" // Default for Docker
	}
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "qwen2.5-coder:7b-instruct-q4_K_M" // Default model
	}

	ollamaClient := providers.NewOllamaClient(ollamaEndpoint, ollamaModel)
	aiAnalyzer := logs_services.NewAIAnalyzer(ollamaClient)
	patternMatcher := logs_services.NewPatternMatcher()
	analysisService := logs_services.NewAnalysisService(aiAnalyzer, patternMatcher)
	analysisHandler := internal_logs_handlers.NewAnalysisHandler(analysisService, logger)

	log.Println("AI analysis services initialized - Ollama:", ollamaEndpoint, "Model:", ollamaModel)

	// Register REST API routes
	router.POST("/api/logs", func(c *gin.Context) {
		resthandlers.PostLogs(restSvc)(c)
	})
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

	// Phase 1: AI-Driven Diagnostics - Analysis endpoints
	router.POST("/api/logs/analyze", analysisHandler.AnalyzeLog)
	router.POST("/api/logs/classify", analysisHandler.ClassifyLog)

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

	// Initialize WebSocket hub
	hub := logs_services.NewWebSocketHub()
	go hub.Run()

	// Register WebSocket routes
	logs_services.RegisterWebSocketRoutes(router, hub)

	// Health check endpoint (system-wide diagnostics)
	router.GET("/api/logs/healthcheck", resthandlers.GetHealthCheck)

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
	go scheduler.Start()

	log.Println("Health intelligence system initialized - scheduler running every 5 minutes")

	log.Printf("Starting logs service on port %s", port)

	// Create an HTTP server with timeouts
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
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
