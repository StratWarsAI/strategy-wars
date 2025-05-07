# <div align="center"><img src="assets/logo.svg" width="200" height="200" alt="strategywars.fun"></div>

# StrategyWarsAI
## 📋 Overview

StrategyWarsAI is an algorithmic trading platform designed for Solana pump.fun tokens. The platform enables you to create, test, and optimize trading strategies through real-time simulation, AI-powered strategy generation, and head-to-head performance comparison.
## Features
- Strategy Management: Create, edit, and manage trading strategies with customizable parameters
- AI Strategy Generation: Automatically generate optimized trading strategies using advanced AI
- Real-time Simulation: Test strategies against actual market data with detailed metrics
- Strategy Comparison: Compare multiple strategies head-to-head with visual performance charts
- WebSocket Integration: Experience live updates of trades and performance metrics
- Performance Analytics: Track metrics like win rate, ROI, profit/loss ratio, and drawdown


## 🏗 System Architecture
#### StrategyWarsAI consists of

- API Server: Handles HTTP requests and WebSocket connections
- Data Collector: Gathers market data via WebSocket
- Simulation Engine: Executes strategy simulations in real-time
- AI Service: Generates and optimizes trading strategies
- Frontend: Visualizes strategies, trades, and performance metrics



### Core Models

#### The system is built around these key data models
```go
// Strategy represents a trading strategy
type Strategy struct {
    ID              int64     
    Name            string    
    Description     string    
    Config          JSONB     // Configuration parameters
    IsPublic        bool      
    VoteCount       int       
    WinCount        int       
    LastWinTime     time.Time 
    Tags            []string  
    ComplexityScore int       
    RiskScore       int       
    AIEnhanced      bool      
    CreatedAt       time.Time 
    UpdatedAt       time.Time 
}

// StrategyConfig contains the operational parameters
type StrategyConfig struct {
    // Entry conditions
    MarketCapThreshold float64 // Minimum market cap in USD
    OnlyNewTokens      bool    
    MinBuysForEntry    int     // Minimum buy trades to trigger entry
    EntryTimeWindowSec int     // Time window for counting buys

    // Exit conditions
    TakeProfitPct  float64 // Take profit percentage
    StopLossPct    float64 // Stop loss percentage
    MaxHoldTimeSec int     // Maximum hold time (seconds)

    // Position sizing
    FixedPositionSizeSol float64 // Position size in SOL
    InitialBalance       float64 // Starting balance
}

// SimulatedTrade represents a simulated trading activity
type SimulatedTrade struct {
    ID                int64    
    StrategyID        int64    
    TokenID           int64    
    SimulationRunID   *int64   
    EntryPrice        float64  
    ExitPrice         *float64 
    EntryTimestamp    int64    
    ExitTimestamp     *int64   
    PositionSize      float64  
    ProfitLoss        *float64 
    Status            string   // 'open', 'closed', 'canceled'
    ExitReason        *string  
    EntryUsdMarketCap float64  
    ExitUsdMarketCap  *float64 
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

// Token represents a Pump.fun token
type Token struct {
    ID                     int64    
    MintAddress            string   
    CreatorAddress         string   
    Name                   string   
    Symbol                 string   
    ImageUrl               string   
    TwitterUrl             string   
    WebsiteUrl             string   
    TelegramUrl            string   
    MetadataUrl            string   
    CreatedTimestamp       int64    
    MarketCap              float64  
    UsdMarketCap           float64  
    Completed              bool     
    KingOfTheHillTimeStamp int64    
    CreatedAt              time.Time
}
```

## Key Components

## 1. Strategy Management
   * Create and customize trading strategies
   * Define entry and exit conditions
   * Set position sizing and risk parameters
   * AI-powered strategy suggestions
## 2. Simulation Engine
   * Real-time strategy execution
   * Market data processing
   * Trade execution simulation
   * Performance tracking and analysis
## 3. Data Collection System
   * WebSocket connection to market data
   * Token and transaction monitoring
   * Market cap and price tracking
   * Real-time data processing


## 🛠 Development Setup
Prerequisites

