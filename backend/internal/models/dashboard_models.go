package models

// PerformanceDataPoint represents a data point for the balance history chart
type PerformanceDataPoint struct {
	Date    string  `json:"date"`
	Balance float64 `json:"balance"`
}

// StrategyDistribution represents a strategy in the profit distribution
type StrategyDistribution struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Trades int     `json:"trades"`
	Profit float64 `json:"profit"`
	Color  string  `json:"color"`
}

// RecentPerformance represents trading performance for a specific time period
type RecentPerformance struct {
	Period    string  `json:"period"`
	Trades    int     `json:"trades"`
	Profit    float64 `json:"profit"`
	WinRate   float64 `json:"win_rate"`
	BestTrade string  `json:"best_trade"`
}

// Dashboard represents all dashboard data
type Dashboard struct {
	TotalBalance         float64                `json:"total_balance"`
	BalanceChange        float64                `json:"balance_change"`
	BalanceChangePercent float64                `json:"balance_change_percent"`
	TotalProfits         float64                `json:"total_profits"`
	TotalTrades          int                    `json:"total_trades"`
	WinningTrades        int                    `json:"winning_trades"`
	LosingTrades         int                    `json:"losing_trades"`
	WinRate              float64                `json:"win_rate"`
	ActiveTradeCount     int                    `json:"active_trade_count"`
	AvgHoldTime          string                 `json:"avg_hold_time"`
	TopPerformer         TopPerformer           `json:"top_performer"`
	MarketStatus         string                 `json:"market_status"`
	VolatilityIndex      int                    `json:"volatility_index"`
	PerformanceData      []PerformanceDataPoint `json:"performance_data"`
	StrategyData         []StrategyDistribution `json:"strategy_data"`
	RecentStats          []RecentPerformance    `json:"recent_stats"`
	LastUpdated          string                 `json:"last_updated"`
}

// TopPerformer represents data about the best performing strategy
type TopPerformer struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	ROI   float64 `json:"roi"`
	Trades int    `json:"trades"`
}