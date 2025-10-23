// Package main starts the logs service for the DevSmith platform.
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
)

var db *sql.DB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	dbURL := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("[ERROR] Failed to close database connection: %v", err)
		}
	}()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Create route registry for debug endpoint
	routeRegistry := debug.NewHTTPRouteRegistry("logs")

	// Register handlers
	http.HandleFunc("/health", healthHandler)
	routeRegistry.Register("GET", "/health")

	http.HandleFunc("/", rootHandler)
	routeRegistry.Register("GET", "/")

	// Register debug routes endpoint (development only)
	http.HandleFunc("/debug/routes", routeRegistry.Handler())

	log.Printf("Starting service on port %s", port)

	// Create an HTTP server with timeouts
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           nil,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[ERROR] Failed to start server: %v", err)
	}
}

// Health check endpoint (REQUIRED for docker-validate)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
			"checks": map[string]bool{
				"database": false,
			},
		}); err != nil {
			log.Printf("[ERROR] Failed to write health check response: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"checks": map[string]bool{
			"database": true,
		},
	}); err != nil {
		log.Printf("[ERROR] Failed to write health check response: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(map[string]string{
		"service": "logs",
		"status":  "running",
	}); err != nil {
		log.Printf("[ERROR] Failed to write root handler response: %v", err)
	}
}
