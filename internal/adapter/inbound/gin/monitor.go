package gin_inbound_adapter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	"prabogo/utils/activity"
)

type monitorAdapter struct {
	domain domain.Domain
}

func NewMonitorAdapter(domain domain.Domain) inbound_port.MonitorPort {
	return &monitorAdapter{
		domain: domain,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

func (h *monitorAdapter) StreamTraffic(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("ws_monitor_traffic")

	interfaceName := c.Param("interface")
	if interfaceName == "" {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "Interface name required"})
		return nil
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrade handles error response
		return nil
	}
	defer ws.Close()

	// Start streaming
	dataChan, cancel, err := h.domain.Monitor().StreamTraffic(ctx, interfaceName)
	if err != nil {
		ws.WriteJSON(map[string]string{"error": err.Error()})
		return nil
	}
	defer cancel()

	// Handle disconnect from client side
	clientGone := make(chan struct{})
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				close(clientGone)
				return
			}
		}
	}()

	// Loop and send data
	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				// Stream ended from source
				ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Stream ended"))
				return nil
			}
			if err := ws.WriteJSON(data); err != nil {
				return nil
			}
		case <-clientGone:
			return nil
		case <-time.After(30 * time.Second):
			// Ping/Keepalive
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return nil
			}
		}
	}
}
