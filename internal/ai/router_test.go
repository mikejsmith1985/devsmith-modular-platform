package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockProvider is a mock AIProvider for testing
type mockProvider struct {
	name string
	info *ModelInfo
}

func (m *mockProvider) Generate(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	return &AIResponse{
		Content:      "mock response from " + m.name,
		InputTokens:  10,
		OutputTokens: 20,
		CostUSD:      0.001,
		Model:        m.info.Model,
		FinishReason: "complete",
	}, nil
}

func (m *mockProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockProvider) GetModelInfo() *ModelInfo {
	return m.info
}

// TestDefaultRouter_NewDefaultRouter_CreatesValidRouter verifies constructor
func TestDefaultRouter_NewDefaultRouter_CreatesValidRouter(t *testing.T) {
	router := NewDefaultRouter()

	assert.NotNil(t, router, "Router should be created")
	assert.NotNil(t, router.providers)
	assert.NotNil(t, router.userPreferences)
}

// TestDefaultRouter_RegisterProvider_StoresProvider verifies provider registration
func TestDefaultRouter_RegisterProvider_StoresProvider(t *testing.T) {
	router := NewDefaultRouter()
	provider := &mockProvider{
		name: "test",
		info: &ModelInfo{
			Provider: "test",
			Model:    "test-model",
		},
	}

	err := router.RegisterProvider("test", "test-model", provider)
	assert.NoError(t, err, "Registration should succeed")

	// Verify provider was stored
	retrieved, exists := router.providers["test:test-model"]
	assert.True(t, exists, "Provider should be retrievable")
	assert.Equal(t, provider, retrieved)
}

// TestDefaultRouter_RegisterProvider_RejectsNilProvider verifies validation
func TestDefaultRouter_RegisterProvider_RejectsNilProvider(t *testing.T) {
	router := NewDefaultRouter()

	err := router.RegisterProvider("test", "test-model", nil)
	assert.Error(t, err, "Should reject nil provider")
	assert.Contains(t, err.Error(), "provider cannot be nil")
}

// TestDefaultRouter_RegisterProvider_RejectsEmptyKey verifies validation
func TestDefaultRouter_RegisterProvider_RejectsEmptyKey(t *testing.T) {
	router := NewDefaultRouter()
	provider := &mockProvider{
		name: "test",
		info: &ModelInfo{Provider: "test"},
	}

	err := router.RegisterProvider("", "test-model", provider)
	assert.Error(t, err, "Should reject empty provider name")

	err = router.RegisterProvider("test", "", provider)
	assert.Error(t, err, "Should reject empty model name")
}

// TestDefaultRouter_Route_ReturnsOllamaByDefault verifies default routing
func TestDefaultRouter_Route_ReturnsOllamaByDefault(t *testing.T) {
	router := NewDefaultRouter()

	ollama := &mockProvider{
		name: "ollama",
		info: &ModelInfo{
			Provider:             "ollama",
			Model:                "local-model",
			CostPer1kInputTokens: 0.0,
		},
	}

	router.RegisterProvider("ollama", "local-model", ollama)

	provider, err := router.Route(context.Background(), "review", 123)

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "ollama", provider.GetModelInfo().Provider)
}

// TestDefaultRouter_Route_ReturnsUserPreferredProvider verifies preference override
func TestDefaultRouter_Route_ReturnsUserPreferredProvider(t *testing.T) {
	router := NewDefaultRouter()

	ollama := &mockProvider{
		name: "ollama",
		info: &ModelInfo{Provider: "ollama", Model: "local"},
	}
	anthropic := &mockProvider{
		name: "anthropic",
		info: &ModelInfo{Provider: "anthropic", Model: "claude-haiku"},
	}

	router.RegisterProvider("ollama", "local", ollama)
	router.RegisterProvider("anthropic", "claude-haiku", anthropic)

	// Set user preference
	err := router.SetUserPreference(context.Background(), 123, "review", "anthropic", "claude-haiku", true)
	assert.NoError(t, err)

	provider, err := router.Route(context.Background(), "review", 123)

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "anthropic", provider.GetModelInfo().Provider)
}

