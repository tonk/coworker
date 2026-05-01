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
	"errors"
	"io/fs"
	"log"
	"mime"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/docs"
	"github.com/tonk/warmdesk/handlers"
	"github.com/tonk/warmdesk/router"
	"github.com/tonk/warmdesk/services"
	"github.com/tonk/warmdesk/staticweb"
	"github.com/tonk/warmdesk/ws"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func init() {
	// Ensure module/script assets are served with browser-accepted MIME types.
	_ = mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".mjs", "text/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".css", "text/css; charset=utf-8")
	_ = mime.AddExtensionType(".json", "application/json; charset=utf-8")
	_ = mime.AddExtensionType(".map", "application/json; charset=utf-8")
}

type overlayFS struct {
	primary  fs.FS
	fallback fs.FS
}

func (o overlayFS) Open(name string) (fs.File, error) {
	if o.primary != nil {
		f, err := o.primary.Open(name)
		if err == nil {
			return f, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	if o.fallback != nil {
		return o.fallback.Open(name)
	}
	return nil, fs.ErrNotExist
}

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

	if cfg.JWTSecret == "change-me-in-production" {
		log.Fatal("refusing to start: jwt_secret is still the default value — set a strong random secret via JWT_SECRET or jwt_secret in your config file")
	}
	if cfg.GinMode == "release" && strings.Contains(cfg.AllowedOrigins, "*") {
		log.Fatal("refusing to start: allowed_origins contains '*' which disables CORS protection — remove the wildcard or restrict to specific origins")
	}

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
	handlers.InitBackup(cfg)
	handlers.InitLiveKit(cfg)

	emailSvc := services.NewEmailService(cfg.SMTP)
	// Allow the email service to read live SMTP settings from the DB after startup
	services.SetSMTPConfigReader(handlers.GetSMTPSettings)
	// Allow the email template system to read live branding/version/URL info
	services.SetAppInfoReader(handlers.GetEmailBranding)
	services.SetEmailService(emailSvc)
	notifSvc := services.NewNotificationService(emailSvc)
	handlers.InitNotifications(notifSvc)
	handlers.SetBaseURL(cfg.BaseURL)

	if cfg.GinMode != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := database.Init(cfg); err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	handlers.StartBackupScheduler()

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

	var trustedProxies []string
	for _, p := range strings.Split(cfg.TrustedProxies, ",") {
		if p = strings.TrimSpace(p); p != "" {
			trustedProxies = append(trustedProxies, p)
		}
	}
	var webFS fs.FS
	if cfg.WebDir != "" {
		if _, err := os.Stat(cfg.WebDir); err == nil {
			primary := os.DirFS(cfg.WebDir)
			if staticweb.FS != nil {
				// Serve from configured web_dir first, but fall back to embedded
				// assets for any missing chunk to avoid hard failures after updates.
				webFS = overlayFS{primary: primary, fallback: staticweb.FS}
			} else {
				webFS = primary
			}
		} else {
			log.Printf("web_dir %q not found, falling back to embedded frontend", cfg.WebDir)
			webFS = staticweb.FS
		}
	} else {
		webFS = staticweb.FS
	}
	handlers.InitReport(cfg, webFS)
	r := router.Setup(authSvc, cfg.AllowedOrigins, webFS, cfg.APILog, cfg.UploadDir, trustedProxies)

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
