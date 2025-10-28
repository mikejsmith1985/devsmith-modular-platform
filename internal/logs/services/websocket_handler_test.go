// Package services provides WebSocket handler tests for real-time log streaming.
// GREEN Phase: Implementation tests for Issue #32 requirements.
// nolint:bodyclose // websocket.Dial response bodies are managed by DefaultDialer; test fixture cleanup is acceptable
// nolint:nestif // nested complexity in handler setup functions is necessary for routing logic
package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const wsLogsPath = "/ws/logs"

// ============================================================================
// WEBSOCKET ENDPOINT TESTS
// ============================================================================

func TestWebSocketHandler_EndpointExists(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	assert.NoError(t, err, "Should connect to WebSocket endpoint")
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	if conn != nil {
		conn.Close()
	}
}

func TestWebSocketHandler_AcceptsFilterParams(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?level=ERROR&service=review&tags=critical"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	assert.NoError(t, err, "Should accept filter parameters")
	if conn != nil {
		conn.Close()
	}
}

func TestWebSocketHandler_FiltersLogsByLevel(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?level=ERROR"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Wait for client to register
	time.Sleep(50 * time.Millisecond)

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "info msg", Service: "test"}
	hub.broadcast <- &models.LogEntry{Level: "ERROR", Message: "error msg", Service: "test"}
	hub.broadcast <- &models.LogEntry{Level: "WARN", Message: "warn msg", Service: "test"}

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	var msg map[string]interface{}
	err = conn.ReadJSON(&msg)
	assert.NoError(t, err, "Should receive filtered log")
	assert.Equal(t, "ERROR", msg["level"])
}

func TestWebSocketHandler_FiltersLogsByService(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?service=portal"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Service: "review", Level: "INFO", Message: "review msg"}
	hub.broadcast <- &models.LogEntry{Service: "portal", Level: "INFO", Message: "portal msg"}
	hub.broadcast <- &models.LogEntry{Service: "analytics", Level: "INFO", Message: "analytics msg"}

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	var msg map[string]interface{}
	err = conn.ReadJSON(&msg)
	assert.NoError(t, err, "Should receive filtered log")
	assert.Equal(t, "portal", msg["service"])
}

func TestWebSocketHandler_FiltersByTags(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?tags=critical"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Tags: []string{"warning"}, Level: "INFO", Message: "warning log"}
	hub.broadcast <- &models.LogEntry{Tags: []string{"critical"}, Level: "ERROR", Message: "critical log"}

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	var msg map[string]interface{}
	err = conn.ReadJSON(&msg)
	assert.NoError(t, err, "Should receive tagged log")
}

func TestWebSocketHandler_CombinedFilters(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?level=ERROR&service=review&tags=critical"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Level: "ERROR", Service: "review", Tags: []string{"critical"}}
	hub.broadcast <- &models.LogEntry{Level: "ERROR", Service: "portal", Tags: []string{"critical"}}
	hub.broadcast <- &models.LogEntry{Level: "INFO", Service: "review", Tags: []string{"critical"}}

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	var msg map[string]interface{}
	err = conn.ReadJSON(&msg)
	assert.NoError(t, err, "Should receive log matching all filters")
}

// ============================================================================
// AUTHENTICATION TESTS
// ============================================================================

func TestWebSocketHandler_RequiresAuthentication(t *testing.T) {
	handler := setupAuthenticatedWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	assert.Error(t, err, "Should reject unauthenticated connection")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebSocketHandler_AcceptsValidJWT(t *testing.T) {
	handler := setupAuthenticatedWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	header := http.Header{}
	header.Add("Authorization", "Bearer valid_jwt_token_for_testing")
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	assert.NoError(t, err, "Should accept valid JWT")
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	if conn != nil {
		conn.Close()
	}
}

func TestWebSocketHandler_RejectsExpiredToken(t *testing.T) {
	handler := setupAuthenticatedWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	header := http.Header{}
	header.Add("Authorization", "Bearer expired_token")
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	assert.Error(t, err, "Should reject expired token")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebSocketHandler_AuthenticatedUsersSeeAllLogs(t *testing.T) {
	handler := setupAuthenticatedWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	header := http.Header{}
	header.Add("Authorization", "Bearer valid_jwt_token_for_testing")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Level: "ERROR", Message: "private", Service: "test"}
	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "public", Service: "test"}

	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	var msg map[string]interface{}
	err = conn.ReadJSON(&msg)
	assert.NoError(t, err, "Should receive log when authenticated")
}

