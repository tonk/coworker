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

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

var attachmentCfg *config.Config
var attachmentAuthSvc *services.AuthService

// InitAttachments stores the config reference for upload settings.
func InitAttachments(cfg *config.Config) {
	attachmentCfg = cfg
}

// InitAttachmentAuth stores the auth service for self-auth in DownloadAttachment.
func InitAttachmentAuth(svc *services.AuthService) {
	attachmentAuthSvc = svc
}

// claimsForAttachment resolves the caller's identity from cookie, Bearer header,
// or short-lived media ticket. Returns nil when the request is unauthenticated.
func claimsForAttachment(c *gin.Context) *services.Claims {
	if attachmentAuthSvc == nil {
		return nil
	}
	if cookieToken, err := c.Cookie("access_token"); err == nil && cookieToken != "" {
		if claims, err := attachmentAuthSvc.ValidateToken(cookieToken); err == nil {
			return claims
		}
	}
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		if claims, err := attachmentAuthSvc.ValidateToken(auth[7:]); err == nil {
			return claims
		}
	}
	if ticket := c.Query("ticket"); ticket != "" {
		if claims, err := attachmentAuthSvc.ValidateMediaTicket(ticket); err == nil {
			return claims
		}
	}
	return nil
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
	validTypes := map[string]bool{"chat_message": true, "conv_message": true, "card_comment": true, "card": true, "ticket_message": true, "ticket": true}
	if !validTypes[ownerType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_type"})
		return
	}
	ownerID, err := strconv.ParseUint(ownerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
		return
	}

	// Verify the uploader has access to the claimed owner entity before accepting the file.
	if middleware.GetGlobalRole(c) != "admin" {
		probe := models.Attachment{OwnerType: ownerType, OwnerID: uint(ownerID)}
		if err := checkAttachmentAccess(probe, userID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
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

	// Detect MIME type from the actual file bytes — ignores the client-supplied Content-Type header.
	mimeType := "application/octet-stream"
	if f, err := os.Open(dest); err == nil {
		if mt, err := mimetype.DetectReader(f); err == nil {
			mimeType = mt.String()
		}
		f.Close()
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
// Auth: httpOnly cookie, Bearer header, or short-lived media ticket (?ticket=).
// This route is registered outside the protected middleware group so it can handle
// all three auth paths without putting the JWT in the URL.
func DownloadAttachment(c *gin.Context) {
	callerClaims := claimsForAttachment(c)
	if callerClaims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
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

	// IDOR check: verify that the requesting user has access to the parent entity.
	// Admins bypass all membership checks.
	if callerClaims.GlobalRole != "admin" {
		if err := checkAttachmentAccess(a, callerClaims.UserID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	uploadDir := "./uploads"
	if attachmentCfg != nil && attachmentCfg.UploadDir != "" {
		uploadDir = attachmentCfg.UploadDir
	}
	path := filepath.Join(uploadDir, a.StoredName)

	safe := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(a.Filename)
	if strings.HasPrefix(a.MimeType, "image/") {
		c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, safe))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, safe))
	}
	c.Header("Content-Type", a.MimeType)
	c.File(path)
}

// checkAttachmentAccess verifies that userID is allowed to access attachment a
// based on its owner_type and owner_id.
func checkAttachmentAccess(a models.Attachment, userID uint) error {
	switch a.OwnerType {
	case "card":
		var card models.Card
		if err := database.DB.Select("project_id").First(&card, a.OwnerID).Error; err != nil {
			return services.ErrForbidden
		}
		return services.RequireProjectRole(card.ProjectID, userID, "", "viewer")
	case "card_comment":
		var comment models.CardComment
		if err := database.DB.Select("card_id").First(&comment, a.OwnerID).Error; err != nil {
			return services.ErrForbidden
		}
		var card models.Card
		if err := database.DB.Select("project_id").First(&card, comment.CardID).Error; err != nil {
			return services.ErrForbidden
		}
		return services.RequireProjectRole(card.ProjectID, userID, "", "viewer")
	case "chat_message":
		var msg models.ChatMessage
		if err := database.DB.Select("project_id").First(&msg, a.OwnerID).Error; err != nil {
			return services.ErrForbidden
		}
		return services.RequireProjectRole(msg.ProjectID, userID, "", "viewer")
	case "conv_message":
		var msg models.ConversationMessage
		if err := database.DB.Select("conversation_id").First(&msg, a.OwnerID).Error; err != nil {
			return services.ErrForbidden
		}
		var count int64
		database.DB.Model(&models.ConversationMember{}).
			Where("conversation_id = ? AND user_id = ?", msg.ConversationID, userID).
			Count(&count)
		if count == 0 {
			return services.ErrForbidden
		}
		return nil
	case "ticket_message":
		var tm models.TicketMessage
		if err := database.DB.Select("ticket_id").First(&tm, a.OwnerID).Error; err != nil {
			return services.ErrForbidden
		}
		var ticket models.Ticket
		if err := database.DB.Select("customer_id").First(&ticket, tm.TicketID).Error; err != nil {
			return services.ErrForbidden
		}
		if ticket.CustomerID == nil {
			return nil
		}
		return requireCustomerAccess(*ticket.CustomerID, userID, "")
	case "ticket":
		var ticket models.Ticket
		if err := database.DB.Select("customer_id").First(&ticket, a.OwnerID).Error; err != nil {
			return services.ErrForbidden
		}
		if ticket.CustomerID == nil {
			return nil
		}
		return requireCustomerAccess(*ticket.CustomerID, userID, "")
	default:
		return services.ErrForbidden
	}
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
