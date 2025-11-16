package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// MockPromptTemplateService is a mock for testing
type MockPromptTemplateService struct {
	mock.Mock
}

func (m *MockPromptTemplateService) GetEffectivePrompt(ctx context.Context, userID int, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error) {
	args := m.Called(ctx, userID, mode, userLevel, outputMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*review_models.PromptTemplate), args.Error(1)
}

func (m *MockPromptTemplateService) SaveCustomPrompt(ctx context.Context, userID int, mode, userLevel, outputMode, promptText string) (*review_models.PromptTemplate, error) {
	args := m.Called(ctx, userID, mode, userLevel, outputMode, promptText)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*review_models.PromptTemplate), args.Error(1)
}

func (m *MockPromptTemplateService) FactoryReset(ctx context.Context, userID int, mode, userLevel, outputMode string) error {
	args := m.Called(ctx, userID, mode, userLevel, outputMode)
	return args.Error(0)
}

func (m *MockPromptTemplateService) GetExecutionHistory(ctx context.Context, userID int, limit int) ([]*review_models.PromptExecution, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*review_models.PromptExecution), args.Error(1)
}

func (m *MockPromptTemplateService) RateExecution(ctx context.Context, userID int, executionID int64, rating int) error {
	args := m.Called(ctx, userID, executionID, rating)
	return args.Error(0)
}

// setupTestRouter creates a test router with authentication middleware mock
func setupTestRouter(handler *PromptHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock authentication middleware that sets user_id in context
	router.Use(func(c *gin.Context) {
		// For tests, set a default user_id unless the test overrides it
		if c.GetInt("user_id") == 0 {
			c.Set("user_id", 1)
		}
		c.Next()
	})

	// Register routes
	router.GET("/api/review/prompts", handler.GetPrompt)
	router.PUT("/api/review/prompts", handler.SavePrompt)
	router.DELETE("/api/review/prompts", handler.ResetPrompt)
	router.GET("/api/review/prompts/history", handler.GetHistory)
	router.POST("/api/review/prompts/:execution_id/rate", handler.RateExecution)

	return router
}

