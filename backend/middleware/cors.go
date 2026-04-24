package middleware

import (
	"log"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// tauriOrigins are always allowed so the desktop client works regardless of
// the server's configured allowed_origins.
// tauri://localhost  — Linux / macOS
// https://tauri.localhost — Windows
var tauriOrigins = []string{"tauri://localhost", "https://tauri.localhost", "http://tauri.localhost"}

// ParseOrigins returns the full set of allowed origins from the comma-separated
// config string, always including the Tauri desktop origins.
func ParseOrigins(allowedOrigins string) map[string]struct{} {
	combined := make(map[string]struct{})
	for _, o := range tauriOrigins {
		combined[o] = struct{}{}
	}
	for _, o := range strings.Split(allowedOrigins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			combined[o] = struct{}{}
		}
	}
	return combined
}

func CORS(allowedOrigins string) gin.HandlerFunc {
	allowed := ParseOrigins(allowedOrigins)

	// Check if wildcard is configured — allow any origin
	_, allowAll := allowed["*"]
	if allowAll {
		log.Printf("WARNING: CORS allowed_origins contains '*' — cross-origin protection is disabled")
	}

	cfg := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if allowAll {
				return true
			}
			_, ok := allowed[origin]
			return ok
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept-Language"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	return cors.New(cfg)
}
