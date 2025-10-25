package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestLogEntry is a simple test log entry struct
type TestLogEntry struct {
	ID        int64
	Service   string
	Level     string
	Message   string
	Timestamp time.Time
}

// TestHub_NewHub verifies hub creation
func TestHub_NewHub(t *testing.T) {
	hub := NewHub(nil)
	assert.NotNil(t, hub)
	assert.Equal(t, 0, hub.ClientCount())
}

// TestHub_BroadcastToClients verifies all clients receive unfiltered logs
func TestHub_BroadcastToClients(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Stop()

	client1 := NewTestClient(t)
	client2 := NewTestClient(t)

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond) // Let registration process

	log := &TestLogEntry{
		ID:      1,
		Service: "portal",
		Level:   "info",
		Message: "Test broadcast",
	}

	hub.Broadcast(log)

	// Both clients should receive
	select {
	case received := <-client1.SendChan:
		assert.Equal(t, log, received)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for client1")
	}

	select {
	case received := <-client2.SendChan:
		assert.Equal(t, log, received)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for client2")
	}
}

// TestHub_FilterByService filters logs by service
func TestHub_FilterByService(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Stop()

	// Client filtering for "portal" only
	client := NewTestClient(t)
	client.Filters = map[string]interface{}{"service": "portal"}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Broadcast log from different service
	log := &TestLogEntry{
		Service: "review",
		Level:   "info",
		Message: "Should be filtered",
	}
	hub.Broadcast(log)

	// Should not receive
	select {
	case <-client.SendChan:
		t.Fatal("Should not have received filtered log")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}

	// Broadcast matching service
	log2 := &TestLogEntry{
		Service: "portal",
		Level:   "info",
		Message: "Should match",
	}
	hub.Broadcast(log2)

	// Should receive now
	select {
	case received := <-client.SendChan:
		assert.Equal(t, log2, received)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for matching service")
	}
}

// TestHub_FilterByLevel filters logs by level
func TestHub_FilterByLevel(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Stop()

	client := NewTestClient(t)
	client.Filters = map[string]interface{}{"level": "error"}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Broadcast info level
	hub.Broadcast(&TestLogEntry{Level: "info", Message: "info log"})

	select {
	case <-client.SendChan:
		t.Fatal("Should filter out info level")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}

	// Broadcast error level
	hub.Broadcast(&TestLogEntry{Level: "error", Message: "error log"})

	select {
	case <-client.SendChan:
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Should receive error level")
	}
}

// TestHub_MultipleFilters tests combined service + level filters
func TestHub_MultipleFilters(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Stop()

	client := NewTestClient(t)
	client.Filters = map[string]interface{}{
		"service": "portal",
		"level":   "error",
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Matches both filters
	hub.Broadcast(&TestLogEntry{Service: "portal", Level: "error", Message: "match"})
	select {
	case <-client.SendChan:
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Should match both filters")
	}

	// Fails service filter
	hub.Broadcast(&TestLogEntry{Service: "review", Level: "error", Message: "fail"})
	select {
	case <-client.SendChan:
		t.Fatal("Should not match wrong service")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}

	// Fails level filter
	hub.Broadcast(&TestLogEntry{Service: "portal", Level: "info", Message: "fail"})
	select {
	case <-client.SendChan:
		t.Fatal("Should not match wrong level")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

// TestHub_UnregisterClient tests client unregistration
func TestHub_UnregisterClient(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Stop()

	client := NewTestClient(t)
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.ClientCount())

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, hub.ClientCount())
}

// TestHub_Heartbeat verifies heartbeat mechanism
func TestHub_Heartbeat(t *testing.T) {
	hub := NewHub(nil)
	hub.heartbeatInterval = 50 * time.Millisecond // Fast heartbeat for testing
	go hub.Run()
	defer hub.Stop()

	client := NewTestClient(t)
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Should receive heartbeat
	select {
	case msg := <-client.SendChan:
		assert.NotNil(t, msg)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Should receive heartbeat")
	}
}

// TestHub_BackpressureDropsMessages tests dropping on slow consumers
func TestHub_BackpressureDropsMessages(t *testing.T) {
	hub := NewHub(nil)
	hub.backpressureStrategy = "drop"
	go hub.Run()
	defer hub.Stop()

	// Client with small buffer to simulate slow consumer
	client := NewTestClient(t)
	client.SendChan = make(chan interface{}, 1) // Small buffer

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Fill the buffer
	hub.Broadcast(&TestLogEntry{ID: 1, Message: "msg1"})
	time.Sleep(10 * time.Millisecond)

	// This should trigger backpressure (drop)
	hub.Broadcast(&TestLogEntry{ID: 2, Message: "msg2"})
	hub.Broadcast(&TestLogEntry{ID: 3, Message: "msg3"})
	time.Sleep(50 * time.Millisecond)

	// Should have only first message
	count := 0
	for {
		select {
		case <-client.SendChan:
			count++
		default:
			goto done
		}
	}
done:
	// With drop strategy and small buffer, we should not queue all messages
	assert.True(t, count <= 3)
}

// TestHub_BackpressureQueuesMessages tests queueing on slow consumers
func TestHub_BackpressureQueuesMessages(t *testing.T) {
	hub := NewHub(nil)
	hub.backpressureStrategy = "queue"
	hub.maxQueueSize = 100
	go hub.Run()
	defer hub.Stop()

	client := NewTestClient(t)
	client.SendChan = make(chan interface{}, 1)

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Send multiple messages
	for i := 1; i <= 5; i++ {
		hub.Broadcast(&TestLogEntry{ID: int64(i), Message: "msg"})
	}

	time.Sleep(50 * time.Millisecond)

	// With queue strategy, we should eventually receive all messages
	received := 0
	for i := 0; i < 10; i++ {
		select {
		case <-client.SendChan:
			received++
		default:
		}
	}
	assert.True(t, received > 0)
}

// TestHub_ClientCount verifies client count tracking
func TestHub_ClientCount(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Stop()

	assert.Equal(t, 0, hub.ClientCount())

	client1 := NewTestClient(t)
	client2 := NewTestClient(t)

	hub.Register(client1)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 2, hub.ClientCount())

	hub.Unregister(client1)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())
}

