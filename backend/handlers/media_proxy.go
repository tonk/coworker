package handlers

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxProxiedImageBytes = 5 * 1024 * 1024 // 5 MiB

var mediaProxyClient = &http.Client{
	Timeout: 8 * time.Second,
}

// ProxyImage fetches an external image and serves it from same-origin to comply
// with strict CSP policies (img-src 'self' ...).
//
// GET /api/v1/media/proxy?url=https://example.com/image.png
func ProxyImage(c *gin.Context) {
	if !IsExternalImageProxyEnabled() {
		c.JSON(http.StatusForbidden, gin.H{"error": "external image proxy is disabled"})
		return
	}

	raw := strings.TrimSpace(c.Query("url"))
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url required"})
		return
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image url"})
		return
	}

	if !isSafeRemoteHost(u.Hostname()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsafe image host"})
		return
	}

	req, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, u.String(), nil)
	req.Header.Set("User-Agent", "WarmDesk-MediaProxy/1.0")
	resp, err := mediaProxyClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "image fetch failed"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "image fetch failed"})
		return
	}

	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is not an image"})
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "private, max-age=3600")
	c.Status(http.StatusOK)
	_, _ = io.CopyN(c.Writer, resp.Body, maxProxiedImageBytes)
}

func isSafeRemoteHost(host string) bool {
	if host == "" {
		return false
	}
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "localhost" {
		return false
	}

	// If hostname is already an IP literal, reject private/link-local/loopback.
	if ip := net.ParseIP(h); ip != nil {
		return isPublicIP(ip)
	}

	// Resolve DNS and ensure all resolved addresses are public.
	ips, err := net.LookupIP(h)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return false
		}
	}
	return true
}

func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return !ip.IsLoopback() &&
		!ip.IsPrivate() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsMulticast() &&
		!ip.IsUnspecified()
}
