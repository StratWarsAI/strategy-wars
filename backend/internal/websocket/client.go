// internal/websocket/client.go
package websocket

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client for Pump.fun
type Client struct {
	URL            string
	Conn           *websocket.Conn
	Logger         *logger.Logger
	TokenChannel   chan map[string]interface{}
	TradeChannel   chan map[string]interface{}
	done           chan struct{}
	reconnectDelay time.Duration
	mu             sync.Mutex
	isConnected    bool
}

// NewClient creates a new WebSocket client
func NewClient(url string, logger *logger.Logger) *Client {
	return &Client{
		URL:            url,
		Logger:         logger,
		TokenChannel:   make(chan map[string]interface{}, 100),
		TradeChannel:   make(chan map[string]interface{}, 100),
		done:           make(chan struct{}),
		reconnectDelay: 5 * time.Second,
		isConnected:    false,
	}
}

// Connect establishes a WebSocket connection
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return nil
	}

	c.Logger.Info("Connecting to WebSocket: %s", c.URL)

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(c.URL, nil)
	if err != nil {
		return fmt.Errorf("websocket connection error: %v", err)
	}
	c.Conn = conn

	// Send the Socket.IO handshake
	err = conn.WriteMessage(websocket.TextMessage, []byte("40"))
	if err != nil {
		conn.Close()
		return fmt.Errorf("websocket handshake error: %v", err)
	}

	c.isConnected = true
	c.Logger.Info("Connected to WebSocket successfully")
	return nil
}

// Listen starts listening for WebSocket messages
func (c *Client) Listen() {
	c.Logger.Info("Starting WebSocket listener")

	// Ping loop
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				if c.isConnected && c.Conn != nil {
					if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						c.Logger.Error("WebSocket ping error: %v", err)
						c.isConnected = false
						c.Conn.Close()
						c.mu.Unlock()
						// Try to reconnect
						go c.reconnect()
						return
					}
				}
				c.mu.Unlock()
			case <-c.done:
				return
			}
		}
	}()

	// Message receive loop
	for {
		select {
		case <-c.done:
			return
		default:
			c.mu.Lock()
			if !c.isConnected || c.Conn == nil {
				c.mu.Unlock()
				time.Sleep(time.Second) // Don't busy-wait
				continue
			}
			conn := c.Conn
			c.mu.Unlock()

			_, message, err := conn.ReadMessage()
			if err != nil {
				c.Logger.Error("WebSocket read error: %v", err)
				c.mu.Lock()
				c.isConnected = false
				if c.Conn != nil {
					c.Conn.Close()
					c.Conn = nil
				}
				c.mu.Unlock()
				// Try to reconnect
				go c.reconnect()
				// Wait a bit before trying to read again
				time.Sleep(time.Second)
				continue
			}

			// Process the message
			c.processMessage(message)
		}
	}
}

// Close closes the WebSocket connection
func (c *Client) Close() {
	// Prevent closing done channel multiple times
	c.mu.Lock()
	select {
	case <-c.done:
		// Channel already closed
	default:
		close(c.done)
	}

	if c.isConnected && c.Conn != nil {
		c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.Conn.Close()
		c.Conn = nil
	}
	c.isConnected = false
	c.mu.Unlock()
}

// reconnect attempts to reconnect to the WebSocket
func (c *Client) reconnect() {
	c.Logger.Info("Attempting to reconnect in %v...", c.reconnectDelay)
	time.Sleep(c.reconnectDelay)

	// Check if we're shutting down
	select {
	case <-c.done:
		return
	default:
		// continue
	}

	if err := c.Connect(); err != nil {
		c.Logger.Error("Failed to reconnect: %v", err)
		// Use exponential backoff
		c.mu.Lock()
		c.reconnectDelay *= 2
		if c.reconnectDelay > 2*time.Minute {
			c.reconnectDelay = 2 * time.Minute // Cap the delay
		}
		c.mu.Unlock()
		go c.reconnect()
		return
	}

	c.mu.Lock()
	c.reconnectDelay = 5 * time.Second // Reset the delay
	c.mu.Unlock()
}

// processMessage processes a WebSocket message
func (c *Client) processMessage(message []byte) {
	messageStr := string(message)

	// Skip non-data messages
	if len(messageStr) < 2 || !strings.HasPrefix(messageStr, "42") {
		return
	}

	// Parse the JSON payload (Socket.IO format: 42["event",{data}])
	var data []interface{}
	if err := json.Unmarshal([]byte(messageStr[2:]), &data); err != nil {
		c.Logger.Error("Failed to parse WebSocket message: %v", err)
		return
	}

	// Check if we have enough data
	if len(data) < 2 {
		return
	}

	// Check the event type
	eventType, ok := data[0].(string)
	if !ok {
		return
	}

	// Extract the data payload
	dataMap, ok := data[1].(map[string]interface{})
	if !ok {
		c.Logger.Error("Invalid data format")
		return
	}

	// Route based on event type
	switch eventType {
	case "tradeCreated":
		c.Logger.Debug("Received trade event: %s", eventType)
		select {
		case c.TradeChannel <- dataMap:
			// Successfully sent to channel
		default:
			c.Logger.Warn("Trade channel full, dropping message")
		}

	case "tokenCreated":
		c.Logger.Debug("Received token event: %s", eventType)
		select {
		case c.TokenChannel <- dataMap:
			// Successfully sent to channel
		default:
			c.Logger.Warn("Token channel full, dropping message")
		}

	default:
		c.Logger.Debug("Received other event: %s", eventType)
	}
}