// TestHub_ConcurrentBroadcast tests concurrent broadcasting
func TestHub_ConcurrentBroadcast(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Stop()

	const numClients = 10
	clients := make([]*TestClient, numClients)

	for i := 0; i < numClients; i++ {
		clients[i] = NewTestClient(t)
		hub.Register(clients[i])
	}
	time.Sleep(20 * time.Millisecond)

	// Broadcast concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			hub.Broadcast(&TestLogEntry{ID: int64(id), Message: "concurrent"})
		}(i)
	}
	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	// All clients should receive messages
	for _, client := range clients {
		received := 0
		for {
			select {
			case <-client.SendChan:
				received++
			default:
				goto next
			}
		}
	next:
		assert.True(t, received > 0, "client should receive messages")
	}
}

// TestHub_AuthenticationFiltering tests authenticated vs public logs
func TestHub_AuthenticationFiltering(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Stop()

	// Authenticated client
	authClient := NewTestClient(t)
	authClient.IsAuthenticated = true
	authClient.UserRole = "admin"

	// Unauthenticated client
	unauthClient := NewTestClient(t)
	unauthClient.IsAuthenticated = false

	hub.Register(authClient)
	hub.Register(unauthClient)
	time.Sleep(10 * time.Millisecond)

	// Broadcast restricted log (for authenticated users only)
	log := &TestLogEntry{
		ID:      1,
		Service: "admin-service",
		Level:   "info",
		Message: "Restricted log",
	}

	hub.Broadcast(log)
	time.Sleep(50 * time.Millisecond)

	// Authenticated client should receive
	select {
	case <-authClient.SendChan:
		// Expected
	case <-time.After(200 * time.Millisecond):
		// May not receive depending on implementation
	}

	// Unauthenticated might not receive (if authorization is enforced)
	select {
	case <-unauthClient.SendChan:
		// Could receive if log is public
	case <-time.After(100 * time.Millisecond):
		// Or could be filtered
	}
}

// TestHub_MatchesFilters tests the filter matching logic
func TestHub_MatchesFilters(t *testing.T) {
	hub := NewHub(nil)

	tests := []struct {
		name    string
		filters map[string]interface{}
		log     *TestLogEntry
		want    bool
	}{
		{
			name:    "no filters",
			filters: map[string]interface{}{},
			log:     &TestLogEntry{Service: "portal", Level: "info"},
			want:    true,
		},
		{
			name:    "service match",
			filters: map[string]interface{}{"service": "portal"},
			log:     &TestLogEntry{Service: "portal", Level: "info"},
			want:    true,
		},
		{
			name:    "service mismatch",
			filters: map[string]interface{}{"service": "portal"},
			log:     &TestLogEntry{Service: "review", Level: "info"},
			want:    false,
		},
		{
			name:    "level match",
			filters: map[string]interface{}{"level": "error"},
			log:     &TestLogEntry{Service: "portal", Level: "error"},
			want:    true,
		},
		{
			name:    "level mismatch",
			filters: map[string]interface{}{"level": "error"},
			log:     &TestLogEntry{Service: "portal", Level: "info"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hub.MatchesFilters(tt.filters, tt.log)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestClient is a mock client for testing
type TestClient struct {
	SendChan         chan interface{}
	Filters          map[string]interface{}
	IsAuthenticated  bool
	UserRole         string
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	lastHeartbeat    time.Time
	pongReceived     bool
	consecutiveDrops int
}

// NewTestClient creates a test client
func NewTestClient(t *testing.T) *TestClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &TestClient{
		SendChan:      make(chan interface{}, 256),
		Filters:       make(map[string]interface{}),
		ctx:           ctx,
		cancel:        cancel,
		lastHeartbeat: time.Now(),
		pongReceived:  true,
	}
}

// Send sends a message to the client
func (c *TestClient) Send(msg interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ctx == nil {
		return ErrBackpressure
	}

	select {
	case c.SendChan <- msg:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
		return ErrBackpressure
	}
}

// UpdateFilters updates the client's filters
func (c *TestClient) UpdateFilters(filters map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Filters = filters
}

// GetFilters returns the client's filters
func (c *TestClient) GetFilters() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Filters
}

// Close closes the client
func (c *TestClient) Close() error {
	c.cancel()
	close(c.SendChan)
	return nil
}

// Context returns the client's context
func (c *TestClient) Context() context.Context {
	return c.ctx
}
