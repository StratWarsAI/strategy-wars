// cmd/api/main.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/api"
	"github.com/StratWarsAI/strategy-wars/internal/config"
	"github.com/StratWarsAI/strategy-wars/internal/database"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

func main() {
	log := logger.New("api-server")
	log.Info("Starting StrategyWars API Server")

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

	// Create and start API server
	apiServer := api.NewServer(cfg.Server.Port, db, cfg, logger.New("api-server"))
	go func() {
		if err := apiServer.Start(cfg.Automation.Enabled); err != nil {
			log.Error("API server error: %v", err)
		}
	}()
	log.Info("API server started on port %d", cfg.Server.Port)
	if cfg.Automation.Enabled {
		log.Info("Automation services are enabled")
	} else {
		log.Info("Automation services are disabled")
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("API server is now running. Press Ctrl+C to exit")
	<-sigChan
	log.Info("Shutting down...")

	// Gracefully shutdown API server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := apiServer.Stop(ctx); err != nil {
		log.Error("Error stopping API server: %v", err)
	}

	// Allow some time for pending operations to complete
	time.Sleep(2 * time.Second)
}
