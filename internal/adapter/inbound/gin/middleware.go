package gin_inbound_adapter

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/activity"
	"MikrOps/utils/logger"
	"MikrOps/utils/redis"
)

type middlewareAdapter struct {
	domain      domain.Domain
	redisClient any // Will be used for rate limiting if available
}

func NewMiddlewareAdapter(
	domain domain.Domain,
) inbound_port.MiddlewareHttpPort {
	return &middlewareAdapter{
		domain: domain,
	}
}

// InternalAuth - Legacy
func (h *middlewareAdapter) InternalAuth(a any) error {
	c := a.(*gin.Context)
	token := c.GetHeader("X-API-Key")
	if token == "" {
		token = c.GetHeader("X-Client-Key")
	}
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			token = authHeader[7:]
		}
	}

	if token == "" {
		SendAbort(c, http.StatusUnauthorized, "Unauthorized")
		return nil
	}

	// 1. Try static internal key first (Legacy/M2M)
	internalKey := os.Getenv("INTERNAL_KEY")
	if internalKey == "" {
		logger.GetLogger().Warn("INTERNAL_KEY environment variable not set, internal authentication disabled")
		SendAbort(c, http.StatusUnauthorized, "Internal authentication not configured")
		return nil
	}
	if len(internalKey) < 32 {
		logger.GetLogger().Warn("INTERNAL_KEY too weak, must be at least 32 characters")
		SendAbort(c, http.StatusInternalServerError, "Internal key too weak")
		return nil
	}

	if token == internalKey {
		// Set a "system" user mock for internal access
		systemUser := &model.User{
			Username:     "system",
			IsSuperadmin: true,
			UserRole:     model.UserRoleSuperAdmin,
			Status:       model.UserStatusActive,
		}
		c.Set("user", systemUser)
		c.Set("is_internal", true)
		c.Next()
		return nil
	}

	// 2. Try JWT validation
	user, err := h.domain.Auth().ValidateToken(activity.NewContext(c.Request.Context(), "http_internal_auth"), token)
	if err == nil {
		c.Set("user", user)
		c.Set("user_id", user.ID)
		if user.TenantID != nil {
			c.Set("tenant_id", *user.TenantID)
		}
		c.Set("user_role", string(user.UserRole))
		c.Next()
		return nil
	}

	SendAbort(c, http.StatusUnauthorized, "Unauthorized")
	return nil
}

// ClientAuth - Legacy
func (h *middlewareAdapter) ClientAuth(a any) error {
	c := a.(*gin.Context)
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		SendAbort(c, http.StatusUnauthorized, "Unauthorized")
		return nil
	}
	token := authHeader[7:]
	exists, _ := h.domain.Client().IsExists(activity.NewContext(c.Request.Context(), "http_client_auth"), token)
	if !exists {
		SendAbort(c, http.StatusUnauthorized, "Unauthorized")
		return nil
	}
	c.Next()
	return nil
}

// UserAuth - Legacy
func (h *middlewareAdapter) UserAuth(a any) error {
	c := a.(*gin.Context)
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		SendAbort(c, http.StatusUnauthorized, "Unauthorized")
		return nil
	}
	token := authHeader[7:]
	user, err := h.domain.Auth().ValidateToken(activity.NewContext(c.Request.Context(), "http_user_auth"), token)
	if err != nil {
		SendAbort(c, http.StatusUnauthorized, err.Error())
		return nil
	}
	c.Set("user", user)
	c.Set("user_id", user.ID)
	c.Set("tenant_id", user.TenantID)
	c.Set("user_role", string(user.UserRole))
	c.Next()
	return nil
}

// RequestID Middleware
func (h *middlewareAdapter) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ZapLogger Middleware
func (h *middlewareAdapter) ZapLogger() gin.HandlerFunc {
	l := logger.GetLogger()
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		latency := time.Since(start)

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		}
		if reqID, exists := c.Get("request_id"); exists {
			fields = append(fields, zap.String("request_id", reqID.(string)))
		}

		if c.Writer.Status() >= 500 {
			l.Error("Server error", fields...)
		} else {
			l.Info("Request handled", fields...)
		}
	}
}

