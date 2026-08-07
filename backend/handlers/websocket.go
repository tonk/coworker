package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
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
				// Same-origin: browser connecting to the backend that also serves
				// the frontend (production mode). Always allow.
				if u, err := url.Parse(origin); err == nil && u.Host == r.Host {
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
	// 2. ?ticket= query param — short-lived WS ticket (Tauri mode; never the long-lived JWT)
	if tokenStr == "" {
		tokenStr = c.Query("ticket")
	}
	// 3. Authorization header
	if tokenStr == "" {
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			tokenStr = auth[7:]
		}
	}

	// Cookie and Bearer paths accept a normal access token; ?ticket= path accepts only a WS ticket.
	var claims *services.Claims
	var err error
	if c.Query("ticket") != "" {
		claims, err = h.authSvc.ValidateWSTicket(tokenStr)
	} else {
		claims, err = h.authSvc.ValidateToken(tokenStr)
	}
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

	// h.upgrader has CheckOrigin set in NewWSHandler — origin validation is performed there.
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil) // nosemgrep: go.gorilla.security.audit.websocket-missing-origin-check.websocket-missing-origin-check
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	var user models.User
	database.DB.First(&user, claims.UserID)

	hub := appws.GetOrCreateHub(project.ID)
	client := appws.NewClient(hub, conn, user.ID, user.Username, user.DisplayName, user.AvatarURL, project.ID, project.Name, project.Slug, claims.GlobalRole, handleIncoming)

	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// HandleUserWS establishes a personal WebSocket connection for receiving user-scoped
