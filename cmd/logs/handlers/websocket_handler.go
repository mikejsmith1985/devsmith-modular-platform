// Package handlers provides HTTP and WebSocket handlers for the logs service.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Restrict in production
	},
}

// WebSocketHandler handles WebSocket connections for the Logs Service.
type WebSocketHandler struct {
	hub *services.WebSocketHub
}

// NewWebSocketHandler creates a new WebSocketHandler instance.
func NewWebSocketHandler(hub *services.WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// HandleWebSocket upgrades HTTP connections to WebSocket and registers clients.
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	client := &services.Client{
		Conn:    conn, // Ensure the connection is passed to the client
		Send:    make(chan *models.LogEntry, 256),
		Filters: make(map[string]string),
	}

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
