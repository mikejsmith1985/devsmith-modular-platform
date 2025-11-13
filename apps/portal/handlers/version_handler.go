package portal_handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/version"
)

// VersionInfo represents the version information response
type VersionInfo struct {
	Service   string `json:"service"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Status    string `json:"status"`
}

// HandleVersion returns version information for the Portal service
// GET /api/portal/version
func HandleVersion(c *gin.Context) {
	info := VersionInfo{
		Service:   "portal",
		Version:   version.Version,
		Commit:    version.CommitHash,
		BuildTime: version.BuildTime,
		GoVersion: runtime.Version(),
		Status:    "healthy",
	}

	c.JSON(http.StatusOK, info)
}

// HandleVersionShort returns a short version string
// GET /version (for quick checks)
func HandleVersionShort(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version.ShortVersion(),
	})
}
