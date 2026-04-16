package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tonk/warmdesk/handlers"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/services"
)

func Setup(authSvc *services.AuthService, allowedOrigins string, webDir string, apiLog bool, uploadDir string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	if apiLog {
		r.Use(gin.Logger())
	}
	r.Use(middleware.CORS(allowedOrigins))

	authHandler := handlers.NewAuthHandler(authSvc)
	wsHandler := handlers.NewWSHandler(authSvc)

	v1 := r.Group("/api/v1")

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public endpoints
	v1.GET("/version", handlers.GetVersion)
	v1.GET("/system/settings", handlers.GetSystemSettings)

	// Auth routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/mfa/verify", authHandler.MFAVerify)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
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
			admin.POST("/users/:id/mfa/disable", handlers.AdminDisableUserMFA)
			admin.GET("/users/:id/projects", handlers.AdminGetUserProjects)
			admin.PUT("/users/:id/projects", handlers.AdminSetUserProjects)
			admin.GET("/users/:id/customers", handlers.AdminGetUserCustomers)
			admin.PUT("/users/:id/customers", handlers.AdminSetUserCustomers)
			admin.GET("/projects", handlers.AdminListProjects)
			admin.POST("/projects", handlers.AdminCreateProject)
			admin.PUT("/projects/:id", handlers.AdminUpdateProject)
			admin.DELETE("/projects/:id", handlers.AdminDeleteProject)
			admin.POST("/projects/:id/restore", handlers.AdminRestoreProject)
			admin.GET("/system", handlers.AdminGetSystemSettings)
			admin.PUT("/system", handlers.AdminUpdateSystemSettings)
			admin.POST("/system/test-email", handlers.AdminSendTestEmail)
			admin.POST("/system/backup", handlers.AdminBackupDatabase)
		}

		// Users (for direct messages / user lookup)
		protected.GET("/users", handlers.ListAllUsers)

		// Online presence (global)
		protected.GET("/online-users", handlers.GetOnlineUsers)

		// Favorite users
		protected.GET("/favorite-users", handlers.ListFavoriteUsers)
		protected.POST("/favorite-users/:userId", handlers.AddFavoriteUser)
		protected.DELETE("/favorite-users/:userId", handlers.RemoveFavoriteUser)

		// Starred projects
		protected.GET("/starred-projects", handlers.ListStarredProjects)

		// Customers and contracts
		customers := protected.Group("/customers")
		{
			customers.GET("", handlers.ListCustomers)
			customers.POST("", handlers.CreateCustomer)
			customers.GET("/:customerId", handlers.GetCustomer)
			customers.PUT("/:customerId", handlers.UpdateCustomer)
			customers.DELETE("/:customerId", handlers.DeleteCustomer)
			customers.POST("/:customerId/favorite", handlers.AddCustomerFavorite)
			customers.DELETE("/:customerId/favorite", handlers.RemoveCustomerFavorite)
			customers.GET("/:customerId/contracts", handlers.ListContracts)
			customers.POST("/:customerId/contracts", handlers.CreateContract)
			customers.PUT("/:customerId/contracts/:contractId", handlers.UpdateContract)
			customers.DELETE("/:customerId/contracts/:contractId", handlers.DeleteContract)
			customers.GET("/:customerId/members", handlers.ListCustomerMembers)
			customers.PUT("/:customerId/members", handlers.SetCustomerMembers)
		}

		// Direct messages (legacy 1-on-1)
		dm := protected.Group("/direct-messages")
		{
			dm.GET("/conversations", handlers.ListConversations)
			dm.GET("/:userId", handlers.ListDirectMessages)
			dm.POST("/:userId", handlers.SendDirectMessage)
			dm.DELETE("/:userId/:msgId", handlers.DeleteDirectMessage)
		}

		// Image upload (avatars, logos)
		protected.POST("/upload/image", handlers.UploadImage)

		// File attachments
		protected.POST("/attachments", handlers.UploadAttachment)
		protected.GET("/attachments/:id", handlers.DownloadAttachment)
		protected.DELETE("/attachments/:id", handlers.DeleteAttachment)

		// Global search
		protected.GET("/search", handlers.GlobalSearch)

		// Reports
		protected.GET("/reports/time", handlers.GetTimeReport)
		protected.GET("/reports/time/pdf", handlers.GetTimeReportPDF)

		// Prometheus metrics (admin or metrics role)
		protected.GET("/metrics", middleware.MetricsAuth(), handlers.GetMetrics)

		// Conversations (1-on-1 and group)
		convs := protected.Group("/conversations")
		{
			convs.GET("", handlers.GetConversations)
			convs.POST("", handlers.CreateConversation)
			convs.GET("/:id/messages", handlers.GetConversationMessages)
			convs.POST("/:id/messages", handlers.SendConversationMessage)
			convs.PATCH("/:id/messages/:msgId", handlers.EditConversationMessage)
			convs.DELETE("/:id/messages/:msgId", handlers.DeleteConversationMessage)
			convs.POST("/:id/members", handlers.AddConversationMember)
			convs.DELETE("/:id/members/:userId", handlers.RemoveConversationMember)
			convs.POST("/:id/avatar", handlers.UploadConversationAvatar)
			convs.POST("/:id/messages/:msgId/reactions", handlers.ToggleConvReaction)
		}

		// Projects
		projects := protected.Group("/projects")
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

			// Card comments
			projects.GET("/:projectSlug/cards/:cardId/comments", handlers.ListComments)
			projects.POST("/:projectSlug/cards/:cardId/comments", handlers.CreateComment)
			projects.PUT("/:projectSlug/cards/:cardId/comments/:commentId", handlers.UpdateComment)
			projects.DELETE("/:projectSlug/cards/:cardId/comments/:commentId", handlers.DeleteComment)

			// Card checklist
			projects.GET("/:projectSlug/cards/:cardId/checklist", handlers.ListChecklistItems)
			projects.POST("/:projectSlug/cards/:cardId/checklist", handlers.CreateChecklistItem)
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
		}
	}

	// Public incoming webhook receivers
	v1.POST("/webhooks/:token", handlers.IncomingWebhook)
	v1.POST("/gitea-webhook/:token", handlers.IncomingGiteaWebhook)
	v1.POST("/github-webhook/:token", handlers.IncomingGitHubWebhook)
	v1.POST("/gitlab-webhook/:token", handlers.IncomingGitLabWebhook)

	// WebSocket (auth via ?token= query param)
	v1.GET("/ws/user", wsHandler.HandleUserWS)
	v1.GET("/ws/:projectSlug", wsHandler.HandleWS)

	// Ticket API — authenticated via X-API-Key header or ?api_key= query param
	ticket := v1.Group("/ticket")
	ticket.Use(middleware.APIKeyAuth())
	{
		ticket.POST("/:projectSlug/cards", handlers.TicketAdd)
		ticket.POST("/:projectSlug/cards/:cardId/comments", handlers.TicketComment)
		ticket.PATCH("/:projectSlug/cards/:cardId/move", handlers.TicketMove)
	}

	// Serve uploaded files
	if uploadDir != "" {
		r.Static("/uploads", uploadDir)
	}

	// Serve frontend SPA from webDir when configured
	if webDir != "" {
		r.Static("/assets", webDir+"/assets")
		r.StaticFile("/favicon.ico", webDir+"/favicon.ico")
		r.StaticFile("/favicon.svg", webDir+"/favicon.svg")
		r.StaticFile("/logo.svg", webDir+"/logo.svg")
		r.StaticFile("/logo-full.svg", webDir+"/logo-full.svg")
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File(webDir + "/index.html")
		})
	}

	return r
}
