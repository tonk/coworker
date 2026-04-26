package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// IPAllowlist blocks requests whose client IP is not in the allowed_ips system setting.
// An empty setting means no restriction — all IPs are allowed.
// Accepts comma-separated IPv4/IPv6 addresses and CIDR ranges, e.g. "10.0.0.0/8, 192.168.1.5".
func IPAllowlist() gin.HandlerFunc {
	return func(c *gin.Context) {
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
