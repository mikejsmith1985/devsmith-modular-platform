package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// ReviewHandler handles HTTP requests for the review service.
type ReviewHandler struct {
	scanService    ScanServiceInterface
	skimService    *services.SkimService
	reviewService  ReviewServiceInterface
	previewService *services.PreviewService
}

// NewReviewHandler creates a new instance of ReviewHandler.
func NewReviewHandler(reviewService ReviewServiceInterface, previewService *services.PreviewService, skimService *services.SkimService, scanService ScanServiceInterface) *ReviewHandler {
	return &ReviewHandler{
		reviewService:  reviewService,
		previewService: previewService,
		skimService:    skimService,
		scanService:    scanService,
	}
}

// GetScanAnalysis handles Scan Mode requests
func (h *ReviewHandler) GetScanAnalysis(c *gin.Context) {
	// Check and handle errors for ParseInt
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	query := c.Query("q") // GET /api/reviews/:id/scan?q=authentication

	// Validate reading mode and query
	readingMode := c.DefaultQuery("mode", "scan")
	if !ValidateRequest(c, func() error {
		return ValidateScanRequest(readingMode, query)
	}) {
		return
	}

	// Check and handle errors for GetReview
	review, err := h.reviewService.GetReview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve review"})
		return
	}

	// Check and handle errors for AnalyzeScan
	output, err := h.scanService.AnalyzeScan(c.Request.Context(), review.ID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze scan"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// CreateReviewSession handles POST /api/review/sessions
func (h *ReviewHandler) CreateReviewSession(c *gin.Context) {
	var req struct {
		Title        string `json:"title"`
		CodeSource   string `json:"code_source"`
		GithubRepo   string `json:"github_repo"`
		GithubBranch string `json:"github_branch"`
		PastedCode   string `json:"pasted_code"`
		UserID       int64  `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Validate all inputs
	if !ValidateRequest(c, func() error {
		return ValidateCreateReviewRequest(req.Title, req.CodeSource, req.PastedCode, req.GithubRepo, req.GithubBranch)
	}) {
		return
	}

	review := &db.Review{
		UserID:       req.UserID,
		Title:        req.Title,
		CodeSource:   req.CodeSource,
		GithubRepo:   req.GithubRepo,
		GithubBranch: req.GithubBranch,
		PastedCode:   req.PastedCode,
	}
	created, err := h.reviewService.CreateReview(c.Request.Context(), review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// ReviewServiceInterface defines the contract for review-related services.
type ReviewServiceInterface interface {
	// GetReview retrieves a review by its ID.
	GetReview(ctx context.Context, id int64) (*models.Review, error)
	// CreateReview creates a new review.
	CreateReview(ctx context.Context, review *db.Review) (*db.Review, error)
}

// ScanServiceInterface defines the contract for scan-related services.
type ScanServiceInterface interface {
	// AnalyzeScan analyzes the scan for a given review.
	AnalyzeScan(ctx context.Context, reviewID int64, query string) (*models.ScanModeOutput, error)
}

// GetSkimAnalysis handles GET /api/reviews/:id/skim
func (h *ReviewHandler) GetSkimAnalysis(c *gin.Context) {
	// Check and handle errors for ParseInt
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	review, err := h.reviewService.GetReview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	output, err := h.skimService.AnalyzeSkim(c.Request.Context(), review.ID, "owner", "repo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}

// ListReviewSessions handles GET /api/review/sessions
func (h *ReviewHandler) ListReviewSessions(c *gin.Context) {
	// TODO: Replace with real DB query
	sessions := []gin.H{
		{"id": 1, "title": "Example Session", "user_id": 42},
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}
