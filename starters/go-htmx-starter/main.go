package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const defaultPort = "8080"

func main() {
	// Initialize logger
	serviceName := getEnv("SERVICE_NAME", "go-htmx-starter")
	logLevel := getEnv("LOG_LEVEL", "INFO")
	logger := NewLogger(serviceName, logLevel)

	// Get port from environment
	port := getEnv("PORT", defaultPort)

	// Setup routes
	mux := setupRoutes(logger)

	// Start server
	addr := ":" + port
	logger.Info(context.Background(), "server_starting", map[string]interface{}{
		"port":    port,
		"service": serviceName,
	})

	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error(context.Background(), "server_error", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatal(err)
	}
}

func setupRoutes(logger *Logger) http.Handler {
	mux := http.NewServeMux()

	// Apply correlation middleware to all routes
	handler := correlationMiddleware(mux, logger)

	// Health check endpoint (required by platform)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"service":   getEnv("SERVICE_NAME", "go-htmx-starter"),
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)

		logger.Debug(r.Context(), "health_check", map[string]interface{}{
			"status": "healthy",
		})
	})

	// Home page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		logger.Info(r.Context(), "page_view", map[string]interface{}{
			"page": "home",
		})

		tmpl := template.Must(template.ParseFiles(
			"templates/base.html",
			"templates/index.html",
		))

		data := struct {
			Title   string
			Message string
		}{
			Title:   "Go + HTMX Starter",
			Message: "Welcome to the DevSmith Platform Starter Template",
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			logger.Error(r.Context(), "template_error", map[string]interface{}{
				"error": err.Error(),
				"page":  "home",
			})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	// HTMX fragment endpoint (demonstrates partial HTML responses)
	mux.HandleFunc("/api/fragment", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(r.Context(), "api_call", map[string]interface{}{
			"endpoint": "/api/fragment",
			"method":   r.Method,
		})

		tmpl := template.Must(template.ParseFiles("templates/fragment.html"))

		data := struct {
			Message   string
			Timestamp string
		}{
			Message:   "This content was loaded dynamically with HTMX!",
			Timestamp: time.Now().Format("15:04:05"),
		}

		if err := tmpl.Execute(w, data); err != nil {
			logger.Error(r.Context(), "template_error", map[string]interface{}{
				"error":    err.Error(),
				"endpoint": "/api/fragment",
			})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	return handler
}

// correlationMiddleware extracts or generates correlation ID for request tracing
func correlationMiddleware(next http.Handler, logger *Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract or generate correlation ID
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)
		r = r.WithContext(ctx)

		// Add to response headers
		w.Header().Set("X-Correlation-ID", correlationID)

		// Log request
		logger.Info(ctx, "http_request", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		})

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
