// Package services provides the implementation of the WebSocket hub for real-time log streaming.
package services

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// WebSocketHub manages WebSocket clients and broadcasts log entries to them.
type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan *models.LogEntry
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client represents a WebSocket client connected to the hub.
type Client struct {
	Conn    *websocket.Conn
	Send    chan *models.LogEntry
	Filters map[string]string
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
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case log := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if h.matchesFilters(client, log) {
					select {
					case client.Send <- log:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// matchesFilters checks if a log entry matches the filters set by a client.
func (h *WebSocketHub) matchesFilters(client *Client, log *models.LogEntry) bool {
	if service, ok := client.Filters["service"]; ok && service != log.Service {
		return false
	}
	if level, ok := client.Filters["level"]; ok && level != log.Level {
		return false
	}
	return true
}

// Register adds a client to the hub.
func (h *WebSocketHub) Register(client *Client) {
	h.register <- client
}

// WritePump sends messages to the WebSocket connection.
func (c *Client) WritePump() {
	defer func() {
		if err := c.Conn.Close(); err != nil {
			// Log the error for debugging purposes
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}()
	for log := range c.Send {
		if err := c.Conn.WriteJSON(log); err != nil {
			break
		}
	}
}

// ReadPump reads messages from the WebSocket connection.
func (c *Client) ReadPump() {
	defer func() {
		if err := c.Conn.Close(); err != nil {
			// Log the error for debugging purposes
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}()
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
