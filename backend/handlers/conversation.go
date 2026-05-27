package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	appws "github.com/tonk/warmdesk/ws"
)

// ConvMessageResponse wraps ConversationMessage with attachments and reactions.
type ConvMessageResponse struct {
	models.ConversationMessage
	Attachments []models.Attachment      `json:"attachments"`
	Reactions   []models.ReactionSummary `json:"reactions"`
}

// CreateConversation POST /conversations
// Body: {"user_ids": [2, 3], "name": "optional group name"}
// If only one other user is given and a 1-on-1 conversation already exists,
// that existing conversation is returned instead of creating a duplicate.
func CreateConversation(c *gin.Context) {
	me := middleware.GetUserID(c)

	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required,min=1"`
		Name    string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Deduplicate and strip current user from the list
	seen := map[uint]bool{me: true}
	var others []uint
	for _, id := range req.UserIDs {
		if !seen[id] {
			seen[id] = true
			others = append(others, id)
		}
	}
	if len(others) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid recipients"})
		return
	}

	isGroup := len(others) > 1

	// For 1-on-1: return existing conversation if one exists
	if !isGroup {
		otherID := others[0]
		var convIDs []uint
		database.DB.Model(&models.ConversationMember{}).
			Where("user_id = ?", me).
			Pluck("conversation_id", &convIDs)
		if len(convIDs) > 0 {
			var existingID uint
			database.DB.Model(&models.ConversationMember{}).
				Where("user_id = ? AND conversation_id IN ?", otherID, convIDs).
				Pluck("conversation_id", &existingID)
			if existingID != 0 {
				var existing models.Conversation
				database.DB.Preload("Members.User").
					Where("id = ? AND is_group = false", existingID).
					First(&existing)
				if existing.ID != 0 {
					c.JSON(http.StatusOK, existing)
					return
				}
			}
		}
	}

	conv := models.Conversation{
		Name:        req.Name,
		IsGroup:     isGroup,
		CreatedByID: me,
	}
	if err := database.DB.Create(&conv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create conversation"})
		return
	}

	now := time.Now()
	allMembers := append([]uint{me}, others...)
	for _, uid := range allMembers {
		database.DB.Create(&models.ConversationMember{
			ConversationID: conv.ID,
			UserID:         uid,
			JoinedAt:       now,
		})
	}

	database.DB.Preload("Members.User").First(&conv, conv.ID)
	c.JSON(http.StatusCreated, conv)
}

// GetConversations GET /conversations
// Returns all conversations the current user is a member of.
func GetConversations(c *gin.Context) {
	me := middleware.GetUserID(c)

	var convIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("user_id = ?", me).
		Pluck("conversation_id", &convIDs)

	var convs []models.Conversation
	if len(convIDs) > 0 {
		database.DB.Preload("Members.User").
			Where("id IN ?", convIDs).
			Order("updated_at DESC").
			Find(&convs)
	}

	c.JSON(http.StatusOK, convs)
}

// GetConversationMessages GET /conversations/:id/messages
func GetConversationMessages(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var msgs []models.ConversationMessage
	database.DB.Preload("Sender").
		Where("conversation_id = ?", convID).
		Order("created_at ASC").
		Limit(200).
		Find(&msgs)

	ids := make([]uint, len(msgs))
	for i, m := range msgs {
		ids[i] = m.ID
	}
	attachMap := LoadAttachments("conv_message", ids)
	reactMap := LoadReactionSummaries("conv_message", ids)

	out := make([]ConvMessageResponse, len(msgs))
	for i, m := range msgs {
		out[i] = ConvMessageResponse{
			ConversationMessage: m,
			Attachments:         attachMap[m.ID],
			Reactions:           reactMap[m.ID],
		}
		if out[i].Attachments == nil {
			out[i].Attachments = []models.Attachment{}
		}
		if out[i].Reactions == nil {
			out[i].Reactions = []models.ReactionSummary{}
		}
	}

	c.JSON(http.StatusOK, out)
}

// SendConversationMessage POST /conversations/:id/messages
func SendConversationMessage(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	msg := models.ConversationMessage{
		ConversationID: uint(convID),
		SenderID:       me,
		Body:           req.Body,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send"})
		return
	}

	// Bump conversation updated_at so it floats to the top of the list
	database.DB.Model(&models.Conversation{}).Where("id = ?", convID).
		Update("updated_at", time.Now())

	database.DB.Preload("Sender").First(&msg, msg.ID)

	if notifSvc != nil {
		go notifSvc.NotifyNewDM(msg, msg.Sender)
		go notifSvc.NotifyMentions(msg.Body, me, "direct message")
	}

	c.JSON(http.StatusCreated, msg)
}

// DeleteConversationMessage DELETE /conversations/:id/messages/:msgId
func DeleteConversationMessage(c *gin.Context) {
	me := middleware.GetUserID(c)
	msgID, _ := strconv.Atoi(c.Param("msgId"))

	var msg models.ConversationMessage
	if err := database.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if msg.SenderID != me {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your message"})
		return
	}

	database.DB.Model(&msg).Update("is_deleted", true)

	// Broadcast to all conversation members
	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", c.Param("id")).
		Pluck("user_id", &memberIDs)
	for _, uid := range memberIDs {
		appws.BroadcastToUser(uid, appws.Message{
			Type: appws.TypeDMMessageDeleted,
			Payload: map[string]interface{}{
				"conversation_id": msg.ConversationID,
				"id":              msg.ID,
			},
		})
	}

	c.JSON(http.StatusOK, msg)
}

