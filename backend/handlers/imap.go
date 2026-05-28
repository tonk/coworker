package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/services"
)

// AdminTestIMAP POST /api/v1/admin/imap/test
// Accepts a JSON body with the current form values so the user can test
// without saving first. Any omitted field falls back to the saved DB value.
func AdminTestIMAP(c *gin.Context) {
	cfg := GetIMAPSettings()
	var body struct {
		Host          *string `json:"host"`
		Port          *int    `json:"port"`
		Username      *string `json:"username"`
		Password      *string `json:"password"`
		UseTLS        *bool   `json:"use_tls"`
		Mailbox       *string `json:"mailbox"`
		AuthMechanism *string `json:"auth_mechanism"`
		AccessToken   *string `json:"access_token"`
	}
	if err := c.ShouldBindJSON(&body); err == nil {
		if body.Host != nil {
			cfg.Host = *body.Host
		}
		if body.Port != nil {
			cfg.Port = *body.Port
		}
		if body.Username != nil {
			cfg.Username = *body.Username
		}
		if body.Password != nil {
			cfg.Password = *body.Password
		}
		if body.UseTLS != nil {
			cfg.UseTLS = *body.UseTLS
		}
		if body.Mailbox != nil {
			cfg.Mailbox = *body.Mailbox
		}
		if body.AuthMechanism != nil {
			cfg.AuthMechanism = *body.AuthMechanism
		}
		if body.AccessToken != nil {
			cfg.AccessToken = *body.AccessToken
		}
	}
	if err := services.TestIMAPConnection(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}

// AdminPollIMAP POST /api/v1/admin/imap/poll
func AdminPollIMAP(c *gin.Context) {
	svc := services.GetDefaultService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "IMAP service not running"})
		return
	}
	n, err := svc.PollOnce()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("imap: admin poll complete — %d message(s) processed", n)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Poll complete — %d message(s) processed", n)})
}
