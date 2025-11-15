// Package logs_services provides the implementation of the WebSocket hub for real-time log streaming.
package logs_services

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// WebSocketHub manages WebSocket clients and broadcasts log entries to them.
// It handles client registration, unregistration, filtering, and heartbeat management.
// nolint:govet // Field order optimized for performance, not memory alignment
type WebSocketHub struct {
	clients         map[*Client]bool
	broadcast       chan *logs_models.LogEntry
	analysisResults chan *AnalysisNotification // Phase 1: AI analysis notifications
	register        chan *Client
	unregister      chan *Client
	stop            chan struct{}
	mu              sync.RWMutex
}

// AnalysisNotification represents an AI analysis result broadcast to clients
type AnalysisNotification struct {
	Type      string          `json:"type"` // "new_issue"
	LogID     int64           `json:"log_id"`
	IssueType string          `json:"issue_type"`
	Analysis  *AnalysisResult `json:"analysis"`
	Timestamp time.Time       `json:"timestamp"`
}

// Client represents a WebSocket client connected to the hub.
// Each client has its own Send channel, filter settings, and authentication state.
// nolint:govet // Field order optimized for readability
type Client struct {
	LastActivity time.Time
	Conn         *websocket.Conn
	Send         chan *logs_models.LogEntry
	Filters      map[string]string
	// Registered channel is closed by the hub after the client
	// has been added to the active clients map. Tests wait on this
	// to ensure registration is complete before sending messages.
	Registered chan struct{}
	// done channel signals WritePump to exit gracefully
	done chan struct{}
	mu   sync.Mutex
	// writeMu serializes concurrent writes to the websocket connection
	writeMu  sync.Mutex
	IsAuth   bool
	IsPublic bool
	// wg tracks goroutines (ReadPump, WritePump) for clean shutdown
	wg sync.WaitGroup
}

// NewWebSocketHub creates and returns a new WebSocketHub instance.
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:         make(map[*Client]bool),
		broadcast:       make(chan *logs_models.LogEntry, 256),
		analysisResults: make(chan *AnalysisNotification, 128),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		stop:            make(chan struct{}),
	}
}

// Run starts the WebSocketHub, handling client registration, unregistration, and broadcasting logs.
// This method blocks and should be run in a separate goroutine.
// It processes four types of events:
//   - client registration: adds client to active clients map
//   - client unregistration: removes client and closes its Send channel
//   - broadcast: routes log entry to matching clients
//   - heartbeat tick: sends ping to clients, disconnects inactive ones
//   - stop: signal to shut down the hub gracefully
func (h *WebSocketHub) Run() {
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-h.stop:
			// Graceful shutdown: close all client connections and exit
			h.mu.Lock()
			for client := range h.clients {
				close(client.Send)
			}
			h.clients = make(map[*Client]bool)
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			// Initialize last activity to now to avoid immediate heartbeat-triggered disconnect
			client.mu.Lock()
			client.LastActivity = time.Now()
			client.mu.Unlock()
			h.mu.Unlock()
			// Signal back to the client that registration is complete
			if client.Registered != nil {
				// close in a goroutine to avoid blocking hub loop if receiver is not waiting
				go func(ch chan struct{}) {
					close(ch)
				}(client.Registered)
			}
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

		case analysisNotif := <-h.analysisResults:
			// Phase 1: Broadcast AI analysis results
			h.broadcastAnalysisNotification(analysisNotif)

		case <-heartbeatTicker.C:
			h.sendHeartbeats()
		}
	}
}

// Stop signals the WebSocketHub to shut down gracefully.
// It closes all client connections and stops the hub goroutine.
// Safe to call multiple times (using recover to catch panic from closing closed channel).
func (h *WebSocketHub) Stop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocketHub already stopped") // Explicitly log instead of empty branch
		}
	}()
	close(h.stop)
}

// broadcastToClients sends a log entry to all clients that match the log's visibility and filters.
func (h *WebSocketHub) broadcastToClients(log *logs_models.LogEntry) {
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

// broadcastAnalysisNotification broadcasts an AI analysis result to all connected clients
// Phase 1: This sends "new_issue" notifications when AI analysis completes
func (h *WebSocketHub) broadcastAnalysisNotification(notif *AnalysisNotification) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	log.Printf("Broadcasting analysis notification: type=%s, issue_type=%s, log_id=%d",
		notif.Type, notif.IssueType, notif.LogID)

	// For analysis notifications, send to all authenticated clients
	// (Public clients typically don't see detailed diagnostic info)
	for client := range h.clients {
		if !client.IsAuth {
			continue
		}

		// Create a LogEntry wrapper for the notification to send through existing channel
		// This allows us to reuse the Send channel infrastructure
		notifLog := &logs_models.LogEntry{
			ID:        notif.LogID,
			Level:     "info",
			Service:   "ai-analyzer",
			Message:   notif.Type, // "new_issue"
			IssueType: notif.IssueType,
			CreatedAt: notif.Timestamp,
			// Store the full analysis in Metadata for client-side parsing
			Metadata: encodeAnalysisToJSON(notif.Analysis),
		}

		// Attempt to send with backpressure handling
		select {
		case client.Send <- notifLog:
			client.mu.Lock()
			client.LastActivity = time.Now()
			client.mu.Unlock()
		default:
			// Backpressure: client buffer full, skip this notification
			log.Printf("Skipped analysis notification for client (buffer full)")
		}
	}
}

