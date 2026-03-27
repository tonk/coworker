package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/handlers"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/services"
)

func Setup(authSvc *services.AuthService, allowedOrigins string, webDir string) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(allowedOrigins))

	authHandler := handlers.NewAuthHandler(authSvc)
	wsHandler := handlers.NewWSHandler(authSvc)

	v1 := r.Group("/api/v1")

	// Public system settings (e.g. registration enabled)
	v1.GET("/system/settings", handlers.GetSystemSettings)

	// Auth routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// Authenticated routes
	protected := v1.Group("")
	protected.Use(middleware.Auth(authSvc))
	{
		// Current user
		protected.GET("/auth/me", authHandler.Me)
		protected.PUT("/auth/me", authHandler.UpdateMe)
		protected.PUT("/auth/me/password", authHandler.ChangePassword)

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
			admin.GET("/projects", handlers.AdminListProjects)
			admin.POST("/projects", handlers.AdminCreateProject)
			admin.PUT("/projects/:id", handlers.AdminUpdateProject)
			admin.DELETE("/projects/:id", handlers.AdminDeleteProject)
			admin.GET("/system", handlers.AdminGetSystemSettings)
			admin.PUT("/system", handlers.AdminUpdateSystemSettings)
		}

		// Users (for direct messages / user lookup)
		protected.GET("/users", handlers.ListAllUsers)

		// Online presence (global)
		protected.GET("/online-users", handlers.GetOnlineUsers)

		// Starred projects
		protected.GET("/starred-projects", handlers.ListStarredProjects)

		// Direct messages (legacy 1-on-1)
		dm := protected.Group("/direct-messages")
		{
			dm.GET("/conversations", handlers.ListConversations)
			dm.GET("/:userId", handlers.ListDirectMessages)
			dm.POST("/:userId", handlers.SendDirectMessage)
			dm.DELETE("/:userId/:msgId", handlers.DeleteDirectMessage)
		}

		// Conversations (1-on-1 and group)
		convs := protected.Group("/conversations")
		{
			convs.GET("", handlers.GetConversations)
			convs.POST("", handlers.CreateConversation)
			convs.GET("/:id/messages", handlers.GetConversationMessages)
			convs.POST("/:id/messages", handlers.SendConversationMessage)
			convs.DELETE("/:id/messages/:msgId", handlers.DeleteConversationMessage)
			convs.POST("/:id/members", handlers.AddConversationMember)
		}

		// Projects
		projects := protected.Group("/projects")
		{
			projects.GET("", handlers.ListProjects)
			projects.POST("", handlers.CreateProject)
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
			projects.POST("/:projectSlug/cards/:cardId/labels/:labelId", handlers.AssignLabel)
			projects.DELETE("/:projectSlug/cards/:cardId/labels/:labelId", handlers.RemoveLabel)
			projects.PUT("/:projectSlug/cards/:cardId/assignee", handlers.UpdateAssignee)

			// Card history
			projects.GET("/:projectSlug/cards/:cardId/history", handlers.GetCardHistory)

			// Card comments
			projects.GET("/:projectSlug/cards/:cardId/comments", handlers.ListComments)
			projects.POST("/:projectSlug/cards/:cardId/comments", handlers.CreateComment)
			projects.PUT("/:projectSlug/cards/:cardId/comments/:commentId", handlers.UpdateComment)
			projects.DELETE("/:projectSlug/cards/:cardId/comments/:commentId", handlers.DeleteComment)

			// Chat history
			projects.GET("/:projectSlug/chat/messages", handlers.ListChatMessages)
			projects.DELETE("/:projectSlug/chat/messages/:msgId", handlers.DeleteChatMessage)

			// Star / unstar project
			projects.POST("/:projectSlug/star", handlers.StarProject)
			projects.DELETE("/:projectSlug/star", handlers.UnstarProject)
		}
	}

	// WebSocket (auth via ?token= query param)
	v1.GET("/ws/:projectSlug", wsHandler.HandleWS)

	// Ticket API — authenticated via X-API-Key header or ?api_key= query param
	ticket := v1.Group("/ticket")
	ticket.Use(middleware.APIKeyAuth())
	{
		ticket.POST("/:projectSlug/cards", handlers.TicketAdd)
		ticket.POST("/:projectSlug/cards/:cardId/comments", handlers.TicketComment)
		ticket.PATCH("/:projectSlug/cards/:cardId/move", handlers.TicketMove)
	}

	// Serve frontend SPA from webDir when configured
	if webDir != "" {
		r.Static("/assets", webDir+"/assets")
		r.StaticFile("/favicon.ico", webDir+"/favicon.ico")
		r.StaticFile("/favicon.svg", webDir+"/favicon.svg")
		r.StaticFile("/logo.svg", webDir+"/logo.svg")
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
