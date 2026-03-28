package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/config"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
)

var attachmentCfg *config.Config

// InitAttachments stores the config reference for upload settings.
func InitAttachments(cfg *config.Config) {
	attachmentCfg = cfg
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// UploadAttachment POST /api/v1/attachments
// Form fields: file, owner_type, owner_id
func UploadAttachment(c *gin.Context) {
	userID := middleware.GetUserID(c)

	ownerType := c.PostForm("owner_type")
	ownerIDStr := c.PostForm("owner_id")
	if ownerType == "" || ownerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_type and owner_id required"})
		return
	}
	validTypes := map[string]bool{"chat_message": true, "conv_message": true, "card_comment": true, "card": true}
	if !validTypes[ownerType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_type"})
		return
	}
	ownerID, err := strconv.ParseUint(ownerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
		return
	}

	maxMB := int64(25)
	if attachmentCfg != nil && attachmentCfg.MaxUploadMB > 0 {
		maxMB = attachmentCfg.MaxUploadMB
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	if fh.Size > maxMB*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("file too large (max %dMB)", maxMB)})
		return
	}

	// Detect MIME type from content type header
	mimeType := fh.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
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
	storedName := randomHex(16) + ext
	dest := filepath.Join(uploadDir, storedName)

	if err := c.SaveUploadedFile(fh, dest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	attachment := models.Attachment{
		OwnerType:  ownerType,
		OwnerID:    uint(ownerID),
		UploaderID: userID,
		Filename:   fh.Filename,
		StoredName: storedName,
		MimeType:   mimeType,
		SizeBytes:  fh.Size,
	}
	if err := database.DB.Create(&attachment).Error; err != nil {
		os.Remove(dest)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save attachment"})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

// DownloadAttachment GET /api/v1/attachments/:id
func DownloadAttachment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var a models.Attachment
	if err := database.DB.First(&a, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	uploadDir := "./uploads"
	if attachmentCfg != nil && attachmentCfg.UploadDir != "" {
		uploadDir = attachmentCfg.UploadDir
	}
	path := filepath.Join(uploadDir, a.StoredName)

	if strings.HasPrefix(a.MimeType, "image/") {
		c.Header("Content-Disposition", "inline; filename=\""+a.Filename+"\"")
	} else {
		c.Header("Content-Disposition", "attachment; filename=\""+a.Filename+"\"")
	}
	c.Header("Content-Type", a.MimeType)
	c.File(path)
}

// DeleteAttachment DELETE /api/v1/attachments/:id
func DeleteAttachment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var a models.Attachment
	if err := database.DB.First(&a, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if a.UploaderID != userID && globalRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	uploadDir := "./uploads"
	if attachmentCfg != nil && attachmentCfg.UploadDir != "" {
		uploadDir = attachmentCfg.UploadDir
	}
	os.Remove(filepath.Join(uploadDir, a.StoredName))
	database.DB.Delete(&a)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// LoadAttachments fetches attachments for a set of owner IDs and groups them by owner_id.
func LoadAttachments(ownerType string, ownerIDs []uint) map[uint][]models.Attachment {
	result := make(map[uint][]models.Attachment)
	if len(ownerIDs) == 0 {
		return result
	}
	var attachments []models.Attachment
	database.DB.Where("owner_type = ? AND owner_id IN ?", ownerType, ownerIDs).Find(&attachments)
	for _, a := range attachments {
		result[a.OwnerID] = append(result[a.OwnerID], a)
	}
	return result
}
