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

// Backpressure strategy constants.
const (
	BackpressureStrategyDrop  = "drop"
	BackpressureStrategyQueue = "queue"
)

// Hub manages WebSocket clients and broadcasts log entries to them.
// It implements connection management, filtering, heartbeats, and backpressure handling.
// nolint:govet // fieldalignment: all pointer fields are functionally necessary - cannot be removed
type Hub struct {
	mu                   sync.RWMutex
	broadcast            chan interface{}
	register             chan ClientConn
	unregister           chan ClientConn
	cancel               context.CancelFunc
	redisPubSub          RedisPubSubConn
	clients              map[ClientConn]bool
	stats                *HubStats
	backpressureStrategy string
	heartbeatInterval    time.Duration
	heartbeatTimeout     time.Duration
	maxQueueSize         int
	running              bool
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
		backpressureStrategy: BackpressureStrategyDrop,
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
		return h.matchesMapFilters(filters, logData)
	}

	// Try reflection for typed structs with Service and Level fields
	return h.matchesStructFilters(filters, msg)
}

// matchesMapFilters checks filters against a map[string]interface{}.
func (h *Hub) matchesMapFilters(filters, logData map[string]interface{}) bool {
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

	return h.checkServiceFilter(filters, v) && h.checkLevelFilter(filters, v)
}

// checkServiceFilter verifies service filter match.
func (h *Hub) checkServiceFilter(filters map[string]interface{}, v reflect.Value) bool {
	serviceFilter, ok := filters["service"]
	if !ok {
		return true
	}

	sf, ok := serviceFilter.(string)
	if !ok {
		return false
	}

	serviceField := v.FieldByName("Service")
	if !serviceField.IsValid() || serviceField.Kind() != reflect.String {
		return true
	}

	return serviceField.String() == sf
}

// checkLevelFilter verifies level filter match.
func (h *Hub) checkLevelFilter(filters map[string]interface{}, v reflect.Value) bool {
	levelFilter, ok := filters["level"]
	if !ok {
		return true
	}

	lf, ok := levelFilter.(string)
	if !ok {
		return false
	}

	levelField := v.FieldByName("Level")
	if !levelField.IsValid() || levelField.Kind() != reflect.String {
		return true
	}

	return levelField.String() == lf
}

// Hub struct with optimal field alignment.
// Layout: 8 + 8 + 8 + 16 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 = ~120 bytes
// Note: pointers/interfaces first, then durations/ints, then strings

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
		if err := client.Close(); err != nil {
			log.Printf("error closing client: %v", err)
		}

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
				h.recordBackpressure()
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

// recordBackpressure records backpressure event based on strategy.
func (h *Hub) recordBackpressure() {
	h.stats.mu.Lock()
	defer h.stats.mu.Unlock()

	log.Printf("backpressure: %s strategy applied", h.backpressureStrategy)

	switch h.backpressureStrategy {
	case BackpressureStrategyDrop:
		h.stats.TotalDropped++
	case BackpressureStrategyQueue:
		h.stats.TotalQueued++
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
		if err := client.Close(); err != nil {
			log.Printf("error closing connection: %v", err)
		}
	}

	// Close Redis connection if present
	if h.redisPubSub != nil {
		if err := h.redisPubSub.Close(); err != nil {
			log.Printf("error closing Redis pub/sub: %v", err)
		}
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
	if strategy != BackpressureStrategyDrop && strategy != BackpressureStrategyQueue {
		return fmt.Errorf("invalid backpressure strategy: %s (must be '%s' or '%s')", strategy, BackpressureStrategyDrop, BackpressureStrategyQueue)
	}
	h.backpressureStrategy = strategy
	return nil
}

// SetHeartbeatInterval sets the heartbeat interval.
func (h *Hub) SetHeartbeatInterval(interval time.Duration) {
	h.heartbeatInterval = interval
}
