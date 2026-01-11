package gin_inbound_adapter

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	"prabogo/utils/activity"
)

type middlewareAdapter struct {
	domain domain.Domain
}

func NewMiddlewareAdapter(
	domain domain.Domain,
) inbound_port.MiddlewareHttpPort {
	return &middlewareAdapter{
		domain: domain,
	}
}

func (h *middlewareAdapter) InternalAuth(a any) error {
	c := a.(*gin.Context)

	// 1. Try to get token from X-API-Key header
	token := c.GetHeader("X-API-Key")

	// 2. Fallback to Authorization: Bearer <token>
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			token = authHeader[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		c.Abort()
		return nil
	}

	// 3. Validate against INTERNAL_KEY
	internalKey := os.Getenv("INTERNAL_KEY")
	if token != internalKey {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		c.Abort()
		return nil
	}

	c.Next()
	return nil
}

func (h *middlewareAdapter) ClientAuth(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_client_auth")
	authHeader := c.GetHeader("Authorization")
	var bearerToken string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		bearerToken = authHeader[7:]
	}

	if bearerToken == "" {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "Unauthorized",
		})
		c.Abort()
		return nil
	}

	exists, err := h.domain.Client().IsExists(ctx, bearerToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		c.Abort()
		return nil
	}

	if !exists {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "Unauthorized",
		})
		c.Abort()
		return nil
	}

	c.Next()
	return nil
}

func (h *middlewareAdapter) UserAuth(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_user_auth")

	authHeader := c.GetHeader("Authorization")
	var bearerToken string
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		bearerToken = authHeader[7:]
	}

	if bearerToken == "" {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "Unauthorized: No token provided",
		})
		c.Abort()
		return nil
	}

	user, err := h.domain.Auth().ValidateToken(ctx, bearerToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "Unauthorized: " + err.Error(),
		})
		c.Abort()
		return nil
	}

	c.Set("user", user)
	c.Set("user_id", user.ID)
	c.Set("role", user.UserRole)

	c.Next()
	return nil
}
