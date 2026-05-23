package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetAppInfoReader() {
	appInfoReader = nil
}

func TestGetAppInfo_fallback(t *testing.T) {
	resetAppInfoReader()
	ver, company, logo, instance := getAppInfo()
	assert.Equal(t, "dev", ver)
	assert.Equal(t, "WarmDesk", company)
	assert.Equal(t, "", logo)
	assert.Equal(t, "", instance)
}

func TestGetAppInfo_custom(t *testing.T) {
	SetAppInfoReader(func() (string, string, string, string) {
		return "v1.2.3", "Acme Inc", "https://example.com/logo.png", "https://acme.warmdesk.com"
	})
	defer resetAppInfoReader()

	ver, company, logo, instance := getAppInfo()
	assert.Equal(t, "v1.2.3", ver)
	assert.Equal(t, "Acme Inc", company)
	assert.Equal(t, "https://example.com/logo.png", logo)
	assert.Equal(t, "https://acme.warmdesk.com", instance)
}

func TestWrapHTML(t *testing.T) {
	resetAppInfoReader()
	html := WrapHTML("Test Subtitle", "<p>Hello World</p>")

	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "WarmDesk")
	assert.Contains(t, html, "Test Subtitle")
	assert.Contains(t, html, "<p>Hello World</p>")
	assert.Contains(t, html, "vdev")
	assert.Contains(t, html, warmDeskLogoDataURI)
	assert.NotContains(t, html, `href="`)
}

func TestWrapHTML_withLogoAndURL(t *testing.T) {
	SetAppInfoReader(func() (string, string, string, string) {
		return "v0.7.7", "Acme Corp", "https://example.com/logo.png", "https://acme.example.com"
	})
	defer resetAppInfoReader()

	html := WrapHTML("Notification", "<p>Body</p>")
	assert.Contains(t, html, "https://example.com/logo.png")
	assert.Contains(t, html, "https://acme.example.com")
	assert.Contains(t, html, "https://example.com/logo.png")
	assert.Contains(t, html, "https://acme.example.com")
	assert.Contains(t, html, "Acme Corp")
	assert.Contains(t, html, "v0.7.7")
}

func TestWrapHTML_stripsLeadingV(t *testing.T) {
	SetAppInfoReader(func() (string, string, string, string) {
		return "v1.0.0", "Test", "", ""
	})
	defer resetAppInfoReader()

	html := WrapHTML("Sub", "<p>Body</p>")
	assert.Contains(t, html, "v1.0.0")
	assert.NotContains(t, html, "vv1")
}

func TestWrapHTML_emptyCompanyName(t *testing.T) {
	SetAppInfoReader(func() (string, string, string, string) {
		return "dev", "", "", ""
	})
	defer resetAppInfoReader()

	html := WrapHTML("Sub", "<p>Body</p>")
	assert.Contains(t, html, "WarmDesk")
}

func TestWrapText(t *testing.T) {
	resetAppInfoReader()
	text := WrapText("Test Subtitle", "Hello World")

	assert.Contains(t, text, "WarmDesk")
	assert.Contains(t, text, "Test Subtitle")
	assert.Contains(t, text, "Hello World")
	assert.Contains(t, text, "vdev")
	assert.NotContains(t, text, "http")
}

func TestWrapText_withInstanceURL(t *testing.T) {
	SetAppInfoReader(func() (string, string, string, string) {
		return "v2.0.0", "My Company", "", "https://my.warmdesk.com"
	})
	defer resetAppInfoReader()

	text := WrapText("Alert", "System update")
	assert.Contains(t, text, "My Company")
	assert.Contains(t, text, "https://my.warmdesk.com")
	assert.Contains(t, text, "v2.0.0")
}

func TestWrapText_emptyCompanyName(t *testing.T) {
	SetAppInfoReader(func() (string, string, string, string) {
		return "dev", "", "", ""
	})
	defer resetAppInfoReader()

	text := WrapText("Sub", "Body")
	assert.Contains(t, text, "WarmDesk")
}

func TestWrapHTML_containsSeparator(t *testing.T) {
	resetAppInfoReader()
	text := WrapText("Sub", "Body")
	assert.True(t, strings.Count(text, "-----------------------------------------------") >= 0)
}
