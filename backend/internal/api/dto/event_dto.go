// dto/event_dto.go
package dto

// BaseEventDTO is the base structure for all WebSocket events
type BaseEventDTO struct {
	Type       string `json:"type"`
	StrategyID int64  `json:"strategy_id"`
	Timestamp  int64  `json:"timestamp"`
}

// SimulationStartedEvent represents a simulation start event
type SimulationStartedEvent struct {
	BaseEventDTO
}

// SimulationCompletedEvent represents a simulation completion event
type SimulationCompletedEvent struct {
	BaseEventDTO
	TotalIterations  int     `json:"totalIterations"`
	ExecutionTimeSec float64 `json:"executionTimeSec"`
}

// SimulationStatusEvent represents current simulation status
type SimulationStatusEvent struct {
	BaseEventDTO
	TotalTrades      int     `json:"totalTrades"`
	ActiveTrades     int     `json:"activeTrades"`
	ProfitableTrades int     `json:"profitableTrades"`
	LosingTrades     int     `json:"losingTrades"`
	WinRate          float64 `json:"winRate"`
	ROI              float64 `json:"roi"`
	CurrentBalance   float64 `json:"currentBalance"`
	InitialBalance   float64 `json:"initialBalance"`
}

// SimulationBalanceDepletedEvent represents when balance is too low for trades
type SimulationBalanceDepletedEvent struct {
	BaseEventDTO
	RemainingBalance float64 `json:"remainingBalance"`
	PositionSize     float64 `json:"positionSize"`
}

// TradeExecutedEvent represents a buy order event
type TradeExecutedEvent struct {
	BaseEventDTO
	TokenID        int64                  `json:"tokenId"`
	TokenSymbol    string                 `json:"tokenSymbol"`
	TokenName      string                 `json:"tokenName"`
	TokenMint      string                 `json:"tokenMint"`
	ImageUrl       string                 `json:"imageUrl"`
	TwitterUrl     string                 `json:"twitterUrl"`
	WebsiteUrl     string                 `json:"websiteUrl"`
	Action         string                 `json:"action"`
	Price          float64                `json:"price"`
	Amount         float64                `json:"amount"`
	EntryMarketCap float64                `json:"entryMarketCap"`
	UsdMarketCap   float64                `json:"usdMarketCap"`
	CurrentBalance float64                `json:"currentBalance"`
	SignalData     map[string]interface{} `json:"signalData,omitempty"`
}

// TradeClosedEvent represents a sell order event
type TradeClosedEvent struct {
	BaseEventDTO
	TokenID        int64   `json:"tokenId"`
	TokenSymbol    string  `json:"tokenSymbol"`
	TokenName      string  `json:"tokenName"`
	TokenMint      string  `json:"tokenMint"`
	ImageUrl       string  `json:"imageUrl"`
	TwitterUrl     string  `json:"twitterUrl"`
	WebsiteUrl     string  `json:"websiteUrl"`
	Action         string  `json:"action"`
	EntryPrice     float64 `json:"entryPrice"`
	ExitPrice      float64 `json:"exitPrice"`
	ProfitLoss     float64 `json:"profitLoss"`
	ProfitLossPct  float64 `json:"profitLossPct"`
	ExitReason     string  `json:"exitReason"`
	EntryMarketCap float64 `json:"entryMarketCap"`
	ExitMarketCap  float64 `json:"exitMarketCap"`
	UsdMarketCap   float64 `json:"usdMarketCap"`
}

// WebSocketMessage is a generic interface for all WebSocket messages
type WebSocketMessage interface {
	GetType() string
	GetStrategyID() int64
	GetTimestamp() int64
}

func (b BaseEventDTO) GetType() string {
	return b.Type
}

func (b BaseEventDTO) GetStrategyID() int64 {
	return b.StrategyID
}

func (b BaseEventDTO) GetTimestamp() int64 {
	return b.Timestamp
}
