package ai

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockFailingProvider fails on first call, succeeds on second
type mockFailingProvider struct {
	name          string
	failCount     int
	callCount     int
	shouldFail    bool
	info          *ModelInfo
}

func (m *mockFailingProvider) Generate(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	m.callCount++
	if m.shouldFail || m.failCount > 0 {
		m.failCount--
		return nil, fmt.Errorf("provider %s failed", m.name)
	}
	return &AIResponse{
		Content: "response from " + m.name,
		Model:   m.info.Model,
	}, nil
}

func (m *mockFailingProvider) HealthCheck(ctx context.Context) error {
	if m.shouldFail {
		return fmt.Errorf("provider %s unhealthy", m.name)
	}
	return nil
}

func (m *mockFailingProvider) GetModelInfo() *ModelInfo {
	return m.info
}

// TestFallbackChain_NewFallbackChain_CreatesValidChain verifies constructor
func TestFallbackChain_NewFallbackChain_CreatesValidChain(t *testing.T) {
	chain := NewFallbackChain()

	assert.NotNil(t, chain, "Chain should be created")
	assert.NotNil(t, chain.providers)
}

// TestFallbackChain_AddProvider_StoresProvider verifies addition
func TestFallbackChain_AddProvider_StoresProvider(t *testing.T) {
	chain := NewFallbackChain()

	provider := &mockFailingProvider{
		name: "primary",
		info: &ModelInfo{Provider: "primary", Model: "model1"},
	}

	chain.AddProvider(provider)
	assert.Equal(t, 1, len(chain.providers))
}

// TestFallbackChain_Generate_UsesPrimaryProvider verifies primary used
func TestFallbackChain_Generate_UsesPrimaryProvider(t *testing.T) {
	chain := NewFallbackChain()

	primary := &mockFailingProvider{
		name:       "primary",
		shouldFail: false,
		info:       &ModelInfo{Provider: "primary", Model: "model1"},
	}
	fallback := &mockFailingProvider{
		name:       "fallback",
		shouldFail: false,
		info:       &ModelInfo{Provider: "fallback", Model: "model2"},
	}

	chain.AddProvider(primary)
	chain.AddProvider(fallback)

	resp, err := chain.Generate(context.Background(), &AIRequest{Prompt: "test"})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, primary.callCount, "Primary should be called once")
	assert.Equal(t, 0, fallback.callCount, "Fallback should not be called")
}

// TestFallbackChain_Generate_FallsBackOnFailure verifies fallback triggered
func TestFallbackChain_Generate_FallsBackOnFailure(t *testing.T) {
	chain := NewFallbackChain()

	primary := &mockFailingProvider{
		name:       "primary",
		shouldFail: true,
		info:       &ModelInfo{Provider: "primary", Model: "model1"},
	}
	fallback := &mockFailingProvider{
		name:       "fallback",
		shouldFail: false,
		info:       &ModelInfo{Provider: "fallback", Model: "model2"},
	}

	chain.AddProvider(primary)
	chain.AddProvider(fallback)

	resp, err := chain.Generate(context.Background(), &AIRequest{Prompt: "test"})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, primary.callCount, "Primary should be attempted")
	assert.Equal(t, 1, fallback.callCount, "Fallback should be called")
	assert.Contains(t, resp.Content, "fallback")
}

// TestFallbackChain_Generate_MultipleFallbacks verifies chain progression
func TestFallbackChain_Generate_MultipleFallbacks(t *testing.T) {
	chain := NewFallbackChain()

	p1 := &mockFailingProvider{name: "p1", shouldFail: true, info: &ModelInfo{Provider: "p1", Model: "m1"}}
	p2 := &mockFailingProvider{name: "p2", shouldFail: true, info: &ModelInfo{Provider: "p2", Model: "m2"}}
	p3 := &mockFailingProvider{name: "p3", shouldFail: false, info: &ModelInfo{Provider: "p3", Model: "m3"}}

	chain.AddProvider(p1)
	chain.AddProvider(p2)
	chain.AddProvider(p3)

	resp, err := chain.Generate(context.Background(), &AIRequest{Prompt: "test"})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, p1.callCount)
	assert.Equal(t, 1, p2.callCount)
	assert.Equal(t, 1, p3.callCount)
	assert.Contains(t, resp.Content, "p3")
}

