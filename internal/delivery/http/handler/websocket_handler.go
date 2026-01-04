package handler

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WebSocketHandler handles WebSocket connections and broadcasting
type WebSocketHandler struct {
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	broadcast chan []byte
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

// GetBroadcastChannel returns the broadcast channel for other components
func (h *WebSocketHandler) GetBroadcastChannel() chan []byte {
	return h.broadcast
}

// GetClientCount returns the number of connected clients
func (h *WebSocketHandler) GetClientCount() int {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	return len(h.clients)
}

// HandleWS handles WebSocket connection requests
func (h *WebSocketHandler) HandleWS(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	clientIP := c.ClientIP()
	log.Printf("New WebSocket client connected from %s", clientIP)

	h.clientsMu.Lock()
	h.clients[ws] = true
	h.clientsMu.Unlock()

	defer func() {
		h.clientsMu.Lock()
		delete(h.clients, ws)
		h.clientsMu.Unlock()
		ws.Close()
		log.Printf("WebSocket client disconnected from %s", clientIP)
	}()

	// Configure WebSocket
	// Set read deadline to detect stale connections
	// ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	// ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

	for {
		// _, _, err := ws.ReadMessage()
		// For extensive debugging, verify what error is returned
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error from %s: %v", clientIP, err)
			} else {
				// Normal closure or expected error
				log.Printf("WebSocket closed by client %s: %v", clientIP, err)
			}
			break
		}
		// If we receive a message? For now we just ignore inputs or log them
		if len(message) > 0 {
			log.Printf("Received message from %s: %s", clientIP, string(message))
		}
	}
}

// Broadcaster runs in a goroutine to broadcast messages to all clients
func (h *WebSocketHandler) Broadcaster() {
	for {
		msg := <-h.broadcast

		// Get all clients safely
		h.clientsMu.RLock()
		clients := make([]*websocket.Conn, 0, len(h.clients))
		for client := range h.clients {
			clients = append(clients, client)
		}
		h.clientsMu.RUnlock()

		// Send to all clients
		for _, client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("Write error: %v", err)
				client.Close()

				// Remove dead client
				h.clientsMu.Lock()
				delete(h.clients, client)
				h.clientsMu.Unlock()
			}
		}
	}
}

// HandleHealthCheck handles health check endpoint
func (h *WebSocketHandler) HandleHealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"clients":   h.GetClientCount(),
	})
}
