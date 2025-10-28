// Package services provides the implementation of the WebSocket hub for real-time log streaming.
package services

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// WebSocketHub manages WebSocket clients and broadcasts log entries to them.
// It handles client registration, unregistration, filtering, and heartbeat management.
// nolint:govet // Field order optimized for performance, not memory alignment
type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan *models.LogEntry
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client represents a WebSocket client connected to the hub.
// Each client has its own Send channel, filter settings, and authentication state.
// nolint:govet // Field order optimized for readability
type Client struct {
	Conn         *websocket.Conn
	Send         chan *models.LogEntry
	Filters      map[string]string
	IsAuth       bool
	IsPublic     bool
	LastActivity time.Time
	mu           sync.Mutex
}

// NewWebSocketHub creates and returns a new WebSocketHub instance.
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *models.LogEntry, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the WebSocketHub, handling client registration, unregistration, and broadcasting logs.
// This method blocks and should be run in a separate goroutine.
// It processes four types of events:
//   - client registration: adds client to active clients map
//   - client unregistration: removes client and closes its Send channel
//   - broadcast: routes log entry to matching clients
//   - heartbeat tick: sends ping to clients, disconnects inactive ones
func (h *WebSocketHub) Run() {
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered, total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered, remaining clients: %d", len(h.clients))

		case log := <-h.broadcast:
			h.broadcastToClients(log)

		case <-heartbeatTicker.C:
			h.sendHeartbeats()
		}
	}
}

// broadcastToClients sends a log entry to all clients that match the log's visibility and filters.
func (h *WebSocketHub) broadcastToClients(log *models.LogEntry) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		// Check visibility: authenticated users see all, unauthenticated see only public
		if !client.IsAuth && !h.isPublicLog(log) {
			continue
		}

		// Check filters
		if !h.matchesFilters(client, log) {
			continue
		}

		// Attempt to send with backpressure handling
		select {
		case client.Send <- log:
			client.mu.Lock()
			client.LastActivity = time.Now()
			client.mu.Unlock()
		default:
			// Backpressure: client buffer full, close connection
			h.mu.RUnlock()
			go h.closeClient(client)
			h.mu.RLock()
		}
	}
}

// sendHeartbeats sends ping messages to all clients and disconnects inactive ones.
// Inactivity is determined by lack of activity for 60 seconds.
func (h *WebSocketHub) sendHeartbeats() {
	h.mu.RLock()
	now := time.Now()
	deadlineTime := now.Add(-60 * time.Second) // Disconnect if no activity for 60s

	for client := range h.clients {
		client.mu.Lock()
		lastActivity := client.LastActivity
		client.mu.Unlock()

		if lastActivity.Before(deadlineTime) {
			// No pong response, close connection
			h.mu.RUnlock()
			go func(c *Client) {
				if err := c.Conn.Close(); err != nil {
					log.Printf("Error closing connection: %v", err)
				}
				h.closeClient(c)
			}(client)
			h.mu.RLock()
		} else {
			// Send heartbeat ping
			if err := client.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("Error writing ping message: %v", err)
			}
		}
	}
	h.mu.RUnlock()
}

// closeClient removes a client from the hub and performs cleanup.
func (h *WebSocketHub) closeClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
	}
	h.mu.Unlock()
}

// matchesFilters checks if a log entry matches all filters set by a client.
// Returns true only if the log matches ALL active filters (AND logic).
func (h *WebSocketHub) matchesFilters(client *Client, log *models.LogEntry) bool {
	// Check level filter
	if level, ok := client.Filters["level"]; ok && level != log.Level {
		return false
	}

	// Check service filter
	if service, ok := client.Filters["service"]; ok && service != log.Service {
		return false
	}

	// Check tags filter
	if tagFilter, ok := client.Filters["tags"]; ok {
		if !h.logHasTag(log, tagFilter) {
			return false
		}
	}

	return true
}

// logHasTag checks if a log entry contains a specific tag.
func (h *WebSocketHub) logHasTag(log *models.LogEntry, tag string) bool {
	if log.Tags == nil {
		return false
	}
	for _, t := range log.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// isPublicLog checks if a log entry should be visible to unauthenticated users.
// By default, only INFO logs are public. For tests, all levels can be public if LOGS_WEBSOCKET_PUBLIC_ALL is set.
func (h *WebSocketHub) isPublicLog(log *models.LogEntry) bool {
	if os.Getenv("LOGS_WEBSOCKET_PUBLIC_ALL") == "1" {
		return true
	}
	return log.Level == "INFO"
}

// Register adds a client to the hub for broadcasting.
func (h *WebSocketHub) Register(client *Client) {
	h.register <- client
}

// WritePump sends messages from the client's Send channel to the WebSocket connection.
// It runs in its own goroutine for each client and closes when the connection is lost.
func (c *Client) WritePump(hub *WebSocketHub) {
	defer func() {
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}()

	for log := range c.Send {
		if err := c.Conn.WriteJSON(log); err != nil {
			break
		}
		c.mu.Lock()
		c.LastActivity = time.Now()
		c.mu.Unlock()
	}

	// Unregister when write pump exits
	hub.unregister <- c
}

// ReadPump reads messages from the WebSocket connection.
// It runs in its own goroutine for each client and handles the pong handler
// for heartbeat responses. It closes when the connection is lost.
func (c *Client) ReadPump(hub *WebSocketHub) {
	defer func() {
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}()

	if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		log.Printf("Error setting read deadline: %v", err)
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.LastActivity = time.Now()
		c.mu.Unlock()
		if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("Error setting pong deadline: %v", err)
		}
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		c.mu.Lock()
		c.LastActivity = time.Now()
		c.mu.Unlock()
		if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("Error setting read deadline: %v", err)
			break
		}
	}

	// Unregister when read pump exits
	hub.unregister <- c
}
