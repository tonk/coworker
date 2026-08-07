package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
}

// UploadImage POST /api/v1/upload/image
// Accepts a multipart image file and returns its public URL.
// Used for user avatars, customer logos, and company branding.
func UploadImage(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}

	// Validate file content against magic bytes — the client-supplied Content-Type
	// header is ignored, matching the convention used for attachments.
	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}
	detected, err := mimetype.DetectReader(f)
	f.Close()
	if err != nil || !allowedImageTypes[detected.String()] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file must be an image (jpeg, png, gif, webp, svg)"})
		return
	}
	mimeType := detected.String()

	maxMB := int64(25)
	if attachmentCfg != nil && attachmentCfg.MaxUploadMB > 0 {
		maxMB = attachmentCfg.MaxUploadMB
	}

	if fh.Size > maxMB*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("image too large (max %d MB)", maxMB)})
		return
	}

	uploadDir := "./uploads"
	if attachmentCfg != nil && attachmentCfg.UploadDir != "" {
		uploadDir = attachmentCfg.UploadDir
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload dir"})
		return
	}

	ext := filepath.Ext(fh.Filename)
	if ext == "" {
		// Derive extension from MIME type
		switch mimeType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		case "image/svg+xml":
			ext = ".svg"
		}
	}

	storedName := randomHex(16) + ext
	dest := filepath.Join(uploadDir, storedName)
	if err := c.SaveUploadedFile(fh, dest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + storedName})
}

// isValidAvatarURL reports whether url is acceptable as an avatar/logo value —
// either a local upload path or a fully-qualified http(s) URL. An empty
// string is valid (it means "no avatar").
func isValidAvatarURL(url string) bool {
	if url == "" {
		return true
	}
	lower := strings.ToLower(url)
	return strings.HasPrefix(url, "/uploads/") ||
		strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://")
}

// deleteOldUpload removes the file behind a previous /uploads/... URL when it's
// being replaced by a different one (or cleared), so avatars/logos don't pile up
// as orphaned files every time they're changed. No-op for empty, unchanged, or
// non-local (external) URLs.
func deleteOldUpload(oldURL, newURL string) {
	if oldURL == "" || oldURL == newURL || !strings.HasPrefix(oldURL, "/uploads/") {
		return
	}
	uploadDir := "./uploads"
	if attachmentCfg != nil && attachmentCfg.UploadDir != "" {
		uploadDir = attachmentCfg.UploadDir
	}
	name := filepath.Base(oldURL)
	_ = os.Remove(filepath.Join(uploadDir, name))
}