// notifications (e.g. @mention alerts) even when not viewing a specific project.
func (h *WSHandler) HandleUserWS(c *gin.Context) {
	// 1. httpOnly cookie (browser)
	tokenStr, _ := c.Cookie("access_token")
	// 2. ?ticket= query param — short-lived WS ticket (Tauri mode; never the long-lived JWT)
	if tokenStr == "" {
		tokenStr = c.Query("ticket")
	}
	// 3. Authorization header
	if tokenStr == "" {
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			tokenStr = auth[7:]
		}
	}

	var claims *services.Claims
	var err error
	if c.Query("ticket") != "" {
		claims, err = h.authSvc.ValidateWSTicket(tokenStr)
	} else {
		claims, err = h.authSvc.ValidateToken(tokenStr)
	}
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// h.upgrader has CheckOrigin set in NewWSHandler — origin validation is performed there.
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil) // nosemgrep: go.gorilla.security.audit.websocket-missing-origin-check.websocket-missing-origin-check
	if err != nil {
		log.Printf("ws user upgrade error: %v", err)
		return
	}

	var user models.User
	database.DB.First(&user, claims.UserID)

	hub := appws.GetOrCreateUserHub(user.ID)
	client := appws.NewClient(hub, conn, user.ID, user.Username, user.DisplayName, user.AvatarURL, 0, "", "", claims.GlobalRole, handleUserIncoming)
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

		// The frontend's title-blink indicator needs to name which project a chat
		// notification came from; ChatMessage.Project is json:"-" (never serialized
		// on its own), so it's added as sibling fields here rather than exposing the
		// full Project on every chat API response.
		appws.BroadcastToProject(client.ProjectID(), appws.Message{
			Type: appws.TypeChatMessageCreated,
			Payload: struct {
				models.ChatMessage
				ProjectName string `json:"project_name"`
				ProjectSlug string `json:"project_slug"`
			}{ChatMessage: chatMsg, ProjectName: client.ProjectName(), ProjectSlug: client.ProjectSlug()},
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

		database.DB.Model(&chatMsg).Updates(map[string]any{"body": payload.Body, "is_edited": true})
		appws.BroadcastToProject(client.ProjectID(), appws.Message{
			Type:    appws.TypeChatMessageUpdated,
			Payload: map[string]any{"id": chatMsg.ID, "body": payload.Body, "is_edited": true},
		})

	case appws.TypeChatTyping:
		// Broadcast typing notification to all other clients in the project.
		appws.BroadcastToProject(client.ProjectID(), appws.Message{
			Type: appws.TypeChatUserTyping,
			Payload: map[string]any{
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

	case appws.TypeCallOffer, appws.TypeCallAnswer, appws.TypeCallICE, appws.TypeCallHangup, appws.TypeCallReject, appws.TypeCallFailed, appws.TypeCallMute, appws.TypeCallScreenShare:
		handleCallSignal(client, msg)
	}
}

// handleUserIncoming processes messages received on the personal user WebSocket.
func handleUserIncoming(client *appws.Client, raw []byte) {
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

	case appws.TypeCallOffer, appws.TypeCallAnswer, appws.TypeCallICE, appws.TypeCallHangup, appws.TypeCallReject, appws.TypeCallFailed, appws.TypeCallMute, appws.TypeCallScreenShare:
		handleCallSignal(client, msg)

	case appws.TypeCallGroupInvite:
		handleGroupCallInvite(client, msg)
	}
}

// callBasePayload holds the fields common to all call signaling messages.
type callBasePayload struct {
	ToUserID       uint   `json:"to_user_id"`
	ConversationID uint   `json:"conversation_id"`
	SDP            string `json:"sdp,omitempty"`
	Candidate      string `json:"candidate,omitempty"`
	HasVideo       bool   `json:"has_video,omitempty"`
	Muted          bool   `json:"muted,omitempty"`
	Sharing        bool   `json:"sharing,omitempty"`
}

// handleCallSignal relays WebRTC signaling messages between two conversation members.
// It verifies membership, rebuilds the outgoing payload from scratch, and never
// forwards raw JSON to prevent payload injection.
func handleCallSignal(client *appws.Client, msg appws.Message) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload callBasePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil || payload.ToUserID == 0 || payload.ConversationID == 0 {
		client.SendError("invalid_payload", "to_user_id and conversation_id required", msg.ID)
		return
	}

	callerID := client.UserID()
	calleeID := payload.ToUserID
	convID := payload.ConversationID

	// Both users must be members of the conversation
	if !isMember(convID, callerID) || !isMember(convID, calleeID) {
		client.SendError("forbidden", "not a conversation member", msg.ID)
		return
	}

	switch msg.Type {
	case appws.TypeCallOffer:
		if !appws.IsUserOnline(calleeID) {
			// Callee is offline — notify caller
			unavail, _ := json.Marshal(appws.Message{
				Type: appws.TypeCallUnavailable,
				Payload: map[string]any{
					"from_user_id":    callerID,
					"conversation_id": convID,
				},
			})
			client.Send(unavail)
			return
		}
		// Look up caller's avatar URL
		var caller models.User
		database.DB.First(&caller, callerID)
		appws.BroadcastToUser(calleeID, appws.Message{
			Type: appws.TypeCallRing,
			Payload: map[string]any{
				"from_user_id":    callerID,
				"from_name":       client.DisplayName(),
				"from_avatar":     caller.AvatarURL,
				"conversation_id": convID,
				"sdp":             payload.SDP,
				"has_video":       payload.HasVideo,
			},
		})

	case appws.TypeCallAnswer:
		appws.BroadcastToUser(calleeID, appws.Message{
			Type: appws.TypeCallAnswer,
			Payload: map[string]any{
				"from_user_id":    callerID,
				"conversation_id": convID,
				"sdp":             payload.SDP,
			},
		})

	case appws.TypeCallICE:
		appws.BroadcastToUser(calleeID, appws.Message{
			Type: appws.TypeCallICE,
			Payload: map[string]any{
				"from_user_id":    callerID,
				"conversation_id": convID,
				"candidate":       payload.Candidate,
			},
		})

	case appws.TypeCallHangup:
		appws.BroadcastToUser(calleeID, appws.Message{
			Type: appws.TypeCallHangup,
			Payload: map[string]any{
				"from_user_id":    callerID,
				"conversation_id": convID,
			},
		})

	case appws.TypeCallReject:
		appws.BroadcastToUser(calleeID, appws.Message{
			Type: appws.TypeCallReject,
			Payload: map[string]any{
				"from_user_id":    callerID,
				"conversation_id": convID,
			},
		})

	case appws.TypeCallFailed:
		appws.BroadcastToUser(calleeID, appws.Message{
			Type: appws.TypeCallFailed,
			Payload: map[string]any{
				"from_user_id":    callerID,
				"conversation_id": convID,
			},
		})

	case appws.TypeCallMute:
		appws.BroadcastToUser(calleeID, appws.Message{
			Type: appws.TypeCallMute,
			Payload: map[string]any{
				"from_user_id":    callerID,
				"conversation_id": convID,
				"muted":           payload.Muted,
			},
		})

	case appws.TypeCallScreenShare:
		appws.BroadcastToUser(calleeID, appws.Message{
			Type: appws.TypeCallScreenShare,
			Payload: map[string]any{
				"from_user_id":    callerID,
				"conversation_id": convID,
				"sharing":         payload.Sharing,
			},
		})
	}
}

// handleGroupCallInvite adds invited users to the conversation (if not already
// members) and relays a call.group_invite notification to each of them so they
// can join the active LiveKit room.
func handleGroupCallInvite(client *appws.Client, msg appws.Message) {
	type groupInvitePayload struct {
		ToUserIDs      []uint `json:"to_user_ids"`
		ConversationID uint   `json:"conversation_id"`
	}
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload groupInvitePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil || len(payload.ToUserIDs) == 0 || payload.ConversationID == 0 {
		client.SendError("invalid_payload", "to_user_ids and conversation_id required", msg.ID)
		return
	}

	inviterID := client.UserID()
	convID := payload.ConversationID

	if !isMember(convID, inviterID) {
		client.SendError("forbidden", "not a conversation member", msg.ID)
		return
	}

	var conv models.Conversation
	database.DB.First(&conv, convID)

	var inviter models.User
	database.DB.First(&inviter, inviterID)

	now := time.Now()
	for _, uid := range payload.ToUserIDs {
		if uid == inviterID {
			continue
		}
		// Only invite users the inviter already has a conversation with,
		// preventing arbitrary users from being silently added to any conversation.
		var sharedCount int64
		database.DB.Raw(`
			SELECT COUNT(*) FROM conversation_members cm1
			JOIN conversation_members cm2 ON cm1.conversation_id = cm2.conversation_id
			WHERE cm1.user_id = ? AND cm2.user_id = ?
		`, inviterID, uid).Scan(&sharedCount)
		if sharedCount == 0 {
			continue
		}
		if !isMember(convID, uid) {
			database.DB.Create(&models.ConversationMember{
				ConversationID: convID,
				UserID:         uid,
				JoinedAt:       now,
			})
			var count int64
			database.DB.Model(&models.ConversationMember{}).
				Where("conversation_id = ?", convID).Count(&count)
			if count > 2 {
				database.DB.Model(&models.Conversation{}).
					Where("id = ?", convID).Update("is_group", true)
			}
		}
		appws.BroadcastToUser(uid, appws.Message{
			Type: appws.TypeCallGroupInvite,
			Payload: map[string]any{
				"from_user_id":    inviterID,
				"from_name":       client.DisplayName(),
				"from_avatar":     inviter.AvatarURL,
				"conversation_id": convID,
				"conv_name":       conv.Name,
			},
		})
	}
}

// Ensure middleware.GetUserID is accessible (used in auth.go)
var _ = middleware.GetUserID
