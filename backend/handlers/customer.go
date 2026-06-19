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

// customerRoleRank maps customer/group roles to a comparable integer.
func customerRoleRank(role string) int {
	switch role {
	case "viewer":
		return 1
	case "member":
		return 2
	case "admin", "owner":
		return 3
	default:
		return 0
	}
}

// getAccessibleCustomerRoles returns a map of customerID → effective role for
// the given user, combining direct CustomerAccess rows with group-based access.
func getAccessibleCustomerRoles(userID uint) map[uint]string {
	roles := make(map[uint]string)
	var direct []models.CustomerAccess
	database.DB.Where("user_id = ?", userID).Find(&direct)
	for _, a := range direct {
		roles[a.CustomerID] = a.Role
	}
	type gaRow struct {
		CustomerID uint
		Role       string
	}
	var groupRows []gaRow
	database.DB.Raw(`
		SELECT gca.customer_id, gca.role
		FROM group_customer_accesses gca
		JOIN group_members gm ON gm.group_id = gca.group_id
		WHERE gm.user_id = ?`, userID).Scan(&groupRows)
	for _, ga := range groupRows {
		if customerRoleRank(ga.Role) > customerRoleRank(roles[ga.CustomerID]) {
			roles[ga.CustomerID] = ga.Role
		}
	}
	return roles
}

// canManageCustomer returns true for global admins, users with direct "admin"
// CustomerAccess, or users whose group has "owner" access to the customer.
func canManageCustomer(c *gin.Context, customerID uint) bool {
	if middleware.GetGlobalRole(c) == "admin" {
		return true
	}
	userID := middleware.GetUserID(c)
	var count int64
	database.DB.Model(&models.CustomerAccess{}).
		Where("user_id = ? AND customer_id = ? AND role = 'admin'", userID, customerID).
		Count(&count)
	if count > 0 {
		return true
	}
	database.DB.Raw(`
		SELECT COUNT(*) FROM group_customer_accesses gca
		JOIN group_members gm ON gm.group_id = gca.group_id
		WHERE gca.customer_id = ? AND gm.user_id = ? AND gca.role = 'owner'`,
		customerID, userID).Scan(&count)
	return count > 0
}

// CustomerListItem wraps Customer with extra UI metadata.
type CustomerListItem struct {
	models.Customer
	IsFavorite     bool   `json:"is_favorite"`
	ProjectCount   int64  `json:"project_count"`
	ContractCount  int64  `json:"contract_count"`
	TicketNew           int64  `json:"ticket_new"`
	TicketPending       int64  `json:"ticket_pending"`
	TicketPendingClose  int64  `json:"ticket_pending_close"`
	TicketClosed        int64  `json:"ticket_closed"`
	MyRole              string `json:"my_role"` // "admin", "member", or "" (unrestricted / global admin)
}

