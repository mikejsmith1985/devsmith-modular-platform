// Package main starts the logs service for the DevSmith platform.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	apphandlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/logs/handlers"
	resthandlers "github.com/mikejsmith1985/devsmith-modular-platform/cmd/logs/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
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

	// Verify connection
	if err := dbConn.Ping(); err != nil {
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("[ERROR] Failed to close database: %v", closeErr)
		}
		log.Fatal("Failed to ping database:", err)
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
	logRepo := db.NewLogRepository(dbConn)
	restSvc := services.NewRestLogService(logRepo, logger)

	// Issue #023: Production Enhancements - Initialize alert and aggregation services
	alertConfigRepo := db.NewAlertConfigRepository(dbConn)
	alertViolationRepo := db.NewAlertViolationRepository(dbConn)

	// Create alert service for threshold management (implements AlertThresholdService interface)
	alertSvc := services.NewAlertService(alertViolationRepo, alertConfigRepo, logRepo, logger)

	// Create validation aggregation service for analytics
	validationAgg := services.NewValidationAggregation(logRepo, logger)

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

	// Initialize WebSocket hub
	hub := services.NewWebSocketHub()
	go hub.Run()

	// Register WebSocket routes
	services.RegisterWebSocketRoutes(router, hub)

	// Health check endpoint (system-wide diagnostics)
	router.GET("/api/logs/healthcheck", resthandlers.GetHealthCheck)

	// Phase 3: Health Intelligence - Initialize services
	storageService := services.NewHealthStorageService(dbConn)
	policyService := services.NewHealthPolicyService(dbConn)
	repairService := services.NewAutoRepairService(dbConn, policyService)

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
	scheduler := services.NewHealthScheduler(5*time.Minute, storageService, repairService)
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
