// Package debug provides utilities for debugging and inspecting application state.
package debug

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

// RouteInfo represents a single route in the application
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	HandlerName string `json:"handler_name"`
}

// RoutesResponse is the JSON response for the debug routes endpoint
type RoutesResponse struct {
	Service string      `json:"service"`
	Count   int         `json:"count"`
	Routes  []RouteInfo `json:"routes"`
}

// RegisterDebugRoutes adds debug endpoints to the router
// These endpoints are only registered in development mode (ENV != production)
func RegisterDebugRoutes(router *gin.Engine, serviceName string) {
	// Only enable in development/testing
	env := os.Getenv("ENV")
	if env == "production" {
		return
	}

	// Register the debug routes endpoint
	router.GET("/debug/routes", func(c *gin.Context) {
		GetRoutesHandler(c, router, serviceName)
	})
}

// GetRoutesHandler returns all registered routes in the application
func GetRoutesHandler(c *gin.Context, router *gin.Engine, serviceName string) {
	routes := router.Routes()

	routeInfos := make([]RouteInfo, 0, len(routes))
	for _, route := range routes {
		// Skip the debug endpoint itself to avoid confusion
		if route.Path == "/debug/routes" {
			continue
		}

		routeInfos = append(routeInfos, RouteInfo{
			Method:      route.Method,
			Path:        route.Path,
			HandlerName: route.Handler,
		})
	}

	response := RoutesResponse{
		Service: serviceName,
		Count:   len(routeInfos),
		Routes:  routeInfos,
	}

	c.JSON(http.StatusOK, response)
}

// HTTPRouteRegistry is a simple registry for net/http based services
type HTTPRouteRegistry struct {
	serviceName string
	routes      []RouteInfo
	mu          sync.RWMutex
}

// NewHTTPRouteRegistry creates a new route registry for net/http services
func NewHTTPRouteRegistry(serviceName string) *HTTPRouteRegistry {
	return &HTTPRouteRegistry{
		serviceName: serviceName,
		routes:      make([]RouteInfo, 0),
	}
}

// Register adds a route to the registry
func (r *HTTPRouteRegistry) Register(method, path string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes = append(r.routes, RouteInfo{
		Method:      method,
		Path:        path,
		HandlerName: "http.HandlerFunc",
	})
}

// Handler returns an http.HandlerFunc that lists all registered routes
func (r *HTTPRouteRegistry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Only enable in development/testing
		env := os.Getenv("ENV")
		if env == "production" {
			http.Error(w, "Not available in production", http.StatusNotFound)
			return
		}

		r.mu.RLock()
		defer r.mu.RUnlock()

		response := RoutesResponse{
			Service: r.serviceName,
			Count:   len(r.routes),
			Routes:  r.routes,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
