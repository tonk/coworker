package handlers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	metaTagRe   = regexp.MustCompile(`(?i)<meta[^>]+>`)
	propAttrRe  = regexp.MustCompile(`(?i)\bproperty=["']([^"']+)["']`)
	nameAttrRe  = regexp.MustCompile(`(?i)\bname=["']([^"']+)["']`)
	contAttrRe  = regexp.MustCompile(`(?i)\bcontent=["']([^"']*)["']`)
	titleElemRe = regexp.MustCompile(`(?i)<title[^>]*>([^<]{1,300})</title>`)
)

// LinkPreview fetches a URL and returns its Open Graph metadata.
func LinkPreview(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url required"})
		return
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}

	// SSRF: reject private/loopback hostnames before DNS lookup
	if isPrivateHostname(parsed.Hostname()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "private addresses not allowed"})
		return
	}

	// SSRF: resolve once, verify the IP is public, then dial by IP to prevent DNS rebinding.
	safeIP, err := resolveAndVerify(parsed.Hostname())
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "private addresses not allowed"})
		return
	}

	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	fetchURL := *parsed
	if strings.Contains(safeIP, ":") {
		fetchURL.Host = fmt.Sprintf("[%s]:%s", safeIP, port)
	} else {
		fetchURL.Host = fmt.Sprintf("%s:%s", safeIP, port)
	}

	req, err := http.NewRequest(http.MethodGet, fetchURL.String(), nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad url"})
		return
	}
	req.Host = parsed.Host
	req.Header.Set("User-Agent", "WarmDesk-LinkPreview/1.0")
	req.Header.Set("Accept", "text/html")

	resp, err := mediaProxyClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch"})
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(strings.ToLower(ct), "text/html") {
		c.JSON(http.StatusOK, gin.H{"url": rawURL, "title": "", "description": "", "image": "", "site_name": parsed.Hostname()})
		return
	}

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	html := string(bodyBytes)
	if i := strings.Index(strings.ToLower(html), "</head>"); i >= 0 {
		html = html[:i]
	}

	metas := extractMetaTags(html)
	title := firstNonEmpty(metas["og:title"], metas["twitter:title"])
	if title == "" {
		if m := titleElemRe.FindStringSubmatch(html); len(m) > 1 {
			title = strings.TrimSpace(m[1])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"url":         rawURL,
		"title":       title,
		"description": firstNonEmpty(metas["og:description"], metas["twitter:description"], metas["description"]),
		"image":       firstNonEmpty(metas["og:image"], metas["twitter:image"]),
		"site_name":   firstNonEmpty(metas["og:site_name"], parsed.Hostname()),
	})
}

func extractMetaTags(html string) map[string]string {
	out := map[string]string{}
	for _, tag := range metaTagRe.FindAllString(html, -1) {
		var key string
		if m := propAttrRe.FindStringSubmatch(tag); len(m) > 1 {
			key = strings.ToLower(m[1])
		} else if m := nameAttrRe.FindStringSubmatch(tag); len(m) > 1 {
			key = strings.ToLower(m[1])
		}
		if key == "" {
			continue
		}
		if m := contAttrRe.FindStringSubmatch(tag); len(m) > 1 && out[key] == "" {
			out[key] = m[1]
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func isPrivateHostname(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0", "[::1]":
		return true
	}
	return false
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, cidr := range []string{
		"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12",
		"192.168.0.0/16", "169.254.0.0/16", "::1/128", "fc00::/7",
	} {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}