// RemoveConversationMember DELETE /conversations/:id/members/:userId
func RemoveConversationMember(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))
	targetID, _ := strconv.Atoi(c.Param("userId"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}
	if uint(targetID) == me {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove yourself"})
		return
	}
	if !isMember(uint(convID), uint(targetID)) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user is not a member"})
		return
	}

	database.DB.Where("conversation_id = ? AND user_id = ?", convID, targetID).
		Delete(&models.ConversationMember{})

	// Check remaining members
	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", convID).
		Pluck("user_id", &memberIDs)

	// If only the creator is left and there are no messages, delete the conversation
	if len(memberIDs) == 1 {
		var conv models.Conversation
		database.DB.First(&conv, convID)
		if memberIDs[0] == conv.CreatedByID {
			var msgCount int64
			database.DB.Model(&models.ConversationMessage{}).
				Where("conversation_id = ? AND is_deleted = false", convID).
				Count(&msgCount)
			if msgCount == 0 {
				database.DB.Where("conversation_id = ?", convID).Delete(&models.ConversationMember{})
				database.DB.Delete(&conv)
				appws.BroadcastToUser(conv.CreatedByID, appws.Message{
					Type:    "dm.conversation_deleted",
					Payload: map[string]interface{}{"conversation_id": convID},
				})
				c.JSON(http.StatusOK, gin.H{"message": "removed", "conversation_deleted": true})
				return
			}
		}
	}

	// Notify remaining members of the removal
	for _, uid := range memberIDs {
		appws.BroadcastToUser(uid, appws.Message{
			Type: "dm.member_removed",
			Payload: map[string]interface{}{
				"conversation_id": convID,
				"user_id":         targetID,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

// AddConversationMember POST /conversations/:id/members
func AddConversationMember(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if isMember(uint(convID), req.UserID) {
		c.JSON(http.StatusOK, gin.H{"message": "already a member"})
		return
	}

	database.DB.Create(&models.ConversationMember{
		ConversationID: uint(convID),
		UserID:         req.UserID,
		JoinedAt:       time.Now(),
	})

	// Promote to group if now more than 2 members
	var count int64
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", convID).Count(&count)
	if count > 2 {
		database.DB.Model(&models.Conversation{}).
			Where("id = ?", convID).Update("is_group", true)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "added"})
}

// EditConversationMessage PATCH /conversations/:id/messages/:msgId
func EditConversationMessage(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))
	msgID, _ := strconv.Atoi(c.Param("msgId"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var msg models.ConversationMessage
	if err := database.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if msg.SenderID != me {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your message"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	database.DB.Model(&msg).Updates(map[string]interface{}{"body": req.Body, "is_edited": true})

	// Broadcast to all conversation members
	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", convID).
		Pluck("user_id", &memberIDs)
	for _, uid := range memberIDs {
		appws.BroadcastToUser(uid, appws.Message{
			Type: appws.TypeDMMessageUpdated,
			Payload: map[string]interface{}{
				"conversation_id": convID,
				"id":              msg.ID,
				"body":            req.Body,
				"is_edited":       true,
			},
		})
	}

	database.DB.Preload("Sender").First(&msg, msg.ID)
	c.JSON(http.StatusOK, msg)
}

// UploadConversationAvatar POST /conversations/:id/avatar
// Accepts a multipart image, stores it in the uploads dir, and saves the URL on the conversation.
func UploadConversationAvatar(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	fh, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
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

	// Validate MIME type from actual file bytes — reject non-images and SVG (XSS risk).
	allowedAvatarTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if f, err := os.Open(dest); err == nil {
		mt, err := mimetype.DetectReader(f)
		f.Close()
		if err != nil || !allowedAvatarTypes[mt.String()] {
			os.Remove(dest)
			c.JSON(http.StatusBadRequest, gin.H{"error": "file must be an image (jpeg, png, gif, webp)"})
			return
		}
	}

	avatarURL := "/uploads/" + storedName
	if err := database.DB.Model(&models.Conversation{}).Where("id = ?", convID).
		Update("avatar", avatarURL).Error; err != nil {
		os.Remove(dest)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update avatar"})
		return
	}

	// Broadcast to all members so their UI updates
	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", convID).
		Pluck("user_id", &memberIDs)
	for _, uid := range memberIDs {
		appws.BroadcastToUser(uid, appws.Message{
			Type: "dm.conversation_updated",
			Payload: map[string]interface{}{
				"conversation_id": convID,
				"avatar":          avatarURL,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"avatar": avatarURL})
}

// LeaveConversation DELETE /conversations/:id
// Current user leaves the conversation.
// For 1-on-1: deletes the conversation for both parties.
// For group: removes only the current user.
func LeaveConversation(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var conv models.Conversation
	if err := database.DB.First(&conv, convID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", convID).
		Pluck("user_id", &memberIDs)

	if !conv.IsGroup {
		// 1-on-1: delete the whole conversation so neither party is left with an orphan
		database.DB.Where("conversation_id = ?", convID).Delete(&models.ConversationMember{})
		database.DB.Delete(&conv)
		for _, uid := range memberIDs {
			appws.BroadcastToUser(uid, appws.Message{
				Type:    "dm.conversation_deleted",
				Payload: map[string]interface{}{"conversation_id": convID},
			})
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		return
	}

	// Group: remove only the current user
	database.DB.Where("conversation_id = ? AND user_id = ?", convID, me).
		Delete(&models.ConversationMember{})

	// Notify remaining members
	for _, uid := range memberIDs {
		if uid == me {
			continue
		}
		appws.BroadcastToUser(uid, appws.Message{
			Type: "dm.member_removed",
			Payload: map[string]interface{}{
				"conversation_id": convID,
				"user_id":         me,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "left"})
}

func isMember(convID, userID uint) bool {
	var m models.ConversationMember
	return database.DB.Where("conversation_id = ? AND user_id = ?", convID, userID).
		First(&m).Error == nil
}
