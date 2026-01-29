package gin_inbound_adapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/palantir/stacktrace"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	"MikrOps/utils/logger"

	"go.uber.org/zap"
)

type CustomerPortalHandler struct {
	domainRegistry domain.Domain
}

func NewCustomerPortalHandler(domainRegistry domain.Domain) *CustomerPortalHandler {
	return &CustomerPortalHandler{
		domainRegistry: domainRegistry,
	}
}

// GetProfile handles getting customer profile with conditional credential visibility
// GET /v1/customer/portal/profile
func (h *CustomerPortalHandler) GetProfile(c *gin.Context) {
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

	// Get profile from domain
	profile, err := h.domainRegistry.CustomerPortal().GetCustomerProfile(c.Request.Context(), customerID)
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

// UpdateProfile handles updating customer profile
// PUT /v1/customer/portal/profile
func (h *CustomerPortalHandler) UpdateProfile(c *gin.Context) {
	// Get customer ID from context
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

	// Parse request body
	var req model.UpdateCustomerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Update profile
	if err := h.domainRegistry.CustomerPortal().UpdateCustomerProfile(c.Request.Context(), customerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update profile",
			"message": stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// GetTraffic handles getting current traffic statistics
// GET /v1/customer/portal/traffic
func (h *CustomerPortalHandler) GetTraffic(c *gin.Context) {
	// Get customer ID from context
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

	// Get traffic stats from domain
	traffic, err := h.domainRegistry.CustomerPortal().GetTrafficStats(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get traffic",
			"message": stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": traffic,
	})
}

// StreamTraffic handles WebSocket streaming of traffic data
// GET /v1/customer/portal/traffic/stream
// This leverages Monitor.StreamTraffic which uses MikroTik LISTEN mode (not polling)
func (h *CustomerPortalHandler) StreamTraffic(c *gin.Context) {
	// Get customer ID from context
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

	// Upgrade to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for now
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// Start monitoring via Monitor domain (uses MikroTik LISTEN, not polling)
	trafficChan, err := h.domainRegistry.Monitor().StreamTraffic(c.Request.Context(), customerID.String())
	if err != nil {
		conn.WriteJSON(map[string]interface{}{
			"error":   "Failed to start streaming",
			"message": err.Error(),
		})
		return
	}

	// Stream traffic data to WebSocket client
	for traffic := range trafficChan {
		if err := conn.WriteJSON(traffic); err != nil {
			logger.Error("Failed to write to WebSocket", zap.Error(err))
			break
		}
	}
}

// ResetSession handles resetting customer's active session
// POST /v1/customer/portal/session/reset
func (h *CustomerPortalHandler) ResetSession(c *gin.Context) {
	// Get customer ID from context
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

	// Reset session
	if err := h.domainRegistry.CustomerPortal().ResetSession(c.Request.Context(), customerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reset session",
			"message": stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session reset request sent successfully",
		"note":    "Disconnection will be processed momentarily",
	})
}

// GetUsageHistory handles getting historical usage data
// GET /v1/customer/portal/usage
func (h *CustomerPortalHandler) GetUsageHistory(c *gin.Context) {
	// Get customer ID from context
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

	// Get days parameter (default 30, max 90)
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	// Get usage history from domain
	usage, err := h.domainRegistry.CustomerPortal().GetUsageHistory(c.Request.Context(), customerID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get usage history",
			"message": stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": usage,
		"meta": gin.H{
			"days": days,
		},
	})
}
