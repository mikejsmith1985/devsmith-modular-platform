// The analytics command starts the analytics service for the DevSmith platform.
package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/analytics/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/db"
	analytics_handlers "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/config"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize instrumentation logger for this service using validated config
	logsServiceURL, err := config.LoadLogsConfig()
	if err != nil {
		log.Fatalf("Failed to load logging configuration: %v", err)
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

	aggregationRepo := db.NewAggregationRepository(dbPool)
	logReader := db.NewLogReader(dbPool)

	aggregatorService := services.NewAggregatorService(aggregationRepo, logReader, logger)
	trendService := services.NewTrendService(aggregationRepo, logger)
	anomalyService := services.NewAnomalyService(aggregationRepo, logger)
	topIssuesService := services.NewTopIssuesService(logReader, logger)
	exportService := services.NewExportService(aggregationRepo, logger)

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

	// Register API routes
	apiHandler.RegisterRoutes(router)

	// Register UI routes
	handlers.RegisterUIRoutes(router, logger)

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
