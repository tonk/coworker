package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// canManageCustomers returns true when the caller is a global admin or has
// owner/admin role in at least one project.
func canManageCustomers(c *gin.Context) bool {
	if middleware.GetGlobalRole(c) == "admin" {
		return true
	}
	userID := middleware.GetUserID(c)
	var count int64
	database.DB.Model(&models.ProjectMember{}).
		Where("user_id = ? AND role IN ('owner','admin')", userID).
		Count(&count)
	return count > 0
}

// CustomerListItem wraps Customer with extra UI metadata.
type CustomerListItem struct {
	models.Customer
	IsFavorite    bool  `json:"is_favorite"`
	ProjectCount  int64 `json:"project_count"`
	ContractCount int64 `json:"contract_count"`
}

// ListCustomers returns all customers with favourite status and counts.
func ListCustomers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var customers []models.Customer
	database.DB.Order("position asc, id asc").Find(&customers)

	var favs []models.CustomerFavorite
	database.DB.Where("user_id = ?", userID).Find(&favs)
	favSet := make(map[uint]bool, len(favs))
	for _, f := range favs {
		favSet[f.CustomerID] = true
	}

	items := make([]CustomerListItem, len(customers))
	for i, cust := range customers {
		var pCount, cCount int64
		database.DB.Model(&models.Project{}).Where("customer_id = ? AND deleted_at IS NULL", cust.ID).Count(&pCount)
		database.DB.Model(&models.Contract{}).Where("customer_id = ?", cust.ID).Count(&cCount)
		items[i] = CustomerListItem{
			Customer:      cust,
			IsFavorite:    favSet[cust.ID],
			ProjectCount:  pCount,
			ContractCount: cCount,
		}
	}

	c.JSON(http.StatusOK, items)
}

// CustomerDetailResponse is the full view returned by GetCustomer.
type CustomerDetailResponse struct {
	Customer  CustomerListItem `json:"customer"`
	Contracts []ContractGroup  `json:"contracts"`
	Projects  []models.Project `json:"projects"` // projects with no contract
}

// ContractGroup bundles a contract with its projects.
type ContractGroup struct {
	models.Contract
	Projects []models.Project `json:"projects"`
}

// GetCustomer returns a customer with its contracts and projects.
func GetCustomer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var cust models.Customer
	if err := database.DB.First(&cust, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	var fav models.CustomerFavorite
	isFav := database.DB.Where("user_id = ? AND customer_id = ?", userID, cust.ID).First(&fav).Error == nil

	var pCount, cCount int64
	database.DB.Model(&models.Project{}).Where("customer_id = ? AND deleted_at IS NULL", cust.ID).Count(&pCount)
	database.DB.Model(&models.Contract{}).Where("customer_id = ?", cust.ID).Count(&cCount)
	custItem := CustomerListItem{Customer: cust, IsFavorite: isFav, ProjectCount: pCount, ContractCount: cCount}

	var contracts []models.Contract
	database.DB.Where("customer_id = ?", cust.ID).Order("id asc").Find(&contracts)

	contractGroups := make([]ContractGroup, len(contracts))
	for i, con := range contracts {
		var projects []models.Project
		database.DB.Where("customer_id = ? AND contract_id = ? AND deleted_at IS NULL", cust.ID, con.ID).
			Order("position asc, id asc").Find(&projects)
		contractGroups[i] = ContractGroup{Contract: con, Projects: projects}
	}

	var unassigned []models.Project
	database.DB.Where("customer_id = ? AND contract_id IS NULL AND deleted_at IS NULL", cust.ID).
		Order("position asc, id asc").Find(&unassigned)

	c.JSON(http.StatusOK, CustomerDetailResponse{
		Customer:  custItem,
		Contracts: contractGroups,
		Projects:  unassigned,
	})
}

