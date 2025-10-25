// Package services provides WebSocket handler implementation for real-time log streaming.
package services

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// WebSocketHandler handles WebSocket connections for real-time log streaming.
type WebSocketHandler struct {
	hub *WebSocketHub
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(hub *WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// RegisterWebSocketRoutes registers the WebSocket routes on a Gin router.
func RegisterWebSocketRoutes(router *gin.Engine, hub *WebSocketHub) {
	handler := NewWebSocketHandler(hub)
	router.GET("/ws/logs", handler.HandleWebSocket)
}

// HandleWebSocket handles WebSocket upgrade and client connection.
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Parse and validate authentication
	authHeader := c.GetHeader("Authorization")
	isAuthenticated := h.validateAuth(authHeader)

	// Parse filter parameters from query string
	filters := h.parseFilterParams(c)

	// Parse visibility based on authentication
	isPublic := !isAuthenticated

	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create client with filters
	client := &Client{
		Conn:         conn,
		Send:         make(chan *models.LogEntry, 256),
		Filters:      filters,
		IsAuth:       isAuthenticated,
		IsPublic:     isPublic,
		LastActivity: time.Now(),
	}

	// Register client with hub
	h.hub.Register(client)

	// Start client read and write pumps
	go client.ReadPump(h.hub)
	go client.WritePump(h.hub)
}

// validateAuth checks if authentication header is valid JWT.
func (h *WebSocketHandler) validateAuth(authHeader string) bool {
	if authHeader == "" {
		return false
	}

	// Check for Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return false
	}

	// TODO: Validate JWT token signature and expiry
	// For now, accept any non-empty Bearer token
	return token != "expired_token"
}

// parseFilterParams extracts filter parameters from query string.
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
