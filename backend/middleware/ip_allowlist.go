package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// IPAllowlist restricts API-key-authenticated requests to the IPs listed in the
// allowed_ips system setting. JWT/cookie-authenticated requests (browser clients)
// are always allowed through — the setting is intended to restrict automation and
// CI/CD tooling, not human users. An empty setting means no restriction at all.
// Accepts comma-separated IPv4/IPv6 addresses and CIDR ranges, e.g. "10.0.0.0/8, 192.168.1.5".
func IPAllowlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only API-key requests are subject to the allowlist.
		// Browser/JWT clients are never blocked regardless of their IP.
		if c.GetHeader("X-API-Key") == "" {
			c.Next()
			return
		}

		var row models.SystemSetting
		if err := database.DB.Where("key = ?", "allowed_ips").First(&row).Error; err != nil || strings.TrimSpace(row.Value) == "" {
			c.Next()
			return
		}

		clientIP := net.ParseIP(strings.TrimSpace(c.ClientIP()))
		if clientIP == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		for _, entry := range strings.Split(row.Value, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			if strings.Contains(entry, "/") {
				_, network, err := net.ParseCIDR(entry)
				if err == nil && network.Contains(clientIP) {
					c.Next()
					return
				}
			} else {
				if allowed := net.ParseIP(entry); allowed != nil && allowed.Equal(clientIP) {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}