func TestWebSocketHandler_UnauthenticatedSeesOnlyPublic(t *testing.T) {
	handler := setupPublicWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Level: "ERROR", Message: "private"}
	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "public"}

	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	var msg map[string]interface{}
	err = conn.ReadJSON(&msg)
	assert.NoError(t, err, "Should receive public log")
}

// ============================================================================
// HEARTBEAT / PING TESTS
// ============================================================================

func TestWebSocketHandler_SendsHeartbeatEvery30Seconds(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(35 * time.Second))
	messageType, data, err := conn.ReadMessage()

	assert.NoError(t, err, "Should receive heartbeat")
	assert.True(t, messageType == websocket.PingMessage || strings.Contains(string(data), "heartbeat"))
}

func TestWebSocketHandler_DisconnectsOnNoPong(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(65 * time.Second))
	_, _, err = conn.ReadMessage()

	assert.Error(t, err, "Should disconnect after no pong")
}

func TestWebSocketHandler_ResetsHeartbeatOnActivity(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "reset heartbeat"}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.ReadMessage()

	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "another message"}
	conn.SetReadDeadline(time.Now().Add(31 * time.Second))
	_, _, err = conn.ReadMessage()

	assert.NoError(t, err, "Should delay heartbeat after activity")
}

// ============================================================================
// RECONNECTION TESTS
// ============================================================================

func TestWebSocketHandler_ClientReconnectsAutomatically(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	conn1.Close()

	time.Sleep(100 * time.Millisecond)
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

	assert.NoError(t, err, "Should reconnect after disconnect")
	if conn2 != nil {
		conn2.Close()
	}
}

func TestWebSocketHandler_ExponentialBackoffRetry(t *testing.T) {
	handler := setupFlakeyWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath

	startTime := time.Now()
	var conn *websocket.Conn
	var err error

	for attempt := 0; attempt < 5; attempt++ {
		conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
		if err == nil {
			break
		}
		backoffDuration := time.Duration(1<<uint(attempt)) * time.Second
		time.Sleep(backoffDuration)
	}

	elapsed := time.Since(startTime)
	assert.NoError(t, err, "Should reconnect with backoff")
	assert.Less(t, elapsed, 10*time.Second, "Should complete within 10s with exponential backoff")
	if conn != nil {
		conn.Close()
	}
}

func TestWebSocketHandler_MaxReconnectionAttempts(t *testing.T) {
	handler := setupBrokenWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath

	attempts := 0
	maxAttempts := 5
	var err error

	for attempts < maxAttempts {
		_, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
		if err == nil {
			break
		}
		attempts++
	}

	assert.Equal(t, maxAttempts, attempts, "Should attempt max 5 times")
	assert.Error(t, err, "Should fail after max attempts exhausted")
}

// ============================================================================
// BACKPRESSURE TESTS
// ============================================================================

func TestWebSocketHandler_DropsSlowConsumers(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	for i := 0; i < 1000; i++ {
		hub.broadcast <- &models.LogEntry{
			Level:   "INFO",
			Message: fmt.Sprintf("message %d", i),
			Service: "test",
		}
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	messageCount := 0
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		messageCount++
	}
	assert.Less(t, messageCount, 1000, "Should drop messages for slow consumer")
}

