package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
)

func ListFavoriteUsers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var favs []models.FavoriteUser
	database.DB.Where("user_id = ?", userID).Find(&favs)

	ids := make([]uint, 0, len(favs))
	for _, f := range favs {
		ids = append(ids, f.FavoriteUserID)
	}
	if len(ids) == 0 {
		c.JSON(http.StatusOK, []models.User{})
		return
	}

	var users []models.User
	database.DB.Where("id IN ?", ids).Find(&users)
	c.JSON(http.StatusOK, users)
}

func AddFavoriteUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil || uint(targetID) == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Ensure target user exists
	var target models.User
	if err := database.DB.First(&target, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	fav := models.FavoriteUser{UserID: userID, FavoriteUserID: uint(targetID)}
	database.DB.FirstOrCreate(&fav, fav)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func RemoveFavoriteUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	database.DB.Delete(&models.FavoriteUser{}, "user_id = ? AND favorite_user_id = ?", userID, targetID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