// ListCustomers returns all customers with favourite status and counts.
func ListCustomers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	isAdmin := middleware.GetGlobalRole(c) == "admin"

	var customers []models.Customer
	database.DB.Where("time_tracking_only = false").Order("position asc, id asc").Find(&customers)

	// myRoles maps customerID → effective role (direct or via group) for the current user.
	myRoles := make(map[uint]string)
	if !isAdmin {
		myRoles = getAccessibleCustomerRoles(userID)
		// Non-admins only see customers they have access to.
		filtered := customers[:0]
		for _, cu := range customers {
			if _, ok := myRoles[cu.ID]; ok {
				filtered = append(filtered, cu)
			}
		}
		customers = filtered
	}

	var favs []models.CustomerFavorite
	database.DB.Where("user_id = ?", userID).Find(&favs)
	favSet := make(map[uint]bool, len(favs))
	for _, f := range favs {
		favSet[f.CustomerID] = true
	}

	items := make([]CustomerListItem, len(customers))
	for i, cust := range customers {
		var pCount, cCount, tNew, tPending, tPendingClose, tClosed int64
		database.DB.Model(&models.Project{}).Where("customer_id = ? AND deleted_at IS NULL", cust.ID).Count(&pCount)
		database.DB.Model(&models.Contract{}).Where("customer_id = ?", cust.ID).Count(&cCount)
		database.DB.Model(&models.Ticket{}).Where("customer_id = ? AND status IN ('new', 'open')", cust.ID).Count(&tNew)
		database.DB.Model(&models.Ticket{}).Where("customer_id = ? AND status = 'pending'", cust.ID).Count(&tPending)
		database.DB.Model(&models.Ticket{}).Where("customer_id = ? AND status = 'pending_close'", cust.ID).Count(&tPendingClose)
		database.DB.Model(&models.Ticket{}).Where("customer_id = ? AND status = 'closed'", cust.ID).Count(&tClosed)
		myRole := myRoles[cust.ID]
		if isAdmin {
			myRole = "admin"
		}
		items[i] = CustomerListItem{
			Customer:           cust,
			IsFavorite:         favSet[cust.ID],
			ProjectCount:       pCount,
			ContractCount:      cCount,
			TicketNew:          tNew,
			TicketPending:      tPending,
			TicketPendingClose: tPendingClose,
			TicketClosed:       tClosed,
			MyRole:             myRole,
		}
	}

	c.JSON(http.StatusOK, items)
}

// CustomerDetailResponse is the full view returned by GetCustomer.
type CustomerDetailResponse struct {
	Customer  CustomerListItem          `json:"customer"`
	Contracts []ContractGroup           `json:"contracts"`
	Projects  []models.Project          `json:"projects"` // projects with no contract
	Contacts  []models.CustomerContact  `json:"contacts"`
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

	if middleware.GetGlobalRole(c) != "admin" {
		accessible := getAccessibleCustomerRoles(userID)
		if _, ok := accessible[cust.ID]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
	}

	var fav models.CustomerFavorite
	isFav := database.DB.Where("user_id = ? AND customer_id = ?", userID, cust.ID).First(&fav).Error == nil

	var pCount, cCount int64
	database.DB.Model(&models.Project{}).Where("customer_id = ? AND deleted_at IS NULL", cust.ID).Count(&pCount)
	database.DB.Model(&models.Contract{}).Where("customer_id = ?", cust.ID).Count(&cCount)

	myRole := ""
	if middleware.GetGlobalRole(c) == "admin" {
		myRole = "admin"
	} else {
		myRole = getAccessibleCustomerRoles(userID)[cust.ID]
	}
	custItem := CustomerListItem{Customer: cust, IsFavorite: isFav, ProjectCount: pCount, ContractCount: cCount, MyRole: myRole}

	var contracts []models.Contract
	database.DB.Preload("TimeSlots").Where("customer_id = ?", cust.ID).Order("id asc").Find(&contracts)

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

	var contacts []models.CustomerContact
	database.DB.Where("customer_id = ?", cust.ID).Order("is_primary desc, id asc").Find(&contacts)

	c.JSON(http.StatusOK, CustomerDetailResponse{
		Customer:  custItem,
		Contracts: contractGroups,
		Projects:  unassigned,
		Contacts:  contacts,
	})
}

