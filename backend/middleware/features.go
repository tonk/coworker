package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// RequireFeature returns middleware that checks the user has the given feature enabled.
// feature is one of "board_enabled", "chat_enabled", "helpdesk_enabled", "time_tracking_enabled".
func RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		role := GetGlobalRole(c)
		if role == "admin" {
			c.Next()
			return
		}
		var user models.User
		if err := database.DB.Select(feature).First(&user, userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "feature not available"})
			return
		}
		var enabled bool
		switch feature {
		case "board_enabled":
			enabled = user.BoardEnabled
		case "chat_enabled":
			enabled = user.ChatEnabled
		case "helpdesk_enabled":
			enabled = user.HelpdeskEnabled
		case "time_tracking_enabled":
			enabled = user.TimeTrackingEnabled
		}
		// For time tracking, also allow time_tracking_viewer
		if !enabled && feature == "time_tracking_enabled" {
			if !user.TimeTrackingViewer {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "time tracking not enabled"})
				return
			}
		} else if !enabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "feature not available"})
			return
		}
		c.Next()
	}
}
