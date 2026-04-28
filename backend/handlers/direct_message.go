package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// ListDirectMessages returns messages between the current user and the given user.
func ListDirectMessages(c *gin.Context) {
	currentUserID := middleware.GetUserID(c)
	otherID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var msgs []models.DirectMessage
	database.DB.
		Preload("Sender").
		Preload("Receiver").
		Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			currentUserID, otherID, otherID, currentUserID,
		).
		Order("created_at asc").
		Limit(100).
		Find(&msgs)

	c.JSON(http.StatusOK, msgs)
}

// SendDirectMessage sends a message from the current user to another user.
func SendDirectMessage(c *gin.Context) {
	currentUserID := middleware.GetUserID(c)
	otherID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if uint(otherID) == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot message yourself"})
		return
	}

	var other models.User
	if err := database.DB.First(&other, otherID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	msg := models.DirectMessage{
		SenderID:   currentUserID,
		ReceiverID: uint(otherID),
		Body:       req.Body,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	database.DB.Preload("Sender").Preload("Receiver").First(&msg, msg.ID)
	c.JSON(http.StatusCreated, msg)
}

// DeleteDirectMessage soft-deletes a message the current user sent.
func DeleteDirectMessage(c *gin.Context) {
	currentUserID := middleware.GetUserID(c)
	msgID, err := strconv.ParseUint(c.Param("msgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	var msg models.DirectMessage
	if err := database.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	if msg.SenderID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	database.DB.Model(&msg).Update("is_deleted", true)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListConversations returns all users the current user has exchanged messages with.
func ListConversations(c *gin.Context) {
	currentUserID := middleware.GetUserID(c)

	type conv struct {
		UserID      uint   `json:"user_id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
	}

	// Get unique user IDs the current user has talked to
	var convs []conv
	database.DB.Raw(`
		SELECT DISTINCT u.id as user_id, u.username, u.display_name, u.avatar_url
		FROM users u
		JOIN direct_messages dm ON (
			(dm.sender_id = ? AND dm.receiver_id = u.id) OR
			(dm.receiver_id = ? AND dm.sender_id = u.id)
		)
		WHERE u.deleted_at IS NULL
		ORDER BY u.username
	`, currentUserID, currentUserID).Scan(&convs)

	c.JSON(http.StatusOK, convs)
}

// ListAllUsers godoc
// @Summary      List all active users
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array}  models.User
// @Router       /users [get]
func ListAllUsers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var users []models.User

	if globalRole == "admin" {
		database.DB.Where("is_active = ?", true).Find(&users)
		c.JSON(http.StatusOK, users)
		return
	}

	// Collect the customer IDs visible to the current user (direct + via groups).
	myRoles := getAccessibleCustomerRoles(userID)
	if len(myRoles) == 0 {
		// No explicit customer access rows → user sees all customers, so see all users.
		database.DB.Where("is_active = ?", true).Find(&users)
		c.JSON(http.StatusOK, users)
		return
	}

	customerIDs := make([]uint, 0, len(myRoles))
	for cid := range myRoles {
		customerIDs = append(customerIDs, cid)
	}

	// Collect user IDs who share at least one of those customers (direct access).
	var directUserIDs []uint
	database.DB.Model(&models.CustomerAccess{}).
		Where("customer_id IN ?", customerIDs).
		Distinct("user_id").
		Pluck("user_id", &directUserIDs)

	// Also collect user IDs who access those customers via a group.
	var groupUserIDs []uint
	database.DB.Raw(`
		SELECT DISTINCT gm.user_id
		FROM group_members gm
		JOIN group_customer_accesses gca ON gca.group_id = gm.group_id
		WHERE gca.customer_id IN ?`, customerIDs).Scan(&groupUserIDs)

	seen := make(map[uint]bool, len(directUserIDs)+len(groupUserIDs)+1)
	seen[userID] = true // always include the requester themselves
	for _, id := range directUserIDs {
		seen[id] = true
	}
	for _, id := range groupUserIDs {
		seen[id] = true
	}
	allIDs := make([]uint, 0, len(seen))
	for id := range seen {
		allIDs = append(allIDs, id)
	}

	database.DB.Where("is_active = ? AND id IN ?", true, allIDs).Find(&users)
	c.JSON(http.StatusOK, users)
}
