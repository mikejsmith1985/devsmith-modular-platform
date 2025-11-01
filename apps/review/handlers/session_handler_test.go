package review_handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestCreateSession_ValidRequest tests successful session creation.
func TestCreateSession_ValidRequest(t *testing.T) {
	// GIVEN: A session handler with mocked logger
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	// Create a test router
	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: POST request with valid session data
	reqBody := review_services.CreateSessionRequest{
		Title:       "Test Session",
		Description: "Test description",
		CodeSource:  "paste",
		CodeContent: "func main() {}",
		Language:    "go",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/review/sessions", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 201 Created
	assert.Equal(t, http.StatusCreated, w.Code)

	var response review_models.CodeReviewSession
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test Session", response.Title)
	assert.Equal(t, "active", response.Status)
}

// TestCreateSession_MissingTitle tests validation of required title field.
func TestCreateSession_MissingTitle(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: POST request without title
	reqBody := review_services.CreateSessionRequest{
		CodeSource: "paste",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/review/sessions", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 400 Bad Request (binding error or validation error)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestListSessions_WithPagination tests session listing with pagination.
func TestListSessions_WithPagination(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: GET request with pagination parameters
	req, err := http.NewRequest("GET", "/api/review/sessions?limit=10&offset=0", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 200 OK with pagination
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "sessions")
	assert.Contains(t, response, "limit")
	assert.Contains(t, response, "offset")
}

// TestGetSession_ValidID tests retrieving a session.
func TestGetSession_ValidID(t *testing.T) {
	// GIVEN: A session handler with a service
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: GET request for a session
	req, err := http.NewRequest("GET", "/api/review/sessions/1", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 500 (no persistence layer yet, service returns error)
	// After Phase 11.7 (repository integration), this should be 404 for missing session
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestGetSession_InvalidID tests session retrieval with invalid ID.
func TestGetSession_InvalidID(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: GET request with invalid session ID
	req, err := http.NewRequest("GET", "/api/review/sessions/invalid", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid session ID")
}

// TestDeleteSession_ValidID tests session deletion.
func TestDeleteSession_ValidID(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: DELETE request for a session
	req, err := http.NewRequest("DELETE", "/api/review/sessions/1", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 204 No Content
	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestUpdateSessionMode_ValidMode tests mode state update.
func TestUpdateSessionMode_ValidMode(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: POST request to update mode state
	reqBody := review_services.ModeUpdateRequest{
		Status:      "completed",
		IsCompleted: true,
		IssuesFound: 5,
		QualityScore: 85,
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/review/sessions/1/modes/preview", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Mode updated successfully")
}

// TestCompleteSession_ValidID tests session completion.
func TestCompleteSession_ValidID(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: POST request to complete session
	req, err := http.NewRequest("POST", "/api/review/sessions/1/complete", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 200 OK with statistics
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestArchiveSession_ValidID tests session archiving.
func TestArchiveSession_ValidID(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: POST request to archive session
	req, err := http.NewRequest("POST", "/api/review/sessions/1/archive", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Session archived successfully")
}

// TestGetSessionHistory_ValidID tests retrieving session history.
func TestGetSessionHistory_ValidID(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: GET request for session history
	req, err := http.NewRequest("GET", "/api/review/sessions/1/history?limit=50", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "history")
}

// TestAddSessionNote_ValidData tests adding notes to session.
func TestAddSessionNote_ValidData(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: POST request to add note
	reqBody := map[string]string{
		"mode": "preview",
		"note": "This code has potential performance issues",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/review/sessions/1/notes", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Note added successfully")
}

// TestAddSessionNote_MissingRequiredFields tests validation of required fields.
func TestAddSessionNote_MissingRequiredFields(t *testing.T) {
	// GIVEN: A session handler
	logger := &testutils.MockLogger{}
	service := review_services.NewSessionService(logger)
	handler := NewSessionHandlers(service, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	// WHEN: POST request without required fields
	reqBody := map[string]string{
		"mode": "preview",
		// note is missing
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/review/sessions/1/notes", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Response should be 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "required")
}
