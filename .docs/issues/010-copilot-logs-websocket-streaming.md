# Issue #010: [COPILOT] Logs Service - WebSocket Real-time Streaming

**Labels:** `copilot`, `logs`, `websocket`, `real-time`
**Created:** 2025-10-19
**Issue:** #10
**Estimated Time:** 90-120 minutes
**Depends On:** Issue #009 (Logs Foundation)

---

# 🚨 STEP 0: CREATE FEATURE BRANCH FIRST 🚨

```bash
git checkout development && git pull origin development
git checkout -b feature/010-copilot-logs-websocket-streaming
git branch --show-current
```

---

## Task Description

Add real-time WebSocket streaming to Logs Service. Users connect via WebSocket, receive logs as they're ingested. Critical for debugging and monitoring.

---

## Success Criteria
- [ ] WebSocket endpoint at WS /ws/logs
- [ ] Clients receive logs in real-time as they're ingested
- [ ] Support filtering (subscribe to specific service/level)
- [ ] Handle disconnections gracefully
- [ ] Broadcast to all connected clients
- [ ] 70%+ test coverage

---

## Implementation

### Phase 1: WebSocket Hub (Pub/Sub)

**File:** `internal/logs/services/websocket_hub.go`
```go
package services

import (
	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"sync"
)

type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan *models.LogEntry
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	hub     *WebSocketHub
	conn    *websocket.Conn
	send    chan *models.LogEntry
	filters map[string]string  // service, level filters
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *models.LogEntry, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

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
				// Filter by client preferences
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

func (h *WebSocketHub) matchesFilters(client *Client, log *models.LogEntry) bool {
	if service, ok := client.filters["service"]; ok && service != log.Service {
		return false
	}
	if level, ok := client.filters["level"]; ok && level != log.Level {
		return false
	}
	return true
}

func (h *WebSocketHub) Broadcast(log *models.LogEntry) {
	h.broadcast <- log
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var filters map[string]string
		err := c.conn.ReadJSON(&filters)
		if err != nil {
			break
		}
		c.filters = filters  // Update client filters
	}
}

func (c *Client) WritePump() {
	defer c.conn.Close()

	for log := range c.send {
		if err := c.conn.WriteJSON(log); err != nil {
			return
		}
	}
}
```

**Commit:** `git add internal/logs/services/websocket_hub.go && git commit -m "feat(logs): add WebSocket hub for real-time streaming"`

---

### Phase 2: WebSocket Handler

**File:** `cmd/logs/handlers/websocket_handler.go`
```go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true  // TODO: Restrict in production
	},
}

type WebSocketHandler struct {
	hub *services.WebSocketHub
}

func NewWebSocketHandler(hub *services.WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &services.Client{
		hub:     h.hub,
		conn:    conn,
		send:    make(chan *models.LogEntry, 256),
		filters: make(map[string]string),
	}

	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
```

**Commit:** `git add cmd/logs/handlers/websocket_handler.go && git commit -m "feat(logs): add WebSocket handler"`

---

### Phase 3: Update LogService to Broadcast

**File:** `internal/logs/services/log_service.go` (modify)
```go
type LogService struct {
	logRepo LogRepositoryInterface
	hub     *WebSocketHub  // Add this
}

func NewLogService(logRepo LogRepositoryInterface, hub *WebSocketHub) *LogService {
	return &LogService{
		logRepo: logRepo,
		hub:     hub,
	}
}

func (s *LogService) IngestLog(ctx context.Context, log *models.LogEntry) error {
	if err := s.logRepo.Create(ctx, log); err != nil {
		return err
	}

	// Broadcast to WebSocket clients
	s.hub.Broadcast(log)

	return nil
}
```

**Commit:** `git add internal/logs/services/log_service.go && git commit -m "feat(logs): broadcast logs to WebSocket clients on ingestion"`

---

### Phase 4: Update Main to Initialize Hub

**File:** `cmd/logs/main.go` (modify)
```go
func main() {
	// ... existing setup

	// Initialize WebSocket hub
	hub := services.NewWebSocketHub()
	go hub.Run()  // Run in goroutine

	logService := services.NewLogService(logRepo, hub)
	logHandler := handlers.NewLogHandler(logService)
	wsHandler := handlers.NewWebSocketHandler(hub)

	// ... routes
	router.GET("/ws/logs", wsHandler.HandleWebSocket)
}
```

**Commit:** `git add cmd/logs/main.go && git commit -m "feat(logs): initialize WebSocket hub in main"`

---

### Phase 5: Update go.mod for WebSocket

```bash
go get github.com/gorilla/websocket
```

**Commit:** `git add go.mod go.sum && git commit -m "deps(logs): add gorilla/websocket"`

---

### Phase 6: Tests

**File:** `internal/logs/services/websocket_hub_test.go`
```go
package services

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

func TestWebSocketHub_BroadcastToClients(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()

	// Create mock client
	client := &Client{
		hub:     hub,
		send:    make(chan *models.LogEntry, 1),
		filters: map[string]string{},
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)  // Let registration process

	// Broadcast log
	log := &models.LogEntry{
		ID:      1,
		Service: "portal",
		Level:   "info",
		Message: "Test log",
	}
	hub.Broadcast(log)

	// Client should receive it
	select {
	case received := <-client.send:
		assert.Equal(t, "Test log", received.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for log")
	}
}

func TestWebSocketHub_FiltersByService(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()

	client := &Client{
		hub:     hub,
		send:    make(chan *models.LogEntry, 1),
		filters: map[string]string{"service": "portal"},
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Broadcast non-matching log
	log := &models.LogEntry{Service: "review", Message: "Should not receive"}
	hub.Broadcast(log)

	select {
	case <-client.send:
		t.Fatal("Should not have received filtered log")
	case <-time.After(100 * time.Millisecond):
		// Expected - no message
	}
}
```

**Commit:** `git add internal/logs/services/websocket* && git commit -m "test(logs): add WebSocket hub tests"`

---

### Phase 7: Update nginx for WebSocket

**File:** `docker/nginx/nginx.conf` (update /logs location)
```nginx
location /logs {
    proxy_pass http://logs:8082;
    proxy_set_header Host $host;

    # WebSocket support
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

**Commit:** `git add docker/nginx/nginx.conf && git commit -m "feat(logs): add WebSocket support to nginx"`

---

### Phase 8: Push

```bash
git push -u origin feature/010-copilot-logs-websocket-streaming
```

---

## References
- `ARCHITECTURE.md` lines 1126-1145 (Logs Service spec)
- Gorilla WebSocket docs: https://github.com/gorilla/websocket

**Time:** 90-120 minutes

**Testing WebSocket:**
```javascript
// In browser console
const ws = new WebSocket('ws://localhost:3000/ws/logs');
ws.onmessage = (event) => console.log(JSON.parse(event.data));
ws.send(JSON.stringify({service: 'portal'}));  // Filter
```
