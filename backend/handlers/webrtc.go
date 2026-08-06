package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
)

var webrtcCfg *config.Config

func InitWebRTC(cfg *config.Config) {
	webrtcCfg = cfg
}

// iceServer mirrors the shape of an RTCIceServer for the frontend.
type iceServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// GetICEServers returns the STUN/TURN servers 1:1 call clients should use to
// build their RTCPeerConnection. STUN-only (two public Google servers) is
// always included; a TURN entry is added on top when turn_urls is configured,
// since STUN alone cannot relay media across symmetric NATs/CGNAT/restrictive
// firewalls.
//
// GET /api/v1/ice-servers
func GetICEServers(c *gin.Context) {
	servers := []iceServer{
		{URLs: []string{"stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302"}},
	}

	if webrtcCfg != nil && webrtcCfg.TurnURLs != "" {
		urls := strings.Split(webrtcCfg.TurnURLs, ",")
		for i := range urls {
			urls[i] = strings.TrimSpace(urls[i])
		}
		servers = append(servers, iceServer{
			URLs:       urls,
			Username:   webrtcCfg.TurnUsername,
			Credential: webrtcCfg.TurnCredential,
		})
	}

	c.JSON(http.StatusOK, gin.H{"ice_servers": servers})
}
