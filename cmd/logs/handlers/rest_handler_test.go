package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// nolint:dupl // MockLogService implements LogService interface - dupl is expected
type MockLogService struct {
	InsertFn     func(ctx context.Context, entry map[string]interface{}) (int64, error)
	QueryFn      func(ctx context.Context, filters map[string]interface{}, page map[string]int) ([]interface{}, error)
	GetByIDFn    func(ctx context.Context, id int64) (interface{}, error)
	StatsFn      func(ctx context.Context) (map[string]interface{}, error)
	DeleteByIDFn func(ctx context.Context, id int64) error
	DeleteFn     func(ctx context.Context, filters map[string]interface{}) (int64, error)
}

func (m *MockLogService) Insert(ctx context.Context, entry map[string]interface{}) (int64, error) {
	if m.InsertFn != nil {
		return m.InsertFn(ctx, entry)
	}
	return 1, nil
}

func (m *MockLogService) Query(ctx context.Context, filters map[string]interface{}, page map[string]int) ([]interface{}, error) {
	if m.QueryFn != nil {
		return m.QueryFn(ctx, filters, page)
	}
	return []interface{}{}, nil
}

func (m *MockLogService) GetByID(ctx context.Context, id int64) (interface{}, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockLogService) Stats(ctx context.Context) (map[string]interface{}, error) {
	if m.StatsFn != nil {
		return m.StatsFn(ctx)
	}
	return map[string]interface{}{}, nil
}

func (m *MockLogService) DeleteByID(ctx context.Context, id int64) error {
	if m.DeleteByIDFn != nil {
		return m.DeleteByIDFn(ctx, id)
	}
	return nil
}

func (m *MockLogService) Delete(ctx context.Context, filters map[string]interface{}) (int64, error) {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, filters)
	}
	return 0, nil
}

func TestPostLogs_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		InsertFn: func(ctx context.Context, entry map[string]interface{}) (int64, error) {
			return 42, nil
		},
	}

	router.POST("/api/logs", PostLogs(mockSvc))

	body := map[string]interface{}{
		"service": "portal",
		"level":   "info",
		"message": "User logged in",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(42), resp["id"])
}

func TestPostLogs_MissingRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockSvc := &MockLogService{}

	router.POST("/api/logs", PostLogs(mockSvc))

	body := map[string]interface{}{
		"level": "info",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetLogs_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		QueryFn: func(ctx context.Context, filters map[string]interface{}, page map[string]int) ([]interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": 1, "service": "portal", "level": "info", "message": "test"},
			}, nil
		},
	}

	router.GET("/api/logs", GetLogs(mockSvc))

	req := httptest.NewRequest("GET", "/api/logs?limit=10&offset=0", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["entries"])
}

func TestGetLogByID_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		GetByIDFn: func(ctx context.Context, id int64) (interface{}, error) {
			return map[string]interface{}{
				"id": id, "service": "portal", "level": "info", "message": "test",
			}, nil
		},
	}

	router.GET("/api/logs/:id", GetLogByID(mockSvc))

	req := httptest.NewRequest("GET", "/api/logs/42", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetStats_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		StatsFn: func(ctx context.Context) (map[string]interface{}, error) {
			return map[string]interface{}{
				"total": 100, "by_level": map[string]int{"info": 50, "error": 50},
			}, nil
		},
	}

	router.GET("/api/logs/stats", GetStats(mockSvc))

	req := httptest.NewRequest("GET", "/api/logs/stats", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteLogs_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		DeleteFn: func(ctx context.Context, filters map[string]interface{}) (int64, error) {
			return 25, nil
		},
	}

	router.DELETE("/api/logs", DeleteLogs(mockSvc))

	body := map[string]interface{}{"before": "2025-01-01"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("DELETE", "/api/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostLogs_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockSvc := &MockLogService{}

	router.POST("/api/logs", PostLogs(mockSvc))

	req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostLogs_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		InsertFn: func(ctx context.Context, entry map[string]interface{}) (int64, error) {
			return 0, assert.AnError
		},
	}

	router.POST("/api/logs", PostLogs(mockSvc))

	body := map[string]interface{}{
		"service": "portal",
		"level":   "info",
		"message": "test",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetLogs_InvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		QueryFn: func(ctx context.Context, filters map[string]interface{}, page map[string]int) ([]interface{}, error) {
			return []interface{}{}, nil
		},
	}

	router.GET("/api/logs", GetLogs(mockSvc))

	req := httptest.NewRequest("GET", "/api/logs?limit=abc&offset=xyz", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetLogs_QueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		QueryFn: func(ctx context.Context, filters map[string]interface{}, page map[string]int) ([]interface{}, error) {
			return nil, assert.AnError
		},
	}

	router.GET("/api/logs", GetLogs(mockSvc))

	req := httptest.NewRequest("GET", "/api/logs", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetLogByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockSvc := &MockLogService{}

	router.GET("/api/logs/:id", GetLogByID(mockSvc))

	req := httptest.NewRequest("GET", "/api/logs/abc", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetLogByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		GetByIDFn: func(ctx context.Context, id int64) (interface{}, error) {
			return nil, assert.AnError
		},
	}

	router.GET("/api/logs/:id", GetLogByID(mockSvc))

	req := httptest.NewRequest("GET", "/api/logs/999", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetStats_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		StatsFn: func(ctx context.Context) (map[string]interface{}, error) {
			return nil, assert.AnError
		},
	}

	router.GET("/api/logs/stats", GetStats(mockSvc))

	req := httptest.NewRequest("GET", "/api/logs/stats", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteLogs_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockSvc := &MockLogService{}

	router.DELETE("/api/logs", DeleteLogs(mockSvc))

	req := httptest.NewRequest("DELETE", "/api/logs", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteLogs_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockSvc := &MockLogService{
		DeleteFn: func(ctx context.Context, filters map[string]interface{}) (int64, error) {
			return 0, assert.AnError
		},
	}

	router.DELETE("/api/logs", DeleteLogs(mockSvc))

	body := map[string]interface{}{"before": "2025-01-01"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("DELETE", "/api/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