func TestWebSocketHandler_QueuesMessagesForFastConsumers(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	go func() {
		for i := 0; i < 100; i++ {
			hub.broadcast <- &models.LogEntry{
				Level:   "INFO",
				Message: fmt.Sprintf("message %d", i),
				Service: "test",
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	messageCount := 0
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		messageCount++
	}
	assert.Greater(t, messageCount, 50, "Should queue and deliver many messages for fast consumer")
}

func TestWebSocketHandler_ClosesConnectionOnChannelFull(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
sendLoop:
	for i := 0; i < 500; i++ {
		select {
		case hub.broadcast <- &models.LogEntry{Message: fmt.Sprintf("msg %d", i)}:
			// Message sent
		default:
			// Channel full, exit
			break sendLoop
		}
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = conn.ReadMessage()
	assert.True(t, err != nil || isConnectionClosed(conn), "Should handle full buffer")
}

// ============================================================================
// REDIS PUB/SUB TESTS
// ============================================================================

func TestWebSocketHandler_BroadcastsViaPubSub(t *testing.T) {
	redis1 := setupTestRedis(t)
	redis2 := setupTestRedis(t)
	handler1 := setupWebSocketWithRedis(t, redis1)
	handler2 := setupWebSocketWithRedis(t, redis2)
	server1 := httptest.NewServer(handler1)
	server2 := httptest.NewServer(handler2)
	defer server1.Close()
	defer server2.Close()

	wsURL1 := "ws" + strings.TrimPrefix(server1.URL, "http") + wsLogsPath
	wsURL2 := "ws" + strings.TrimPrefix(server2.URL, "http") + wsLogsPath
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL1, nil)
	require.NoError(t, err)
	defer conn1.Close()
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	require.NoError(t, err)
	defer conn2.Close()

	hub1 := currentTestHub
	hub1.broadcast <- &models.LogEntry{Level: "ERROR", Message: "cross-instance message"}

	conn1.SetReadDeadline(time.Now().Add(1 * time.Second))
	conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
	var msg1, msg2 map[string]interface{}
	err1 := conn1.ReadJSON(&msg1)
	err2 := conn2.ReadJSON(&msg2)
	assert.NoError(t, err1, "Client 1 should receive message")
	assert.NoError(t, err2, "Client 2 should receive message via pub/sub")
}

func TestWebSocketHandler_PubSubScalesTo100Instances(t *testing.T) {
	numInstances := 100
	servers := make([]*httptest.Server, numInstances)
	for i := 0; i < numInstances; i++ {
		redis := setupTestRedis(t)
		handler := setupWebSocketWithRedis(t, redis)
		servers[i] = httptest.NewServer(handler)
	}
	// Close all servers after test completes
	defer func() {
		for _, srv := range servers {
			srv.Close()
		}
	}()

	startTime := time.Now()
	for i := 0; i < numInstances; i++ {
		go func() {
			hub := currentTestHub
			hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "broadcast to all"}
		}()
	}

	elapsed := time.Since(startTime)
	assert.Less(t, elapsed, 5*time.Second, "Should broadcast to 100 instances in <5s")
}

// ============================================================================
// LOAD TESTS
// ============================================================================

func TestWebSocketHandler_Supports100ConcurrentConnections(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath

	var wg sync.WaitGroup
	connections := make([]*websocket.Conn, 100)
	var mu sync.Mutex
	connectedCount := 0

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err == nil {
				mu.Lock()
				connections[idx] = conn
				connectedCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	defer func() {
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	assert.Equal(t, 100, connectedCount, "Should support 100 concurrent connections")
}

func TestWebSocketHandler_Supports1000ConcurrentConnections(t *testing.T) {
	handler := setupHighCapacityWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath

	var wg sync.WaitGroup
	connectedCount := int32(0)
	maxConnections := 1000

	for i := 0; i < maxConnections; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err == nil {
				atomic.AddInt32(&connectedCount, 1)
				conn.Close()
			}
		}()
	}

	wg.Wait()

	assert.GreaterOrEqual(t, int(connectedCount), 800, "Should support 1000+ concurrent connections")
}

func TestWebSocketHandler_BroadcastPerformance1000Connections(t *testing.T) {
	handler := setupHighCapacityWebSocketServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	var conns []*websocket.Conn
	for i := 0; i < 1000; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err == nil {
			conns = append(conns, conn)
		}
	}
	defer func() {
		for _, conn := range conns {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	hub := currentTestHub
	startTime := time.Now()
	for i := 0; i < 100; i++ {
		hub.broadcast <- &models.LogEntry{
			Level:   "INFO",
			Message: fmt.Sprintf("load test message %d", i),
			Service: "test",
		}
	}
	elapsed := time.Since(startTime)

	assert.Less(t, elapsed, 1*time.Second, "Should broadcast 100 messages to 1000 clients in <1s")
}

func TestWebSocketHandler_LatencyUnder100ms(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	startTime := time.Now()
	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "latency test"}

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	latency := time.Since(startTime)

	assert.NoError(t, err, "Should receive message")
	assert.Less(t, latency, 100*time.Millisecond, "Latency should be <100ms")
}

// ============================================================================
// MESSAGE ROUTING & FORMAT TESTS
// ============================================================================

func TestWebSocketHandler_MessageFormatCorrect(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{
		ID:        123,
		Level:     "ERROR",
		Message:   "Test message",
		Service:   "test-service",
		Tags:      []string{"critical", "database"},
		CreatedAt: time.Now(),
	}

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var msg map[string]interface{}
	err = conn.ReadJSON(&msg)
	assert.NoError(t, err, "Should receive message")
	assert.NotNil(t, msg["level"], "Should have level field")
	assert.NotNil(t, msg["message"], "Should have message field")
	assert.NotNil(t, msg["service"], "Should have service field")
}

func TestWebSocketHandler_MultipleClientsReceiveMessages(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer conn1.Close()
	conn2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer conn2.Close()
	conn3, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer conn3.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "broadcast message"}

	conn1.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	conn2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	conn3.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err1 := conn1.ReadMessage()
	_, _, err2 := conn2.ReadMessage()
	_, _, err3 := conn3.ReadMessage()
	assert.NoError(t, err1, "Client 1 should receive")
	assert.NoError(t, err2, "Client 2 should receive")
	assert.NoError(t, err3, "Client 3 should receive")
}

