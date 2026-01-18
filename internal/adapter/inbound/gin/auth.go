package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/activity"
)

type authAdapter struct {
	domain domain.Domain
}

func NewAuthAdapter(domain domain.Domain) inbound_port.AuthHttpPort {
	return &authAdapter{
		domain: domain,
	}
}

func (a *authAdapter) Login(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_login")

	request := struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	identifier := request.Email
	if identifier == "" {
		identifier = request.Username
	}

	user, accessToken, refreshToken, err := a.domain.Auth().Login(ctx, identifier, request.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return enhanced login response with absolute expiry
	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data: model.EnhancedLoginResponse{
			AccessToken:       accessToken,
			RefreshToken:      refreshToken,
			TokenType:         "Bearer",
			ExpiresIn:         int64(15 * 60),           // 15 minutes in seconds
			RefreshExpiresIn:  int64(30 * 24 * 60 * 60), // 30 days
			AbsoluteExpiresIn: int64(90 * 24 * 60 * 60), // 90 days
			User:              user,
		},
	})
}

func (a *authAdapter) Register(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_register")

	var input model.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	user, err := a.domain.Auth().Register(ctx, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data:    user,
	})
}

func (a *authAdapter) Logout(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_logout")

	var request model.LogoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Get user ID from context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	if err := a.domain.Auth().Logout(ctx, userID.(string), request.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "logged out successfully"},
	})
}

func (a *authAdapter) RefreshToken(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_refresh_token")

	var request model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	user, accessToken, newRefreshToken, err := a.domain.Auth().RefreshToken(ctx, request.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Check if token was rotated (newRefreshToken will be non-empty if rotated)
	rotated := newRefreshToken != ""

	response := model.RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(15 * 60), // 15 minutes
		Rotation:    rotated,
		User:        user,
	}

	if rotated {
		// Token was rotated - include new refresh token
		response.RefreshToken = &newRefreshToken
		refreshExpiry := int64(30 * 24 * 60 * 60) // 30 days
		response.RefreshExpiresIn = &refreshExpiry
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    response,
	})
}

func (a *authAdapter) GetProfile(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_get_profile")

	// Get user ID from context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	user, err := a.domain.Auth().GetUserProfile(ctx, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    user.ToResponse(),
	})
}
