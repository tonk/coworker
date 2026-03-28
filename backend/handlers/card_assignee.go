package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
	appws "github.com/tonk/coworker/ws"
)

func loadCardAssignees(cardID uint) []models.User {
	var users []models.User
	database.DB.Joins("JOIN card_assignees ON card_assignees.user_id = users.id").
		Where("card_assignees.card_id = ?", cardID).Find(&users)
	if users == nil {
		users = []models.User{}
	}
	return users
}

// AddCardAssignee POST /projects/:projectSlug/cards/:cardId/assignees/:userId
func AddCardAssignee(c *gin.Context) {
	me := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}
	targetUID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, me, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	assn := models.CardAssignee{CardID: uint(cardID), UserID: uint(targetUID)}
	database.DB.Where(assn).FirstOrCreate(&assn)

	assignees := loadCardAssignees(uint(cardID))
	appws.BroadcastToProject(project.ID, appws.Message{
		Type:    appws.TypeBoardCardUpdated,
		Payload: gin.H{"id": cardID, "assignees": assignees},
	})
	c.JSON(http.StatusOK, assignees)
}

// RemoveCardAssignee DELETE /projects/:projectSlug/cards/:cardId/assignees/:userId
func RemoveCardAssignee(c *gin.Context) {
	me := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}
	targetUID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, me, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	database.DB.Where("card_id = ? AND user_id = ?", cardID, targetUID).Delete(&models.CardAssignee{})

	assignees := loadCardAssignees(uint(cardID))
	appws.BroadcastToProject(project.ID, appws.Message{
		Type:    appws.TypeBoardCardUpdated,
		Payload: gin.H{"id": cardID, "assignees": assignees},
	})
	c.JSON(http.StatusOK, assignees)
}
