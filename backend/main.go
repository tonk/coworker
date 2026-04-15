// Package main is the entry point for the WarmDesk server.
//
// @title           WarmDesk API
// @version         1.0
// @description     Self-hosted project management tool — Kanban boards, team chat, discussions, and time reporting.
//
// @contact.name    WarmDesk
// @contact.url     https://github.com/tonk/warmdesk
//
// @license.name    MIT
//
// @host            localhost:8080
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer " followed by your JWT access token.
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for the Ticket API (CI/CD automation).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/docs"
	"github.com/tonk/warmdesk/handlers"
	"github.com/tonk/warmdesk/router"
	"github.com/tonk/warmdesk/services"
	"github.com/tonk/warmdesk/ws"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func main() {
	configFile := flag.String("config", "", "path to config file (overrides CONFIG_FILE env var)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	log.Printf("Starting WarmDesk %s", version)

	cfg := config.Load(*configFile)

	if cfg.BaseURL != "" {
		// Strip scheme so Swagger host is just "host" or "host:port"
		host := cfg.BaseURL
		if i := strings.Index(host, "://"); i >= 0 {
			host = host[i+3:]
		}
		host = strings.TrimRight(host, "/")
		docs.SwaggerInfo.Host = host
	}

	handlers.SetVersion(version)
	handlers.InitSystemDefaults(cfg)
	handlers.InitAttachments(cfg)
	handlers.InitReport(cfg)

	emailSvc := services.NewEmailService(cfg.SMTP)
	// Allow the email service to read live SMTP settings from the DB after startup
	services.SetSMTPConfigReader(handlers.GetSMTPSettings)
	services.SetEmailService(emailSvc)
	notifSvc := services.NewNotificationService(emailSvc)
	handlers.InitNotifications(notifSvc)
	handlers.SetBaseURL(cfg.BaseURL)

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

	addr := ":" + cfg.Port
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		log.Printf("Starting server (HTTPS) on %s", addr)
		if err := r.RunTLS(addr, cfg.TLSCert, cfg.TLSKey); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	} else {
		log.Printf("Starting server on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}
}
