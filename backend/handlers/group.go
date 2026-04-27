package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

// ─── Response types ──────────────────────────────────────────────────────────

type GroupListItem struct {
	models.UserGroup
	MemberCount   int64 `json:"member_count"`
	ProjectCount  int64 `json:"project_count"`
	CustomerCount int64 `json:"customer_count"`
}

type GroupMemberEntry struct {
	GroupID uint        `json:"group_id"`
	UserID  uint        `json:"user_id"`
	User    models.User `json:"user"`
}

type GroupProjectEntry struct {
	GroupID   uint           `json:"group_id"`
	ProjectID uint           `json:"project_id"`
	Role      string         `json:"role"`
	Project   models.Project `json:"project"`
}

type GroupCustomerEntry struct {
	GroupID    uint            `json:"group_id"`
	CustomerID uint            `json:"customer_id"`
	Role       string          `json:"role"`
	Customer   models.Customer `json:"customer"`
}

type GroupDetail struct {
	models.UserGroup
	Members        []GroupMemberEntry   `json:"members"`
	ProjectAccess  []GroupProjectEntry  `json:"project_access"`
	CustomerAccess []GroupCustomerEntry `json:"customer_access"`
}

type ProjectGroupEntry struct {
	GroupID uint               `json:"group_id"`
	Role    string             `json:"role"`
	Group   models.UserGroup   `json:"group"`
	Members []GroupMemberEntry `json:"members"`
}

type CustomerGroupEntry struct {
	GroupID uint               `json:"group_id"`
	Role    string             `json:"role"`
	Group   models.UserGroup   `json:"group"`
	Members []GroupMemberEntry `json:"members"`
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func parseGroupID(c *gin.Context) (uint, bool) {
	id, err := strconv.Atoi(c.Param("groupId"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return 0, false
	}
	return uint(id), true
}

func loadGroupMembers(groupID uint) []GroupMemberEntry {
	var gms []models.GroupMember
	database.DB.Where("group_id = ?", groupID).Find(&gms)
	out := make([]GroupMemberEntry, 0, len(gms))
	for _, gm := range gms {
		var u models.User
		if database.DB.First(&u, gm.UserID).Error == nil {
			out = append(out, GroupMemberEntry{GroupID: gm.GroupID, UserID: gm.UserID, User: u})
		}
	}
	return out
}

func loadGroupDetail(groupID uint) (*GroupDetail, bool) {
	var g models.UserGroup
	if err := database.DB.First(&g, groupID).Error; err != nil {
		return nil, false
	}

	var gpas []models.GroupProjectAccess
	database.DB.Where("group_id = ?", groupID).Find(&gpas)
	projAccess := make([]GroupProjectEntry, 0, len(gpas))
	for _, gpa := range gpas {
		var p models.Project
		if database.DB.Where("id = ? AND deleted_at IS NULL", gpa.ProjectID).First(&p).Error == nil {
			projAccess = append(projAccess, GroupProjectEntry{
				GroupID: gpa.GroupID, ProjectID: gpa.ProjectID, Role: gpa.Role, Project: p,
			})
		}
	}

	var gcas []models.GroupCustomerAccess
	database.DB.Where("group_id = ?", groupID).Find(&gcas)
	custAccess := make([]GroupCustomerEntry, 0, len(gcas))
	for _, gca := range gcas {
		var cu models.Customer
		if database.DB.First(&cu, gca.CustomerID).Error == nil {
			custAccess = append(custAccess, GroupCustomerEntry{
				GroupID: gca.GroupID, CustomerID: gca.CustomerID, Role: gca.Role, Customer: cu,
			})
		}
	}

	return &GroupDetail{
		UserGroup:      g,
		Members:        loadGroupMembers(groupID),
		ProjectAccess:  projAccess,
		CustomerAccess: custAccess,
	}, true
}

func validGroupRole(role string) bool {
	return role == "viewer" || role == "member" || role == "owner"
}

// ─── Admin: group CRUD ────────────────────────────────────────────────────────

func AdminListGroups(c *gin.Context) {
	var groups []models.UserGroup
	database.DB.Order("name asc").Find(&groups)
	result := make([]GroupListItem, len(groups))
	for i, g := range groups {
		var mc, pc, cc int64
		database.DB.Model(&models.GroupMember{}).Where("group_id = ?", g.ID).Count(&mc)
		database.DB.Model(&models.GroupProjectAccess{}).Where("group_id = ?", g.ID).Count(&pc)
		database.DB.Model(&models.GroupCustomerAccess{}).Where("group_id = ?", g.ID).Count(&cc)
		result[i] = GroupListItem{UserGroup: g, MemberCount: mc, ProjectCount: pc, CustomerCount: cc}
	}
	c.JSON(http.StatusOK, result)
}

func AdminGetGroup(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	detail, ok := loadGroupDetail(groupID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func AdminCreateGroup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=200"`
		Description string `json:"description"`
		Avatar      string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	g := models.UserGroup{Name: req.Name, Description: req.Description, Avatar: req.Avatar}
	if err := database.DB.Create(&g).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "a group with that name already exists"})
		return
	}
	c.JSON(http.StatusCreated, g)
}