// CORS Middleware
func (h *middlewareAdapter) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from environment
		allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
		origin := c.Request.Header.Get("Origin")

		// Default to allow all in development if not configured
		if allowedOriginsEnv == "" {
			logger.GetLogger().Warn("CORS_ALLOWED_ORIGINS not set, defaulting to allow all origins (not recommended for production)")
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Check if origin is allowed
			allowedOrigins := strings.Split(allowedOriginsEnv, ",")
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				allowedOrigin = strings.TrimSpace(allowedOrigin)
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// RateLimit Middleware
func (h *middlewareAdapter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		userID := getUserIDFromContext(c)

		var key string
		if userID != "" {
			key = fmt.Sprintf("ratelimit:user:%s", userID)
		} else {
			key = fmt.Sprintf("ratelimit:ip:%s", clientIP)
		}

		// Simple implementation using utils/redis
		ctx := c.Request.Context()
		limit := int64(60) // Default limit per minute

		// Role-based limits
		if role, exists := c.Get("user_role"); exists {
			switch strings.ToLower(role.(string)) {
			case string(model.UserRoleSuperAdmin):
				limit = 1000
			case string(model.UserRoleAdmin):
				limit = 200
			case string(model.UserRoleTechnician):
				limit = 120
			}
		}

		// Placeholder for actual increment logic (using Get/Set for now since Incr is not in utils/redis)
		// Ideally we would add Incr to utils/redis
		countStr, _ := redis.Get(ctx, key)
		var count int64
		fmt.Sscanf(countStr, "%d", &count)

		if count >= limit {
			SendAbort(c, http.StatusTooManyRequests, "Too many requests. Please try again later.")
			return
		}

		redis.Set(ctx, key, fmt.Sprintf("%d", count+1))

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-(count+1)))

		// Reset every minute (placeholder calculation)
		resetTime := time.Now().Add(time.Minute).Truncate(time.Minute).Unix()
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		c.Next()
	}
}

// RequireTenantAccess Middleware
// Now simplified - tenant context is already resolved and validated by TenantContext()
// This middleware just ensures tenant_id exists in context as a safety check
func (h *middlewareAdapter) RequireTenantAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Tenant should already be in context from TenantContext middleware
		tenantIDValue, exists := c.Get("tenant_id")
		if !exists {
			SendAbort(c, http.StatusForbidden, "Tenant context not found")
			return
		}

		// Verify it's a valid UUID
		if _, ok := tenantIDValue.(uuid.UUID); !ok {
			SendAbort(c, http.StatusForbidden, "Invalid tenant context")
			return
		}

		c.Next()
	}
}

// RequireRole Middleware
func (h *middlewareAdapter) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Role not found"})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		role := model.UserRole(roleStr)

		if role == model.UserRoleSuperAdmin {
			c.Next()
			return
		}

		for _, r := range roles {
			if strings.EqualFold(roleStr, r) {
				c.Next()
				return
			}
		}

		SendAbort(c, http.StatusForbidden, fmt.Sprintf("Required role: %s", strings.Join(roles, " or ")))
	}
}

// Validator Middleware
func (h *middlewareAdapter) Validator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Example validation: Ensure Content-Type is application/json for non-GET requests
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodDelete &&
			c.Request.Method != http.MethodOptions {
			contentType := c.GetHeader("Content-Type")
			// Allow empty body (Content-Length: 0) to skip Content-Type check
			if c.Request.ContentLength > 0 && !strings.HasPrefix(contentType, "application/json") {
				SendAbort(c, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}
		}
		c.Next()
	}
}

// Helper functions (formerly in separate files)

func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uuid.UUID); ok {
			return id.String()
		}
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

func extractTenantIDFromRequest(c *gin.Context) string {
	tenantID := c.Param("tenant_id")
	if tenantID != "" {
		return tenantID
	}
	return c.Query("tenant_id")
}

// CustomerAuth Middleware - Validates customer JWT tokens for customer portal
func (h *middlewareAdapter) CustomerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			SendAbort(c, http.StatusUnauthorized, "Unauthorized - missing token")
			return
		}

		token := authHeader[7:]

		// Import customer_auth domain to validate customer token
		// Note: We need to add customer auth domain to middleware dependencies
		// For now, we'll use a simpler validation approach
		// Extract claims from JWT manually
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "default_secret_please_change"
		}

		// Parse JWT and extract customer_id and tenant_id
		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !parsedToken.Valid {
			SendAbort(c, http.StatusUnauthorized, "Invalid token")
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			SendAbort(c, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		// Verify this is a customer token (not admin token)
		tokenType, _ := claims["type"].(string)
		if tokenType != "customer" {
			SendAbort(c, http.StatusUnauthorized, "Invalid token type")
			return
		}

		// Extract customer ID and tenant ID
		customerID, ok := claims["sub"].(string)
		if !ok {
			SendAbort(c, http.StatusUnauthorized, "Invalid customer ID in token")
			return
		}

		tenantIDStr, ok := claims["tenant_id"].(string)
		if !ok {
			SendAbort(c, http.StatusUnauthorized, "Invalid tenant ID in token")
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			SendAbort(c, http.StatusUnauthorized, "Invalid tenant ID format")
			return
		}

		// Set customer context
		c.Set("customer_id", customerID)
		c.Set("tenant_id", tenantID)
		c.Set("customer_email", claims["email"])
		c.Set("is_customer", true)

		c.Next()
	}
}
