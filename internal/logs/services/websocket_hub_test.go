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

	// Mock clients
	client1 := &Client{send: make(chan *models.LogEntry, 1)}
	client2 := &Client{send: make(chan *models.LogEntry, 1)}

	hub.register <- client1
	hub.register <- client2

	// Broadcast log
	testLog := &models.LogEntry{Message: "test"}
	hub.broadcast <- testLog

	// Both clients should receive
	assert.Equal(t, testLog, <-client1.send)
	assert.Equal(t, testLog, <-client2.send)
}

func TestWebSocketHub_FiltersByService(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()

	client := &Client{
		filters: map[string]string{"service": "portal"},
		send:    make(chan *models.LogEntry, 1),
	}

	hub.register <- client
	hub.broadcast <- &models.LogEntry{Service: "review"}

	// Should not receive (filtered out)
	select {
	case <-client.send:
		t.Fatal("Client received filtered log")
	default:
		// Expected - no message
	}
}