// The analytics command starts the analytics service for the DevSmith platform.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	app_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/analytics/handlers"
	analytics_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/db"
	analytics_handlers "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/handlers"
	analytics_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/config"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize instrumentation logger for this service using validated config
	logsServiceURL, logsEnabled, err := config.LoadLogsConfigWithFallbackFor("analytics")
	if err != nil {
		log.Fatalf("Failed to load logging configuration: %v", err)
	}
	if !logsEnabled {
		log.Printf("Instrumentation/logging disabled: continuing startup without external logs")
		logsServiceURL = ""
	}
	instrLogger := instrumentation.NewServiceInstrumentationLogger("analytics", logsServiceURL)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatal("DATABASE_URL environment variable is required")
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to the database")
	}
	defer dbPool.Close()

	aggregationRepo := analytics_db.NewAggregationRepository(dbPool)
	logReader := analytics_db.NewLogReader(dbPool)

	aggregatorService := analytics_services.NewAggregatorService(aggregationRepo, logReader, logger)
	trendService := analytics_services.NewTrendService(aggregationRepo, logger)
	anomalyService := analytics_services.NewAnomalyService(aggregationRepo, logger)
	topIssuesService := analytics_services.NewTopIssuesService(logReader, logger)
	exportService := analytics_services.NewExportService(aggregationRepo, logger)

	apiHandler := analytics_handlers.NewAnalyticsHandler(aggregatorService, trendService, anomalyService, topIssuesService, exportService, logger)

	router := gin.Default()

	// Middleware for logging requests (skip health checks)
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

	// Serve static files (CSS, JS)
	router.Static("/static", "./apps/analytics/static")

	// Health endpoint for nginx and orchestration
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "analytics",
		})
	})

	// Register API routes
	apiHandler.RegisterRoutes(router)

	// Register UI routes
	app_handlers.RegisterUIRoutes(router, logger)

	// Register debug routes (development only)
	debug.RegisterDebugRoutes(router, "analytics")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	logger.Infof("Analytics service starting on port %s...", port)

	if err := router.Run(":" + port); err != nil {
		logger.WithError(err).Fatalf("Failed to start server: %v", err)
	}
}
