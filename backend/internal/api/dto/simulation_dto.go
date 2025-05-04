// internal/models/dto/simulation_dto.go
package dto

// SimulationStatusDTO represents the current status and metrics of a simulation
type SimulationStatusDTO struct {
	StrategyID       int64         `json:"strategyId"`
	StrategyName     string        `json:"strategyName"`
	IsRunning        bool          `json:"isRunning"`
	StartTime        int64         `json:"startTime"`
	ExecutionTimeSec float64       `json:"executionTimeSec"`
	TotalTrades      int           `json:"totalTrades"`
	ActiveTrades     int           `json:"activeTrades"`
	ProfitableTrades int           `json:"profitableTrades"`
	LosingTrades     int           `json:"losingTrades"`
	WinRate          float64       `json:"winRate"`
	TotalProfit      float64       `json:"totalProfit"`
	TotalLoss        float64       `json:"totalLoss"`
	AvgProfit        float64       `json:"avgProfit"`
	AvgLoss          float64       `json:"avgLoss"`
	MaxDrawdown      float64       `json:"maxDrawdown"`
	NetPnL           float64       `json:"netPnl"`
	InitialBalance   float64       `json:"initialBalance"`
	CurrentBalance   float64       `json:"currentBalance"`
	ROI              float64       `json:"roi"`
	SimConfig        *SimConfigDTO `json:"simConfig,omitempty"`
}

// SimConfigDTO represents the configuration parameters of a strategy simulation
type SimConfigDTO struct {
	InitialBalance       float64 `json:"initialBalance"`
	FixedPositionSizeSol float64 `json:"fixedPositionSizeSol"`
	MarketCapThreshold   float64 `json:"marketCapThreshold"`
	TakeProfitPct        float64 `json:"takeProfitCct"`
	StopLossPct          float64 `json:"stopLossPct"`
	MaxHoldTimeSec       int     `json:"maxHoldTimeCec"`
	EntryTimeWindowSec   int     `json:"entryTimeWindow_sec"`
	MinBuysForEntry      int     `json:"minBuysForEntry"`
}
