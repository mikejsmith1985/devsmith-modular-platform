// Package services provides the implementation of the WebSocket hub for real-time log streaming.
package services

import (
	"sync"

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
	send    chan *models.LogEntry
	filters map[string]string
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
				close(client.send)
			}
			h.mu.Unlock()

		case log := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if h.matchesFilters(client, log) {
					select {
					case client.send <- log:
					default:
						close(client.send)
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
	if service, ok := client.filters["service"]; ok && service != log.Service {
		return false
	}
	if level, ok := client.filters["level"]; ok && level != log.Level {
		return false
	}
	return true
}
