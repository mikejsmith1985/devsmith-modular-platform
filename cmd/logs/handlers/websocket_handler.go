// Package handlers provides HTTP handlers for the Logs service.
//
// nolint:govet // fieldalignment: struct layout optimized for semantic clarity and concurrency safety
package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// WebSocketConnection wraps a gorilla websocket connection and implements ClientConn.
type WebSocketConnection struct {
	conn             *websocket.Conn        // 8 bytes
	send             chan interface{}       // 8 bytes
	hub              *services.Hub          // 8 bytes
	ctx              context.Context        // 16 bytes (interface)
	cancel           context.CancelFunc     // 8 bytes
	filters          map[string]interface{} // 8 bytes
	heartbeatTimeout time.Duration          // 8 bytes
	maxMessageSize   int64                  // 8 bytes
	userID           int64                  // 8 bytes
	lastPong         time.Time              // 24 bytes
	userRole         string                 // 16 bytes
	isAuthenticated  bool                   // 1 byte (padded to 8)
}

// WebSocketHandler handles WebSocket connections for real-time log streaming.
type WebSocketHandler struct {
	hub            *services.Hub
	upgrader       websocket.Upgrader
	idleTimeout    time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	maxMessageSize int64
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(hub *services.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		idleTimeout:    30 * time.Second,
		readTimeout:    10 * time.Second,
		writeTimeout:   10 * time.Second,
		maxMessageSize: 65536, // 64KB
	}
}

// HandleWebSocket handles WebSocket upgrade and connection lifecycle.
// Endpoint: GET /ws/logs?level=ERROR&service=portal
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create context for this connection
	ctx, cancel := context.WithCancel(context.Background())

	// Parse authentication (from header or URL param)
	userID, userRole, isAuth := h.parseAuth(c)

	// Parse filter parameters from URL query
	filters := h.parseFilters(c)

	// Create WebSocket connection wrapper
	wsConn := &WebSocketConnection{
		conn:             conn,
		send:             make(chan interface{}, 256),
		filters:          filters,
		hub:              h.hub,
		ctx:              ctx,
		cancel:           cancel,
		heartbeatTimeout: 90 * time.Second,
		lastPong:         time.Now(),
		isAuthenticated:  isAuth,
		userID:           userID,
		userRole:         userRole,
		maxMessageSize:   h.maxMessageSize,
	}

	// Configure timeouts
	if err := conn.SetReadDeadline(time.Now().Add(h.readTimeout)); err != nil {
		log.Printf("failed to set read deadline: %v", err)
		if err := conn.Close(); err != nil {
			log.Printf("failed to close conn: %v", err)
		}
		return
	}
	if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
		log.Printf("failed to set write deadline: %v", err)
		if err := conn.Close(); err != nil {
			log.Printf("failed to close conn: %v", err)
		}
		return
	}

	conn.SetPongHandler(func(string) error {
		wsConn.lastPong = time.Now()
		if err := conn.SetReadDeadline(time.Now().Add(h.idleTimeout)); err != nil {
			log.Printf("failed to set read deadline in pong handler: %v", err)
		}
		return nil
	})

	// Register with hub
	h.hub.Register(wsConn)

	// Start read/write pumps
	go wsConn.readPump()
	go wsConn.writePump()
}

// Send sends a message to the WebSocket connection.
func (wc *WebSocketConnection) Send(msg interface{}) error {
	select {
	case wc.send <- msg:
		return nil
	case <-wc.ctx.Done():
		return wc.ctx.Err()
	default:
		return services.ErrBackpressure
	}
}

// UpdateFilters updates the connection's message filters.
func (wc *WebSocketConnection) UpdateFilters(filters map[string]interface{}) {
	wc.filters = filters
}

// GetFilters returns the connection's current filters.
func (wc *WebSocketConnection) GetFilters() map[string]interface{} {
	return wc.filters
}

// Close closes the WebSocket connection.
func (wc *WebSocketConnection) Close() error {
	wc.cancel()
	return wc.conn.Close()
}

// Context returns the connection's context.
func (wc *WebSocketConnection) Context() context.Context {
	return wc.ctx
}

// readPump reads messages from the WebSocket connection.
// It handles filter updates and control messages.
func (wc *WebSocketConnection) readPump() {
	defer func() {
		wc.hub.Unregister(wc)
		if err := wc.conn.Close(); err != nil {
			log.Printf("error closing connection in readPump: %v", err)
		}
	}()

	for {
		var message map[string]interface{}
		err := wc.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			return
		}

		wc.handleIncomingMessage(message)
	}
}

// handleIncomingMessage processes incoming WebSocket messages.
func (wc *WebSocketConnection) handleIncomingMessage(message map[string]interface{}) {
	msgType, ok := message["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "filters":
		if filters, ok := message["filters"].(map[string]interface{}); ok {
			wc.UpdateFilters(filters)
		}

	case "ping":
		wc.send <- map[string]interface{}{"type": "pong", "timestamp": time.Now().Unix()}

	case "auth":
		wc.handleAuthMessage(message)
	}
}

