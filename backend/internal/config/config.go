// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config represents the entire application configuration
type Config struct {
	WebSocket struct {
		URL string
	}

	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		SSLMode  string
	}

	Monitoring struct {
		Interval int
	}
	Server struct {
		Port int
	}
}

// LoadConfig loads configuration from .env file
func LoadConfig() (*Config, error) {
	// Determine the root directory (assuming the function is called from the root)
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}

	// Load .env file from root
	envPath := filepath.Join(rootDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Create config struct
	var config Config

	// WebSocket
	config.WebSocket.URL = os.Getenv("WEBSOCKET_URL")

	// Database Configuration
	config.Database.Host = os.Getenv("DB_HOST")

	// Convert port to int with error handling
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_PORT: %v", err)
		}
		config.Database.Port = port
	}

	config.Database.User = os.Getenv("DB_USER")
	config.Database.Password = os.Getenv("DB_PASS")
	config.Database.Name = os.Getenv("DB_NAME")
	config.Database.SSLMode = os.Getenv("DB_SSLMODE")

	// Monitoring Configuration
	if intervalStr := os.Getenv("MONITORING_INTERVAL"); intervalStr != "" {
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid MONITORING_INTERVAL: %v", err)
		}
		config.Monitoring.Interval = interval
	}

	// Server Configuration
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SERVER_PORT: %v", err)
		}
		config.Server.Port = port
	} else {
		config.Server.Port = 8080 // Default port
	}

	// Validate required configurations
	if config.WebSocket.URL == "" {
		return nil, fmt.Errorf("WEBSOCKET_URL is required")
	}
	if config.Database.Host == "" || config.Database.User == "" || config.Database.Name == "" {
		return nil, fmt.Errorf("database configuration is incomplete")
	}

	return &config, nil
}
