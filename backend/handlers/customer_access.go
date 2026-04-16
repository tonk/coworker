package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// CustomerMemberItem is returned by the member list endpoint.
type CustomerMemberItem struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	GravatarURL string `json:"gravatar_url"`
	Role        string `json:"role"` // "member" | "admin"
}

// ListCustomerMembers returns all users with explicit access to this customer.
// Accessible to global admins and to users who hold the "admin" role for this customer.
func ListCustomerMembers(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	if !canManageCustomer(c, uint(custID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var rows []models.CustomerAccess
	database.DB.Where("customer_id = ?", custID).Find(&rows)

	if len(rows) == 0 {
		c.JSON(http.StatusOK, []CustomerMemberItem{})
		return
	}

	userIDs := make([]uint, len(rows))
	roleByUser := make(map[uint]string, len(rows))
	for i, r := range rows {
		userIDs[i] = r.UserID
		roleByUser[r.UserID] = r.Role
	}

	var users []models.User
	database.DB.Where("id IN ?", userIDs).Find(&users)

	items := make([]CustomerMemberItem, 0, len(users))
	for _, u := range users {
		items = append(items, CustomerMemberItem{
			UserID:      u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			AvatarURL:   u.AvatarURL,
			GravatarURL: u.GravatarURL,
			Role:        roleByUser[u.ID],
		})
	}

	c.JSON(http.StatusOK, items)
}

// SetCustomerMembers syncs the member list for a customer.
// Accessible to global admins and to users who hold the "admin" role for this customer.
// Passing an empty members list removes all restrictions (everyone can see the customer).
// Request: { "members": [{"user_id": 1, "role": "admin"}, ...] }
func SetCustomerMembers(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	if !canManageCustomer(c, uint(custID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// A customer-admin cannot remove themselves (that would lock them out).
	callerID := middleware.GetUserID(c)
	isGlobalAdmin := middleware.GetGlobalRole(c) == "admin"

	var req struct {
		Members []struct {
			UserID uint   `json:"user_id"`
			Role   string `json:"role"`
		} `json:"members"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	desired := make(map[uint]string, len(req.Members))
	for _, m := range req.Members {
		role := m.Role
		if role != "admin" && role != "member" {
			role = "member"
		}
		desired[m.UserID] = role
	}

	// Non-global-admins cannot remove their own admin row or downgrade themselves.
	if !isGlobalAdmin {
		desired[callerID] = "admin"
	}

	var current []models.CustomerAccess
	database.DB.Where("customer_id = ?", custID).Find(&current)

	currentSet := make(map[uint]bool, len(current))
	for _, r := range current {
		currentSet[r.UserID] = true
	}

	// Remove rows no longer desired.
	for _, r := range current {
		if _, ok := desired[r.UserID]; !ok {
			database.DB.Where("user_id = ? AND customer_id = ?", r.UserID, custID).
				Delete(&models.CustomerAccess{})
		}
	}

	// Upsert desired rows.
	for uid, role := range desired {
		if currentSet[uid] {
			database.DB.Model(&models.CustomerAccess{}).
				Where("user_id = ? AND customer_id = ?", uid, custID).
				Update("role", role)
		} else {
			database.DB.Create(&models.CustomerAccess{
				UserID:     uid,
				CustomerID: uint(custID),
				Role:       role,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
