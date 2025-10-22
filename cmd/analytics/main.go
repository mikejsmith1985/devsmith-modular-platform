// The analytics command starts the analytics service for the DevSmith platform.
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

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

	handler := handlers.NewAnalyticsHandler(aggregatorService, trendService, anomalyService, topIssuesService, exportService, logger)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "analytics",
			"status":  "healthy",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "DevSmith Analytics",
			"version": "0.1.0",
			"message": "Analytics service is running",
		})
	})

	handler.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	logger.Infof("Analytics service starting on port %s...", port)

	if err := router.Run(":" + port); err != nil {
		logger.WithError(err).Fatalf("Failed to start server: %v", err)
	}
}
