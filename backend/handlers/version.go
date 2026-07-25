package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var serverVersion = "dev"
var serverAppMode = ""
var serverInstanceMode = ""

// SetVersion is called by main to inject the build-time version string.
func SetVersion(v string) { serverVersion = v }

// SetAppMode is called by main to inject the configured application mode.
func SetAppMode(m string) { serverAppMode = m }

// SetInstanceMode is called by main to inject the configured instance mode
// ("" for production, "test" for a test instance).
func SetInstanceMode(m string) { serverInstanceMode = m }

// GetVersion returns the server version, application mode, and instance mode.
//
//	@Summary	Server version
//	@Tags		system
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Router		/version [get]
func GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": serverVersion, "app_mode": serverAppMode, "instance_mode": serverInstanceMode})
}
