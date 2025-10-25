package services

import (
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketHub_BroadcastToClients(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()

	// Mock clients
	client1 := &Client{Send: make(chan *models.LogEntry, 1)}
	client2 := &Client{Send: make(chan *models.LogEntry, 1)}

	hub.register <- client1
	hub.register <- client2

	// Broadcast log
	testLog := &models.LogEntry{Message: "test"}
	hub.broadcast <- testLog

	// Both clients should receive
	assert.Equal(t, testLog, <-client1.Send)
	assert.Equal(t, testLog, <-client2.Send)
}

func TestWebSocketHub_FiltersByService(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()

	client := &Client{
		Filters: map[string]string{"service": "portal"},
		Send:    make(chan *models.LogEntry, 1),
	}

	hub.register <- client
	hub.broadcast <- &models.LogEntry{Service: "review"}

	// Should not receive (filtered out)
	select {
	case <-client.Send:
		t.Fatal("Client received filtered log")
	default:
		// Expected - no message
	}
}