// Test: GET /api/review/prompts - Returns effective prompt with metadata
func TestPromptHandler_GetPrompt_Success(t *testing.T) {
	// GIVEN: A mock service that returns a user custom prompt
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	userTemplate := &review_models.PromptTemplate{
		ID:         "custom-preview-beginner-quick-1",
		UserID:     intPtr(1),
		Mode:       "preview",
		UserLevel:  "beginner",
		OutputMode: "quick",
		PromptText: "My custom preview prompt with {{code}}",
		Variables:  []string{"{{code}}"},
		IsDefault:  false,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockService.On("GetEffectivePrompt", mock.Anything, 1, "preview", "beginner", "quick").
		Return(userTemplate, nil)

	// WHEN: User requests the prompt
	req := httptest.NewRequest("GET", "/api/review/prompts?mode=preview&user_level=beginner&output_mode=quick", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 200 with prompt and metadata
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "custom-preview-beginner-quick-1", response["id"])
	assert.Equal(t, "My custom preview prompt with {{code}}", response["prompt_text"])
	assert.Equal(t, "preview", response["mode"])
	assert.True(t, response["is_custom"].(bool), "Should indicate this is a custom prompt")
	assert.True(t, response["can_reset"].(bool), "Should indicate reset is available")

	mockService.AssertExpectations(t)
}

// Test: GET /api/review/prompts - Returns system default when no custom exists
func TestPromptHandler_GetPrompt_SystemDefault(t *testing.T) {
	// GIVEN: A mock service that returns a system default prompt
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	systemTemplate := &review_models.PromptTemplate{
		ID:         "default-skim-intermediate-detailed",
		UserID:     nil, // System default
		Mode:       "skim",
		UserLevel:  "intermediate",
		OutputMode: "detailed",
		PromptText: "System default skim prompt with {{code}}",
		Variables:  []string{"{{code}}"},
		IsDefault:  true,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockService.On("GetEffectivePrompt", mock.Anything, 1, "skim", "intermediate", "detailed").
		Return(systemTemplate, nil)

	// WHEN: User requests the prompt
	req := httptest.NewRequest("GET", "/api/review/prompts?mode=skim&user_level=intermediate&output_mode=detailed", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should indicate system default
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["is_custom"].(bool), "Should indicate this is NOT a custom prompt")
	assert.False(t, response["can_reset"].(bool), "Should indicate reset is NOT available")
	assert.True(t, response["is_default"].(bool), "Should indicate this is default")

	mockService.AssertExpectations(t)
}

// Test: GET /api/review/prompts - Missing query parameters
func TestPromptHandler_GetPrompt_MissingParams(t *testing.T) {
	// GIVEN: Handler setup
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	// WHEN: Request without required parameters
	req := httptest.NewRequest("GET", "/api/review/prompts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "mode")
}

// Test: GET /api/review/prompts - Service error
func TestPromptHandler_GetPrompt_ServiceError(t *testing.T) {
	// GIVEN: Service returns error
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	mockService.On("GetEffectivePrompt", mock.Anything, 1, "preview", "beginner", "quick").
		Return(nil, errors.New("database error"))

	// WHEN: Request prompt
	req := httptest.NewRequest("GET", "/api/review/prompts?mode=preview&user_level=beginner&output_mode=quick", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test: PUT /api/review/prompts - Successfully saves custom prompt
func TestPromptHandler_SavePrompt_Success(t *testing.T) {
	// GIVEN: Mock service that saves custom prompt
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	savedTemplate := &review_models.PromptTemplate{
		ID:         "custom-critical-expert-comprehensive-1",
		UserID:     intPtr(1),
		Mode:       "critical",
		UserLevel:  "expert",
		OutputMode: "comprehensive",
		PromptText: "My custom critical analysis prompt with {{code}} and special instructions",
		Variables:  []string{"{{code}}"},
		IsDefault:  false,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockService.On("SaveCustomPrompt", mock.Anything, 1, "critical", "expert", "comprehensive",
		"My custom critical analysis prompt with {{code}} and special instructions").
		Return(savedTemplate, nil)

	// WHEN: User saves custom prompt
	requestBody := map[string]string{
		"mode":        "critical",
		"user_level":  "expert",
		"output_mode": "comprehensive",
		"prompt_text": "My custom critical analysis prompt with {{code}} and special instructions",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("PUT", "/api/review/prompts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 200 with saved prompt
	assert.Equal(t, http.StatusOK, w.Code)

	var response review_models.PromptTemplate
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "custom-critical-expert-comprehensive-1", response.ID)
	assert.NotNil(t, response.UserID)
	assert.Equal(t, 1, *response.UserID)

	mockService.AssertExpectations(t)
}

// Test: PUT /api/review/prompts - Missing required variables
func TestPromptHandler_SavePrompt_MissingVariables(t *testing.T) {
	// GIVEN: Service validates and returns error for missing {{code}}
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	mockService.On("SaveCustomPrompt", mock.Anything, 1, "preview", "beginner", "quick",
		"Invalid prompt without code variable").
		Return(nil, errors.New("prompt must contain {{code}} variable"))

	// WHEN: User tries to save prompt without required variable
	requestBody := map[string]string{
		"mode":        "preview",
		"user_level":  "beginner",
		"output_mode": "quick",
		"prompt_text": "Invalid prompt without code variable",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("PUT", "/api/review/prompts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "{{code}}")

	mockService.AssertExpectations(t)
}

// Test: PUT /api/review/prompts - Invalid JSON body
func TestPromptHandler_SavePrompt_InvalidJSON(t *testing.T) {
	// GIVEN: Handler setup
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	// WHEN: Request with invalid JSON
	req := httptest.NewRequest("PUT", "/api/review/prompts", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test: DELETE /api/review/prompts - Successfully resets to factory default
func TestPromptHandler_ResetPrompt_Success(t *testing.T) {
	// GIVEN: Mock service that successfully resets
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	mockService.On("FactoryReset", mock.Anything, 1, "detailed", "intermediate", "quick").
		Return(nil)

	// WHEN: User resets prompt to factory default
	req := httptest.NewRequest("DELETE", "/api/review/prompts?mode=detailed&user_level=intermediate&output_mode=quick", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 200 with success message
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "reset")

	mockService.AssertExpectations(t)
}

// Test: DELETE /api/review/prompts - No custom prompt exists
func TestPromptHandler_ResetPrompt_NoCustomExists(t *testing.T) {
	// GIVEN: Service returns error indicating no custom prompt
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	mockService.On("FactoryReset", mock.Anything, 1, "scan", "beginner", "detailed").
		Return(errors.New("no custom prompt to delete"))

	// WHEN: User tries to reset when already using default
	req := httptest.NewRequest("DELETE", "/api/review/prompts?mode=scan&user_level=beginner&output_mode=detailed", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 404 Not Found
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

// Test: GET /api/review/prompts/history - Returns execution history
func TestPromptHandler_GetHistory_Success(t *testing.T) {
	// GIVEN: Mock service returns execution history
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	executions := []*review_models.PromptExecution{
		{
			ID:             1,
			TemplateID:     "template-1",
			UserID:         1,
			RenderedPrompt: "Rendered prompt 1",
			Response:       "AI response 1",
			ModelUsed:      "claude-3-5-sonnet",
			LatencyMs:      1500,
			TokensUsed:     2000,
			UserRating:     intPtr(5),
			CreatedAt:      time.Now().Add(-1 * time.Hour),
		},
		{
			ID:             2,
			TemplateID:     "template-2",
			UserID:         1,
			RenderedPrompt: "Rendered prompt 2",
			Response:       "AI response 2",
			ModelUsed:      "gpt-4",
			LatencyMs:      2000,
			TokensUsed:     3000,
			UserRating:     nil,
			CreatedAt:      time.Now().Add(-2 * time.Hour),
		},
	}

	mockService.On("GetExecutionHistory", mock.Anything, 1, 50).
		Return(executions, nil)

	// WHEN: User requests execution history
	req := httptest.NewRequest("GET", "/api/review/prompts/history?limit=50", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 200 with executions
	assert.Equal(t, http.StatusOK, w.Code)

	var response []*review_models.PromptExecution
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, int64(1), response[0].ID)
	assert.Equal(t, "claude-3-5-sonnet", response[0].ModelUsed)

	mockService.AssertExpectations(t)
}

// Test: GET /api/review/prompts/history - Default limit applied
func TestPromptHandler_GetHistory_DefaultLimit(t *testing.T) {
	// GIVEN: Mock service
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	mockService.On("GetExecutionHistory", mock.Anything, 1, 50).
		Return([]*review_models.PromptExecution{}, nil)

	// WHEN: User requests history without limit parameter
	req := httptest.NewRequest("GET", "/api/review/prompts/history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should use default limit of 50
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertCalled(t, "GetExecutionHistory", mock.Anything, 1, 50)
}

// Test: POST /api/review/prompts/:execution_id/rate - Successfully rates execution
func TestPromptHandler_RateExecution_Success(t *testing.T) {
	// GIVEN: Mock service that accepts rating
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	mockService.On("RateExecution", mock.Anything, 1, int64(123), 4).
		Return(nil)

	// WHEN: User rates an execution
	requestBody := map[string]int{
		"rating": 4,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/review/prompts/123/rate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 200 with success message
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "Rating")

	mockService.AssertExpectations(t)
}

// Test: POST /api/review/prompts/:execution_id/rate - Invalid rating value
func TestPromptHandler_RateExecution_InvalidRating(t *testing.T) {
	// GIVEN: Handler setup
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)
	router := setupTestRouter(handler)

	// WHEN: User provides rating outside 1-5 range
	requestBody := map[string]int{
		"rating": 6, // Invalid
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/review/prompts/123/rate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Rating")
}

// Test: All endpoints require authentication
func TestPromptHandler_RequiresAuthentication(t *testing.T) {
	// GIVEN: Router WITHOUT authentication middleware
	mockService := new(MockPromptTemplateService)
	handler := NewPromptHandler(mockService)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// No auth middleware - user_id not set
	router.GET("/api/review/prompts", handler.GetPrompt)

	// WHEN: Request without authentication
	req := httptest.NewRequest("GET", "/api/review/prompts?mode=preview&user_level=beginner&output_mode=quick", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
