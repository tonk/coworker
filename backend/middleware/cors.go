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

	// Check if wildcard is configured — refuse to set CORS headers at runtime.
	_, hasWildcard := allowed["*"]
	if hasWildcard {
		log.Printf("WARNING: CORS allowed_origins contains '*' — wildcard CORS is blocked at runtime; no CORS headers will be set")
	}

	cfg := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if hasWildcard {
				// Wildcard is never honoured at request time regardless of config.
				return false
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
