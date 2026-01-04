// internal/middleware/cors.go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func CORSMiddleware() gin.HandlerFunc {
	// Konfigurasi rs/cors yang support WebSocket
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:5500",
			"http://127.0.0.1:5500",
			"http://localhost:5501",
			"http://127.0.0.1:5501",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"Upgrade",
			"Connection",
			"Sec-WebSocket-Key",
			"Sec-WebSocket-Version",
			"Sec-WebSocket-Extensions",
			"Sec-WebSocket-Protocol",
		},
		ExposedHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           43200, // 12 hours in seconds
	})

	return func(ctx *gin.Context) {
		c.HandlerFunc(ctx.Writer, ctx.Request)
		ctx.Next()
	}
}
