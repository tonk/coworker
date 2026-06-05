package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// AuthenticateAPIKey resolves raw against the database and populates the gin
// context with the same keys as JWT Auth. Returns false (and aborts the
// request) on any auth failure so callers can return immediately.
func AuthenticateAPIKey(c *gin.Context, raw string) bool {
	sum := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(sum[:])

	var key models.APIKey
	if err := database.DB.Preload("User").Where("key_hash = ?", hash).First(&key).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
		return false
	}

	if !key.User.IsActive {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account deactivated"})
		return false
	}

	// Enforce project scope when the key is project-scoped
	if key.ProjectID != nil {
		slug := c.Param("projectSlug")
		if slug == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "key is scoped to a specific project"})
			return false
		}
		var project models.Project
		if err := database.DB.Where("slug = ?", slug).First(&project).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return false
		}
		if project.ID != *key.ProjectID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "key not valid for this project"})
			return false
		}
	}

	// Update last used timestamp (best-effort)
	database.DB.Model(&key).Update("last_used_at", time.Now())

	c.Set(ContextUserID, key.UserID)
	c.Set(ContextUsername, key.User.Username)
	c.Set(ContextGlobalRole, key.User.GlobalRole)
	return true
}

// APIKeyAuth authenticates requests using the X-API-Key header.
// On success it sets the same context keys as JWT Auth so handlers work unchanged.
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-API-Key")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			return
		}
		if AuthenticateAPIKey(c, raw) {
			c.Next()
		}
	}
}