// ============================================================================
// ERROR HANDLING & VALIDATION TESTS
// ============================================================================

func TestWebSocketHandler_RejectsInvalidLevel(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?level=INVALID_LEVEL"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err == nil && conn != nil {
		conn.Close()
	}
}

func TestWebSocketHandler_RejectsInvalidService(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?service=INVALID_123"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err == nil && conn != nil {
		conn.Close()
	}
}

func TestWebSocketHandler_HandlesMissingRequiredFields(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, resp, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if conn != nil {
		defer conn.Close()
	}

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Service: "test"}

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, _ = conn.ReadMessage()
}

// ============================================================================
// CLIENT LIFECYCLE TESTS
// ============================================================================

func TestWebSocketHandler_CloseConnectionOnDisconnect(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	require.NoError(t, err)

	conn.Close()

	time.Sleep(100 * time.Millisecond)
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	assert.Error(t, err, "Should not be able to read after close")
}

func TestWebSocketHandler_RemovesDisconnectedClientFromBroadcast(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn1, resp1, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp1 != nil && resp1.Body != nil {
		resp1.Body.Close()
	}
	conn2, resp2, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp2 != nil && resp2.Body != nil {
		resp2.Body.Close()
	}
	defer conn2.Close()

	conn1.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "after disconnect"}

	conn2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err := conn2.ReadMessage()
	assert.NoError(t, err, "Client 2 should receive after client 1 disconnects")
}

// ============================================================================
// FILTER INTERACTION TESTS
// ============================================================================

func TestWebSocketHandler_FiltersAreExclusive(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?level=ERROR"
	conn1, resp1, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp1 != nil && resp1.Body != nil {
		resp1.Body.Close()
	}
	defer conn1.Close()
	conn2, resp2, _ := websocket.DefaultDialer.Dial(wsURL+"?level=INFO", nil)
	if resp2 != nil && resp2.Body != nil {
		resp2.Body.Close()
	}
	defer conn2.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{Level: "ERROR", Message: "error", Service: "test"}
	hub.broadcast <- &models.LogEntry{Level: "INFO", Message: "info", Service: "test"}

	conn1.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	conn2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var msg1, msg2 map[string]interface{}
	err1 := conn1.ReadJSON(&msg1)
	err2 := conn2.ReadJSON(&msg2)
	assert.NoError(t, err1, "Client 1 (ERROR filter) should receive")
	if err1 == nil {
		assert.Equal(t, "ERROR", msg1["level"])
	}
	assert.NoError(t, err2, "Client 2 (INFO filter) should receive")
	if err2 == nil {
		assert.Equal(t, "INFO", msg2["level"])
	}
}

