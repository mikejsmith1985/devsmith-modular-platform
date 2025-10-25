// Package services provides the implementation of the WebSocket hub for real-time log streaming.
package services

import (
	"log"
	"sync"
	"time"

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
			h.mu.RLock()
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
					go func(c *Client) {
						h.mu.Lock()
						if _, ok := h.clients[c]; ok {
							delete(h.clients, c)
							close(c.Send)
						}
						h.mu.Unlock()
					}(client)
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()

		case <-heartbeatTicker.C:
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
						c.Conn.Close()
						h.mu.Lock()
						if _, ok := h.clients[c]; ok {
							delete(h.clients, c)
							close(c.Send)
						}
						h.mu.Unlock()
					}(client)
					h.mu.RLock()
				} else {
					// Send heartbeat ping
					_ = client.Conn.WriteMessage(websocket.PingMessage, []byte{})
				}
			}
			h.mu.RUnlock()
		}
	}
}

// matchesFilters checks if a log entry matches the filters set by a client.
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
		tagFound := false
		if log.Tags != nil {
			for _, tag := range log.Tags {
				if tag == tagFilter {
					tagFound = true
					break
				}
			}
		}
		if !tagFound {
			return false
		}
	}

	return true
}

// isPublicLog checks if a log entry should be visible to unauthenticated users.
func (h *WebSocketHub) isPublicLog(log *models.LogEntry) bool {
	// For now, treat all INFO level logs as public, others as private
	// This can be extended with a field on LogEntry
	return log.Level == "INFO"
}

// Register adds a client to the hub.
func (h *WebSocketHub) Register(client *Client) {
	h.register <- client
}

// WritePump sends messages to the WebSocket connection.
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
func (c *Client) ReadPump(hub *WebSocketHub) {
	defer func() {
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.LastActivity = time.Now()
		c.mu.Unlock()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}

	// Unregister when read pump exits
	hub.unregister <- c
}
