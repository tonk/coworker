package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	appws "github.com/tonk/warmdesk/ws"
)

type WSHandler struct {
	authSvc  *services.AuthService
	upgrader websocket.Upgrader
}

func NewWSHandler(authSvc *services.AuthService, allowedOrigins string) *WSHandler {
	allowed := middleware.ParseOrigins(allowedOrigins)
	_, allowAll := allowed["*"]
	return &WSHandler{
		authSvc: authSvc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				if allowAll {
					return true
				}
				origin := r.Header.Get("Origin")
				if origin == "" {
					// No Origin header — direct (non-browser) client; allow.
					return true
				}
				_, ok := allowed[origin]
				return ok
			},
		},
	}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	slug := c.Param("projectSlug")

	// 1. httpOnly cookie (browser)
	tokenStr, _ := c.Cookie("access_token")
	// 2. ?token= query param (Tauri / direct clients)
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}
	// 3. Authorization header
	if tokenStr == "" {
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			tokenStr = auth[7:]
		}
	}

	claims, err := h.authSvc.ValidateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if err := services.RequireProjectRole(project.ID, claims.UserID, claims.GlobalRole, "viewer"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	var user models.User
	database.DB.First(&user, claims.UserID)

	hub := appws.GetOrCreateHub(project.ID)
	client := appws.NewClient(hub, conn, user.ID, user.Username, user.DisplayName, user.AvatarURL, project.ID, claims.GlobalRole, handleIncoming)

	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// HandleUserWS establishes a personal WebSocket connection for receiving user-scoped
// notifications (e.g. @mention alerts) even when not viewing a specific project.
func (h *WSHandler) HandleUserWS(c *gin.Context) {
	// 1. httpOnly cookie (browser)
	tokenStr, _ := c.Cookie("access_token")
	// 2. ?token= query param (Tauri / direct clients)
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}
	// 3. Authorization header
	if tokenStr == "" {
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			tokenStr = auth[7:]
		}
	}

	claims, err := h.authSvc.ValidateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws user upgrade error: %v", err)
		return
	}

	var user models.User
	database.DB.First(&user, claims.UserID)

	hub := appws.GetOrCreateUserHub(user.ID)
	client := appws.NewClient(hub, conn, user.ID, user.Username, user.DisplayName, user.AvatarURL, 0, claims.GlobalRole, nil)
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func handleIncoming(client *appws.Client, raw []byte) {
	var msg appws.Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		client.SendError("parse_error", "invalid JSON", "")
		return
	}

	switch msg.Type {
	case appws.TypePing:
		pong := appws.Message{
			Type:    appws.TypePong,
			Payload: map[string]string{"server_time": time.Now().UTC().Format(time.RFC3339)},
		}
		data, _ := json.Marshal(pong)
		client.Send(data)
		return
	}

	// Viewers are read-only — block all write operations
	if client.GlobalRole() == "viewer" {
		client.SendError("forbidden", "viewers are read-only", msg.ID)
		return
	}

	switch msg.Type {
	case appws.TypeChatSend:
		payloadBytes, _ := json.Marshal(msg.Payload)
		var payload appws.ChatSendPayload
		if err := json.Unmarshal(payloadBytes, &payload); err != nil || payload.Body == "" {
			client.SendError("invalid_payload", "body required", msg.ID)
			return
		}

		chatMsg := models.ChatMessage{
			ProjectID: client.ProjectID(),
			UserID:    client.UserID(),
			Body:      payload.Body,
		}
		database.DB.Create(&chatMsg)
		database.DB.Preload("User").First(&chatMsg, chatMsg.ID)

		appws.BroadcastToProject(client.ProjectID(), appws.Message{
			Type:    appws.TypeChatMessageCreated,
			Payload: chatMsg,
		})

		if notifSvc != nil {
			go notifSvc.NotifyMentions(payload.Body, client.UserID(), "project chat")
		}

	case appws.TypeChatEdit:
		payloadBytes, _ := json.Marshal(msg.Payload)
		var payload appws.ChatEditPayload
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			client.SendError("invalid_payload", "invalid payload", msg.ID)
			return
		}

		var chatMsg models.ChatMessage
		if err := database.DB.Where("id = ? AND project_id = ?", payload.MessageID, client.ProjectID()).First(&chatMsg).Error; err != nil {
			client.SendError("not_found", "message not found", msg.ID)
			return
		}
		if chatMsg.UserID != client.UserID() {
			client.SendError("forbidden", "not your message", msg.ID)
			return
		}

		database.DB.Model(&chatMsg).Updates(map[string]interface{}{"body": payload.Body, "is_edited": true})
		appws.BroadcastToProject(client.ProjectID(), appws.Message{
			Type:    appws.TypeChatMessageUpdated,
			Payload: map[string]interface{}{"id": chatMsg.ID, "body": payload.Body, "is_edited": true},
		})

	case appws.TypeChatTyping:
		// Broadcast typing notification to all other clients in the project.
		appws.BroadcastToProject(client.ProjectID(), appws.Message{
			Type: appws.TypeChatUserTyping,
			Payload: map[string]interface{}{
				"user_id":      client.UserID(),
				"username":     client.Username(),
				"display_name": client.DisplayName(),
			},
		})

	case appws.TypeChatDelete:
		payloadBytes, _ := json.Marshal(msg.Payload)
		var payload appws.ChatDeletePayload
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			client.SendError("invalid_payload", "invalid payload", msg.ID)
			return
		}

		var chatMsg models.ChatMessage
		if err := database.DB.Where("id = ? AND project_id = ?", payload.MessageID, client.ProjectID()).First(&chatMsg).Error; err != nil {
			client.SendError("not_found", "message not found", msg.ID)
			return
		}
		if chatMsg.UserID != client.UserID() {
			// Check if owner
			role := services.GetMemberRole(client.ProjectID(), client.UserID())
			if role != "owner" {
				client.SendError("forbidden", "not your message", msg.ID)
				return
			}
		}

		database.DB.Model(&chatMsg).Update("is_deleted", true)
		appws.BroadcastToProject(client.ProjectID(), appws.Message{
			Type:    appws.TypeChatMessageDeleted,
			Payload: map[string]uint{"id": payload.MessageID},
		})
	}
}

// Ensure middleware.GetUserID is accessible (used in auth.go)
var _ = middleware.GetUserID
