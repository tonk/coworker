package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/middleware"
)

var livekitCfg *config.Config

func InitLiveKit(cfg *config.Config) {
	livekitCfg = cfg
}

// livekitVideoClaims mirrors the LiveKit "video" JWT grant.
type livekitVideoClaims struct {
	Room           string `json:"room"`
	RoomJoin       bool   `json:"roomJoin"`
	CanPublish     bool   `json:"canPublish"`
	CanSubscribe   bool   `json:"canSubscribe"`
	CanPublishData bool   `json:"canPublishData"`
}

type livekitClaims struct {
	Video *livekitVideoClaims `json:"video,omitempty"`
	jwt.RegisteredClaims
}

// GetLiveKitToken returns a short-lived LiveKit access token for the caller.
// The room name is derived from the conversation ID so all members join the
// same room automatically.
//
// GET /api/v1/conversations/:id/livekit-token
func GetLiveKitToken(c *gin.Context) {
	if livekitCfg.LiveKitAPIKey == "" || livekitCfg.LiveKitAPISecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "livekit not configured"})
		return
	}

	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || convID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	userID := middleware.GetUserID(c)
	if !isMember(uint(convID), userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a conversation member"})
		return
	}

	roomName := fmt.Sprintf("conv-%d", convID)
	if livekitCfg.LiveKitRoomPrefix != "" {
		roomName = livekitCfg.LiveKitRoomPrefix + "-" + roomName
	}

	now := time.Now()
	claims := livekitClaims{
		Video: &livekitVideoClaims{
			Room:           roomName,
			RoomJoin:       true,
			CanPublish:     true,
			CanSubscribe:   true,
			CanPublishData: true,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    livekitCfg.LiveKitAPIKey,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(6 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(livekitCfg.LiveKitAPISecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    signed,
		"url":      livekitCfg.LiveKitURL,
		"room":     roomName,
		"identity": fmt.Sprintf("%d", userID),
	})
}
