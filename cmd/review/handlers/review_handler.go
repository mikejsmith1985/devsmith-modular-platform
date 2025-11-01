package cmd_review_handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	review_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/templates"
)

// ReviewHandler handles HTTP requests for the review service.
type ReviewHandler struct {
	scanService    ScanServiceInterface
	skimService    *review_services.SkimService
	reviewService  ReviewServiceInterface
	previewService *review_services.PreviewService
	instrLogger    *instrumentation.ServiceInstrumentationLogger
}

// NewReviewHandler creates a new instance of ReviewHandler.
func NewReviewHandler(reviewService ReviewServiceInterface, previewService *review_services.PreviewService, skimService *review_services.SkimService, scanService ScanServiceInterface, instrLogger *instrumentation.ServiceInstrumentationLogger) *ReviewHandler {
	return &ReviewHandler{
		reviewService:  reviewService,
		previewService: previewService,
		skimService:    skimService,
		scanService:    scanService,
		instrLogger:    instrLogger,
	}
}

// GetScanAnalysis handles Scan Mode requests
func (h *ReviewHandler) GetScanAnalysis(c *gin.Context) {
	// Check and handle errors for ParseInt
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
		h.instrLogger.LogValidationFailure(c.Request.Context(), "invalid_review_id", "review ID must be a valid integer", nil)
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
		//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
		h.instrLogger.LogError(c.Request.Context(), "retrieve_review_failed", "failed to retrieve review from database", map[string]interface{}{
			"review_id": id,
			"error":     err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve review"})
		return
	}

	// Check and handle errors for AnalyzeScan
	output, err := h.scanService.AnalyzeScan(c.Request.Context(), review.ID, query)
	if err != nil {
		//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
		h.instrLogger.LogError(c.Request.Context(), "scan_analysis_failed", "failed to perform scan analysis", map[string]interface{}{
			"review_id": review.ID,
			"query":     query,
			"error":     err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze scan"})
		return
	}

	// Log successful scan completion
	//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
	h.instrLogger.LogEvent(c.Request.Context(), "scan_analysis_completed", map[string]interface{}{
		"review_id":    review.ID,
		"query":        query,
		"reading_mode": readingMode,
	})

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
		//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
		h.instrLogger.LogValidationFailure(c.Request.Context(), "invalid_request_format", "request body could not be parsed as JSON", nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Validate all inputs
	if !ValidateRequest(c, func() error {
		return ValidateCreateReviewRequest(req.Title, req.CodeSource, req.PastedCode, req.GithubRepo, req.GithubBranch)
	}) {
		return
	}

	review := &review_db.Review{
		UserID:       req.UserID,
		Title:        req.Title,
		CodeSource:   req.CodeSource,
		GithubRepo:   req.GithubRepo,
		GithubBranch: req.GithubBranch,
		PastedCode:   req.PastedCode,
	}
	created, err := h.reviewService.CreateReview(c.Request.Context(), review)
	if err != nil {
		//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
		h.instrLogger.LogError(c.Request.Context(), "create_review_failed", "failed to create review session", map[string]interface{}{
			"title":       req.Title,
			"code_source": req.CodeSource,
			"error":       err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review session"})
		return
	}

	// Log successful session creation
	//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
	h.instrLogger.LogEvent(c.Request.Context(), "review_session_created", map[string]interface{}{
		"review_id":   created.ID,
		"user_id":     req.UserID,
		"title":       req.Title,
		"code_source": req.CodeSource,
		"code_size":   len(req.PastedCode),
	})

	c.JSON(http.StatusCreated, created)
}

// ReviewServiceInterface defines the contract for review-related review_services.
type ReviewServiceInterface interface {
	// GetReview retrieves a review by its ID.
	GetReview(ctx context.Context, id int64) (*review_models.Review, error)
	// CreateReview creates a new review.
	CreateReview(ctx context.Context, review *review_db.Review) (*review_db.Review, error)
}

// ScanServiceInterface defines the contract for scan-related review_services.
type ScanServiceInterface interface {
	// AnalyzeScan analyzes the scan for a given review.
	AnalyzeScan(ctx context.Context, reviewID int64, query string) (*review_models.ScanModeOutput, error)
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

	// TODO: Get code from request body or session. For now, pass empty code.
	output, err := h.skimService.AnalyzeSkim(c.Request.Context(), review.ID, "")
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

// HandlePreviewMode handles POST /api/review/sessions/modes/preview (HTMX)
func (h *ReviewHandler) HandlePreviewMode(c *gin.Context) {
	var req struct {
		Code string `form:"pasted_code" json:"code"`
	}

	// Try form binding first (HTMX), then JSON
	if err := c.ShouldBindForm(&req); err != nil {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.String(http.StatusBadRequest, "Code required")
			return
		}
	}

	if req.Code == "" {
		c.String(http.StatusBadRequest, "Code required")
		return
	}

	// Call preview service
	result, err := h.previewService.AnalyzePreview(c.Request.Context(), req.Code)
	if err != nil {
		c.String(http.StatusInternalServerError, "Preview analysis failed")
		return
	}

	// Render Templ component as HTML response
	component := templates.PreviewModeHtmxResponse(templates.PreviewResult{
		FileTree:            result.FileTree,
		BoundedContexts:     result.BoundedContexts,
		TechStack:           result.TechStack,
		ArchitecturePattern: result.ArchitecturePattern,
		EntryPoints:         result.EntryPoints,
		ExternalDependencies: result.ExternalDependencies,
		Summary:             result.Summary,
	})
	component.Render(c.Request.Context(), c.Writer)
}