// CreateCustomer creates a new customer (admin only).
func CreateCustomer(c *gin.Context) {
	if middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}
	var req struct {
		Name              string `json:"name" binding:"required,min=1,max=200"`
		Description       string `json:"description"`
		LogoURL           string `json:"logo_url"`
		BillingStreet     string `json:"billing_street"`
		BillingCity       string `json:"billing_city"`
		BillingPostalCode string `json:"billing_postal_code"`
		BillingCountry    string `json:"billing_country"`
		VATNumber         string `json:"vat_number"`
		POReference       string `json:"po_reference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var maxPos int
	database.DB.Model(&models.Customer{}).Select("COALESCE(MAX(position),0)").Scan(&maxPos)

	cust := models.Customer{
		Name:              req.Name,
		Description:       req.Description,
		LogoURL:           req.LogoURL,
		BillingStreet:     req.BillingStreet,
		BillingCity:       req.BillingCity,
		BillingPostalCode: req.BillingPostalCode,
		BillingCountry:    req.BillingCountry,
		VATNumber:         req.VATNumber,
		POReference:       req.POReference,
		Position:          maxPos + 1,
	}
	if err := database.DB.Create(&cust).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, cust)
}

// UpdateCustomer updates a customer's metadata.
func UpdateCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if !canManageCustomer(c, uint(id)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	var cust models.Customer
	if err := database.DB.First(&cust, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}
	var req struct {
		Name              string `json:"name"`
		Description       string `json:"description"`
		LogoURL           string `json:"logo_url"`
		BillingStreet     string `json:"billing_street"`
		BillingCity       string `json:"billing_city"`
		BillingPostalCode string `json:"billing_postal_code"`
		BillingCountry    string `json:"billing_country"`
		VATNumber         string `json:"vat_number"`
		POReference       string `json:"po_reference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := map[string]interface{}{
		"description":         req.Description,
		"logo_url":            req.LogoURL,
		"billing_street":      req.BillingStreet,
		"billing_city":        req.BillingCity,
		"billing_postal_code": req.BillingPostalCode,
		"billing_country":     req.BillingCountry,
		"vat_number":          req.VATNumber,
		"po_reference":        req.POReference,
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
		Updates(map[string]any{"customer_id": nil, "contract_id": nil})
	database.DB.Where("customer_id = ?", cust.ID).Delete(&models.Contract{})
	database.DB.Where("customer_id = ?", cust.ID).Delete(&models.CustomerFavorite{})
	database.DB.Where("customer_id = ?", cust.ID).Delete(&models.CustomerAccess{})
	database.DB.Where("customer_id = ?", cust.ID).Delete(&models.GroupCustomerAccess{})
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
	userID := middleware.GetUserID(c)
	role := middleware.GetGlobalRole(c)
	if err := requireCustomerAccess(uint(id), userID, role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var contracts []models.Contract
	database.DB.Preload("TimeSlots").Where("customer_id = ?", id).Order("id asc").Find(&contracts)
	c.JSON(http.StatusOK, contracts)
}

// CreateContract creates a contract under a customer.
func CreateContract(c *gin.Context) {
	custID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if !canManageCustomer(c, uint(custID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	var cust models.Customer
	if err := database.DB.First(&cust, custID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}
	var req struct {
		Name          string   `json:"name" binding:"required,min=1,max=200"`
		Description   string   `json:"description"`
		StartDate     string   `json:"start_date"`
		EndDate       string   `json:"end_date"`
		PricePerHour  *float64 `json:"price_per_hour"`
		PricePerKm    *float64 `json:"price_per_km"`
		Currency      string   `json:"currency"`
		TimeSlots     []struct {
			Label                string   `json:"label"`
			StartTime            string   `json:"start_time"`
			EndTime              string   `json:"end_time"`
			DayType              string   `json:"day_type"`
			EndDayOffset         int      `json:"end_day_offset"`
			MultiplicationFactor *float64 `json:"multiplication_factor"`
			HourlyRate           *float64 `json:"hourly_rate"`
		} `json:"time_slots"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	currency := req.Currency
	if currency == "" {
		currency = "€"
	}
	con := models.Contract{
		CustomerID:   uint(custID),
		Name:         req.Name,
		Description:  req.Description,
		PricePerHour: req.PricePerHour,
		PricePerKm:   req.PricePerKm,
		Currency:     currency,
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
	for _, s := range req.TimeSlots {
		if s.StartTime == "" || s.EndTime == "" {
			continue
		}
		dayType := s.DayType
		if dayType == "" {
			dayType = "all"
		}
		database.DB.Create(&models.ContractTimeSlot{
			ContractID:           con.ID,
			Label:                s.Label,
			StartTime:            s.StartTime,
			EndTime:              s.EndTime,
			DayType:              dayType,
			EndDayOffset:         s.EndDayOffset,
			MultiplicationFactor: s.MultiplicationFactor,
			HourlyRate:           s.HourlyRate,
		})
	}
	database.DB.Preload("TimeSlots").First(&con, con.ID)
	c.JSON(http.StatusCreated, con)
}

// UpdateContract updates a contract.
func UpdateContract(c *gin.Context) {
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
	if !canManageCustomer(c, con.CustomerID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	var req struct {
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		StartDate     string   `json:"start_date"`
		EndDate       string   `json:"end_date"`
		PricePerHour  *float64 `json:"price_per_hour"`
		PricePerKm    *float64 `json:"price_per_km"`
		Currency      string   `json:"currency"`
		TimeSlots     []struct {
			Label                string   `json:"label"`
			StartTime            string   `json:"start_time"`
			EndTime              string   `json:"end_time"`
			DayType              string   `json:"day_type"`
			EndDayOffset         int      `json:"end_day_offset"`
			MultiplicationFactor *float64 `json:"multiplication_factor"`
			HourlyRate           *float64 `json:"hourly_rate"`
		} `json:"time_slots"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
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
	updates["price_per_hour"] = req.PricePerHour
	updates["price_per_km"] = req.PricePerKm
	if req.Currency != "" {
		updates["currency"] = req.Currency
	}
	database.DB.Model(&con).Updates(updates)
	database.DB.Where("contract_id = ?", con.ID).Delete(&models.ContractTimeSlot{})
	for _, s := range req.TimeSlots {
		if s.StartTime == "" || s.EndTime == "" {
			continue
		}
		dayType := s.DayType
		if dayType == "" {
			dayType = "all"
		}
		database.DB.Create(&models.ContractTimeSlot{
			ContractID:           con.ID,
			Label:                s.Label,
			StartTime:            s.StartTime,
			EndTime:              s.EndTime,
			DayType:              dayType,
			EndDayOffset:         s.EndDayOffset,
			MultiplicationFactor: s.MultiplicationFactor,
			HourlyRate:           s.HourlyRate,
		})
	}
	database.DB.Preload("TimeSlots").First(&con, contractID)
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
	database.DB.Where("contract_id = ?", con.ID).Delete(&models.ContractTimeSlot{})
	database.DB.Delete(&con)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// CustomerContractRates groups a customer with contracts that have at least one time slot.
type CustomerContractRates struct {
	CustomerID   uint              `json:"customer_id"`
	CustomerName string            `json:"customer_name"`
	Contracts    []models.Contract `json:"contracts"`
}

// ListAllContractRates returns all accessible customers with contracts that have time slots.
func ListAllContractRates(c *gin.Context) {
	userID := middleware.GetUserID(c)
	isAdmin := middleware.GetGlobalRole(c) == "admin"

	var customers []models.Customer
	database.DB.Order("position asc, id asc").Find(&customers)

	if !isAdmin {
		myRoles := getAccessibleCustomerRoles(userID)
		filtered := customers[:0]
		for _, cu := range customers {
			if _, ok := myRoles[cu.ID]; ok {
				filtered = append(filtered, cu)
			}
		}
		customers = filtered
	}

	result := make([]CustomerContractRates, 0)
	for _, cust := range customers {
		var contracts []models.Contract
		database.DB.Preload("TimeSlots").Where("customer_id = ?", cust.ID).Order("id asc").Find(&contracts)
		withSlots := make([]models.Contract, 0)
		for _, con := range contracts {
			if len(con.TimeSlots) > 0 {
				withSlots = append(withSlots, con)
			}
		}
		if len(withSlots) > 0 {
			result = append(result, CustomerContractRates{
				CustomerID:   cust.ID,
				CustomerName: cust.Name,
				Contracts:    withSlots,
			})
		}
	}

	c.JSON(http.StatusOK, result)
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