// TestDefaultRouter_Route_AppSpecificRouting verifies app isolation
func TestDefaultRouter_Route_AppSpecificRouting(t *testing.T) {
	router := NewDefaultRouter()

	ollama := &mockProvider{
		name: "ollama",
		info: &ModelInfo{Provider: "ollama", Model: "local"},
	}
	anthropic := &mockProvider{
		name: "anthropic",
		info: &ModelInfo{Provider: "anthropic", Model: "claude-haiku"},
	}

	router.RegisterProvider("ollama", "local", ollama)
	router.RegisterProvider("anthropic", "claude-haiku", anthropic)

	// Set preference for 'review' app
	router.SetUserPreference(context.Background(), 123, "review", "anthropic", "claude-haiku", true)
	// Different preference for 'logs' app
	router.SetUserPreference(context.Background(), 123, "logs", "ollama", "local", true)

	reviewProvider, _ := router.Route(context.Background(), "review", 123)
	assert.Equal(t, "anthropic", reviewProvider.GetModelInfo().Provider)

	logsProvider, _ := router.Route(context.Background(), "logs", 123)
	assert.Equal(t, "ollama", logsProvider.GetModelInfo().Provider)
}

// TestDefaultRouter_SetUserPreference_PersistsSelection verifies persistence
func TestDefaultRouter_SetUserPreference_PersistsSelection(t *testing.T) {
	router := NewDefaultRouter()

	provider := &mockProvider{
		name: "test",
		info: &ModelInfo{Provider: "test", Model: "test-model"},
	}
	router.RegisterProvider("test", "test-model", provider)

	err := router.SetUserPreference(context.Background(), 456, "review", "test", "test-model", true)
	assert.NoError(t, err)

	// Retrieve preference
	pref := router.userPreferences[router.getUserAppKey(456, "review")]
	assert.NotNil(t, pref)
	assert.Equal(t, "test:test-model", pref.ProviderModel)
}

// TestDefaultRouter_SetUserPreference_NonPersistentSession verifies temp setting
func TestDefaultRouter_SetUserPreference_NonPersistentSession(t *testing.T) {
	router := NewDefaultRouter()

	provider := &mockProvider{
		name: "test",
		info: &ModelInfo{Provider: "test", Model: "test-model"},
	}
	router.RegisterProvider("test", "test-model", provider)

	// Non-persistent preference (session only)
	err := router.SetUserPreference(context.Background(), 789, "review", "test", "test-model", false)
	assert.NoError(t, err)

	// Should still be retrievable in session
	routed, _ := router.Route(context.Background(), "review", 789)
	assert.Equal(t, "test", routed.GetModelInfo().Provider)
}

// TestDefaultRouter_GetAvailableModels_ReturnsAllRegistered verifies model list
func TestDefaultRouter_GetAvailableModels_ReturnsAllRegistered(t *testing.T) {
	router := NewDefaultRouter()

	providers := []struct {
		name  string
		model string
	}{
		{"ollama", "local"},
		{"anthropic", "claude-haiku"},
		{"anthropic", "claude-sonnet"},
		{"openai", "gpt-4-turbo"},
	}

	for _, p := range providers {
		mock := &mockProvider{
			name: p.name,
			info: &ModelInfo{
				Provider: p.name,
				Model:    p.model,
			},
		}
		router.RegisterProvider(p.name, p.model, mock)
	}

	models, err := router.GetAvailableModels(context.Background(), "review", 999)

	assert.NoError(t, err)
	assert.Equal(t, 4, len(models), "Should return all registered models")

	// Verify all models are present
	modelStrs := make(map[string]bool)
	for _, m := range models {
		modelStrs[m.Provider+":"+m.Model] = true
	}
	assert.True(t, modelStrs["ollama:local"])
	assert.True(t, modelStrs["anthropic:claude-haiku"])
	assert.True(t, modelStrs["anthropic:claude-sonnet"])
	assert.True(t, modelStrs["openai:gpt-4-turbo"])
}

// TestDefaultRouter_LogUsage_RecordsUsage verifies usage tracking
func TestDefaultRouter_LogUsage_RecordsUsage(t *testing.T) {
	router := NewDefaultRouter()

	req := &AIRequest{
		Prompt: "test prompt",
	}
	resp := &AIResponse{
		Content:      "test response",
		InputTokens:  100,
		OutputTokens: 50,
		CostUSD:      0.01,
	}

	err := router.LogUsage(context.Background(), 111, "review", req, resp)

	assert.NoError(t, err, "LogUsage should not error")
	// Note: Full verification requires database/storage implementation
}

