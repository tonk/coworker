package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOrigins_empty(t *testing.T) {
	origins := ParseOrigins("")
	assert.Contains(t, origins, "tauri://localhost")
	assert.Contains(t, origins, "https://tauri.localhost")
	assert.Contains(t, origins, "http://tauri.localhost")
	assert.Len(t, origins, 3)
}

func TestParseOrigins_single(t *testing.T) {
	origins := ParseOrigins("http://localhost:5173")
	assert.Contains(t, origins, "tauri://localhost")
	assert.Contains(t, origins, "http://localhost:5173")
	assert.Len(t, origins, 4)
}

func TestParseOrigins_multiple(t *testing.T) {
	origins := ParseOrigins("http://localhost:5173, https://example.com")
	assert.Contains(t, origins, "http://localhost:5173")
	assert.Contains(t, origins, "https://example.com")
	assert.Len(t, origins, 5)
}

func TestParseOrigins_trimsSpaces(t *testing.T) {
	origins := ParseOrigins("  http://a.com , http://b.com  ")
	assert.Contains(t, origins, "http://a.com")
	assert.Contains(t, origins, "http://b.com")
}

func TestParseOrigins_skipsEmpty(t *testing.T) {
	origins := ParseOrigins("http://a.com,,,http://b.com")
	assert.Contains(t, origins, "http://a.com")
	assert.Contains(t, origins, "http://b.com")
	assert.Len(t, origins, 5)
}

func TestParseOrigins_alwaysIncludesTauri(t *testing.T) {
	origins := ParseOrigins("http://example.com")
	for _, o := range tauriOrigins {
		assert.Contains(t, origins, o)
	}
}
