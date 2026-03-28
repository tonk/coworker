package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/config"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/handlers"
	"github.com/tonk/coworker/router"
	"github.com/tonk/coworker/services"
	"github.com/tonk/coworker/ws"
)

func main() {
	configFile := flag.String("config", "", "path to config file (overrides CONFIG_FILE env var)")
	flag.Parse()

	cfg := config.Load(*configFile)
	handlers.InitSystemDefaults(cfg)
	handlers.InitAttachments(cfg)

	emailSvc := services.NewEmailService(cfg.SMTP)
	// Allow the email service to read live SMTP settings from the DB after startup
	services.SetSMTPConfigReader(handlers.GetSMTPSettings)
	notifSvc := services.NewNotificationService(emailSvc)
	handlers.InitNotifications(notifSvc)

	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := database.Init(cfg); err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	// Initialise pub/sub backend (Redis for horizontal scaling, memory for single instance)
	if cfg.RedisURL != "" {
		rps, err := ws.NewRedisPubSub(cfg.RedisURL)
		if err != nil {
			log.Fatalf("Redis pub/sub init failed: %v", err)
		}
		ws.InitPubSub(rps)
	}
	ws.StartPubSubListener()

	authSvc := services.NewAuthService(cfg.JWTSecret)

	r := router.Setup(authSvc, cfg.AllowedOrigins, cfg.WebDir, cfg.APILog, cfg.UploadDir)

	log.Printf("Starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
