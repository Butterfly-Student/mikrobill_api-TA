package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/activity"
)

type logAdapter struct {
	domain domain.Domain
}

func NewLogAdapter(domain domain.Domain) inbound_port.MikrotikLogPort {
	return &logAdapter{
		domain: domain,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func (h *logAdapter) StreamLogs(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_stream_logs")

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("Failed to upgrade to websocket: %v", err)
		return err
	}
	defer conn.Close()

	// Get log stream channel
	logChan, err := h.domain.MikrotikLog().StreamLogs(ctx)
	if err != nil {
		conn.WriteJSON(model.Response{Success: false, Error: err.Error()})
		return nil
	}

	log.Printf("[WebSocket] Client connected for log streaming")

	// Stream logs via WebSocket
	for logData := range logChan {
		if err := conn.WriteJSON(logData); err != nil {
			log.Printf("[WebSocket] Error writing to client: %v", err)
			break
		}
	}

	log.Printf("[WebSocket] Client disconnected from log streaming")
	return nil
}
