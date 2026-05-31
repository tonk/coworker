package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// GetMetrics GET /api/v1/metrics — Prometheus-format metrics.
// Requires admin or metrics global role (enforced by MetricsAuth middleware).
func GetMetrics(c *gin.Context) {
	var projects []models.Project
	database.DB.Where("is_archived = ? AND deleted_at IS NULL", false).Find(&projects)

	var buf strings.Builder

	// warmdesk_projects_total
	fmt.Fprintf(&buf, "# HELP warmdesk_projects_total Number of active (non-archived) projects\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_projects_total gauge\n")
	fmt.Fprintf(&buf, "warmdesk_projects_total %d\n", len(projects))

	// warmdesk_columns_total per project
	fmt.Fprintf(&buf, "\n# HELP warmdesk_columns_total Number of columns per project\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_columns_total gauge\n")
	for _, p := range projects {
		var count int64
		database.DB.Model(&models.Column{}).Where("project_id = ?", p.ID).Count(&count)
		fmt.Fprintf(&buf, "warmdesk_columns_total{project=%q,project_name=%q} %d\n", p.Slug, p.Name, count)
	}

	// warmdesk_cards_total per column, by status
	fmt.Fprintf(&buf, "\n# HELP warmdesk_cards_total Number of cards per column and status\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_cards_total gauge\n")
	for _, p := range projects {
		var columns []models.Column
		database.DB.Where("project_id = ?", p.ID).Find(&columns)
		for _, col := range columns {
			var open, closed int64
			database.DB.Model(&models.Card{}).
				Where("column_id = ? AND closed = false AND deleted_at IS NULL", col.ID).
				Count(&open)
			database.DB.Model(&models.Card{}).
				Where("column_id = ? AND closed = true AND deleted_at IS NULL", col.ID).
				Count(&closed)
			fmt.Fprintf(&buf, "warmdesk_cards_total{project=%q,column=%q,status=\"open\"} %d\n",
				p.Slug, col.Name, open)
			fmt.Fprintf(&buf, "warmdesk_cards_total{project=%q,column=%q,status=\"closed\"} %d\n",
				p.Slug, col.Name, closed)
		}
	}

	// warmdesk_users_total — active and inactive counts per global role
	fmt.Fprintf(&buf, "\n# HELP warmdesk_users_total Number of users per global role and active state\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_users_total gauge\n")
	for _, role := range []string{"admin", "user", "viewer", "metrics", "backup", "customer"} {
		var active, inactive int64
		database.DB.Model(&models.User{}).Where("global_role = ? AND is_active = true AND deleted_at IS NULL", role).Count(&active)
		database.DB.Model(&models.User{}).Where("global_role = ? AND is_active = false AND deleted_at IS NULL", role).Count(&inactive)
		fmt.Fprintf(&buf, "warmdesk_users_total{role=%q,active=\"true\"} %d\n", role, active)
		fmt.Fprintf(&buf, "warmdesk_users_total{role=%q,active=\"false\"} %d\n", role, inactive)
	}

	// warmdesk_customers_total
	var customerTotal int64
	database.DB.Model(&models.Customer{}).Count(&customerTotal)
	fmt.Fprintf(&buf, "\n# HELP warmdesk_customers_total Total number of customers\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_customers_total gauge\n")
	fmt.Fprintf(&buf, "warmdesk_customers_total %d\n", customerTotal)

	// warmdesk_tickets_total — by status and by priority
	fmt.Fprintf(&buf, "\n# HELP warmdesk_tickets_total Number of tickets by status\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_tickets_total gauge\n")
	for _, status := range []string{"new", "open", "pending", "pending_close", "closed"} {
		var count int64
		database.DB.Model(&models.Ticket{}).Where("status = ?", status).Count(&count)
		fmt.Fprintf(&buf, "warmdesk_tickets_total{status=%q} %d\n", status, count)
	}

	fmt.Fprintf(&buf, "\n# HELP warmdesk_tickets_by_priority_total Number of open tickets by priority\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_tickets_by_priority_total gauge\n")
	for _, pri := range []string{"low", "medium", "high", "critical"} {
		var count int64
		database.DB.Model(&models.Ticket{}).
			Where("priority = ? AND status NOT IN ('closed','pending_close')", pri).
			Count(&count)
		fmt.Fprintf(&buf, "warmdesk_tickets_by_priority_total{priority=%q} %d\n", pri, count)
	}

	// warmdesk_sla_breaches_total — response and resolution breaches on open tickets
	var respBreaches, resolBreaches int64
	database.DB.Model(&models.Ticket{}).
		Where("sla_response_breached = true AND status NOT IN ('closed','pending_close')").
		Count(&respBreaches)
	database.DB.Model(&models.Ticket{}).
		Where("sla_resolution_breached = true AND status NOT IN ('closed','pending_close')").
		Count(&resolBreaches)
	fmt.Fprintf(&buf, "\n# HELP warmdesk_sla_breaches_total Number of open tickets currently breaching SLA\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_sla_breaches_total gauge\n")
	fmt.Fprintf(&buf, "warmdesk_sla_breaches_total{type=\"response\"} %d\n", respBreaches)
	fmt.Fprintf(&buf, "warmdesk_sla_breaches_total{type=\"resolution\"} %d\n", resolBreaches)

	// warmdesk_ticket_messages_total — public vs private
	var publicMsgs, privateMsgs int64
	database.DB.Model(&models.TicketMessage{}).Where("is_private = false OR is_private IS NULL").Count(&publicMsgs)
	database.DB.Model(&models.TicketMessage{}).Where("is_private = true").Count(&privateMsgs)
	fmt.Fprintf(&buf, "\n# HELP warmdesk_ticket_messages_total Total ticket messages by visibility\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_ticket_messages_total gauge\n")
	fmt.Fprintf(&buf, "warmdesk_ticket_messages_total{visibility=\"public\"} %d\n", publicMsgs)
	fmt.Fprintf(&buf, "warmdesk_ticket_messages_total{visibility=\"private\"} %d\n", privateMsgs)

	// warmdesk_backup_* metrics
	all := loadAllSettings()

	var lastRunTS float64
	if lr := all[settingBackupLastRun]; lr != "" {
		if t, err := time.Parse(time.RFC3339, lr); err == nil {
			lastRunTS = float64(t.Unix())
		}
	}
	lastSuccess := -1.0 // -1 = never run
	if s := all[settingBackupLastSuccess]; s == "true" {
		lastSuccess = 1
	} else if s == "false" {
		lastSuccess = 0
	}

	var backupFileCount int
	if entries, err := os.ReadDir(backupsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), "warmdesk_db_") {
				backupFileCount++
			}
		}
	}

	fmt.Fprintf(&buf, "\n# HELP warmdesk_backup_last_run_timestamp_seconds Unix timestamp of the last backup attempt (0 if never)\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_backup_last_run_timestamp_seconds gauge\n")
	fmt.Fprintf(&buf, "warmdesk_backup_last_run_timestamp_seconds %g\n", lastRunTS)

	fmt.Fprintf(&buf, "\n# HELP warmdesk_backup_last_success Last backup result: 1=success 0=failed -1=never run\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_backup_last_success gauge\n")
	fmt.Fprintf(&buf, "warmdesk_backup_last_success %g\n", lastSuccess)

	fmt.Fprintf(&buf, "\n# HELP warmdesk_backup_files_total Number of backup files currently stored\n")
	fmt.Fprintf(&buf, "# TYPE warmdesk_backup_files_total gauge\n")
	fmt.Fprintf(&buf, "warmdesk_backup_files_total %d\n", backupFileCount)

	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(http.StatusOK, buf.String())
}
