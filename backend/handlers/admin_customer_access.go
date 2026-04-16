package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// AdminGetUserCustomers returns the customer IDs and roles the user has explicit
// access to, plus a restricted flag (true = has any rows = not unrestricted).
func AdminGetUserCustomers(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var rows []models.CustomerAccess
	database.DB.Where("user_id = ?", userID).Find(&rows)

	ids := make([]uint, len(rows))
	roles := make(map[uint]string, len(rows))
	for i, r := range rows {
		ids[i] = r.CustomerID
		roles[r.CustomerID] = r.Role
	}
	c.JSON(http.StatusOK, gin.H{
		"customer_ids":   ids,
		"customer_roles": roles,
	})
}

// AdminSetUserCustomers syncs a user's customer access to exactly the given list.
// customer_roles maps customer ID to "member" or "admin"; omitted IDs default to "member".
// Passing an empty customer_ids removes all rows (user reverts to unrestricted access).
func AdminSetUserCustomers(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		CustomerIDs   []uint            `json:"customer_ids"`
		CustomerRoles map[uint]string   `json:"customer_roles"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	desired := make(map[uint]string, len(req.CustomerIDs))
	for _, id := range req.CustomerIDs {
		role := "member"
		if req.CustomerRoles != nil {
			if r, ok := req.CustomerRoles[id]; ok && (r == "admin" || r == "member") {
				role = r
			}
		}
		desired[id] = role
	}

	var current []models.CustomerAccess
	database.DB.Where("user_id = ?", userID).Find(&current)

	currentSet := make(map[uint]bool, len(current))
	for _, r := range current {
		currentSet[r.CustomerID] = true
	}

	// Remove rows no longer desired
	for _, r := range current {
		if _, ok := desired[r.CustomerID]; !ok {
			database.DB.Where("user_id = ? AND customer_id = ?", userID, r.CustomerID).
				Delete(&models.CustomerAccess{})
		}
	}

	// Upsert desired rows
	for cid, role := range desired {
		if currentSet[cid] {
			database.DB.Model(&models.CustomerAccess{}).
				Where("user_id = ? AND customer_id = ?", userID, cid).
				Update("role", role)
		} else {
			database.DB.Create(&models.CustomerAccess{
				UserID:     uint(userID),
				CustomerID: cid,
				Role:       role,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_ids":   req.CustomerIDs,
		"customer_roles": desired,
	})
}
