package gin_inbound_adapter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/palantir/stacktrace"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
)

type monitorAdapter struct {
	domainRegistry domain.Domain
	upgrader       websocket.Upgrader
}

func NewMonitorAdapter(domainRegistry domain.Domain) inbound_port.MonitorPort {
	return &monitorAdapter{
		domainRegistry: domainRegistry,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

func (a *monitorAdapter) StreamTraffic(ctx any) error {
	c := ctx.(*gin.Context)
	customerID := c.Param("id")

	// Upgrade HTTP connection to WebSocket
	ws, err := a.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// upgrader.Upgrade handles error response
		return stacktrace.Propagate(err, "failed to upgrade to websocket")
	}
	defer ws.Close()

	// Get streaming channel from domain
	trafficChan, err := a.domainRegistry.Monitor().StreamTraffic(c, customerID)
	if err != nil {
		ws.WriteJSON(map[string]string{"error": err.Error()})
		return stacktrace.Propagate(err, "failed to start traffic stream")
	}

	// 1. Listen for client disconnect (read loop)
	// We need a read loop to detect close frames, even if we only write.
	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// 2. Stream data to client (write loop)
	for data := range trafficChan {
		// Using JSON marshaling for simplicity.
		// For high performance, we might pre-marshal in domain or use raw bytes if optimized.
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := ws.WriteJSON(data); err != nil {
			// Client disconnected or error
			return nil
		}
	}

	// Channel closed by domain (e.g. max restarts reached or context cancelled)
	ws.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Stream ended"),
		time.Now().Add(time.Second))

	return nil
}

func (a *monitorAdapter) PingCustomer(ctx any) error {
	c := ctx.(*gin.Context)
	customerID := c.Param("id")

	stats, err := a.domainRegistry.Monitor().PingCustomer(c, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: stats})
	return nil
}

func (a *monitorAdapter) StreamPing(ctx any) error {
	c := ctx.(*gin.Context)
	customerID := c.Param("id")

	// Upgrade
	ws, err := a.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil
	}
	defer ws.Close()

	// Start Ping Stream
	dataChan, err := a.domainRegistry.Monitor().StreamPing(c, customerID)
	if err != nil {
		ws.WriteJSON(map[string]string{"type": "error", "error": err.Error()})
		return nil
	}

	// Client disconnect handler
	clientGone := make(chan struct{})
	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				close(clientGone)
				return
			}
		}
	}()

	// Stream Loop
	for {
		select {
		case resp, ok := <-dataChan:
			if !ok {
				ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Ping finished"))
				return nil
			}

			// If it's summary, send summary type, else update
			if resp.IsSummary {
				// User wants specific summary format
				// "sent": sent, "received": received, "packet_loss": ...
				// Our PingResponse has these fields.

				// Send as summary object
				ws.WriteJSON(map[string]interface{}{
					"type": "summary",
					"summary": map[string]interface{}{
						"sent":        resp.Sent,
						"received":    resp.Received,
						"packet_loss": resp.PacketLoss,
						"avg_rtt":     resp.AvgRtt,
					},
				})
				return nil // Summary usually means done?
			} else {
				ws.WriteJSON(map[string]interface{}{
					"type": "update",
					"data": resp,
				})
			}

		case <-clientGone:
			return nil
		}
	}
}
