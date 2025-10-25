// Package main starts the logs service for the DevSmith platform.
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/logs/handlers"
	restHandlers "github.com/mikejsmith1985/devsmith-modular-platform/cmd/logs/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/service"
	"github.com/sirupsen/logrus"
)

var dbConn *sql.DB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

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

	// Setup repository and service for logs REST API
	logRepo := db.NewLogRepository(dbConn)
	logService := service.NewLogService(logRepo)

	// Register REST API routes
	router.POST("/api/logs", restHandlers.PostLogs(logService))
	router.GET("/api/logs", restHandlers.GetLogs(logService))
	router.GET("/api/logs/:id", restHandlers.GetLogByID(logService))
	router.GET("/api/logs/stats", restHandlers.GetStats(logService))
	router.DELETE("/api/logs", restHandlers.DeleteLogs(logService))

	// Serve static files for logs dashboard
	router.Static("/static", "./apps/logs/static")

	// Register UI routes for dashboard
	handlers.RegisterUIRoutes(router, logger)

	// Register debug routes (development only)
	debug.RegisterDebugRoutes(router, "logs")

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