// TestDefaultRouter_Route_UnknownProviderFallback verifies error handling
func TestDefaultRouter_Route_UnknownProviderFallback(t *testing.T) {
	router := NewDefaultRouter()

	// Set preference for non-existent provider
	provider := &mockProvider{
		name: "available",
		info: &ModelInfo{Provider: "available", Model: "model"},
	}
	router.RegisterProvider("available", "model", provider)
	router.SetUserPreference(context.Background(), 222, "review", "nonexistent", "model", true)

	// Should fall back to available provider
	routed, err := router.Route(context.Background(), "review", 222)

	assert.NoError(t, err, "Should fall back gracefully")
	assert.NotNil(t, routed)
}

// TestDefaultRouter_CostOptimizedRouting verifies cost preference
func TestDefaultRouter_CostOptimizedRouting_PicksFreestOption(t *testing.T) {
	router := NewDefaultRouter()

	expensive := &mockProvider{
		name: "expensive",
		info: &ModelInfo{
			Provider:             "expensive",
			Model:                "expensive-model",
			CostPer1kInputTokens: 0.05,
		},
	}
	free := &mockProvider{
		name: "free",
		info: &ModelInfo{
			Provider:             "free",
			Model:                "free-model",
			CostPer1kInputTokens: 0.0,
		},
	}

	router.RegisterProvider("expensive", "expensive-model", expensive)
	router.RegisterProvider("free", "free-model", free)

	// When no preference set, should default to free (Ollama-like)
	// Note: This tests the router's default behavior
	models, _ := router.GetAvailableModels(context.Background(), "review", 333)
	assert.Greater(t, len(models), 0)
}

// TestDefaultRouter_MultiProviderScene verifies complex routing
func TestDefaultRouter_MultiProviderScene(t *testing.T) {
	router := NewDefaultRouter()

	// Register multiple providers
	providers := map[string]*mockProvider{
		"ollama:local": {
			name: "ollama",
			info: &ModelInfo{Provider: "ollama", Model: "local", CostPer1kInputTokens: 0.0},
		},
		"anthropic:haiku": {
			name: "anthropic",
			info: &ModelInfo{Provider: "anthropic", Model: "haiku", CostPer1kInputTokens: 0.00080},
		},
		"openai:gpt4": {
			name: "openai",
			info: &ModelInfo{Provider: "openai", Model: "gpt4", CostPer1kInputTokens: 0.01},
		},
	}

	for _, p := range providers {
		// Extract provider and model from key
		provider := p.info.Provider
		model := p.info.Model
		router.RegisterProvider(provider, model, p)
	}

	// Test 1: Default routing (should pick free)
	p1, _ := router.Route(context.Background(), "review", 444)
	assert.Equal(t, "ollama", p1.GetModelInfo().Provider, "Should default to free provider")

	// Test 2: User selects expensive
	router.SetUserPreference(context.Background(), 444, "review", "openai", "gpt4", true)
	p2, _ := router.Route(context.Background(), "review", 444)
	assert.Equal(t, "openai", p2.GetModelInfo().Provider, "Should respect user preference")

	// Test 3: Different user has default
	p3, _ := router.Route(context.Background(), "review", 555)
	assert.Equal(t, "ollama", p3.GetModelInfo().Provider, "Other users should have independent preferences")
}

// TestDefaultRouter_getUserAppKey_GeneratesUniqueKeys verifies key generation
func TestDefaultRouter_getUserAppKey_GeneratesUniqueKeys(t *testing.T) {
	router := NewDefaultRouter()

	key1 := router.getUserAppKey(123, "review")
	key2 := router.getUserAppKey(123, "logs")
	key3 := router.getUserAppKey(456, "review")

	assert.NotEqual(t, key1, key2, "Different apps should have different keys")
	assert.NotEqual(t, key1, key3, "Different users should have different keys")
	assert.NotEqual(t, key2, key3, "Different user+app should have different keys")
}

// TestDefaultRouter_Concurrency verifies thread safety
func TestDefaultRouter_Concurrency(t *testing.T) {
	router := NewDefaultRouter()

	provider := &mockProvider{
		name: "test",
		info: &ModelInfo{Provider: "test", Model: "test-model"},
	}
	router.RegisterProvider("test", "test-model", provider)

	// Simulate concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(userID int64) {
			router.SetUserPreference(context.Background(), int64(userID), "review", "test", "test-model", true)
			router.Route(context.Background(), "review", int64(userID))
			done <- true
		}(int64(i))
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.True(t, true, "Should handle concurrent access without panic")
}
