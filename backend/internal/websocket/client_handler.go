// internal/websocket/client_handler.go
package websocket

import (
	"encoding/json"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	//maxMessageSize = 1024
)

// ClientWSHandler handles WebSocket connections from clients
type ClientWSHandler struct {
	hub    *WSHub
	logger *logger.Logger
}

// NewClientWSHandler creates a new WebSocket client handler
func NewClientWSHandler(hub *WSHub, logger *logger.Logger) *ClientWSHandler {
	return &ClientWSHandler{
		hub:    hub,
		logger: logger,
	}
}

// ServeWS handles WebSocket requests from clients
func (h *ClientWSHandler) ServeWS() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		client := &WSClient{
			hub:  h.hub,
			conn: c,
			send: make(chan []byte, 256),
		}

		client.hub.register <- client

		go h.readPump(client)
		go h.writePump(client)

		h.logger.Info("New WebSocket client connected")

		// Blocking, connection will be closed when this function returns
		select {}
	})
}

// readPump pumps messages from the WebSocket connection to the hub
func (h *ClientWSHandler) readPump(client *WSClient) {
	defer func() {
		client.hub.unregister <- client
		if err := client.conn.Close(); err != nil {
			h.logger.Error("Error closing connection: %v", err)
		}
		h.logger.Info("WebSocket client disconnected")
	}()

	if err := client.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		h.logger.Error("SetReadDeadline error: %v", err)
	}
	client.conn.SetPongHandler(func(string) error {
		err := client.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			h.logger.Error("Failed to set read deadline: %v", err)
		}
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket error: %v", err)
			}
			break
		}

		// Handle messages from client
		h.handleClientMessage(client, message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (h *ClientWSHandler) writePump(client *WSClient) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := client.conn.Close(); err != nil {
			h.logger.Error("Error closing connection: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-client.send:
			if err := client.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				h.logger.Error("SetWriteDeadline error: %v", err)
				return
			}
			if !ok {
				// The hub closed the channel
				if err := client.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					h.logger.Debug("Error sending close message: %v", err)
				}
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				h.logger.Error("Error writing message: %v", err)
				return
			}

			// Add queued messages to the current websocket message
			n := len(client.send)
			for i := 0; i < n; i++ {
				_, err := w.Write([]byte{'\n'})
				if err != nil {
					h.logger.Error("Failed to write newline: %v", err)
					return
				}
				msg := <-client.send
				_, err = w.Write(msg)
				if err != nil {
					h.logger.Error("Failed to write message: %v", err)
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := client.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				h.logger.Error("SetWriteDeadline error: %v", err)
				return
			}
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage processes messages from clients
func (h *ClientWSHandler) handleClientMessage(client *WSClient, message []byte) {
	// Parse the message
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		h.logger.Error("Error parsing client message: %v", err)
		return
	}

	// Get the message type
	messageType, ok := data["type"].(string)
	if !ok {
		h.logger.Error("Invalid message format: missing 'type' field")
		return
	}

	// Handle different message types
	switch messageType {
	case "subscribe":
		// Handle subscription to specific strategy updates
		h.handleSubscribe(client, data)

	case "unsubscribe":
		// Handle un subscription from strategy updates
		h.handleUnsubscribe(client, data)

	case "ping":
		// Handle ping messages
		h.handlePing(client, data)

	default:
		h.logger.Debug("Unknown message type: %s", messageType)
	}
}

// handleSubscribe handles subscription requests
func (h *ClientWSHandler) handleSubscribe(client *WSClient, data map[string]interface{}) {
	// In a real implementation, you'd store subscription info for each client
	// and only send them relevant updates

	// This is a simple acknowledgment
	response := map[string]interface{}{
		"type":    "subscribe_ack",
		"success": true,
	}

	if strategyID, ok := data["strategy_id"].(float64); ok {
		response["strategy_id"] = strategyID
		h.logger.Info("Client subscribed to strategy %d", int64(strategyID))
	}

	// Send acknowledgment
	jsonResponse, _ := json.Marshal(response)
	client.send <- jsonResponse
}

// handleUnsubscribe handles un subscription requests
func (h *ClientWSHandler) handleUnsubscribe(client *WSClient, data map[string]interface{}) {
	// In a real implementation, you'd remove the subscription

	// This is a simple acknowledgment
	response := map[string]interface{}{
		"type":    "unsubscribe_ack",
		"success": true,
	}

	if strategyID, ok := data["strategy_id"].(float64); ok {
		response["strategy_id"] = strategyID
		h.logger.Info("Client unsubscribed from strategy %d", int64(strategyID))
	}

	// Send acknowledgment
	jsonResponse, _ := json.Marshal(response)
	client.send <- jsonResponse
}

// handlePing handles ping messages
func (h *ClientWSHandler) handlePing(client *WSClient, _ map[string]interface{}) {
	// Send pong response
	response := map[string]interface{}{
		"type": "pong",
		"time": time.Now().Unix(),
	}

	jsonResponse, _ := json.Marshal(response)
	client.send <- jsonResponse
}
