package debug

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterDebugRoutes_Development(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set development environment
	os.Setenv("ENV", "development")
	defer os.Unsetenv("ENV")

	router := gin.New()
	RegisterDebugRoutes(router, "test-service")

	// Verify debug route is registered
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/debug/routes" && route.Method == "GET" {
			found = true
			break
		}
	}

	assert.True(t, found, "Debug route should be registered in development mode")
}

func TestRegisterDebugRoutes_Production(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set production environment
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	router := gin.New()
	RegisterDebugRoutes(router, "test-service")

	// Verify debug route is NOT registered
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/debug/routes" {
			found = true
			break
		}
	}

	assert.False(t, found, "Debug route should NOT be registered in production mode")
}

func TestGetRoutesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set development environment
	os.Setenv("ENV", "development")
	defer os.Unsetenv("ENV")

	router := gin.New()
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"test": "data"})
	})
	router.POST("/api/create", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"created": true})
	})

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/debug/routes", http.NoBody)

	GetRoutesHandler(c, router, "test-service")

	assert.Equal(t, http.StatusOK, w.Code)

	var response RoutesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "test-service", response.Service)
	assert.Equal(t, 2, response.Count)
	assert.Len(t, response.Routes, 2)

	// Check that routes are present
	routeMap := make(map[string]string)
	for _, route := range response.Routes {
		routeMap[route.Path] = route.Method
	}

	assert.Equal(t, "GET", routeMap["/api/test"])
	assert.Equal(t, "POST", routeMap["/api/create"])
}

func TestNewHTTPRouteRegistry(t *testing.T) {
	registry := NewHTTPRouteRegistry("test-service")

	assert.NotNil(t, registry)
	assert.Equal(t, "test-service", registry.serviceName)
	assert.Empty(t, registry.routes)
}

func TestHTTPRouteRegistry_Register(t *testing.T) {
	registry := NewHTTPRouteRegistry("test-service")

	registry.Register("GET", "/api/test")
	registry.Register("POST", "/api/create")

	assert.Len(t, registry.routes, 2)
	assert.Equal(t, "GET", registry.routes[0].Method)
	assert.Equal(t, "/api/test", registry.routes[0].Path)
	assert.Equal(t, "POST", registry.routes[1].Method)
	assert.Equal(t, "/api/create", registry.routes[1].Path)
}

func TestHTTPRouteRegistry_Handler_Development(t *testing.T) {
	// Set development environment
	os.Setenv("ENV", "development")
	defer os.Unsetenv("ENV")

	registry := NewHTTPRouteRegistry("test-service")
	registry.Register("GET", "/api/test")
	registry.Register("POST", "/api/create")

	handler := registry.Handler()

	req := httptest.NewRequest(http.MethodGet, "/debug/routes", http.NoBody)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response RoutesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "test-service", response.Service)
	assert.Equal(t, 2, response.Count)
	assert.Len(t, response.Routes, 2)
}

func TestHTTPRouteRegistry_Handler_Production(t *testing.T) {
	// Set production environment
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	registry := NewHTTPRouteRegistry("test-service")
	registry.Register("GET", "/api/test")

	handler := registry.Handler()

	req := httptest.NewRequest(http.MethodGet, "/debug/routes", http.NoBody)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Not available in production")
}

func TestRouteInfo(t *testing.T) {
	route := RouteInfo{
		Method:      "GET",
		Path:        "/api/test",
		HandlerName: "TestHandler",
	}

	assert.Equal(t, "GET", route.Method)
	assert.Equal(t, "/api/test", route.Path)
	assert.Equal(t, "TestHandler", route.HandlerName)
}

func TestRoutesResponse(t *testing.T) {
	response := RoutesResponse{
		Service: "test-service",
		Count:   2,
		Routes: []RouteInfo{
			{Method: "GET", Path: "/api/test", HandlerName: "TestHandler"},
			{Method: "POST", Path: "/api/create", HandlerName: "CreateHandler"},
		},
	}

	assert.Equal(t, "test-service", response.Service)
	assert.Equal(t, 2, response.Count)
	assert.Len(t, response.Routes, 2)
}