// TestFallbackChain_Generate_AllFail verifies error when all fail
func TestFallbackChain_Generate_AllFail(t *testing.T) {
	chain := NewFallbackChain()

	p1 := &mockFailingProvider{name: "p1", shouldFail: true, info: &ModelInfo{Provider: "p1", Model: "m1"}}
	p2 := &mockFailingProvider{name: "p2", shouldFail: true, info: &ModelInfo{Provider: "p2", Model: "m2"}}

	chain.AddProvider(p1)
	chain.AddProvider(p2)

	resp, err := chain.Generate(context.Background(), &AIRequest{Prompt: "test"})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "all providers failed")
}

// TestFallbackChain_Generate_NoProviders verifies error when empty
func TestFallbackChain_Generate_NoProviders(t *testing.T) {
	chain := NewFallbackChain()

	resp, err := chain.Generate(context.Background(), &AIRequest{Prompt: "test"})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "no providers")
}

// TestFallbackChain_GetSuccessfulProvider_ReturnsWorking verifies provider selection
func TestFallbackChain_GetSuccessfulProvider_ReturnsWorking(t *testing.T) {
	chain := NewFallbackChain()

	p1 := &mockFailingProvider{name: "p1", shouldFail: true, info: &ModelInfo{Provider: "p1", Model: "m1"}}
	p2 := &mockFailingProvider{name: "p2", shouldFail: false, info: &ModelInfo{Provider: "p2", Model: "m2"}}

	chain.AddProvider(p1)
	chain.AddProvider(p2)

	provider, err := chain.GetSuccessfulProvider(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "p2", provider.GetModelInfo().Provider)
}

// TestFallbackChain_HealthCheck_SkipsFailedProviders verifies health check
func TestFallbackChain_HealthCheck_SkipsFailedProviders(t *testing.T) {
	chain := NewFallbackChain()

	p1 := &mockFailingProvider{name: "p1", shouldFail: true, info: &ModelInfo{Provider: "p1", Model: "m1"}}
	p2 := &mockFailingProvider{name: "p2", shouldFail: false, info: &ModelInfo{Provider: "p2", Model: "m2"}}

	chain.AddProvider(p1)
	chain.AddProvider(p2)

	healthyProvider, err := chain.GetHealthyProvider(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, healthyProvider)
	assert.Equal(t, "p2", healthyProvider.GetModelInfo().Provider)
}

// TestFallbackChain_GetHealthyProvider_AllUnhealthy verifies error handling
func TestFallbackChain_GetHealthyProvider_AllUnhealthy(t *testing.T) {
	chain := NewFallbackChain()

	p1 := &mockFailingProvider{name: "p1", shouldFail: true, info: &ModelInfo{Provider: "p1", Model: "m1"}}
	p2 := &mockFailingProvider{name: "p2", shouldFail: true, info: &ModelInfo{Provider: "p2", Model: "m2"}}

	chain.AddProvider(p1)
	chain.AddProvider(p2)

	provider, err := chain.GetHealthyProvider(context.Background())

	assert.Error(t, err)
	assert.Nil(t, provider)
}

// TestFallbackChain_SetMaxRetries_RespectsSetting verifies retry limit
func TestFallbackChain_SetMaxRetries_RespectsSetting(t *testing.T) {
	chain := NewFallbackChain()
	chain.SetMaxRetries(2)

	p1 := &mockFailingProvider{name: "p1", failCount: 5, info: &ModelInfo{Provider: "p1", Model: "m1"}}

	chain.AddProvider(p1)

	resp, err := chain.Generate(context.Background(), &AIRequest{Prompt: "test"})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestFallbackChain_RecordFailure_TracksFailures verifies failure tracking
func TestFallbackChain_RecordFailure_TracksFailures(t *testing.T) {
	chain := NewFallbackChain()

	p1 := &mockFailingProvider{name: "p1", info: &ModelInfo{Provider: "p1", Model: "m1"}}

	chain.AddProvider(p1)
	chain.RecordFailure(context.Background(), "p1")

	failures := chain.GetFailureCount("p1")
	assert.Greater(t, failures, int64(0))
}
