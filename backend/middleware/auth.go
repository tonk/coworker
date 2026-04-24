package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/services"
)

const (
	ContextUserID     = "user_id"
	ContextUsername   = "username"
	ContextGlobalRole = "global_role"
)

func Auth(authSvc *services.AuthService) gin.HandlerFunc {
	apiKeyAuth := APIKeyAuth()
	return func(c *gin.Context) {
		// 1. httpOnly cookie — browser clients
		if cookieToken, err := c.Cookie("access_token"); err == nil && cookieToken != "" {
			claims, err := authSvc.ValidateToken(cookieToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.Set(ContextUserID, claims.UserID)
			c.Set(ContextUsername, claims.Username)
			c.Set(ContextGlobalRole, claims.GlobalRole)
			c.Next()
			return
		}

		// 2. Authorization: Bearer header (API / Tauri clients)
		header := c.GetHeader("Authorization")
		// Also accept ?token= query param (needed for <img> tags which can't set headers)
		if !strings.HasPrefix(header, "Bearer ") {
			if t := c.Query("token"); t != "" {
				header = "Bearer " + t
			}
		}
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := authSvc.ValidateToken(tokenStr)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.Set(ContextUserID, claims.UserID)
			c.Set(ContextUsername, claims.Username)
			c.Set(ContextGlobalRole, claims.GlobalRole)
			c.Next()
			return
		}

		// 3. X-API-Key header (CI/CD ticket API)
		if c.GetHeader("X-API-Key") != "" {
			apiKeyAuth(c)
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ContextGlobalRole)
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		c.Next()
	}
}

// MetricsAuth allows admin and metrics roles (read-only Prometheus scraper accounts).
func MetricsAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ContextGlobalRole)
		if role != "admin" && role != "metrics" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// BackupAuth allows admin and backup roles (automated backup accounts).
func BackupAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ContextGlobalRole)
		if role != "admin" && role != "backup" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	v, _ := c.Get(ContextUserID)
	id, _ := v.(uint)
	return id
}

func GetGlobalRole(c *gin.Context) string {
	v, _ := c.Get(ContextGlobalRole)
	role, _ := v.(string)
	return role
}
