// internal/websocket/server.go
package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/gofiber/websocket/v2"
)

// WSClient represents a WebSocket client connected to our server
type WSClient struct {
	hub  *WSHub
	conn *websocket.Conn
	send chan []byte
}

// WSHub maintains the set of active WebSocket clients
type WSHub struct {
	clients    map[*WSClient]bool
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan []byte
	logger     *logger.Logger
	mu         sync.Mutex
}

// NewWSHub creates a new WebSocket hub for the server
func NewWSHub(logger *logger.Logger) *WSHub {
	return &WSHub{
		clients:    make(map[*WSClient]bool),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan []byte, 256),
		logger:     logger,
	}
}

// Run starts the hub
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("Client connected to WebSocket")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Info("Client disconnected from WebSocket")

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastJSON broadcasts a JSON message to all clients
func (h *WSHub) BroadcastJSON(v interface{}) {
	// Convert v to JSON
	jsonData, err := json.Marshal(v)
	if err != nil {
		h.logger.Error("Error marshaling JSON: %v", err)
		return
	}

	// Log what we're broadcasting
	h.logger.Debug("Broadcasting WebSocket message: %s", string(jsonData))

	h.mu.Lock()
	clientCount := len(h.clients)
	h.mu.Unlock()
	h.logger.Debug("Broadcasting to %d connected clients", clientCount)

	h.broadcast <- jsonData
}

// writePump pumps messages from the hub to the websocket connection
func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()

		// Safely close the connection
		if err := c.conn.Close(); err != nil {
			c.hub.logger.Error("Error closing WebSocket connection: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					c.hub.logger.Error("Error writing close message: %v", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.hub.logger.Error("Error getting next writer: %v", err)
				return
			}

			// Safely write the message
			if _, err = w.Write(message); err != nil {
				c.hub.logger.Error("Failed to write message: %v", err)
				if closeErr := w.Close(); closeErr != nil {
					c.hub.logger.Error("Additional error closing writer: %v", closeErr)
				}
				return
			}

			// Process additional queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				// Safely write newline
				if _, err := w.Write([]byte{'\n'}); err != nil {
					c.hub.logger.Error("Failed to write newline: %v", err)
					if closeErr := w.Close(); closeErr != nil {
						c.hub.logger.Error("Additional error closing writer: %v", closeErr)
					}
					return
				}

				// Safely write queued message
				queuedMessage := <-c.send
				if _, err := w.Write(queuedMessage); err != nil {
					c.hub.logger.Error("Failed to write queued message: %v", err)
					if closeErr := w.Close(); closeErr != nil {
						c.hub.logger.Error("Additional error closing writer: %v", closeErr)
					}
					return
				}
			}

			// Safely close the writer
			if err := w.Close(); err != nil {
				c.hub.logger.Error("Failed to close writer: %v", err)
				return
			}

		case <-ticker.C:
			// Safely send ping message
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.logger.Error("Failed to send ping: %v", err)
				return
			}
		}
	}
}
