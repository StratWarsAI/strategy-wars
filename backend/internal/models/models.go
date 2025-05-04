// internal/models/models.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Strategy represents a trading strategy
type Strategy struct {
	ID              int64     `json:"-"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	Config          JSONB     `json:"config"`
	IsPublic        bool      `json:"is_public"`
	VoteCount       int       `json:"vote_count"`
	WinCount        int       `json:"win_count"`
	LastWinTime     time.Time `json:"last_win_time,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
	ComplexityScore int       `json:"complexity_score"`
	RiskScore       int       `json:"risk_score"`
	AIEnhanced      bool      `json:"ai_enhanced"`
	CreatedAt       time.Time `json:"-"`
	UpdatedAt       time.Time `json:"-"`
}

type StrategyConfig struct {
	// Entry conditions
	MarketCapThreshold float64 `json:"marketCapThreshold"` // Minimum market cap in USD
	OnlyNewTokens      bool    `json:"only_new_tokens"`
	MinBuysForEntry    int     `json:"minBuysForEntry"`    // Minimum buy trades to trigger entry
	EntryTimeWindowSec int     `json:"entryTimeWindowSec"` // Time window for counting buys (seconds)

	// Exit conditions
	TakeProfitPct  float64 `json:"takeProfitPct"`  // Take profit percentage
	StopLossPct    float64 `json:"stopLossPct"`    // Stop loss percentage
	MaxHoldTimeSec int     `json:"maxHoldTimeSec"` // Maximum hold time (seconds)

	// Position sizing
	FixedPositionSizeSol float64 `json:"fixedPositionSizeSol"` // Fixed position size in SOL

	InitialBalance   float64 `json:"initialBalance"`      // Initial balance in SOL
	MaxTokensToTrack int     `json:"max_tokens_to_track"` // Maximum number of tokens to track
}

// Simulation runs a trading simulation
type SimulationRun struct {
	ID                   int64     `json:"-"`
	StartTime            time.Time `json:"start_time"`
	EndTime              time.Time `json:"end_time"`
	WinnerStrategyID     int64     `json:"winner_strategy_id,omitempty"`
	Status               string    `json:"status"` // 'preparing', 'running', 'completed', 'failed'
	SimulationParameters JSONB     `json:"simulation_parameters,omitempty"`
	CreatedAt            time.Time `json:"-"`
	UpdatedAt            time.Time `json:"-"`
}

// SimulationResult represents the result of a simulation
type SimulationResult struct {
	ID                int64     `json:"-"`
	SimulationRunID   int64     `json:"-"`
	StrategyID        int64     `json:"-"`
	ROI               float64   `json:"roi"`
	TradeCount        int       `json:"trade_count"`
	WinRate           float64   `json:"win_rate"`
	MaxDrawdown       float64   `json:"max_drawdown"`
	PerformanceRating string    `json:"performance_rating"` // 'excellent', 'good', 'average', 'poor', 'very_poor'
	Analysis          string    `json:"analysis,omitempty"`
	Rank              int       `json:"rank"`
	CreatedAt         time.Time `json:"-"`
}

// StrategyMetric represents AI analysis of a strategy
type StrategyMetric struct {
	ID               int64     `json:"-"`
	StrategyID       int64     `json:"-"`
	SimulationRunID  *int64    `json:"-"`
	WinRate          float64   `json:"win_rate"`
	AvgProfit        float64   `json:"avg_profit"`
	AvgLoss          float64   `json:"avg_loss"`
	MaxDrawdown      float64   `json:"max_drawdown"`
	TotalTrades      int       `json:"total_trades"`
	SuccessfulTrades int       `json:"successful_trades"`
	RiskScore        int       `json:"risk_score"`
	CreatedAt        time.Time `json:"-"`
}

// SimulationEvent represents events that occur during a simulation
type SimulationEvent struct {
	ID              int64     `json:"-"`
	StrategyID      int64     `json:"-"`
	SimulationRunID int64     `json:"-"`
	EventType       string    `json:"event_type"` // "simulation_started", "trade_executed", "trade_closed", etc.
	EventData       JSONB     `json:"event_data"` // Flexible JSON data structure for event-specific data
	Timestamp       time.Time `json:"timestamp"`
	CreatedAt       time.Time `json:"-"`
}

// Strategy Generation represents a generation of strategies
type StrategyGeneration struct {
	ID                int64     `json:"-"`
	GenerationNumber  int       `json:"generation_number"`
	ParentStrategyID  int64     `json:"parent_strategy_id"`
	ChildStrategyID   int64     `json:"child_strategy_id"`
	ImprovementReason string    `json:"improvement_reason,omitempty"`
	CreatedAt         time.Time `json:"-"`
}

// Token represents a Pump.fun token
type Token struct {
	ID                     int64     `json:"-"`
	MintAddress            string    `json:"mint"`
	CreatorAddress         string    `json:"creator"`
	Name                   string    `json:"name"`
	Symbol                 string    `json:"symbol"`
	ImageUrl               string    `json:"image_url"`
	TwitterUrl             string    `json:"twitter_url"`
	WebsiteUrl             string    `json:"website_url"`
	TelegramUrl            string    `json:"telegram_url"`
	MetadataUrl            string    `json:"metadata_url"`
	CreatedTimestamp       int64     `json:"created_timestamp"`
	MarketCap              float64   `json:"market_cap"`
	UsdMarketCap           float64   `json:"usd_market_cap"`
	Completed              bool      `json:"completed"`
	KingOfTheHillTimeStamp int64     `json:"king_of_the_hill_timestamp"`
	CreatedAt              time.Time `json:"-"`
}

// Trade represents a Pump.fun trade
type Trade struct {
	ID          int64   `json:"-"`
	TokenID     int64   `json:"-"`
	MintAddress string  `json:"mint"`
	Signature   string  `json:"signature"`
	SolAmount   float64 `json:"sol_amount"`
	TokenAmount float64 `json:"token_amount"`
	IsBuy       bool    `json:"is_buy"`
	UserAddress string  `json:"user"`
	Timestamp   int64   `json:"timestamp"`
}

// SimulatedTrade represents a simulated trading activity
type SimulatedTrade struct {
	ID                int64     `json:"-"`
	StrategyID        int64     `json:"-"`
	TokenID           int64     `json:"-"`
	SimulationRunID   *int64    `json:"-"`
	EntryPrice        float64   `json:"entry_price"`
	ExitPrice         *float64  `json:"exit_price,omitempty"`
	EntryTimestamp    int64     `json:"entry_timestamp"`
	ExitTimestamp     *int64    `json:"exit_timestamp,omitempty"`
	PositionSize      float64   `json:"position_size"`
	ProfitLoss        *float64  `json:"profit_loss,omitempty"`
	Status            string    `json:"status"` // 'open', 'closed', 'canceled'
	ExitReason        *string   `json:"exit_reason,omitempty"`
	EntryUsdMarketCap float64   `json:"entry_usd_market_cap"`
	ExitUsdMarketCap  *float64  `json:"exit_usd_market_cap,omitempty"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

// Custom type for JSONB handling
type JSONB map[string]interface{}

// Implement Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Implement Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type for JSONB")
	}

	return json.Unmarshal(data, j)
}
