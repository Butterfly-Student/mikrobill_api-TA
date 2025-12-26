// File: internal/adapter/handler/auth_handler.go
package handler

import (
	"mikrobill/internal/model"
	"mikrobill/internal/usecase"
	pkg_logger "mikrobill/pkg/logger"
	"mikrobill/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}


func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg_logger.Warn("Invalid login request body", zap.Error(err))
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get client IP for login tracking
	req.IP = c.ClientIP()

	pkg_logger.Debug("Login attempt",
		zap.String("email", req.Email),
		zap.String("ip", req.IP),
	)

	// Execute login usecase
	resp, err := h.authUsecase.Login(c.Request.Context(), req)
	if err != nil {
		pkg_logger.Error("Login failed",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("ip", req.IP),
		)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	pkg_logger.Info("User logged in successfully",
		zap.String("email", req.Email),
		zap.String("ip", req.IP),
	)

	utils.SuccessResponse(c, http.StatusOK, "Login successful", resp)
}


func (h *AuthHandler) Register(c *gin.Context) {
	var req model.CreateUserRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg_logger.Warn("Invalid registration request body", zap.Error(err))
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	pkg_logger.Debug("Registration attempt",
		zap.String("email", req.Email),
		zap.String("username", req.Username),
	)

	// Execute registration usecase
	user, err := h.authUsecase.Register(c.Request.Context(), req)
	if err != nil {
		pkg_logger.Error("Registration failed",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		utils.ErrorResponse(c, http.StatusBadRequest, "Registration failed", err)
		return
	}

	pkg_logger.Info("User registered successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
	)

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", user)
}


func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req model.ChangePasswordRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg_logger.Warn("Invalid change password request body", zap.Error(err))
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from JWT middleware context
	userID, exists := c.Get("user_id")
	if !exists {
		pkg_logger.Error("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		pkg_logger.Error("Invalid user ID type in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	pkg_logger.Debug("Change password attempt",
		zap.Int64("user_id", userIDInt64),
	)

	// Execute change password usecase
	if err := h.authUsecase.ChangePassword(c.Request.Context(), userIDInt64, req); err != nil {
		pkg_logger.Error("Password change failed",
			zap.Error(err),
			zap.Int64("user_id", userIDInt64),
		)
		utils.ErrorResponse(c, http.StatusBadRequest, "Password change failed", err)
		return
	}

	pkg_logger.Info("Password changed successfully",
		zap.Int64("user_id", userIDInt64),
	)

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}


func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user data from JWT middleware context
	userID, _ := c.Get("user_id")
	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")

	pkg_logger.Debug("Get profile request",
		zap.Any("user_id", userID),
		zap.Any("email", email),
	)

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", gin.H{
		"id":    userID,
		"email": email,
		"role":  role,
	})
}


func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")

	pkg_logger.Info("User logged out",
		zap.Any("user_id", userID),
	)

	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}