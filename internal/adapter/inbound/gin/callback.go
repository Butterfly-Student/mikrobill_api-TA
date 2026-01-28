package gin_inbound_adapter

import (
	"encoding/json"
	"fmt"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/redis"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type callbackHandler struct {
	domain domain.Domain
}

func NewCallbackAdapter(domain domain.Domain) inbound_port.CallbackHttpPort {
	return &callbackHandler{
		domain: domain,
	}
}

func (h *callbackHandler) HandlePPPoEUp(c *gin.Context) {
	var input model.PPPoEEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	err := h.domain.Customer().HandlePPPoEUp(c, input)

	// Always Invalidate PPP active/inactive cache and broadcast to WebSocket
	// This ensures real-time updates happen even if DB logging fails
	h.invalidatePPPCacheAndBroadcast(c, "up", input)

	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func (h *callbackHandler) HandlePPPoEDown(c *gin.Context) {
	var input model.PPPoEEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	err := h.domain.Customer().HandlePPPoEDown(c, input)

	// Always Invalidate PPP active/inactive cache and broadcast to WebSocket
	h.invalidatePPPCacheAndBroadcast(c, "down", input)

	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

// invalidatePPPCacheAndBroadcast invalidates cache dan broadcast event ke WebSocket subscribers
func (h *callbackHandler) invalidatePPPCacheAndBroadcast(c *gin.Context, action string, event model.PPPoEEventInput) {
	// Get tenant ID from context
	tenantId, exists := c.Get("tenant_id")
	if !exists {
		log.Warn("tenant_id not found in context, skipping cache invalidation")
		return
	}

	tenantIdStr := fmt.Sprintf("%v", tenantId)

	// Broadcast to WebSocket subscribers via Redis pub/sub
	channelActive := "ppp:active:" + tenantIdStr
	channelInactive := "ppp:inactive:" + tenantIdStr

	// Create event message
	eventData := map[string]interface{}{
		"action":         action,
		"name":           event.Name,
		"remote_address": event.RemoteAddress,
		"caller_id":      event.CallerID,
		"interface":      event.Interface,
	}

	messageBytes, _ := json.Marshal(eventData)
	message := string(messageBytes)

	// Publish to both channels (active state changed, inactive state also changed)
	if err := redis.Publish(c, channelActive, message); err != nil {
		log.Errorf("Failed to publish to %s: %v", channelActive, err)
	}

	if err := redis.Publish(c, channelInactive, message); err != nil {
		log.Errorf("Failed to publish to %s: %v", channelInactive, err)
	}

	log.Infof("PPPoE %s event broadcasted for user %s (tenant: %s)", action, event.Name, tenantIdStr)
}