- Go 1.23+
- PostgreSQL
- Node.js and pnpm for frontend

## Environment Variables
```bash
# WebSocket
WEBSOCKET_URL=ws://example.com/socket

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=password
DB_NAME=strategywars
DB_SSLMODE=disable

# Server
SERVER_PORT=8080

# AI Service
AI_API_KEY=your_api_key
AI_MODEL=model
AI_ENDPOINT=https://api.example.com

# Automation
AUTOMATION_ENABLED=true
STRATEGY_GEN_INTERVAL=60
PERFORMANCE_ANALYSIS_INTERVAL=15
STRATEGIES_PER_INTERVAL=2
MAX_CONCURRENT_SIMULATIONS=2
```

## Installation
Prerequisites

* git clone https://github.com/StratWarsAI/strategy-wars.git
* cd strategy-wars
* go mod download
* go run cmd/api/main.go
* go run cmd/collector/main.go

### Project Structure
```bash
├── cmd/
│   ├── api/                 # API server entry point
│   │   └── main.go
│   └── collector/           # Data collector entry point
│       └── main.go
│
├── internal/
│   ├── api/                 # API server implementation
│   │   ├── handlers/        # Request handlers
│   │   └── server.go        # Server configuration
│   │
│   ├── config/              # Configuration management
│   │   └── config.go
│   │
│   ├── database/            # Database connection
│   │   └── postgres.go
│   │
│   ├── models/              # Data models
│   │   └── models.go
│   │
│   ├── pkg/                 # Shared packages
│   │   └── logger/          # Logging utilities
│   │
│   ├── repository/          # Data access layer
│   │   └── *.go             # Repository implementations
│   │
│   ├── service/             # Business logic
│   │   ├── ai_service.go
│   │   ├── automation_service.go
│   │   ├── data_service.go
│   │   ├── simulation_service.go
│   │   └── strategy_service.go
│   │
│   └── websocket/           # WebSocket implementation
│       ├── client.go
│       └── server.go
│
├── frontend/                     # Frontend 
└── go.mod                   # Go dependencies                    
```
## Real-time WebSocket Communication
```bash
// WebSocket client connection example
wsClient := websocket.NewClient(cfg.WebSocket.URL, logger.New("websocket"))
if err := wsClient.Connect(); err != nil {
    log.Error("Failed to connect to WebSocket: %v", err)
    os.Exit(1)
}

// Start WebSocket listener
go wsClient.Listen()

// Process incoming messages
go func() {
    for {
        select {
        case tokenData := <-wsClient.TokenChannel:
            // Process token data
        case tradeData := <-wsClient.TradeChannel:
            // Process trade data
        }
    }
}()
```

## 📚 Key Libraries
- Go Fiber (Web Framework)
- Gorilla WebSocket
- PostgreSQL (database/sql)
- React/Next.js (Frontend)
- Recharts (Data Visualization)
- Tailwind CSS
- React Query

## 🔐 Security
- Database connection pooling
- Context-based timeouts
- Error handling and logging
- Goroutine management
- API rate limiting
- Input validation

## 📦 Deployment
The application is designed to be deployed using Docker:
```bash
# Vercel configuration
# Docker Compose configuration
docker-compose.yml
{
  "services": {
    "api": {
      "build": {
        "context": ".",
        "dockerfile": "Dockerfile.api"
      },
      "ports": ["8080:8080"],
      "depends_on": ["postgres"]
    },
    "collector": {
      "build": {
        "context": ".",
        "dockerfile": "Dockerfile.collector"
      },
      "depends_on": ["postgres"]
    },
    "postgres": {
      "image": "postgres:14",
      "environment": {
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_USER": "postgres",
        "POSTGRES_DB": "strategywars"
      },
      "volumes": ["postgres_data:/var/lib/postgresql/data"]
    }
  },
  "volumes": {
    "postgres_data": {}
  }
}
```
## 🤝 Contributing
   * Fork the repository
   * Create your feature branch
   * Commit your changes
   * Push to the branch
   * Real-time updates
   * Create a Pull Request

## 📄 License
This project is licensed under the MIT License.