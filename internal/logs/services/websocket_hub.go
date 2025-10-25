// Package services provides the implementation of the WebSocket hub for real-time log streaming.
package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

// ErrBackpressure indicates the send channel is full (client is slow).
var ErrBackpressure = errors.New("backpressure: client send channel full")

// Hub manages WebSocket clients and broadcasts log entries to them.
// It implements connection management, filtering, heartbeats, and backpressure handling.
type Hub struct {
	clients              map[ClientConn]bool
	register             chan ClientConn
	unregister           chan ClientConn
	broadcast            chan interface{}
	mu                   sync.RWMutex
	running              bool
	cancel               context.CancelFunc
	heartbeatInterval    time.Duration
	heartbeatTimeout     time.Duration
	backpressureStrategy string // "drop" or "queue"
	maxQueueSize         int
	redisPubSub          RedisPubSubConn
	stats                *HubStats
}

// HubStats tracks hub metrics for monitoring.
type HubStats struct {
	TotalBroadcasts int64
	TotalDropped    int64
	TotalQueued     int64
	ActiveClients   int
	mu              sync.RWMutex
}

// ClientConn represents a WebSocket client connection.
type ClientConn interface {
	Send(msg interface{}) error
	UpdateFilters(filters map[string]interface{})
	GetFilters() map[string]interface{}
	Close() error
	Context() context.Context
}

// RedisPubSubConn represents a Redis pub/sub connection for multi-instance support.
type RedisPubSubConn interface {
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) (<-chan interface{}, error)
	Close() error
}

// NewHub creates and returns a new Hub instance.
func NewHub(redisPubSub RedisPubSubConn) *Hub {
	return &Hub{
		clients:              make(map[ClientConn]bool),
		register:             make(chan ClientConn, 256),
		unregister:           make(chan ClientConn, 256),
		broadcast:            make(chan interface{}, 256),
		heartbeatInterval:    30 * time.Second,
		heartbeatTimeout:     90 * time.Second,
		backpressureStrategy: "drop",
		maxQueueSize:         1000,
		redisPubSub:          redisPubSub,
		stats: &HubStats{
			ActiveClients: 0,
		},
	}
}

// Run starts the Hub's main event loop.
// It handles client registration, unregistration, and message broadcasting.
// This should be called in a goroutine: go hub.Run()
func (h *Hub) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	h.mu.Lock()
	h.running = true
	h.mu.Unlock()

	// Heartbeat ticker
	heartbeatTicker := time.NewTicker(h.heartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)

		case <-heartbeatTicker.C:
			h.sendHeartbeats()

		case <-ctx.Done():
			h.shutdown()
			return
		}
	}
}

// Stop gracefully stops the hub.
func (h *Hub) Stop() {
	h.mu.Lock()
	if h.running && h.cancel != nil {
		h.cancel()
	}
	h.mu.Unlock()
}

// Register adds a client to the hub.
func (h *Hub) Register(client ClientConn) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client ClientConn) {
	h.unregister <- client
}

// Broadcast sends a message to all matching clients.
func (h *Hub) Broadcast(msg interface{}) {
	select {
	case h.broadcast <- msg:
	default:
		// Broadcast channel full, log and drop
		log.Printf("warning: broadcast channel full, dropping message")
	}
}

// ClientCount returns the current number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// MatchesFilters checks if a message matches a client's filters.
// Returns true if the message should be sent to the client.
func (h *Hub) MatchesFilters(filters map[string]interface{}, msg interface{}) bool {
	// If no filters, match all messages
	if len(filters) == 0 {
		return true
	}

	// Try map-based approach first
	if logData, ok := msg.(map[string]interface{}); ok {
		if !h.matchesMapFilters(filters, logData) {
			return false
		}
		return true
	}

	// Try reflection for typed structs with Service and Level fields
	return h.matchesStructFilters(filters, msg)
}

// matchesMapFilters checks filters against a map[string]interface{}.
func (h *Hub) matchesMapFilters(filters map[string]interface{}, logData map[string]interface{}) bool {
	// Check service filter
	if serviceFilter, ok := filters["service"]; ok {
		if service, ok := logData["service"]; ok {
			if service != serviceFilter {
				return false
			}
		}
	}

	// Check level filter
	if levelFilter, ok := filters["level"]; ok {
		if level, ok := logData["level"]; ok {
			if level != levelFilter {
				return false
			}
		}
	}

	return true
}

