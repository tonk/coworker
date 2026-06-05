package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

func generateAPIKey() (plain, hash, prefix string, err error) {
	b := make([]byte, 24)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = "cwk_" + hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(sum[:])
	if len(plain) > 12 {
		prefix = plain[:12]
	} else {
		prefix = plain
	}
	return
}

// ListAPIKeys godoc
// @Summary      List the current user's API keys
// @Tags         api-keys
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array}  models.APIKey
// @Router       /auth/api-keys [get]
func ListAPIKeys(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var keys []models.APIKey
	database.DB.Where("user_id = ?", userID).Find(&keys)
	c.JSON(http.StatusOK, keys)
}

// CreateAPIKey godoc
// @Summary      Create a new API key
// @Tags         api-keys
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body map[string]string true "Key name"
// @Success      201 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Router       /auth/api-keys [post]
func CreateAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Name string `json:"name" binding:"required,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	plain, hash, prefix, err := generateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	key := models.APIKey{
		UserID:    userID,
		Name:      req.Name,
		KeyHash:   hash,
		KeyPrefix: prefix,
	}
	if err := database.DB.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Return the plain text key ONLY on creation
	c.JSON(http.StatusCreated, gin.H{
		"id":         key.ID,
		"name":       key.Name,
		"key_prefix": key.KeyPrefix,
		"key":        plain, // shown once
		"created_at": key.CreatedAt,
	})
}

func DeleteAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var key models.APIKey
	if err := database.DB.Where("id = ? AND user_id = ? AND project_id IS NULL", id, userID).First(&key).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	database.DB.Delete(&key)
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

// ListProjectAPIKeys godoc
// @Summary      List API keys for a project
// @Tags         api-keys
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Success      200 {array}  models.APIKey
// @Router       /projects/{projectSlug}/api-keys [get]
func ListProjectAPIKeys(c *gin.Context) {
	userID := middleware.GetUserID(c)
	project, err := services.GetProjectBySlug(c.Param("projectSlug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var keys []models.APIKey
	database.DB.Where("project_id = ?", project.ID).Find(&keys)
	c.JSON(http.StatusOK, keys)
}

// CreateProjectAPIKey godoc
// @Summary      Create a project-scoped API key
// @Tags         api-keys
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Param        body body map[string]string true "Key name"
// @Success      201 {object} map[string]interface{}
// @Router       /projects/{projectSlug}/api-keys [post]
func CreateProjectAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	project, err := services.GetProjectBySlug(c.Param("projectSlug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	plain, hash, prefix, err := generateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	key := models.APIKey{
		UserID:    userID,
		ProjectID: &project.ID,
		Name:      req.Name,
		KeyHash:   hash,
		KeyPrefix: prefix,
	}
	if err := database.DB.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":         key.ID,
		"name":       key.Name,
		"key_prefix": key.KeyPrefix,
		"key":        plain,
		"created_at": key.CreatedAt,
	})
}

// AdminListUserAPIKeys lists all personal API keys for a specific user.
func AdminListUserAPIKeys(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var keys []models.APIKey
	database.DB.Where("user_id = ? AND project_id IS NULL", uint(id)).Find(&keys)
	c.JSON(http.StatusOK, keys)
}

// AdminCreateUserAPIKey creates a personal API key on behalf of a specific user.
func AdminCreateUserAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	plain, hash, prefix, err := generateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	key := models.APIKey{
		UserID:    uint(id),
		Name:      req.Name,
		KeyHash:   hash,
		KeyPrefix: prefix,
	}
	if err := database.DB.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":         key.ID,
		"name":       key.Name,
		"key_prefix": key.KeyPrefix,
		"key":        plain, // shown once
		"created_at": key.CreatedAt,
	})
}

// AdminDeleteUserAPIKey deletes a personal API key for a specific user.
func AdminDeleteUserAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	keyId, err := strconv.ParseUint(c.Param("keyId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}
	var key models.APIKey
	if err := database.DB.Where("id = ? AND user_id = ? AND project_id IS NULL", keyId, uint(id)).First(&key).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	database.DB.Delete(&key)
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

func DeleteProjectAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	project, err := services.GetProjectBySlug(c.Param("projectSlug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	id, err := strconv.ParseUint(c.Param("keyId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var key models.APIKey
	if err := database.DB.Where("id = ? AND project_id = ?", id, project.ID).First(&key).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	database.DB.Delete(&key)
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}