func AdminUpdateGroup(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var g models.UserGroup
	if err := database.DB.First(&g, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Avatar      *string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := map[string]any{"description": req.Description}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	database.DB.Model(&g).Updates(updates)
	database.DB.First(&g, groupID)
	c.JSON(http.StatusOK, g)
}

func AdminDeleteGroup(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var g models.UserGroup
	if err := database.DB.First(&g, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	database.DB.Where("group_id = ?", groupID).Delete(&models.GroupMember{})
	database.DB.Where("group_id = ?", groupID).Delete(&models.GroupProjectAccess{})
	database.DB.Where("group_id = ?", groupID).Delete(&models.GroupCustomerAccess{})
	database.DB.Delete(&g)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ─── Admin: group membership ──────────────────────────────────────────────────

func AdminAddGroupMember(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var g models.UserGroup
	if err := database.DB.First(&g, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var u models.User
	if err := database.DB.First(&u, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var count int64
	database.DB.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, req.UserID).Count(&count)
	if count == 0 {
		if err := database.DB.Create(&models.GroupMember{GroupID: groupID, UserID: req.UserID}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminRemoveGroupMember(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	database.DB.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Admin: group access on projects ─────────────────────────────────────────

func AdminSetGroupProjectAccess(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if !validGroupRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be viewer, member, or owner"})
		return
	}
	var gpa models.GroupProjectAccess
	if database.DB.Where("group_id = ? AND project_id = ?", groupID, projectID).First(&gpa).Error != nil {
		gpa = models.GroupProjectAccess{GroupID: groupID, ProjectID: uint(projectID), Role: req.Role}
		database.DB.Create(&gpa)
	} else {
		database.DB.Model(&gpa).Update("role", req.Role)
	}
	c.JSON(http.StatusOK, gpa)
}

func AdminRemoveGroupProjectAccess(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	database.DB.Where("group_id = ? AND project_id = ?", groupID, projectID).Delete(&models.GroupProjectAccess{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Admin: group access on customers ────────────────────────────────────────

func AdminSetGroupCustomerAccess(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	customerID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if !validGroupRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be viewer, member, or owner"})
		return
	}
	var gca models.GroupCustomerAccess
	if database.DB.Where("group_id = ? AND customer_id = ?", groupID, customerID).First(&gca).Error != nil {
		gca = models.GroupCustomerAccess{GroupID: groupID, CustomerID: uint(customerID), Role: req.Role}
		database.DB.Create(&gca)
	} else {
		database.DB.Model(&gca).Update("role", req.Role)
	}
	c.JSON(http.StatusOK, gca)
}

func AdminRemoveGroupCustomerAccess(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	customerID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	database.DB.Where("group_id = ? AND customer_id = ?", groupID, customerID).Delete(&models.GroupCustomerAccess{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Project-scoped: group member management (project owners) ─────────────────

func ListProjectGroups(c *gin.Context) {
	userID := middleware.GetUserID(c)
	project, err := services.GetProjectBySlug(c.Param("projectSlug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}
	var gpas []models.GroupProjectAccess
	database.DB.Where("project_id = ?", project.ID).Find(&gpas)
	result := make([]ProjectGroupEntry, 0, len(gpas))
	for _, gpa := range gpas {
		var g models.UserGroup
		if database.DB.First(&g, gpa.GroupID).Error != nil {
			continue
		}
		result = append(result, ProjectGroupEntry{
			GroupID: gpa.GroupID, Role: gpa.Role, Group: g,
			Members: loadGroupMembers(gpa.GroupID),
		})
	}
	c.JSON(http.StatusOK, result)
}

func ProjectAddGroupMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	project, err := services.GetProjectBySlug(c.Param("projectSlug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	// Verify the group has access to this project
	var gpa models.GroupProjectAccess
	if database.DB.Where("group_id = ? AND project_id = ?", groupID, project.ID).First(&gpa).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group does not have access to this project"})
		return
	}
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var count int64
	database.DB.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, req.UserID).Count(&count)
	if count == 0 {
		if err := database.DB.Create(&models.GroupMember{GroupID: groupID, UserID: req.UserID}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ProjectRemoveGroupMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	project, err := services.GetProjectBySlug(c.Param("projectSlug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "owner access required"})
		return
	}
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var gpa models.GroupProjectAccess
	if database.DB.Where("group_id = ? AND project_id = ?", groupID, project.ID).First(&gpa).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group does not have access to this project"})
		return
	}
	memberID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	database.DB.Where("group_id = ? AND user_id = ?", groupID, memberID).Delete(&models.GroupMember{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Customer-scoped: group member management (customer owners) ───────────────

func ListCustomerGroups(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	if !canManageCustomer(c, uint(customerID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	var gcas []models.GroupCustomerAccess
	database.DB.Where("customer_id = ?", customerID).Find(&gcas)
	result := make([]CustomerGroupEntry, 0, len(gcas))
	for _, gca := range gcas {
		var g models.UserGroup
		if database.DB.First(&g, gca.GroupID).Error != nil {
			continue
		}
		result = append(result, CustomerGroupEntry{
			GroupID: gca.GroupID, Role: gca.Role, Group: g,
			Members: loadGroupMembers(gca.GroupID),
		})
	}
	c.JSON(http.StatusOK, result)
}

func CustomerAddGroupMember(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	if !canManageCustomer(c, uint(customerID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var gca models.GroupCustomerAccess
	if database.DB.Where("group_id = ? AND customer_id = ?", groupID, customerID).First(&gca).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group does not have access to this customer"})
		return
	}
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var count int64
	database.DB.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, req.UserID).Count(&count)
	if count == 0 {
		if err := database.DB.Create(&models.GroupMember{GroupID: groupID, UserID: req.UserID}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func CustomerRemoveGroupMember(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	if !canManageCustomer(c, uint(customerID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var gca models.GroupCustomerAccess
	if database.DB.Where("group_id = ? AND customer_id = ?", groupID, customerID).First(&gca).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group does not have access to this customer"})
		return
	}
	memberID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	database.DB.Where("group_id = ? AND user_id = ?", groupID, memberID).Delete(&models.GroupMember{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