// encodeAnalysisToJSON converts AnalysisResult to JSON bytes
func encodeAnalysisToJSON(analysis *AnalysisResult) []byte {
	if analysis == nil {
		return []byte("{}")
	}
	data, err := json.Marshal(analysis)
	if err != nil {
		log.Printf("Failed to encode analysis to JSON: %v", err)
		return []byte("{}")
	}
	return data
}

// BroadcastAnalysisResult sends an AI analysis notification to all connected clients
// This is called when AI analysis completes for a log entry
func (h *WebSocketHub) BroadcastAnalysisResult(logID int64, issueType string, analysis *AnalysisResult) {
	notif := &AnalysisNotification{
		Type:      "new_issue",
		LogID:     logID,
		IssueType: issueType,
		Analysis:  analysis,
		Timestamp: time.Now(),
	}

	// Non-blocking send to avoid blocking the caller
	select {
	case h.analysisResults <- notif:
		// Successfully queued
	default:
		// Channel full, log warning but don't block
		log.Printf("Warning: Analysis notification channel full, dropping notification for log %d", logID)
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
			h.closeInactiveClient(client)
			h.mu.RLock()
		} else {
			// Send heartbeat
			h.sendHeartbeat(client)
		}
	}
	h.mu.RUnlock()
}

// closeInactiveClient closes an inactive client connection
func (h *WebSocketHub) closeInactiveClient(client *Client) {
	go func(c *Client) {
		// Close the connection to force Read/Write pumps to exit
		c.writeMu.Lock()
		if err := c.Conn.Close(); err != nil {
			log.Printf("error closing inactive client connection: %v", err)
		}
		c.writeMu.Unlock()
		
		// Use unregister channel to ensure thread-safe removal from hub
		// This prevents race condition with hub.Run() goroutine
		select {
		case h.unregister <- c:
		default:
			// If unregister channel is full, hub is shutting down
		}
	}(client)
}

// sendHeartbeat sends ping and text heartbeat messages to a client
func (h *WebSocketHub) sendHeartbeat(client *Client) {
	// Send heartbeat ping + text marker. Ping triggers pong handling
	// on clients; text message ensures tests using ReadMessage see a payload.
	client.writeMu.Lock()
	if err := client.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		log.Printf("Error writing ping message: %v", err)
	}
	if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("heartbeat")); err != nil {
		log.Printf("Error writing heartbeat message: %v", err)
	}
	client.writeMu.Unlock()
}

// closeClient removes a client from the hub and performs cleanup.
func (h *WebSocketHub) closeClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
		// Best-effort close of the underlying connection to speed teardown
		if client.Conn != nil {
			client.writeMu.Lock()
			if err := client.Conn.Close(); err != nil {
				log.Printf("error closing client connection during cleanup: %v", err)
			}
			client.writeMu.Unlock()
		}
	}
	h.mu.Unlock()
}

// matchesFilters checks if a log entry matches all filters set by a client.
// Returns true only if the log matches ALL active filters (AND logic).
func (h *WebSocketHub) matchesFilters(client *Client, log *logs_models.LogEntry) bool {
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
func (h *WebSocketHub) logHasTag(log *logs_models.LogEntry, tag string) bool {
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
func (h *WebSocketHub) isPublicLog(log *logs_models.LogEntry) bool {
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
	c.wg.Add(1)
	defer c.wg.Done()
	defer func() {
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
		// Unregister when write pump exits (non-blocking to avoid deadlock during shutdown)
		select {
		case hub.unregister <- c:
			// Successfully unregistered
		default:
			// Hub stopped, registration no longer possible (this is OK during test cleanup)
		}
	}()

	for {
		select {
		case log, ok := <-c.Send:
			if !ok {
				// Send channel closed, exit gracefully
				return
			}
			// Serialize writes to avoid concurrent WriteMessage/WriteJSON calls
			c.writeMu.Lock()
			if err := c.Conn.WriteJSON(log); err != nil {
				c.writeMu.Unlock()
				return
			}
			c.writeMu.Unlock()
			c.mu.Lock()
			c.LastActivity = time.Now()
			c.mu.Unlock()
		case <-c.done:
			// Client disconnected, exit gracefully
			return
		}
	}
}

// ReadPump reads messages from the WebSocket connection.
// It runs in its own goroutine for each client and handles the pong handler
// for heartbeat responses. It closes when the connection is lost.
func (c *Client) ReadPump(hub *WebSocketHub) {
	c.wg.Add(1)
	defer c.wg.Done()
	defer func() {
		// Close done channel to signal WritePump to exit
		close(c.done)
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
		// Unregister when read pump exits (non-blocking to avoid deadlock during shutdown)
		select {
		case hub.unregister <- c:
			// Successfully unregistered
		default:
			// Hub stopped, registration no longer possible (this is OK during test cleanup)
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
}
