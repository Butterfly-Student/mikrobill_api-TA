package gin_inbound_adapter

import (
	"net/http"

	"MikrOps/internal/domain/customer_auth"
	"MikrOps/internal/model"
	contextutil "MikrOps/utils/context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
)

type CustomerAuthHandler struct {
	customerAuthDomain customer_auth.CustomerAuthDomain
}

func NewCustomerAuthHandler(customerAuthDomain customer_auth.CustomerAuthDomain) *CustomerAuthHandler {
	return &CustomerAuthHandler{
		customerAuthDomain: customerAuthDomain,
	}
}

// CustomerLogin handles customer portal login
// POST /v1/customer/auth/login
func (h *CustomerAuthHandler) CustomerLogin(c *gin.Context) {
	var req model.CustomerLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Get tenant ID from context (should be set by public tenant middleware or slug)
	tenantID, err := contextutil.GetTenantID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing tenant context",
			"message": "Tenant identification required",
		})
		return
	}

	// Attempt login
	_, accessToken, refreshToken, err := h.customerAuthDomain.CustomerLogin(
		c.Request.Context(),
		tenantID,
		req.Email,
		req.Password,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"message": stacktrace.RootCause(err).Error(),
		})
		return
	}

	// Return token response
	c.JSON(http.StatusOK, gin.H{
		"data": model.CustomerLoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(customer_auth.CustomerAccessTokenTTL.Seconds()),
		},
	})
}

// CustomerLogout handles customer portal logout
// POST /v1/customer/auth/logout
func (h *CustomerAuthHandler) CustomerLogout(c *gin.Context) {
	// Get customer ID from context (set by CustomerAuth middleware)
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Customer not authenticated",
		})
		return
	}

	// Get refresh token from request body
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "refresh_token is required",
		})
		return
	}

	// Logout
	if err := h.customerAuthDomain.CustomerLogout(
		c.Request.Context(),
		customerID.(string),
		req.RefreshToken,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Logout failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// RefreshToken handles customer token refresh
// POST /v1/customer/auth/refresh
func (h *CustomerAuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "refresh_token is required",
		})
		return
	}

	// Refresh token
	customer, accessToken, newRefreshToken, err := h.customerAuthDomain.RefreshToken(
		c.Request.Context(),
		req.RefreshToken,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token refresh failed",
			"message": stacktrace.RootCause(err).Error(),
		})
		return
	}

	// Build response
	response := gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int64(customer_auth.CustomerAccessTokenTTL.Seconds()),
	}

	// Include new refresh token if rotated
	if newRefreshToken != "" {
		response["refresh_token"] = newRefreshToken
		response["refresh_expires_in"] = int64(customer_auth.CustomerRefreshTokenTTL.Seconds())
		response["rotated"] = true
	} else {
		response["rotated"] = false
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"customer": gin.H{
			"id":   customer.ID,
			"name": customer.Name,
		},
	})
}

// GetProfile handles getting customer profile
// GET /v1/customer/auth/profile
func (h *CustomerAuthHandler) GetProfile(c *gin.Context) {
	// Get customer ID from context (set by CustomerAuth middleware)
	customerIDStr, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Customer not authenticated",
		})
		return
	}

	customerID, err := uuid.Parse(customerIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid customer ID",
			"message": err.Error(),
		})
		return
	}

	// Get profile
	profile, err := h.customerAuthDomain.GetCustomerProfile(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get profile",
			"message": stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": profile,
	})
}
