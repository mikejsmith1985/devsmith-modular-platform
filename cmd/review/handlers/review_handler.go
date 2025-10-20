package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// CreateReviewSession handles POST /api/review/sessions
func (h *ReviewHandler) CreateReviewSession(c *gin.Context) {
	var req struct {
		UserID       int64  `json:"user_id"`
		Title        string `json:"title"`
		CodeSource   string `json:"code_source"`
		GithubRepo   string `json:"github_repo"`
		GithubBranch string `json:"github_branch"`
		PastedCode   string `json:"pasted_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

type ReviewHandler struct {
	reviewService  *services.ReviewService
	previewService *services.PreviewService
	skimService    *services.SkimService
}

func NewReviewHandler(reviewService *services.ReviewService, previewService *services.PreviewService, skimService *services.SkimService) *ReviewHandler {
	return &ReviewHandler{
		reviewService:  reviewService,
		previewService: previewService,
		skimService:    skimService,
	}
}

// GetSkimAnalysis handles GET /api/reviews/:id/skim
func (h *ReviewHandler) GetSkimAnalysis(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

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
