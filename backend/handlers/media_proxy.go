package handlers

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxProxiedImageBytes = 5 * 1024 * 1024 // 5 MiB

// mediaProxyClient uses a custom dialer that connects to a pre-resolved IP
// address, preventing DNS-rebinding attacks: the IP is checked for safety
// before dialing, so a DNS change after the check cannot redirect the
// connection to a private address.
var mediaProxyClient = &http.Client{
	Timeout: 8 * time.Second,
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// addr is host:port — the host is already the safe IP chosen by
			// resolveAndVerify, so we dial it directly.
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, addr)
		},
	},
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

	// Re-encode the query string so that spaces, commas, etc. in parameter
	// values (e.g. DiceBear seed names) don't produce an invalid HTTP request.
	if u.RawQuery != "" {
		if q, qErr := url.ParseQuery(u.RawQuery); qErr == nil {
			u.RawQuery = q.Encode()
		}
	}

	// Resolve once, verify safety, then build a dial-target URL that uses the
	// resolved IP.  This prevents DNS-rebinding: a DNS change after the check
	// cannot redirect the connection to a private address.
	safeIP, err := resolveAndVerify(u.Hostname())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsafe image host"})
		return
	}

	// Build a fetch URL that dials the resolved IP directly.
	fetchURL := *u
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	fetchURL.Host = fmt.Sprintf("%s:%s", safeIP, port)

	req, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, fetchURL.String(), nil)
	req.Host = u.Host // preserve original Host header for SNI / virtual hosting
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

// resolveAndVerify resolves host to an IP, verifies every resolved address is
// a public routable address, and returns the first IP as a string.  Callers
// should use the returned IP as the dial target (not the hostname) to prevent
// DNS-rebinding: once we verify the IP here it cannot change under us.
func resolveAndVerify(host string) (string, error) {
	if host == "" {
		return "", fmt.Errorf("empty host")
	}
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "localhost" {
		return "", fmt.Errorf("localhost not allowed")
	}

	// IP literal — verify directly without a DNS round-trip.
	if ip := net.ParseIP(h); ip != nil {
		if !isPublicIP(ip) {
			return "", fmt.Errorf("private ip")
		}
		return ip.String(), nil
	}

	// Resolve DNS and ensure every address is public.
	ips, err := net.LookupIP(h)
	if err != nil || len(ips) == 0 {
		return "", fmt.Errorf("dns lookup failed")
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return "", fmt.Errorf("private ip in dns response")
		}
	}
	return ips[0].String(), nil
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
