package gin_inbound_adapter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
)

type DirectMonitorAdapter struct {
	domainRegistry domain.Domain
	upgrader       websocket.Upgrader
}

func NewDirectMonitorAdapter(domainRegistry domain.Domain) inbound_port.DirectMonitorPort {
	return &DirectMonitorAdapter{
		domainRegistry: domainRegistry,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (a *DirectMonitorAdapter) StreamTrafficByInterface(ctx any) {
	c := ctx.(*gin.Context)
	interfaceName := c.Param("interface")

	ws, err := a.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	trafficChan, err := a.domainRegistry.DirectMonitor().StreamTrafficByInterface(c, interfaceName)
	if err != nil {
		ws.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for data := range trafficChan {
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := ws.WriteJSON(data); err != nil {
			return
		}
	}

	ws.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Stream ended"),
		time.Now().Add(time.Second))
}

func (a *DirectMonitorAdapter) PingHost(ctx any) {
	c := ctx.(*gin.Context)
	var request struct {
		IP string `json:"ip" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "Invalid request body"})
		return
	}

	stats, err := a.domainRegistry.DirectMonitor().PingHost(c.Request.Context(), request.IP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: stats})
}

func (a *DirectMonitorAdapter) StreamPingHost(ctx any) {
	c := ctx.(*gin.Context)
	ip := c.Param("ip")

	ws, err := a.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	dataChan, err := a.domainRegistry.DirectMonitor().StreamPingHost(c, ip)
	if err != nil {
		ws.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	clientGone := make(chan struct{})
	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				close(clientGone)
				return
			}
		}
	}()

	for {
		select {
		case resp, ok := <-dataChan:
			if !ok {
				ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Ping finished"))
				return
			}

			if resp.IsSummary {
				ws.WriteJSON(map[string]interface{}{
					"type": "summary",
					"summary": map[string]interface{}{
						"sent":        resp.Sent,
						"received":    resp.Received,
						"packet_loss": resp.PacketLoss,
						"avg_rtt":     resp.AvgRtt,
					},
				})
				return
			} else {
				ws.WriteJSON(map[string]interface{}{
					"type": "update",
					"data": resp,
				})
			}

		case <-clientGone:
			return
		}
	}
}