// matchesStructFilters checks filters against a struct by reflection.
func (h *Hub) matchesStructFilters(filters map[string]interface{}, msg interface{}) bool {
	// Extract service and level using reflection (generic approach)
	v := reflect.ValueOf(msg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Check service filter
	if serviceFilter, ok := filters["service"]; ok {
		serviceField := v.FieldByName("Service")
		if serviceField.IsValid() && serviceField.Kind() == reflect.String {
			if serviceField.String() != serviceFilter.(string) {
				return false
			}
		}
	}

	// Check level filter
	if levelFilter, ok := filters["level"]; ok {
		levelField := v.FieldByName("Level")
		if levelField.IsValid() && levelField.Kind() == reflect.String {
			if levelField.String() != levelFilter.(string) {
				return false
			}
		}
	}

	return true
}

// handleRegister processes client registration.
func (h *Hub) handleRegister(client ClientConn) {
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	h.stats.mu.Lock()
	h.stats.ActiveClients++
	h.stats.mu.Unlock()

	log.Printf("client registered: %d clients active", h.ClientCount())
}

// handleUnregister processes client unregistration.
func (h *Hub) handleUnregister(client ClientConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client]; exists {
		delete(h.clients, client)
		client.Close()

		h.stats.mu.Lock()
		h.stats.ActiveClients--
		h.stats.mu.Unlock()
	}
}

// handleBroadcast sends a message to all matching clients.
func (h *Hub) handleBroadcast(msg interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.stats.mu.Lock()
	h.stats.TotalBroadcasts++
	h.stats.mu.Unlock()

	// Send to all connected clients
	for client := range h.clients {
		// Check if client matches filters
		filters := client.GetFilters()
		if !h.MatchesFilters(filters, msg) {
			continue
		}

		// Send with backpressure handling
		err := client.Send(msg)
		if err != nil {
			if errors.Is(err, ErrBackpressure) {
				h.handleBackpressure(client)
			} else {
				// Client is disconnected or error occurred
				go h.Unregister(client)
			}
		}
	}

	// Publish to Redis for multi-instance support (if configured)
	if h.redisPubSub != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := h.redisPubSub.Publish(ctx, "logs:broadcast", msg)
		cancel()
		if err != nil {
			log.Printf("error publishing to Redis: %v", err)
		}
	}
}

// handleBackpressure handles slow clients based on configured strategy.
func (h *Hub) handleBackpressure(client ClientConn) {
	h.stats.mu.Lock()
	h.stats.TotalDropped++
	h.stats.mu.Unlock()

	log.Printf("backpressure: %s strategy applied", h.backpressureStrategy)

	if h.backpressureStrategy == "drop" {
		// Drop the message (default)
		return
	}

	if h.backpressureStrategy == "queue" {
		// Queue messages up to maxQueueSize
		// Implementation would buffer on a per-client basis
		h.stats.mu.Lock()
		h.stats.TotalQueued++
		h.stats.mu.Unlock()
	}
}

// sendHeartbeats sends heartbeat messages to all connected clients.
func (h *Hub) sendHeartbeats() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	heartbeat := map[string]interface{}{
		"type":      "heartbeat",
		"timestamp": time.Now().Unix(),
	}

	for client := range h.clients {
		err := client.Send(heartbeat)
		if err != nil {
			log.Printf("error sending heartbeat: %v", err)
			go h.Unregister(client)
		}
	}
}

// shutdown gracefully closes the hub.
func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.running = false

	// Close all client connections
	for client := range h.clients {
		client.Close()
	}

	// Close Redis connection if present
	if h.redisPubSub != nil {
		h.redisPubSub.Close()
	}

	log.Printf("hub shutdown: %d clients disconnected", len(h.clients))
}

// GetStats returns current hub statistics.
func (h *Hub) GetStats() map[string]interface{} {
	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()

	return map[string]interface{}{
		"active_clients":        h.stats.ActiveClients,
		"total_broadcasts":      h.stats.TotalBroadcasts,
		"total_dropped":         h.stats.TotalDropped,
		"total_queued":          h.stats.TotalQueued,
		"heartbeat_interval":    h.heartbeatInterval.String(),
		"backpressure_strategy": h.backpressureStrategy,
	}
}

// SetBackpressureStrategy sets the backpressure handling strategy.
func (h *Hub) SetBackpressureStrategy(strategy string) error {
	if strategy != "drop" && strategy != "queue" {
		return fmt.Errorf("invalid backpressure strategy: %s (must be 'drop' or 'queue')", strategy)
	}
	h.backpressureStrategy = strategy
	return nil
}

// SetHeartbeatInterval sets the heartbeat interval.
func (h *Hub) SetHeartbeatInterval(interval time.Duration) {
	h.heartbeatInterval = interval
}
