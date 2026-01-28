package gin_inbound_adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/palantir/stacktrace"
	log "github.com/sirupsen/logrus"

	"MikrOps/internal/domain"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/redis"
)

type pppRealtimeAdapter struct {
	domainRegistry domain.Domain
	upgrader       websocket.Upgrader
}

func NewPPPRealtimeAdapter(domainRegistry domain.Domain) inbound_port.PPPRealtimePort {
	return &pppRealtimeAdapter{
		domainRegistry: domainRegistry,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: Implement proper CORS in production
			},
		},
	}
}

// StreamPPPActive streams real-time PPP active connection updates via WebSocket
func (a *pppRealtimeAdapter) StreamPPPActive(c *gin.Context) error {
	// Get tenant ID from context
	tenantId, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return stacktrace.NewError("tenant_id not found")
	}

	// Upgrade to WebSocket
	ws, err := a.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return stacktrace.Propagate(err, "failed to upgrade to websocket")
	}
	defer ws.Close()

	// Create context for subscription
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Subscribe to Redis pub/sub channel
	channel := "ppp:active:" + tenantId.(string)

	// Client disconnect detection
	clientGone := make(chan struct{})
	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				close(clientGone)
				cancel()
				return
			}
		}
	}()

	// Send initial data
	initialData, err := a.domainRegistry.MikrotikPPPActive().MikrotikListActive(ctx)
	if err != nil {
		ws.WriteJSON(map[string]string{"type": "error", "error": err.Error()})
		return stacktrace.Propagate(err, "failed to get initial PPP active data")
	}

	ws.WriteJSON(map[string]interface{}{
		"type": "initial",
		"data": initialData,
	})

	// Subscribe and stream updates
	go redis.Subscribe(ctx, channel, func(message string) {
		// Parse message
		var update map[string]interface{}
		if err := json.Unmarshal([]byte(message), &update); err != nil {
			log.Errorf("Failed to unmarshal pub/sub message: %v", err)
			return
		}

		// Send to WebSocket client
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := ws.WriteJSON(map[string]interface{}{
			"type": "update",
			"data": update,
		}); err != nil {
			log.Errorf("Failed to write to websocket: %v", err)
			cancel()
		}
	})

	// Wait for client disconnect or context cancel
	<-clientGone

	return nil
}

// StreamPPPInactive streams real-time PPP inactive connection updates via WebSocket
func (a *pppRealtimeAdapter) StreamPPPInactive(c *gin.Context) error {
	tenantId, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return stacktrace.NewError("tenant_id not found")
	}

	ws, err := a.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return stacktrace.Propagate(err, "failed to upgrade to websocket")
	}
	defer ws.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	channel := "ppp:inactive:" + tenantId.(string)

	clientGone := make(chan struct{})
	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				close(clientGone)
				cancel()
				return
			}
		}
	}()

	// Send initial data (inactive = all secrets not in active)
	// For now, we'll send empty array - logic can be implemented later
	ws.WriteJSON(map[string]interface{}{
		"type": "initial",
		"data": []interface{}{},
	})

	go redis.Subscribe(ctx, channel, func(message string) {
		var update map[string]interface{}
		if err := json.Unmarshal([]byte(message), &update); err != nil {
			log.Errorf("Failed to unmarshal pub/sub message: %v", err)
			return
		}

		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := ws.WriteJSON(map[string]interface{}{
			"type": "update",
			"data": update,
		}); err != nil {
			log.Errorf("Failed to write to websocket: %v", err)
			cancel()
		}
	})

	<-clientGone

	return nil
}
