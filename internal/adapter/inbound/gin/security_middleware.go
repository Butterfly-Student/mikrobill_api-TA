package gin_inbound_adapter

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders - Adds security headers to all responses
func (h *middlewareAdapter) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// XSS Protection (for older browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (basic - can be customized)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Strict Transport Security (HSTS) - enforce HTTPS
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// OriginValidation - Validates request origin for customer portal endpoints
func (h *middlewareAdapter) OriginValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from environment
		allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
		if allowedOriginsEnv == "" {
			allowedOriginsEnv = "http://localhost:3000,http://localhost:5173,http://localhost:8080"
		}

		allowedOrigins := strings.Split(allowedOriginsEnv, ",")
		origin := c.Request.Header.Get("Origin")
		referer := c.Request.Header.Get("Referer")

		// For non-GET requests, validate origin
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodOptions {
			if origin == "" && referer == "" {
				c.Next()
				return
			}

			originValid := false
			for _, allowed := range allowedOrigins {
				allowed = strings.TrimSpace(allowed)
				if origin == allowed {
					originValid = true
					break
				}

				if origin == "" && referer != "" && strings.HasPrefix(referer, allowed) {
					originValid = true
					break
				}
			}

			if !originValid && origin != "" {
				SendAbort(c, http.StatusForbidden,
					fmt.Sprintf("Origin '%s' is not allowed", origin))
				return
			}
		}

		c.Next()
	}
}
