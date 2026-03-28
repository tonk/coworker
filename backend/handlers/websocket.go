package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
	appws "github.com/tonk/coworker/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS is handled by middleware
	},
}

type WSHandler struct {
	authSvc *services.AuthService
}

func NewWSHandler(authSvc *services.AuthService) *WSHandler {
	return &WSHandler{authSvc: authSvc}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	slug := c.Param("projectSlug")

	// Auth via query param token
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Also check Authorization header as fallback
		tokenStr = c.GetHeader("Authorization")
		if len(tokenStr) > 7 {
			tokenStr = tokenStr[7:]
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

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
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
