package dto

// DashboardStatsResponseDto represents aggregated dashboard statistics
type DashboardStatsResponseDto struct {
	// Portfolio stats
	TotalBalance         float64 `json:"total_balance"`
	BalanceChange        float64 `json:"balance_change"`
	BalanceChangePercent float64 `json:"balance_change_percent"`
	TotalProfits         float64 `json:"total_profits"`

	// Trading stats
	TotalTrades      int     `json:"total_trades"`
	WinningTrades    int     `json:"winning_trades"`
	LosingTrades     int     `json:"losing_trades"`
	WinRate          float64 `json:"win_rate"`
	ActiveTradeCount int     `json:"active_trade_count"`
	AvgHoldTime      string  `json:"avg_hold_time"`

	// Top strategy
	TopPerformer struct {
		ID     int64   `json:"id"`
		Name   string  `json:"name"`
		ROI    float64 `json:"roi"`
		Trades int     `json:"trades"`
	} `json:"top_performer"`

	// Market conditions
	MarketStatus    string `json:"market_status"`
	VolatilityIndex int    `json:"volatility_index"`
	LastUpdated     string `json:"last_updated"`
}

// PerformanceDataPoint represents a data point for the balance history chart
type PerformanceDataPoint struct {
	Date    string  `json:"date"`
	Balance float64 `json:"balance"`
}

// StrategyDistributionItem represents a strategy in the distribution chart
type StrategyDistributionItem struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Trades int     `json:"trades"`
	Profit float64 `json:"profit"`
	Color  string  `json:"color"`
}

// RecentPerformanceStats represents performance stats for a time period
type RecentPerformanceStats struct {
	Period    string  `json:"period"`
	Trades    int     `json:"trades"`
	Profit    float64 `json:"profit"`
	WinRate   float64 `json:"win_rate"`
	BestTrade string  `json:"best_trade"`
}

// DashboardChartsResponseDto represents dashboard chart data
type DashboardChartsResponseDto struct {
	PerformanceData []PerformanceDataPoint     `json:"performance_data"`
	StrategyData    []StrategyDistributionItem `json:"strategy_data"`
	RecentStats     []RecentPerformanceStats   `json:"recent_stats"`
}

// CompleteDashboardResponseDto represents the complete dashboard response
type CompleteDashboardResponseDto struct {
	Stats  DashboardStatsResponseDto  `json:"stats"`
	Charts DashboardChartsResponseDto `json:"charts"`
}