// CreateCustomer creates a new customer.
func CreateCustomer(c *gin.Context) {
	if !canManageCustomers(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=200"`
		Description string `json:"description"`
		LogoURL     string `json:"logo_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var maxPos int
	database.DB.Model(&models.Customer{}).Select("COALESCE(MAX(position),0)").Scan(&maxPos)

	cust := models.Customer{
		Name:        req.Name,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		Position:    maxPos + 1,
	}
	if err := database.DB.Create(&cust).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, cust)
}

// UpdateCustomer updates a customer's metadata.
func UpdateCustomer(c *gin.Context) {
	if !canManageCustomers(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	id, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var cust models.Customer
	if err := database.DB.First(&cust, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		LogoURL     string `json:"logo_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{
		"description": req.Description,
		"logo_url":    req.LogoURL,
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	database.DB.Model(&cust).Updates(updates)
	database.DB.First(&cust, id)
	c.JSON(http.StatusOK, cust)
}

// DeleteCustomer deletes a customer (admin only). Projects are detached first.
func DeleteCustomer(c *gin.Context) {
	if middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}
	id, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var cust models.Customer
	if err := database.DB.First(&cust, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}
	database.DB.Model(&models.Project{}).Where("customer_id = ?", cust.ID).
		Updates(map[string]interface{}{"customer_id": nil, "contract_id": nil})
	database.DB.Where("customer_id = ?", cust.ID).Delete(&models.Contract{})
	database.DB.Where("customer_id = ?", cust.ID).Delete(&models.CustomerFavorite{})
	database.DB.Delete(&cust)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// AddCustomerFavorite stars a customer for the current user.
func AddCustomerFavorite(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	fav := models.CustomerFavorite{UserID: userID, CustomerID: uint(id)}
	database.DB.Where(fav).FirstOrCreate(&fav)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// RemoveCustomerFavorite unstars a customer.
func RemoveCustomerFavorite(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	database.DB.Where("user_id = ? AND customer_id = ?", userID, id).Delete(&models.CustomerFavorite{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListContracts lists all contracts for a customer.
func ListContracts(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var contracts []models.Contract
	database.DB.Where("customer_id = ?", id).Order("id asc").Find(&contracts)
	c.JSON(http.StatusOK, contracts)
}

// CreateContract creates a contract under a customer.
func CreateContract(c *gin.Context) {
	if !canManageCustomers(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	custID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var cust models.Customer
	if err := database.DB.First(&cust, custID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=200"`
		Description string `json:"description"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	con := models.Contract{
		CustomerID:  uint(custID),
		Name:        req.Name,
		Description: req.Description,
	}
	if t, err := parseContractDate(req.StartDate); err == nil && t != nil {
		con.StartDate = t
	}
	if t, err := parseContractDate(req.EndDate); err == nil && t != nil {
		con.EndDate = t
	}
	if err := database.DB.Create(&con).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, con)
}

// UpdateContract updates a contract.
func UpdateContract(c *gin.Context) {
	if !canManageCustomers(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	contractID, err := strconv.Atoi(c.Param("contractId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var con models.Contract
	if err := database.DB.First(&con, contractID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{"description": req.Description}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if t, err := parseContractDate(req.StartDate); err == nil {
		updates["start_date"] = t // nil clears the field
	}
	if t, err := parseContractDate(req.EndDate); err == nil {
		updates["end_date"] = t
	}
	database.DB.Model(&con).Updates(updates)
	database.DB.First(&con, contractID)
	c.JSON(http.StatusOK, con)
}

// DeleteContract deletes a contract (admin only). Projects are detached first.
func DeleteContract(c *gin.Context) {
	if middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}
	contractID, err := strconv.Atoi(c.Param("contractId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var con models.Contract
	if err := database.DB.First(&con, contractID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	database.DB.Model(&models.Project{}).Where("contract_id = ?", con.ID).Update("contract_id", nil)
	database.DB.Delete(&con)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// parseContractDate parses an optional YYYY-MM-DD date string.
// Returns (nil, nil) for an empty string (clearing the field).
func parseContractDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
