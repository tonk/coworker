package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// tauriOrigins are always allowed so the desktop client works regardless of
// the server's configured allowed_origins.
// tauri://localhost  — Linux / macOS
// https://tauri.localhost — Windows
var tauriOrigins = []string{"tauri://localhost", "https://tauri.localhost"}

func CORS(allowedOrigins string) gin.HandlerFunc {
	origins := []string{}
	for _, o := range strings.Split(allowedOrigins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}

	// Build a combined set for fast lookup. gin-contrib/cors rejects non-http(s)
	// schemes in AllowOrigins, so we use AllowOriginFunc for everything.
	allowed := make(map[string]struct{}, len(origins)+len(tauriOrigins))
	for _, o := range origins {
		allowed[o] = struct{}{}
	}
	for _, o := range tauriOrigins {
		allowed[o] = struct{}{}
	}

	cfg := cors.Config{
		AllowOriginFunc: func(origin string) bool {
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