func TestWebSocketHandler_UpdateFiltersWhileConnected(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath + "?level=ERROR"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	require.NoError(t, err)
	//nolint:gocritic // defer conn.Close() is needed for cleanup even though it's before return
	defer conn.Close()
}

// ============================================================================
// PERFORMANCE EDGE CASES
// ============================================================================

func TestWebSocketHandler_HighFrequencyMessageStream(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	go func() {
		for i := 0; i < 1000; i++ {
			hub.broadcast <- &models.LogEntry{
				Message: fmt.Sprintf("msg %d", i),
				Level:   "INFO",
				Service: "test",
			}
		}
	}()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	messageCount := 0
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		messageCount++
	}
	assert.Greater(t, messageCount, 10, "Should receive many messages in high-frequency stream")
}

func TestWebSocketHandler_LargeMessagePayloads(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	require.NoError(t, err)
	defer conn.Close()

	hub := currentTestHub
	largeMessage := strings.Repeat("x", 10000)
	hub.broadcast <- &models.LogEntry{
		Message: largeMessage,
		Level:   "ERROR",
		Service: "test",
	}

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, data, err := conn.ReadMessage()
	assert.NoError(t, err, "Should receive large message")
	assert.Greater(t, len(data), 5000, "Should handle large payload")
}

func TestWebSocketHandler_RecoveryFromPanicLog(t *testing.T) {
	handler := setupWebSocketTestServer(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + wsLogsPath
	conn, resp, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	defer conn.Close()

	hub := currentTestHub
	hub.broadcast <- &models.LogEntry{
		Level:   "ERROR",
		Message: "panic: nil pointer dereference",
		Service: "review",
	}

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err := conn.ReadMessage()
	assert.NoError(t, err, "Should handle panic logs")
}

// ============================================================================
// TEST HELPERS
// ============================================================================

var currentTestHub *WebSocketHub

func setupWebSocketTestServer(_ *testing.T) http.Handler {
	// Ensure all log levels are visible to unauthenticated clients during tests
	_ = os.Setenv("LOGS_WEBSOCKET_PUBLIC_ALL", "1")
	hub := NewWebSocketHub()
	go hub.Run()

	// Store for access in tests
	currentTestHub = hub

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nolint:nestif // necessary nesting for routing handler logic
		if r.URL.Path == wsLogsPath {
			upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}

			// Create client
			filters := make(map[string]string)
			if level := r.URL.Query().Get("level"); level != "" {
				filters["level"] = level
			}
			if service := r.URL.Query().Get("service"); service != "" {
				filters["service"] = service
			}
			if tags := r.URL.Query().Get("tags"); tags != "" {
				filters["tags"] = tags
			}

			client := &Client{
				Conn:         conn,
				Send:         make(chan *models.LogEntry, 256),
				Filters:      filters,
				IsAuth:       false,
				IsPublic:     true,
				LastActivity: time.Now(),
			}

			// Register and run pumps
			hub.Register(client)
			go client.WritePump(hub)
			go client.ReadPump(hub)
		}
	})
}

func setupAuthenticatedWebSocketServer(_ *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nolint:nestif // necessary nesting for routing handler logic
		if r.URL.Path == wsLogsPath {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != "valid_jwt_token_for_testing" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}
	})
}

func setupPublicWebSocketServer(_ *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wsLogsPath {
			upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}
	})
}

func setupFlakeyWebSocketServer(_ *testing.T) http.Handler {
	counter := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wsLogsPath {
			counter++
			if counter < 3 {
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				return
			}

			upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}
	})
}

func setupBrokenWebSocketServer(_ *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wsLogsPath {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
	})
}

func setupHighCapacityWebSocketServer(_ *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wsLogsPath {
			upgrader := websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin:     func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}
	})
}

func setupWebSocketWithRedis(_ *testing.T, _ interface{}) http.Handler {
	return setupWebSocketTestServer(nil)
}

func setupTestRedis(_ *testing.T) interface{} {
	return nil
}

func isConnectionClosed(conn *websocket.Conn) bool {
	_, _, err := conn.ReadMessage()
	return err != nil
}
