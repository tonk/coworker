package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/services"
)

// AdminTestIMAP POST /api/v1/admin/imap/test
func AdminTestIMAP(c *gin.Context) {
	svc := services.GetDefaultService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "IMAP service not running"})
		return
	}
	cfg := GetIMAPSettings()
	if err := svc.TestConnection(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}

// AdminPollIMAP POST /api/v1/admin/imap/poll
func AdminPollIMAP(c *gin.Context) {
	svc := services.GetDefaultService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "IMAP service not running"})
		return
	}
	if err := svc.PollOnce(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Poll complete"})
}
