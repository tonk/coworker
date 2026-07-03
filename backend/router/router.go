package router

import (
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tonk/warmdesk/handlers"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/services"
)

func Setup(authSvc *services.AuthService, allowedOrigins string, webFS fs.FS, apiLog bool, uploadDir string, trustedProxies []string, appMode string) *gin.Engine {
	ttMode := appMode == "timetracking"
	r := gin.New()
	r.SetTrustedProxies(trustedProxies) //nolint
	r.Use(gin.Recovery())
	if apiLog {
		r.Use(gin.Logger())
	}
	r.Use(middleware.CORS(allowedOrigins))
	r.Use(middleware.SecurityHeaders())

	authHandler := handlers.NewAuthHandler(authSvc)
	passkeyHandler := handlers.NewPasskeyHandler(authSvc)
	wsHandler := handlers.NewWSHandler(authSvc, allowedOrigins)
	handlers.InitAttachmentAuth(authSvc)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.IPAllowlist())

	// Swagger UI — protected by the same IP allowlist as the API
	swagger := r.Group("/swagger")
	swagger.Use(middleware.IPAllowlist())
	{
		swagger.GET("", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/swagger/index.html") })
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Public endpoints
	v1.GET("/version", handlers.GetVersion)
	v1.GET("/system/settings", handlers.GetSystemSettings)
	v1.GET("/media/proxy", middleware.MediaProxyRateLimit(), handlers.ProxyImage) // no auth — img tags can't send Bearer tokens

	// Auth routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", middleware.RegisterRateLimit(), authHandler.Register)
		auth.POST("/login", middleware.AuthRateLimit(), authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", middleware.AuthRateLimit(), authHandler.Refresh)
		auth.POST("/mfa/verify", middleware.AuthRateLimit(), authHandler.MFAVerify)
		auth.POST("/forgot-password", middleware.ResetRateLimit(), authHandler.ForgotPassword)
		auth.POST("/reset-password", middleware.ResetRateLimit(), authHandler.ResetPassword)

		// Passkey authentication (public — identity resolved from credential)
		auth.POST("/passkey/login/begin", middleware.AuthRateLimit(), passkeyHandler.PasskeyLoginBegin)
		auth.POST("/passkey/login/finish", middleware.AuthRateLimit(), passkeyHandler.PasskeyLoginFinish)
	}

	// Authenticated routes
	protected := v1.Group("")
	protected.Use(middleware.Auth(authSvc))
	{
		// Current user
		protected.GET("/auth/me", authHandler.Me)
		protected.PUT("/auth/me", authHandler.UpdateMe)
		protected.PUT("/auth/me/password", authHandler.ChangePassword)

		// MFA management
		protected.GET("/auth/mfa/setup", authHandler.MFASetup)
		protected.POST("/auth/mfa/enable", authHandler.MFAEnable)
		protected.POST("/auth/mfa/disable", authHandler.MFADisable)
		protected.GET("/auth/trusted-devices", authHandler.ListTrustedDevices)
		protected.DELETE("/auth/trusted-devices/:id", authHandler.RevokeTrustedDevice)
		protected.DELETE("/auth/trusted-devices", authHandler.RevokeAllTrustedDevices)

		// Passkey management (registration requires an authenticated session)
		protected.GET("/auth/passkey/register/begin", passkeyHandler.PasskeyRegisterBegin)
		protected.POST("/auth/passkey/register/finish", passkeyHandler.PasskeyRegisterFinish)
		protected.GET("/auth/passkeys", passkeyHandler.PasskeyList)
		protected.DELETE("/auth/passkeys/:id", passkeyHandler.PasskeyDelete)

		// Short-lived purpose-limited tickets (keep long-lived JWTs out of URLs)
		protected.POST("/auth/ws-ticket", authHandler.IssueWSTicket)
		protected.POST("/auth/media-ticket", authHandler.IssueMediaTicket)

		// API keys (personal tokens)
		protected.GET("/auth/api-keys", handlers.ListAPIKeys)
		protected.POST("/auth/api-keys", handlers.CreateAPIKey)
		protected.DELETE("/auth/api-keys/:id", handlers.DeleteAPIKey)

		// Admin
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/users", handlers.AdminListUsers)
			admin.POST("/users", handlers.AdminCreateUser)
			admin.GET("/users/:id", handlers.AdminGetUser)
			admin.PUT("/users/:id", handlers.AdminUpdateUser)
			admin.DELETE("/users/:id", handlers.AdminDeleteUser)
			admin.POST("/users/:id/restore", handlers.AdminRestoreUser)
			admin.DELETE("/users/:id/purge", handlers.AdminPurgeUser)
			admin.POST("/users/:id/mfa/disable", handlers.AdminDisableUserMFA)
			if !ttMode {
				admin.GET("/users/:id/projects", handlers.AdminGetUserProjects)
				admin.PUT("/users/:id/projects", handlers.AdminSetUserProjects)
			}
			admin.GET("/users/:id/customers", handlers.AdminGetUserCustomers)
			admin.PUT("/users/:id/customers", handlers.AdminSetUserCustomers)
			admin.GET("/users/:id/groups", handlers.AdminGetUserGroups)
			admin.PUT("/users/:id/groups", handlers.AdminSetUserGroups)
			admin.GET("/users/:id/login-history", handlers.AdminGetUserLoginHistory)
			admin.GET("/users/:id/api-keys", handlers.AdminListUserAPIKeys)
			admin.POST("/users/:id/api-keys", handlers.AdminCreateUserAPIKey)
			admin.DELETE("/users/:id/api-keys/:keyId", handlers.AdminDeleteUserAPIKey)
			if !ttMode {
				admin.GET("/projects", handlers.AdminListProjects)
				admin.POST("/projects", handlers.AdminCreateProject)
				admin.PUT("/projects/:id", handlers.AdminUpdateProject)
				admin.DELETE("/projects/:id", handlers.AdminDeleteProject)
				admin.POST("/projects/:id/restore", handlers.AdminRestoreProject)
				admin.DELETE("/projects/:id/purge", handlers.AdminPurgeProject)
			}
			admin.GET("/system", handlers.AdminGetSystemSettings)
			admin.PUT("/system", handlers.AdminUpdateSystemSettings)
			admin.POST("/system/test-email", handlers.AdminSendTestEmail)
			admin.POST("/system/backup", handlers.AdminBackupDatabase)
			admin.GET("/system/backups", handlers.AdminListBackups)
			admin.POST("/system/backups/restore", handlers.AdminRestoreBackup)
			admin.GET("/system/backups/:filename", handlers.AdminDownloadBackup)
			admin.DELETE("/system/backups/:filename", handlers.AdminDeleteBackup)

			// Groups (admin-only CRUD + access management)
			admin.GET("/groups", handlers.AdminListGroups)
			admin.POST("/groups", handlers.AdminCreateGroup)
			admin.GET("/groups/:groupId", handlers.AdminGetGroup)
			admin.PATCH("/groups/:groupId", handlers.AdminUpdateGroup)
			admin.DELETE("/groups/:groupId", handlers.AdminDeleteGroup)
			admin.POST("/groups/:groupId/members", handlers.AdminAddGroupMember)
			admin.DELETE("/groups/:groupId/members/:userId", handlers.AdminRemoveGroupMember)
			admin.PUT("/groups/:groupId/projects/:projectId", handlers.AdminSetGroupProjectAccess)
			admin.DELETE("/groups/:groupId/projects/:projectId", handlers.AdminRemoveGroupProjectAccess)
			admin.PUT("/groups/:groupId/customers/:customerId", handlers.AdminSetGroupCustomerAccess)
			admin.DELETE("/groups/:groupId/customers/:customerId", handlers.AdminRemoveGroupCustomerAccess)

			// News items
			admin.GET("/news", handlers.AdminListNews)
			admin.POST("/news", handlers.AdminCreateNews)
			admin.PUT("/news/:id", handlers.AdminUpdateNews)
			admin.DELETE("/news/:id", handlers.AdminDeleteNews)

			if !ttMode {
				// SLA policies (helpdesk)
				admin.GET("/sla-policies", handlers.AdminListSlaPolicies)
				admin.POST("/sla-policies", handlers.AdminCreateSlaPolicy)
				admin.PUT("/sla-policies/:id", handlers.AdminUpdateSlaPolicy)
				admin.DELETE("/sla-policies/:id", handlers.AdminDeleteSlaPolicy)

				// Macros (helpdesk)
				admin.GET("/macros", handlers.AdminListMacros)
				admin.POST("/macros", handlers.AdminCreateMacro)
				admin.PUT("/macros/:id", handlers.AdminUpdateMacro)
				admin.DELETE("/macros/:id", handlers.AdminDeleteMacro)

				// Ticket checklist templates (helpdesk)
				admin.GET("/ticket-checklist-templates", handlers.AdminListTicketChecklistTemplates)
				admin.POST("/ticket-checklist-templates", handlers.AdminCreateTicketChecklistTemplate)
				admin.PUT("/ticket-checklist-templates/:id", handlers.AdminUpdateTicketChecklistTemplate)
				admin.DELETE("/ticket-checklist-templates/:id", handlers.AdminDeleteTicketChecklistTemplate)

				// Invoice templates
				admin.POST("/invoice-templates", handlers.AdminCreateInvoiceTemplate)
				admin.PUT("/invoice-templates/:id", handlers.AdminUpdateInvoiceTemplate)
				admin.DELETE("/invoice-templates/:id", handlers.AdminDeleteInvoiceTemplate)

				// IMAP test & poll
				admin.POST("/imap/test", handlers.AdminTestIMAP)
				admin.POST("/imap/poll", handlers.AdminPollIMAP)
				// IMAP OAuth2 authorization
				admin.GET("/imap/oauth2/auth-url", handlers.AdminIMAPOAuth2AuthURL)
				admin.GET("/imap/oauth2/callback", handlers.AdminIMAPOAuth2Callback)
				admin.GET("/imap/oauth2/status", handlers.AdminIMAPOAuth2Status)
				admin.POST("/imap/oauth2/disconnect", handlers.AdminIMAPOAuth2Disconnect)
			}
		}

		// News (active items visible to all authenticated users)
		protected.GET("/news", handlers.ListActiveNews)

		protected.GET("/docs/user-guide.pdf", handlers.DownloadUserGuide)
		protected.GET("/docs/admin-guide.pdf", middleware.AdminOnly(), handlers.DownloadAdminGuide)

		if !ttMode {
			// Macros — active list visible to all helpdesk users
			protected.GET("/macros", handlers.ListMacros)

			// Ticket checklist templates — active list visible to all helpdesk users
			protected.GET("/ticket-checklist-templates", handlers.ListTicketChecklistTemplates)
		}

		// Users (for direct messages / user lookup)
		protected.GET("/users", handlers.ListAllUsers)

		if !ttMode {
			// Online presence (global)
			protected.GET("/online-users", handlers.GetOnlineUsers)

			// Favorite users
			protected.GET("/favorite-users", handlers.ListFavoriteUsers)
			protected.POST("/favorite-users/:userId", handlers.AddFavoriteUser)
			protected.DELETE("/favorite-users/:userId", handlers.RemoveFavoriteUser)

			// Starred projects
			protected.GET("/starred-projects", handlers.ListStarredProjects)
		}

		// Customers and contracts
		customers := protected.Group("/customers")
		{
			customers.GET("", handlers.ListCustomers)
			customers.POST("", handlers.CreateCustomer)
			customers.GET("/rates", handlers.ListAllContractRates)
			customers.GET("/:customerId", handlers.GetCustomer)
			customers.PUT("/:customerId", handlers.UpdateCustomer)
			customers.DELETE("/:customerId", handlers.DeleteCustomer)
			customers.POST("/:customerId/favorite", handlers.AddCustomerFavorite)
			customers.DELETE("/:customerId/favorite", handlers.RemoveCustomerFavorite)
			customers.GET("/:customerId/contracts", handlers.ListContracts)
			customers.POST("/:customerId/contracts", handlers.CreateContract)
			customers.PUT("/:customerId/contracts/:contractId", handlers.UpdateContract)
			customers.DELETE("/:customerId/contracts/:contractId", handlers.DeleteContract)
			customers.GET("/:customerId/invoices", handlers.ListInvoices)
			customers.POST("/:customerId/invoices", handlers.CreateInvoice)
			customers.GET("/:customerId/invoices/:invoiceId", handlers.GetInvoice)
			customers.PUT("/:customerId/invoices/:invoiceId", handlers.UpdateInvoice)
			customers.DELETE("/:customerId/invoices/:invoiceId", handlers.DeleteInvoice)
			customers.GET("/:customerId/invoices/:invoiceId/pdf", handlers.GetInvoicePDF)
			customers.POST("/:customerId/invoices/:invoiceId/send", handlers.SendInvoice)
			customers.POST("/:customerId/invoices/:invoiceId/credit-note", handlers.CreateCreditNote)
			customers.GET("/:customerId/contacts", handlers.ListContacts)
			customers.POST("/:customerId/contacts", handlers.CreateContact)
			customers.PUT("/:customerId/contacts/:contactId", handlers.UpdateContact)
			customers.DELETE("/:customerId/contacts/:contactId", handlers.DeleteContact)
			customers.GET("/:customerId/members", handlers.ListCustomerMembers)
			customers.PUT("/:customerId/members", handlers.SetCustomerMembers)
			// Group membership management delegated to customer owners
			customers.GET("/:customerId/groups", handlers.ListCustomerGroups)
			customers.POST("/:customerId/groups/:groupId/members", handlers.CustomerAddGroupMember)
			customers.DELETE("/:customerId/groups/:groupId/members/:userId", handlers.CustomerRemoveGroupMember)

			if !ttMode {
				inbox := protected.Group("/tickets")
				inbox.Use(middleware.RequireFeature("helpdesk_enabled"))
				{
					inbox.GET("/inbox", handlers.ListInboxTickets)
					inbox.POST("/inbox", handlers.CreateInboxTicket)
					inbox.GET("/inbox/:ticketId", handlers.GetInboxTicket)
					inbox.GET("/inbox/:ticketId/viewers", handlers.GetInboxTicketViewers)
					inbox.PUT("/inbox/:ticketId", handlers.UpdateInboxTicket)
					inbox.POST("/inbox/:ticketId/messages", handlers.CreateInboxTicketMessage)
					inbox.PATCH("/inbox/:ticketId/messages/:msgId", handlers.UpdateInboxTicketMessage)
					inbox.DELETE("/inbox/:ticketId", handlers.DeleteInboxTicket)
					inbox.POST("/inbox/:ticketId/macros/:macroId", handlers.ApplyInboxMacro)
					inbox.POST("/inbox/:ticketId/checklist/templates/:templateId", handlers.ApplyInboxTicketChecklistTemplate)
					inbox.PUT("/inbox/:ticketId/checklist/:itemId", handlers.UpdateInboxTicketChecklistItem)
					inbox.DELETE("/inbox/:ticketId/checklist/:itemId", handlers.DeleteInboxTicketChecklistItem)
					inbox.PATCH("/inbox/:ticketId/checklist/reorder", handlers.ReorderInboxTicketChecklistItems)
					inbox.POST("/inbox/:ticketId/spam", handlers.MarkInboxSpam)
					inbox.DELETE("/inbox/:ticketId/spam", handlers.UnmarkInboxSpam)
				}

				// Tickets (helpdesk)
				tickets := customers.Group("/:customerId/tickets")
				tickets.Use(middleware.RequireFeature("helpdesk_enabled"))
				{
					tickets.GET("", handlers.ListTickets)
					tickets.POST("", handlers.CreateTicket)
					tickets.GET("/:ticketId", handlers.GetTicket)
					tickets.PUT("/:ticketId", handlers.UpdateTicket)
					tickets.PUT("/:ticketId/move", handlers.MoveTicket)
					tickets.DELETE("/:ticketId", handlers.DeleteTicket)
					tickets.POST("/:ticketId/messages", handlers.CreateTicketMessage)
					tickets.PATCH("/:ticketId/messages/:msgId", handlers.UpdateTicketMessage)
					tickets.POST("/:ticketId/tags", handlers.AddTicketTag)
					tickets.DELETE("/:ticketId/tags/:tagId", handlers.RemoveTicketTag)
					tickets.GET("/:ticketId/links", handlers.ListTicketLinks)
					tickets.POST("/:ticketId/links", handlers.CreateTicketLink)
					tickets.DELETE("/:ticketId/links/:linkId", handlers.DeleteTicketLink)
					tickets.GET("/:ticketId/cards", handlers.ListTicketCardLinks)
					tickets.POST("/:ticketId/cards", handlers.CreateTicketCardLink)
					tickets.POST("/:ticketId/create-card", handlers.CreateCardFromTicket)
					tickets.DELETE("/:ticketId/cards/:linkId", handlers.DeleteTicketCardLink)
					tickets.GET("/:ticketId/viewers", handlers.GetTicketViewers)
					tickets.GET("/:ticketId/history", handlers.GetTicketHistory)
					tickets.GET("/:ticketId/raw-email", handlers.GetTicketRawEmail)
					tickets.POST("/:ticketId/macros/:macroId", handlers.ApplyMacro)
					tickets.POST("/:ticketId/checklist/templates/:templateId", handlers.ApplyTicketChecklistTemplate)
					tickets.PUT("/:ticketId/checklist/:itemId", handlers.UpdateTicketChecklistItem)
					tickets.DELETE("/:ticketId/checklist/:itemId", handlers.DeleteTicketChecklistItem)
					tickets.PATCH("/:ticketId/checklist/reorder", handlers.ReorderTicketChecklistItems)
					tickets.POST("/:ticketId/spam", handlers.MarkSpam)
					tickets.DELETE("/:ticketId/spam", handlers.UnmarkSpam)
				}
			}
		}

		if !ttMode {
			// Direct messages (legacy 1-on-1)
			dm := protected.Group("/direct-messages")
			dm.Use(middleware.BlockCustomerRole(), middleware.RequireFeature("chat_enabled"))
			{
				dm.GET("/conversations", handlers.ListConversations)
				dm.GET("/:userId", handlers.ListDirectMessages)
				dm.POST("/:userId", middleware.MessageRateLimit(), handlers.SendDirectMessage)
				dm.DELETE("/:userId/:msgId", handlers.DeleteDirectMessage)
			}
		}

		// Image upload (avatars, logos)
		protected.POST("/upload/image", handlers.UploadImage)

		// File attachments
		protected.POST("/attachments", handlers.UploadAttachment)
		protected.DELETE("/attachments/:id", handlers.DeleteAttachment)

		if !ttMode {
			// Global search
			protected.GET("/search", handlers.GlobalSearch)

			// Card reference resolver — resolves "PRJ-42" to project slug + card ID
			protected.GET("/cards/resolve/:ref", handlers.ResolveCardRef)

			// Link preview — fetches OG metadata for a URL
			protected.GET("/link-preview", handlers.LinkPreview)
		}

		// Reports
		protected.GET("/reports/time", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeReport)
		protected.GET("/reports/time/pdf", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeReportPDF)
		protected.GET("/reports/time/xlsx", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeReportXLSX)

		// Time-tracking-only customers and projects (no board/CRM created)
		protected.GET("/time-tracking-customers", middleware.RequireFeature("time_tracking_enabled"), handlers.ListTimeTrackingCustomers)
		protected.POST("/time-tracking-customers", middleware.RequireFeature("time_tracking_enabled"), handlers.CreateTimeTrackingCustomer)
		protected.PUT("/time-tracking-customers/:id", middleware.RequireFeature("time_tracking_enabled"), handlers.UpdateTimeTrackingCustomer)
		protected.DELETE("/time-tracking-customers/:id", middleware.RequireFeature("time_tracking_enabled"), handlers.DeleteTimeTrackingCustomer)

		// Time-tracking-only projects (no board created)
		protected.GET("/time-tracking-projects", middleware.RequireFeature("time_tracking_enabled"), handlers.ListTimeTrackingProjects)
		protected.POST("/time-tracking-projects", middleware.RequireFeature("time_tracking_enabled"), handlers.CreateTimeTrackingProject)
		protected.PUT("/time-tracking-projects/:id", middleware.RequireFeature("time_tracking_enabled"), handlers.UpdateTimeTrackingProject)
		protected.DELETE("/time-tracking-projects/:id", middleware.RequireFeature("time_tracking_enabled"), handlers.DeleteTimeTrackingProject)

		// Global invoice list (across all accessible customers)
		protected.GET("/invoices", handlers.ListAllInvoices)

		// Invoice templates (readable by all, writeable by admin only — see admin group above)
		protected.GET("/invoice-templates", handlers.ListInvoiceTemplates)

		// Time entries (personal time registration)
		protected.GET("/time-entries", middleware.RequireFeature("time_tracking_enabled"), handlers.ListTimeEntries)
		protected.POST("/time-entries", middleware.RequireFeature("time_tracking_enabled"), handlers.CreateTimeEntry)
		protected.PUT("/time-entries/:id", middleware.RequireFeature("time_tracking_enabled"), handlers.UpdateTimeEntry)
		protected.DELETE("/time-entries/:id", middleware.RequireFeature("time_tracking_enabled"), handlers.DeleteTimeEntry)
		protected.GET("/time-entries/report", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeEntryReport)
		protected.GET("/time-entries/report/pdf", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeEntryReportPDF)
		protected.GET("/time-entries/report/chart-pdf", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeEntryReportChartPDF)
		protected.GET("/time-entries/report/xlsx", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeEntryReportXLSX)
		protected.GET("/time-entries/sheet/xlsx", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeEntrySheetXLSX)
		protected.GET("/time-entries/grid/pdf", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeSheetGridPDF)
		protected.POST("/time-entries/holidays", middleware.RequireFeature("time_tracking_enabled"), handlers.AddHolidays)
		protected.GET("/time-entries/row-order", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeEntryRowOrder)
		protected.PUT("/time-entries/row-order", middleware.RequireFeature("time_tracking_enabled"), handlers.UpdateTimeEntryRowOrder)
		protected.GET("/time-entries/macro-library", middleware.RequireFeature("time_tracking_enabled"), handlers.GetTimeMacroLibrary)
		protected.PUT("/time-entries/macro-library", middleware.RequireFeature("time_tracking_enabled"), handlers.UpdateTimeMacroLibrary)

		// Prometheus metrics (admin or metrics role)
		protected.GET("/metrics", middleware.MetricsAuth(), handlers.GetMetrics)

		// Automated backup trigger (admin or backup role)
		protected.POST("/backup", middleware.BackupAuth(), handlers.AdminBackupDatabase)

		if !ttMode {
		// Conversations (1-on-1 and group)
		convs := protected.Group("/conversations")
		convs.Use(middleware.BlockCustomerRole(), middleware.RequireFeature("chat_enabled"))
		{
			convs.GET("", handlers.GetConversations)
			convs.POST("", handlers.CreateConversation)
			convs.GET("/:id/messages", handlers.GetConversationMessages)
			convs.POST("/:id/messages", middleware.MessageRateLimit(), handlers.SendConversationMessage)
			convs.PATCH("/:id/messages/:msgId", handlers.EditConversationMessage)
			convs.DELETE("/:id/messages/:msgId", handlers.DeleteConversationMessage)
			convs.DELETE("/:id", handlers.LeaveConversation)
			convs.POST("/:id/members", handlers.AddConversationMember)
			convs.DELETE("/:id/members/:userId", handlers.RemoveConversationMember)
			convs.POST("/:id/avatar", handlers.UploadConversationAvatar)
			convs.POST("/:id/messages/:msgId/reactions", handlers.ToggleConvReaction)
			convs.GET("/:id/livekit-token", handlers.GetLiveKitToken)
		}

		// Projects
		projects := protected.Group("/projects")
		projects.Use(middleware.BlockCustomerRole(), middleware.RequireFeature("board_enabled"))
		{
			projects.GET("", handlers.ListProjects)
			projects.POST("", handlers.CreateProject)
			projects.PATCH("/reorder", handlers.ReorderProjects)
			projects.GET("/:projectSlug", handlers.GetProject)
			projects.PUT("/:projectSlug", handlers.UpdateProject)
			projects.DELETE("/:projectSlug", handlers.DeleteProject)

			// Members
			projects.GET("/:projectSlug/members", handlers.ListMembers)
			projects.POST("/:projectSlug/members", handlers.InviteMember)
			projects.PUT("/:projectSlug/members/:userId/role", handlers.UpdateMemberRole)
			// Group membership management delegated to project owners
			projects.GET("/:projectSlug/groups", handlers.ListProjectGroups)
			projects.POST("/:projectSlug/groups/:groupId/members", handlers.ProjectAddGroupMember)
			projects.DELETE("/:projectSlug/groups/:groupId/members/:userId", handlers.ProjectRemoveGroupMember)
			projects.DELETE("/:projectSlug/members/:userId", handlers.RemoveMember)

			// Labels
			projects.GET("/:projectSlug/labels", handlers.ListLabels)
			projects.POST("/:projectSlug/labels", handlers.CreateLabel)
			projects.PUT("/:projectSlug/labels/:labelId", handlers.UpdateLabel)
			projects.DELETE("/:projectSlug/labels/:labelId", handlers.DeleteLabel)

			// Columns
			projects.GET("/:projectSlug/columns", handlers.ListColumns)
			projects.POST("/:projectSlug/columns", handlers.CreateColumn)
			projects.PUT("/:projectSlug/columns/:columnId", handlers.UpdateColumn)
			projects.DELETE("/:projectSlug/columns/:columnId", handlers.DeleteColumn)
			projects.PATCH("/:projectSlug/columns/reorder", handlers.ReorderColumns)

			// Cards
			projects.GET("/:projectSlug/columns/:columnId/cards", handlers.ListCards)
			projects.POST("/:projectSlug/columns/:columnId/cards", handlers.CreateCard)
			projects.PATCH("/:projectSlug/columns/:columnId/cards/reorder", handlers.ReorderCards)
			projects.GET("/:projectSlug/cards/:cardId", handlers.GetCard)
			projects.PUT("/:projectSlug/cards/:cardId", handlers.UpdateCard)
			projects.DELETE("/:projectSlug/cards/:cardId", handlers.DeleteCard)
			projects.PATCH("/:projectSlug/cards/:cardId/move", handlers.MoveCard)
			projects.POST("/:projectSlug/cards/:cardId/copy", handlers.CopyCard)
			projects.POST("/:projectSlug/cards/:cardId/transfer", handlers.TransferCard)
			projects.GET("/:projectSlug/cards/deleted", handlers.ListDeletedCards)
			projects.DELETE("/:projectSlug/cards/:cardId/permanent", handlers.PermanentDeleteCard)
			projects.POST("/:projectSlug/cards/:cardId/restore", handlers.RestoreCard)
			projects.POST("/:projectSlug/cards/:cardId/labels/:labelId", handlers.AssignLabel)
			projects.DELETE("/:projectSlug/cards/:cardId/labels/:labelId", handlers.RemoveLabel)
			projects.PUT("/:projectSlug/cards/:cardId/assignee", handlers.UpdateAssignee)
			projects.POST("/:projectSlug/cards/:cardId/watchers/:userId", handlers.AddWatcher)
			projects.DELETE("/:projectSlug/cards/:cardId/watchers/:userId", handlers.RemoveWatcher)
			projects.POST("/:projectSlug/cards/:cardId/tags", handlers.AddCardTag)
			projects.DELETE("/:projectSlug/cards/:cardId/tags/:tagId", handlers.RemoveCardTag)

			// Sub-cards
			projects.GET("/:projectSlug/cards/:cardId/subcards", handlers.ListSubCards)
			projects.POST("/:projectSlug/cards/:cardId/subcards", handlers.CreateSubCard)

			// Card history
			projects.GET("/:projectSlug/cards/:cardId/history", handlers.GetCardHistory)

			// Card git links
			projects.GET("/:projectSlug/cards/:cardId/links", handlers.ListCardLinks)

			// Card cross-references
			projects.GET("/:projectSlug/cards/:cardId/refs", handlers.ListCardRefs)
			projects.POST("/:projectSlug/cards/:cardId/refs", handlers.CreateCardRef)
			projects.DELETE("/:projectSlug/cards/:cardId/refs/:refId", handlers.DeleteCardRef)

			// Ticket links (card side)
			projects.GET("/:projectSlug/cards/:cardId/tickets", handlers.ListCardTicketLinks)

			// Card comments
			projects.GET("/:projectSlug/cards/:cardId/comments", handlers.ListComments)
			projects.POST("/:projectSlug/cards/:cardId/comments", handlers.CreateComment)
			projects.PUT("/:projectSlug/cards/:cardId/comments/:commentId", handlers.UpdateComment)
			projects.DELETE("/:projectSlug/cards/:cardId/comments/:commentId", handlers.DeleteComment)

			// Card checklist
			projects.GET("/:projectSlug/cards/:cardId/checklist", handlers.ListChecklistItems)
			projects.POST("/:projectSlug/cards/:cardId/checklist", handlers.CreateChecklistItem)
			projects.PATCH("/:projectSlug/cards/:cardId/checklist/reorder", handlers.ReorderChecklistItems)
			projects.PUT("/:projectSlug/cards/:cardId/checklist/:itemId", handlers.UpdateChecklistItem)
			projects.DELETE("/:projectSlug/cards/:cardId/checklist/:itemId", handlers.DeleteChecklistItem)

			// Card assignees (multiple)
			projects.POST("/:projectSlug/cards/:cardId/assignees/:userId", handlers.AddCardAssignee)
			projects.DELETE("/:projectSlug/cards/:cardId/assignees/:userId", handlers.RemoveCardAssignee)

			// Topics (threaded project discussions)
			projects.GET("/:projectSlug/topics", handlers.ListTopics)
			projects.POST("/:projectSlug/topics", handlers.CreateTopic)
			projects.GET("/:projectSlug/topics/:topicId", handlers.GetTopic)
			projects.PUT("/:projectSlug/topics/:topicId", handlers.UpdateTopic)
			projects.DELETE("/:projectSlug/topics/:topicId", handlers.DeleteTopic)
			projects.POST("/:projectSlug/topics/:topicId/replies", handlers.CreateTopicReply)
			projects.PUT("/:projectSlug/topics/:topicId/replies/:replyId", handlers.UpdateTopicReply)
			projects.DELETE("/:projectSlug/topics/:topicId/replies/:replyId", handlers.DeleteTopicReply)

			// Chat history
			projects.GET("/:projectSlug/chat/messages", handlers.ListChatMessages)
			projects.DELETE("/:projectSlug/chat/messages/:msgId", handlers.DeleteChatMessage)
			projects.POST("/:projectSlug/chat/messages/:msgId/reactions", handlers.ToggleChatReaction)

			// Webhooks
			projects.GET("/:projectSlug/webhooks", handlers.ListWebhooks)
			projects.POST("/:projectSlug/webhooks", handlers.CreateWebhook)
			projects.DELETE("/:projectSlug/webhooks/:webhookId", handlers.DeleteWebhook)
			projects.POST("/:projectSlug/webhooks/:webhookId/regenerate", handlers.RegenerateWebhookToken)

			// Project-scoped API keys
			projects.GET("/:projectSlug/api-keys", handlers.ListProjectAPIKeys)
			projects.POST("/:projectSlug/api-keys", handlers.CreateProjectAPIKey)
			projects.DELETE("/:projectSlug/api-keys/:keyId", handlers.DeleteProjectAPIKey)

			// Star / unstar project
			projects.POST("/:projectSlug/star", handlers.StarProject)
			projects.DELETE("/:projectSlug/star", handlers.UnstarProject)

			// Sprints (Scrum)
			projects.GET("/:projectSlug/sprints", handlers.ListSprints)
			projects.POST("/:projectSlug/sprints", handlers.CreateSprint)
			projects.PATCH("/:projectSlug/sprints/reorder", handlers.ReorderSprints)
			projects.PUT("/:projectSlug/sprints/:sprintId", handlers.UpdateSprint)
			projects.DELETE("/:projectSlug/sprints/:sprintId", handlers.DeleteSprint)
			projects.POST("/:projectSlug/sprints/:sprintId/start", handlers.StartSprint)
			projects.POST("/:projectSlug/sprints/:sprintId/complete", handlers.CompleteSprint)
			projects.POST("/:projectSlug/sprints/:sprintId/cards/:cardId", handlers.AddCardToSprint)
			projects.DELETE("/:projectSlug/sprints/:sprintId/cards/:cardId", handlers.RemoveCardFromSprint)
			projects.PATCH("/:projectSlug/sprints/:sprintId/cards/reorder", handlers.ReorderSprintCards)

			// Epics
			projects.GET("/:projectSlug/epics", handlers.ListEpics)
			projects.POST("/:projectSlug/epics", handlers.CreateEpic)
			projects.PATCH("/:projectSlug/epics/reorder", handlers.ReorderEpics)
			projects.GET("/:projectSlug/epics/:epicId/cards", handlers.ListEpicCards)
			projects.PUT("/:projectSlug/epics/:epicId", handlers.UpdateEpic)
			projects.DELETE("/:projectSlug/epics/:epicId", handlers.DeleteEpic)

			// Backlog (Scrum)
			projects.GET("/:projectSlug/backlog", handlers.ListBacklog)
			projects.PATCH("/:projectSlug/backlog/reorder", handlers.ReorderBacklog)

			// Releases (Scrum)
			projects.GET("/:projectSlug/releases", handlers.ListReleases)
			projects.POST("/:projectSlug/releases", handlers.CreateRelease)
			projects.PUT("/:projectSlug/releases/:releaseId", handlers.UpdateRelease)
			projects.DELETE("/:projectSlug/releases/:releaseId", handlers.DeleteRelease)
			projects.POST("/:projectSlug/releases/:releaseId/sprints/:sprintId", handlers.AddSprintToRelease)
			projects.DELETE("/:projectSlug/releases/:releaseId/sprints/:sprintId", handlers.RemoveSprintFromRelease)

			// Charts
			projects.GET("/:projectSlug/charts/velocity", handlers.GetVelocityChart)
			projects.GET("/:projectSlug/charts/burndown/:sprintId", handlers.GetBurndownChart)
			projects.GET("/:projectSlug/charts/burnup/:sprintId", handlers.GetBurnupChart)
			projects.GET("/:projectSlug/charts/cfd", handlers.GetCFDChart)
			projects.GET("/:projectSlug/charts/cycle-time", handlers.GetCycleTimeChart)
			projects.GET("/:projectSlug/charts/throughput", handlers.GetThroughputChart)
			projects.GET("/:projectSlug/charts/release-burndown/:releaseId", handlers.GetReleaseBurndownChart)
			projects.GET("/:projectSlug/charts/sprint-report/:sprintId", handlers.GetSprintReport)
			projects.GET("/:projectSlug/epics/:epicId/burndown", handlers.GetEpicBurndown)
		}
		} // end if !ttMode (conversations + projects)
	}

	if !ttMode {
		// Public incoming webhook receivers
		v1.POST("/webhooks/:token", handlers.IncomingWebhook)
		v1.POST("/gitea-webhook/:token", handlers.IncomingGiteaWebhook)
		v1.POST("/github-webhook/:token", handlers.IncomingGitHubWebhook)
		v1.POST("/gitlab-webhook/:token", handlers.IncomingGitLabWebhook)
	}

	// Attachment download — self-auth (cookie, Bearer, or media ticket); outside protected group
	v1.GET("/attachments/:id", handlers.DownloadAttachment)

	// WebSocket — self-auth (cookie, Bearer, or WS ticket); outside protected group
	v1.GET("/ws/user", wsHandler.HandleUserWS)
	if !ttMode {
		v1.GET("/ws/:projectSlug", wsHandler.HandleWS)

		// Ticket API — authenticated via X-API-Key header or ?api_key= query param
		ticket := v1.Group("/ticket")
		ticket.Use(middleware.APIKeyAuth())
		{
			ticket.POST("/:projectSlug/cards", handlers.TicketAdd)
			ticket.POST("/:projectSlug/cards/:cardId/comments", handlers.TicketComment)
			ticket.PATCH("/:projectSlug/cards/:cardId/move", handlers.TicketMove)
		}
	}

	// Serve uploaded files
	if uploadDir != "" {
		r.Static("/uploads", uploadDir)
	}

	// Serve frontend SPA from embedded or filesystem webFS when available
	if webFS != nil {
		fileServer := http.FileServer(http.FS(webFS))
		r.GET("/assets/*filepath", gin.WrapH(fileServer))
		r.GET("/fonts/*filepath", gin.WrapH(fileServer))
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			// Serve matching static file from web root; fall back to index.html for SPA routes
			path := strings.TrimPrefix(c.Request.URL.Path, "/")
			// In timetracking mode, serve the time-tracking logo variants instead of the
			// default WarmDesk logos so every client (browser, Tauri) gets the right branding.
			if ttMode {
				switch path {
				case "logo.svg":
					path = "timetracking.svg"
				case "logo-full.svg":
					path = "timetracking-full.svg"
				}
			}
			if path != "" {
				if f, err := webFS.Open(path); err == nil {
					if st, err := f.Stat(); err == nil && !st.IsDir() {
						f.Close()
						if strings.HasSuffix(path, ".html") {
							c.Header("Cache-Control", "no-store")
						}
						c.Request.URL.Path = "/" + path
						fileServer.ServeHTTP(c.Writer, c.Request)
						return
					}
					f.Close()
				}
			}
			f, err := webFS.Open("index.html")
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "frontend not available"})
				return
			}
			defer f.Close()
			data, _ := io.ReadAll(f)
			c.Header("Cache-Control", "no-store")
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	}

	return r
}
