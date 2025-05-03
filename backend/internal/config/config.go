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

	AI struct {
		APIKey   string
		Model    string
		Endpoint string
	}

	Monitoring struct {
		Interval int
	}

	Server struct {
		Port int
	}

	Automation struct {
		Enabled                     bool
		StrategyGenerationInterval  int // In minutes
		PerformanceAnalysisInterval int // In minutes
		StrategiesPerInterval       int
		MaxConcurrentSimulations    int
	}
}

// LoadConfig loads configuration from .env file
func LoadConfig() (*Config, error) {
	// Determine the root directory (assuming the function is called from the root)
	rootDir, err := filepath.Abs("..")
	if err != nil {
		return nil, fmt.Errorf("failed to determine project root: %v", err)
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

	// AI Configuration
	config.AI.APIKey = os.Getenv("AI_API_KEY")
	config.AI.Model = os.Getenv("AI_MODEL")
	config.AI.Endpoint = os.Getenv("AI_ENDPOINT")

	// Automation Configuration
	if enabledStr := os.Getenv("AUTOMATION_ENABLED"); enabledStr != "" {
		enabled, err := strconv.ParseBool(enabledStr)
		if err != nil {
			return nil, fmt.Errorf("invalid AUTOMATION_ENABLED: %v", err)
		}
		config.Automation.Enabled = enabled
	} else {
		config.Automation.Enabled = false // Disabled by default
	}

	if intervalStr := os.Getenv("STRATEGY_GEN_INTERVAL"); intervalStr != "" {
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid STRATEGY_GEN_INTERVAL: %v", err)
		}
		config.Automation.StrategyGenerationInterval = interval
	} else {
		config.Automation.StrategyGenerationInterval = 60 // Default 60 minutes
	}

	if intervalStr := os.Getenv("PERFORMANCE_ANALYSIS_INTERVAL"); intervalStr != "" {
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PERFORMANCE_ANALYSIS_INTERVAL: %v", err)
		}
		config.Automation.PerformanceAnalysisInterval = interval
	} else {
		config.Automation.PerformanceAnalysisInterval = 15 // Default 15 minutes
	}

	if countStr := os.Getenv("STRATEGIES_PER_INTERVAL"); countStr != "" {
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid STRATEGIES_PER_INTERVAL: %v", err)
		}
		config.Automation.StrategiesPerInterval = count
	} else {
		config.Automation.StrategiesPerInterval = 2 // Default 2 strategies
	}

	if maxStr := os.Getenv("MAX_CONCURRENT_SIMULATIONS"); maxStr != "" {
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_CONCURRENT_SIMULATIONS: %v", err)
		}
		config.Automation.MaxConcurrentSimulations = max
	} else {
		config.Automation.MaxConcurrentSimulations = 3 // Default 3 concurrent simulations
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
