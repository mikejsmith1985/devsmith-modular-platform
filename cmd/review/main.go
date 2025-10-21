// DevSmith Review service main entry point.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/review/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// Dependency stubs for local dev/demo

// OllamaClientStub is a stub implementation of the OllamaClient for local development and testing.
type OllamaClientStub struct{}

// Generate simulates the generation of a response by the OllamaClientStub.
func (o *OllamaClientStub) Generate(_ context.Context, _ string) (string, error) {
	return `{"functions":[],"interfaces":[],"data_models":[],"workflows":[],"summary":"Stubbed AI output"}`, nil
}

// MockAnalysisRepository is a mock implementation of the AnalysisRepository for testing purposes.
type MockAnalysisRepository struct{}

// FindByReviewAndMode retrieves a mock analysis result based on the review ID and mode.
func (m *MockAnalysisRepository) FindByReviewAndMode(_ context.Context, _ int64, _ string) (*models.AnalysisResult, error) {
	return nil, fmt.Errorf("not found")
}

// Create saves a mock analysis result.
func (m *MockAnalysisRepository) Create(_ context.Context, _ *models.AnalysisResult) error {
	return nil
}

func main() {
	router := gin.Default()

	// Health and root endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "review",
			"status":  "healthy",
		})
	})
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "DevSmith Review",
			"version": "0.1.0",
			"message": "Review service is running",
		})
	})

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
		// Replace os.Exit with proper error handling
		log.Printf("Failed to ping DB: %v", err)
		return
	}

	reviewRepo := db.NewReviewRepository(sqlDB)

	ollamaClient := &OllamaClientStub{}
	analysisRepo := &MockAnalysisRepository{}
	skimService := services.NewSkimService(ollamaClient, analysisRepo)
	scanService := services.NewScanService(ollamaClient, analysisRepo)
	reviewService := services.NewReviewService(skimService, reviewRepo)
	previewService := services.NewPreviewService()
	reviewHandler := handlers.NewReviewHandler(reviewService, previewService, skimService, scanService)

	// Skim Mode endpoint
	router.GET("/api/reviews/:id/skim", reviewHandler.GetSkimAnalysis)

	// Scan Mode endpoint
	router.GET("/api/reviews/:id/scan", reviewHandler.GetScanAnalysis)

	// Create review session endpoint
	router.POST("/api/review/sessions", reviewHandler.CreateReviewSession)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Review service starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Printf("Failed to start server: %v", err)
		return
	}
}
