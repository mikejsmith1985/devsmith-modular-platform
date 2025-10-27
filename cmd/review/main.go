// DevSmith Review service main entry point.
package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/review/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

func main() {
	router := gin.Default()

	// Initialize structured logger for this service
	logURL := os.Getenv("LOGS_SERVICE_URL")
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

	// Middleware: Log all requests (async, non-blocking)
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path != "/health" {
			reviewLogger.Info("Incoming request", "method", c.Request.Method, "path", c.Request.URL.Path)
		}
		c.Next()
	})

	// Health and root endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "review",
			"status":  "healthy",
		})
	})
	router.HEAD("/health", func(c *gin.Context) {
		reviewLogger.Info("HEAD /health endpoint hit")
		c.Status(200)
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
	analysisRepo := db.NewAnalysisRepository(sqlDB)

	// TODO: Replace with real Ollama client implementation
	ollamaClient := &services.OllamaClientStub{}

	// Wire up services (if needed for future handler expansion)
	_ = services.NewSkimService(ollamaClient, analysisRepo, reviewLogger)
	_ = services.NewScanService(ollamaClient, analysisRepo, reviewLogger)
	_ = services.NewDetailedService(ollamaClient, analysisRepo, reviewLogger)
	_ = services.NewPreviewService(reviewLogger)

	// Handler setup (UIHandler currently only takes logger)
	uiHandler := handlers.NewUIHandler(reviewLogger)

	// Register endpoints
	router.GET("/", uiHandler.HomeHandler)
	router.GET("/analysis", uiHandler.AnalysisResultHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	reviewLogger.Info("Review service starting", "port", port)
	if err := router.Run(":" + port); err != nil {
		reviewLogger.Error("Failed to start server", "error", err)
		return
	}
}
