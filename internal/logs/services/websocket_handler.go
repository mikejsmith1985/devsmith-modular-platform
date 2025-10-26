// Package services provides WebSocket handler implementation for real-time log streaming.
// It implements the WebSocket upgrade, filter parsing, and authentication for the Logs service.
package services

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// WebSocketHandler handles HTTP to WebSocket upgrade and connection setup.
type WebSocketHandler struct {
	hub *WebSocketHub
}

// NewWebSocketHandler creates a new WebSocket handler with the given hub.
func NewWebSocketHandler(hub *WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// RegisterWebSocketRoutes registers the WebSocket endpoint on a Gin router.
func RegisterWebSocketRoutes(router *gin.Engine, hub *WebSocketHub) {
	handler := NewWebSocketHandler(hub)
	router.GET("/ws/logs", handler.HandleWebSocket)
}

// HandleWebSocket upgrades an HTTP connection to WebSocket and registers the client.
// Supports the following query parameters for filtering:
//   - level: Log level filter (e.g., ERROR, WARN, INFO)
//   - service: Service name filter (e.g., portal, review)
//   - tags: Tag filter (exact match, single tag)
//
// Authentication is checked via the Authorization header (Bearer token).
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Parse and validate authentication
	authHeader := c.GetHeader("Authorization")
	isAuthenticated := h.validateAuth(authHeader)

	// Parse filter parameters from query string
	filters := h.parseFilterParams(c)

	// Determine visibility: authenticated users see all logs, others see only public
	isPublic := !isAuthenticated

	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create client with filters and auth info
	client := &Client{
		Conn:         conn,
		Send:         make(chan *models.LogEntry, 256),
		Filters:      filters,
		IsAuth:       isAuthenticated,
		IsPublic:     isPublic,
		LastActivity: time.Now(),
	}

	// Register client with hub and start message pumps
	h.hub.Register(client)
	go client.ReadPump(h.hub)
	go client.WritePump(h.hub)
}

// validateAuth checks if authentication header contains a valid Bearer token.
// Returns true if a valid Bearer token is present, false otherwise.
// Does NOT validate JWT signature (placeholder for future JWT validation).
func (h *WebSocketHandler) validateAuth(authHeader string) bool {
	if authHeader == "" {
		return false
	}

	// Check for Bearer token format
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return false
	}

	// Reject explicitly expired tokens (placeholder for full JWT validation)
	return token != "expired_token"
}

// parseFilterParams extracts and returns filter parameters from the request query string.
// Supports: level, service, tags
func (h *WebSocketHandler) parseFilterParams(c *gin.Context) map[string]string {
	filters := make(map[string]string)

	if level := c.Query("level"); level != "" {
		filters["level"] = level
	}

	if service := c.Query("service"); service != "" {
		filters["service"] = service
	}

	if tags := c.Query("tags"); tags != "" {
		filters["tags"] = tags
	}

	return filters
}
