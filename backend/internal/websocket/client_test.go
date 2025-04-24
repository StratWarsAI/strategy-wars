// internal/websocket/client_test.go
package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// setupWebSocketServer creates a test WebSocket server
func setupWebSocketServer(t *testing.T, handler func(conn *websocket.Conn)) *httptest.Server {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Create a test server that will handle WebSocket connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
			return
		}

		// Wrap the handler to ensure connection is closed
		func() {
			defer func() {
				if closeErr := conn.Close(); closeErr != nil {
					t.Logf("Error closing WebSocket connection: %v", closeErr)
				}
			}()

			// Run the provided handler with the connection
			handler(conn)
		}()
	}))

	return server
}

func TestClientConnect(t *testing.T) {
	// Setup WebSocket server that expects a handshake
	server := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Expect the Socket.IO handshake message "40"
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}
		assert.Equal(t, "40", string(message))

		// Send back a successful handshake response
		err = conn.WriteMessage(websocket.TextMessage, []byte("40"))
		if err != nil {
			t.Fatalf("Failed to write message: %v", err)
		}
	})
	defer server.Close()

	// Replace "ws://" with "http://" and "wss://" with "https://"
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create a WebSocket client
	client := NewClient(wsURL, logger.New("test"))

	// Connect to the server
	err := client.Connect()
	assert.NoError(t, err)
	assert.True(t, client.isConnected)
	assert.NotNil(t, client.Conn)

	// Close the connection
	client.Close()
}

func TestClientProcessMessage(t *testing.T) {
	// Create a client without actually connecting to a WebSocket server
	client := NewClient("ws://localhost:8080", logger.New("test"))

	// Test cases
	testCases := []struct {
		name        string
		message     string
		expectToken bool
		expectTrade bool
	}{
		{
			name:        "Token Created Event",
			message:     `42["tokenCreated",{"mint":"test-mint","creator":"test-creator","name":"Test Token"}]`,
			expectToken: true,
			expectTrade: false,
		},
		{
			name:        "Trade Created Event",
			message:     `42["tradeCreated",{"mint":"test-mint","signature":"test-sig","is_buy":true}]`,
			expectToken: false,
			expectTrade: true,
		},
		{
			name:        "Invalid Event",
			message:     `42["unknownEvent",{"some":"data"}]`,
			expectToken: false,
			expectTrade: false,
		},
		{
			name:        "Non-data message",
			message:     `2`,
			expectToken: false,
			expectTrade: false,
		},
		{
			name:        "Invalid JSON",
			message:     `42["tokenCreated",{invalid-json]`,
			expectToken: false,
			expectTrade: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create channels to receive processed messages
			tokenChan := make(chan map[string]interface{}, 1)
			tradeChan := make(chan map[string]interface{}, 1)
			doneChan := make(chan struct{})

			client.TokenChannel = tokenChan
			client.TradeChannel = tradeChan
			client.done = doneChan

			// Process the message
			client.processMessage([]byte(tc.message))

			// Check if we received a token message
			if tc.expectToken {
				select {
				case tokenData := <-tokenChan:
					assert.NotNil(t, tokenData)
					if tc.name == "Token Created Event" {
						assert.Equal(t, "test-mint", tokenData["mint"])
					}
				case <-time.After(100 * time.Millisecond):
					t.Fatal("Expected token message not received")
				}
			} else {
				select {
				case <-tokenChan:
					t.Fatal("Unexpected token message received")
				case <-time.After(100 * time.Millisecond):
					// This is expected
				}
			}

			// Check if we received a trade message
			if tc.expectTrade {
				select {
				case tradeData := <-tradeChan:
					assert.NotNil(t, tradeData)
					if tc.name == "Trade Created Event" {
						assert.Equal(t, "test-mint", tradeData["mint"])
					}
				case <-time.After(100 * time.Millisecond):
					t.Fatal("Expected trade message not received")
				}
			} else {
				select {
				case <-tradeChan:
					t.Fatal("Unexpected trade message received")
				case <-time.After(100 * time.Millisecond):
					// This is expected
				}
			}
		})
	}
}

func TestClientReconnectLogic(t *testing.T) {
	// Create a client with a non-existent server
	client := NewClient("ws://non-existent-server:12345", logger.New("test"))

	// Set a shorter reconnect delay for testing
	client.reconnectDelay = 100 * time.Millisecond

	// Attempt to connect (should fail)
	err := client.Connect()
	assert.Error(t, err)
	assert.False(t, client.isConnected)

	// Verify reconnect delay increases
	client.reconnect()
	assert.Greater(t, client.reconnectDelay, 100*time.Millisecond)

	// Clean up
	client.Close()
}

func TestClientListen(t *testing.T) {
	// Mutex and counters for received messages
	var mu sync.Mutex
	receivedTradeCount := 0
	receivedTokenCount := 0

	// Setup WebSocket server that sends some test messages
	server := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Accept the handshake
		_, _, err := conn.ReadMessage()
		if err != nil {
			t.Logf("Failed to read handshake: %v", err)
			return
		}

		// Send a token created event
		err = conn.WriteMessage(websocket.TextMessage, []byte(`42["tokenCreated",{"mint":"test-mint","name":"Test Token"}]`))
		if err != nil {
			t.Logf("Failed to send token event: %v", err)
			return
		}

		// Wait a bit
		time.Sleep(50 * time.Millisecond)

		// Send a trade created event
		err = conn.WriteMessage(websocket.TextMessage, []byte(`42["tradeCreated",{"mint":"test-mint","signature":"test-sig"}]`))
		if err != nil {
			t.Logf("Failed to send trade event: %v", err)
			return
		}

		// Keep the connection open for a bit
		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	// Replace "ws://" with "http://" and "wss://" with "https://"
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create channels with counters
	tokenChan := make(chan map[string]interface{}, 10)
	tradeChan := make(chan map[string]interface{}, 10)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-tokenChan:
				mu.Lock() // Lock
				receivedTokenCount++
				mu.Unlock() // Unlock
			case <-tradeChan:
				mu.Lock() // Lock
				receivedTradeCount++
				mu.Unlock() // Unlock
			case <-done:
				return
			}
		}
	}()

	// Create a client
	client := NewClient(wsURL, logger.New("test"))
	client.TokenChannel = tokenChan
	client.TradeChannel = tradeChan
	client.done = done

	// Connect and start listening
	err := client.Connect()
	assert.NoError(t, err)

	go client.Listen()

	// Wait for messages to be processed
	time.Sleep(500 * time.Millisecond)

	// Close everything
	close(done)
	client.Close()

	// Check counts, sync with mutex
	mu.Lock()
	assert.GreaterOrEqual(t, receivedTokenCount, 1, "Should receive at least one token message")
	assert.GreaterOrEqual(t, receivedTradeCount, 1, "Should receive at least one trade message")
	mu.Unlock()
}
