// cmd/collector/main.go
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/config"
	"github.com/StratWarsAI/strategy-wars/internal/database"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/service"
	"github.com/StratWarsAI/strategy-wars/internal/websocket"
)

func main() {
	log := logger.New("data-collector")
	log.Info("Starting Strategy Wars Data Collector")

	// Load configuration from .env
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	log.Info("Configuration loaded successfully")

	// Connect to database
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Name,
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Error closing db: %v", err)
		}
	}()
	log.Info("Connected to database successfully")

	// Create data service
	dataService := service.NewDataService(db, logger.New("data-service"))

	// Create and connect WebSocket client
	wsClient := websocket.NewClient(cfg.WebSocket.URL, logger.New("websocket"))
	if err := wsClient.Connect(); err != nil {
		log.Error("Failed to connect to WebSocket: %v", err)
		os.Exit(1)
	}
	defer wsClient.Close()

	// Start WebSocket listener
	go wsClient.Listen()
	log.Info("WebSocket listener started")

	// Process incoming WebSocket messages
	go processWebSocketMessages(wsClient, dataService, log)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Data collector is now running. Press Ctrl+C to exit")
	<-sigChan
	log.Info("Shutting down...")

	// Allow some time for pending operations to complete
	time.Sleep(2 * time.Second)
}

// processWebSocketMessages processes messages from WebSocket channels
func processWebSocketMessages(wsClient *websocket.Client, dataService *service.DataService, log *logger.Logger) {
	for {
		select {
		case tokenData := <-wsClient.TokenChannel:
			// Process token data
			if err := dataService.ProcessTokenData(tokenData); err != nil {
				log.Error("Failed to process token data: %v", err)
			}

		case tradeData := <-wsClient.TradeChannel:
			// Process trade data
			if err := dataService.ProcessTradeData(tradeData); err != nil {
				log.Error("Failed to process trade data: %v", err)
			}
		}
	}
}