// handleAuthMessage handles authentication messages.
func (wc *WebSocketConnection) handleAuthMessage(message map[string]interface{}) {
	if wc.isAuthenticated {
		return
	}

	token, ok := message["token"].(string)
	if !ok {
		return
	}

	userID, role, valid := validateToken(token)
	if valid {
		wc.isAuthenticated = true
		wc.userID = userID
		wc.userRole = role
		wc.send <- map[string]interface{}{"type": "auth_success", "user_id": userID, "role": role}
	} else {
		wc.send <- map[string]interface{}{"type": "auth_failed", "error": "Invalid token"}
	}
}

// writePump sends messages to the WebSocket connection.
// It handles heartbeats and checks for connection timeout.
func (wc *WebSocketConnection) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		if err := wc.conn.Close(); err != nil {
			log.Printf("error closing connection in writePump: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-wc.send:
			if err := wc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				log.Printf("failed to set write deadline: %v", err)
				return
			}

			if !ok {
				// Channel closed
				if err := wc.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("failed to write close message: %v", err)
				}
				return
			}

			// Write JSON message
			if err := wc.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			wc.handleHeartbeat()

		case <-wc.ctx.Done():
			return
		}
	}
}

// handleHeartbeat handles heartbeat timeout and sending.
func (wc *WebSocketConnection) handleHeartbeat() {
	if err := wc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Printf("failed to set write deadline: %v", err)
		return
	}

	// Check heartbeat timeout
	if time.Since(wc.lastPong) > wc.heartbeatTimeout {
		log.Printf("heartbeat timeout for user %d", wc.userID)
		if err := wc.conn.WriteMessage(websocket.CloseMessage, []byte("heartbeat timeout")); err != nil {
			log.Printf("failed to write timeout message: %v", err)
		}
		return
	}

	// Send heartbeat
	if err := wc.conn.WriteJSON(map[string]interface{}{
		"type":      "heartbeat",
		"timestamp": time.Now().Unix(),
	}); err != nil {
		log.Printf("failed to write heartbeat: %v", err)
	}
}

// parseAuth extracts authentication info from request.
// Returns userID, userRole, and isAuthenticated boolean.
func (h *WebSocketHandler) parseAuth(c *gin.Context) (userID int64, userRole string, isAuth bool) {
	// Try to get auth from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		id, role, valid := validateToken(authHeader)
		if valid {
			return id, role, true
		}
	}

	// Try to get from query parameter (less secure, fallback only)
	token := c.Query("token")
	if token != "" {
		id, role, valid := validateToken(token)
		if valid {
			return id, role, true
		}
	}

	// Not authenticated
	return 0, "", false
}

// parseFilters extracts filter parameters from URL query.
func (h *WebSocketHandler) parseFilters(c *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})

	// Service filter
	if service := c.Query("service"); service != "" {
		filters["service"] = service
	}

	// Level filter
	if level := c.Query("level"); level != "" {
		filters["level"] = level
	}

	// Custom filters can be added here
	return filters
}

// validateToken validates an authentication token.
// Returns userID, userRole, and isValid boolean.
// This is a placeholder - implement proper JWT validation in production.
func validateToken(token string) (userID int64, userRole string, isValid bool) {
	// TODO: Implement proper JWT/auth validation
	// For now, return false to indicate token validation needed
	if token == "" {
		return 0, "", false
	}

	// Placeholder: Parse token and extract user info
	// In production, use JWT library to validate and extract claims
	id := int64(0)
	if field := parseTokenField(token, "user_id"); field != "" {
		if parsed, err := strconv.ParseInt(field, 10, 64); err == nil {
			id = parsed
		}
	}

	role := "viewer" // Default role
	if r := parseTokenField(token, "role"); r != "" {
		role = r
	}

	return id, role, id > 0
}

// parseTokenField extracts a field value from a token string (placeholder).
func parseTokenField(token, field string) string {
	// Placeholder implementation
	// In production, properly decode and validate JWT
	return ""
}

// WebSocketStats returns current hub statistics endpoint.
// Endpoint: GET /ws/logs/stats
func (h *WebSocketHandler) WebSocketStats(c *gin.Context) {
	stats := h.hub.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   stats,
	})
}

// SetBackpressureStrategy updates the hub's backpressure strategy.
// Endpoint: POST /ws/logs/config
// Body: {"backpressure_strategy": "drop|queue"}
func (h *WebSocketHandler) SetBackpressureStrategy(c *gin.Context) {
	var req struct {
		BackpressureStrategy string `json:"backpressure_strategy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.hub.SetBackpressureStrategy(req.BackpressureStrategy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid strategy: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated", "strategy": req.BackpressureStrategy})
}
