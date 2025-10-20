package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/review/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// Dependency stubs for local dev/demo
type OllamaClientStub struct{}

func (o *OllamaClientStub) Generate(ctx context.Context, prompt string) (string, error) {
	return `{"functions":[],"interfaces":[],"data_models":[],"workflows":[],"summary":"Stubbed AI output"}`, nil
}

type MockAnalysisRepository struct{}

func (m *MockAnalysisRepository) FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*models.AnalysisResult, error) {
	return nil, fmt.Errorf("not found")
}
func (m *MockAnalysisRepository) Create(ctx context.Context, result *models.AnalysisResult) error {
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
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	reviewRepo := db.NewReviewRepository(sqlDB)


	ollamaClient := &OllamaClientStub{}
	analysisRepo := &MockAnalysisRepository{}
	skimService := services.NewSkimService(ollamaClient, analysisRepo)
	scanService := services.NewScanService(ollamaClient, analysisRepo)
	reviewService := services.NewReviewService(skimService, reviewRepo)
	previewService := services.NewPreviewService()
	reviewHandler := handlers.NewReviewHandler(reviewService, previewService, skimService)
	reviewHandler.scanService = scanService // Inject scanService if needed


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
	// ...existing code...
	fmt.Printf("Review service starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
