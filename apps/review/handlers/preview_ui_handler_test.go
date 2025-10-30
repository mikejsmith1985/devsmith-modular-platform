package review_handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/stretchr/testify/assert"
)

func TestRegisterPreviewUIRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a nil preview service for testing route registration
	var previewService *review_services.PreviewService

	RegisterPreviewUIRoutes(router, previewService)

	// Verify route is registered
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/review/preview" && route.Method == "GET" {
			found = true
			break
		}
	}

	assert.True(t, found, "Preview route should be registered")
}
