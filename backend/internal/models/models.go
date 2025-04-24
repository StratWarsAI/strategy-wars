// internal/models/models.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID            int64     `json:"-"`
	Username      string    `json:"username"`
	Email         string    `json:"email,omitempty"`
	PasswordHash  string    `json:"-"`
	WalletAddress string    `json:"wallet_address,omitempty"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

// Strategy represents a trading strategy
type Strategy struct {
	ID              int64     `json:"-"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	Config          JSONB     `json:"config"`
	UserID          int64     `json:"-"`
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

// Duel represents a 10-minute trading battle
type Duel struct {
	ID               int64     `json:"-"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	VotingEndTime    time.Time `json:"voting_end_time"`
	WinnerStrategyID int64     `json:"-"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

// Vote represents a vote for a strategy in a duel
type Vote struct {
	ID         int64     `json:"-"`
	DuelID     int64     `json:"-"`
	StrategyID int64     `json:"-"`
	UserID     int64     `json:"-"`
	CreatedAt  time.Time `json:"-"`
}

// UserScore represents a user's performance metrics
type UserScore struct {
	UserID        int64     `json:"-"`
	TotalPoints   int       `json:"total_points"`
	WinCount      int       `json:"win_count"`
	StrategyCount int       `json:"strategy_count"`
	VoteCount     int       `json:"vote_count"`
	LastUpdated   time.Time `json:"last_updated"`
}

// Comment represents a comment on a strategy
type Comment struct {
	ID         int64     `json:"-"`
	StrategyID int64     `json:"-"`
	UserID     int64     `json:"-"`
	ParentID   *int64    `json:"parent_id,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

// Notification represents a system notification
type Notification struct {
	ID        int64     `json:"-"`
	UserID    int64     `json:"-"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	RelatedID *int64    `json:"related_id,omitempty"`
	CreatedAt time.Time `json:"-"`
}

// StrategyMetric represents AI analysis of a strategy
type StrategyMetric struct {
	ID               int64     `json:"-"`
	StrategyID       int64     `json:"-"`
	DuelID           *int64    `json:"-"`
	WinRate          float64   `json:"win_rate"`
	AvgProfit        float64   `json:"avg_profit"`
	AvgLoss          float64   `json:"avg_loss"`
	MaxDrawdown      float64   `json:"max_drawdown"`
	TotalTrades      int       `json:"total_trades"`
	SuccessfulTrades int       `json:"successful_trades"`
	RiskScore        int       `json:"risk_score"`
	CreatedAt        time.Time `json:"-"`
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
	DuelID            *int64    `json:"-"`
	EntryPrice        float64   `json:"entry_price"`
	ExitPrice         *float64  `json:"exit_price,omitempty"`
	EntryTimestamp    int64     `json:"entry_timestamp"`
	ExitTimestamp     *int64    `json:"exit_timestamp,omitempty"`
	PositionSize      float64   `json:"position_size"`
	ProfitLoss        *float64  `json:"profit_loss,omitempty"`
	Status            string    `json:"status"`
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
