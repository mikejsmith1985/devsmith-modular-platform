package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubClient_ExchangeCodeForToken_Error(t *testing.T) {
	client := NewGitHubClient("fakeid", "fakesecret")
	// Use an invalid code to simulate error
	token, err := client.ExchangeCodeForToken(context.Background(), "invalid_code")
	assert.Error(t, err)
	assert.Empty(t, token)
}
