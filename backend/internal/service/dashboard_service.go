package service

import (
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
)

// DashboardService handles dashboard business logic
type DashboardService struct {
	dashboardRepo      repository.DashboardRepositoryInterface
	simulatedTradeRepo repository.SimulatedTradeRepositoryInterface
	logger             *logger.Logger
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	dashboardRepo repository.DashboardRepositoryInterface,
	simulatedTradeRepo repository.SimulatedTradeRepositoryInterface,
	logger *logger.Logger,
) *DashboardService {
	return &DashboardService{
		dashboardRepo:      dashboardRepo,
		simulatedTradeRepo: simulatedTradeRepo,
		logger:             logger,
	}
}

// GetDashboardStats retrieves basic dashboard statistics
func (s *DashboardService) GetDashboardStats(timeframe string) (*models.Dashboard, error) {
	s.logger.Info("Getting dashboard stats for timeframe: %s", timeframe)

	dashboard := &models.Dashboard{
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	// Get total balance
	totalBalance, err := s.dashboardRepo.GetTotalBalance()
	if err != nil {
		s.logger.Error("Error getting total balance: %v", err)
		totalBalance = 0
	}
	dashboard.TotalBalance = totalBalance

	// Get balance change
	balanceChange, balanceChangePercent, err := s.dashboardRepo.GetBalanceChange(timeframe)
	if err != nil {
		s.logger.Error("Error getting balance change: %v", err)
		balanceChange, balanceChangePercent = 0, 0
	}
	dashboard.BalanceChange = balanceChange
	dashboard.BalanceChangePercent = balanceChangePercent

	// Get trading stats
	totalTrades, winningTrades, losingTrades, winRate, err := s.dashboardRepo.GetTradingStats()
	if err != nil {
		s.logger.Error("Error getting trading stats: %v", err)
		totalTrades, winningTrades, losingTrades, winRate = 0, 0, 0, 0
	}
	dashboard.TotalTrades = totalTrades
	dashboard.WinningTrades = winningTrades
	dashboard.LosingTrades = losingTrades
	dashboard.WinRate = winRate

	// Set total profits (assuming this is the sum of all profitable trades)
	dashboard.TotalProfits = balanceChange
	if dashboard.TotalProfits < 0 {
		dashboard.TotalProfits = 0
	}

	// Get active trade count
	activeTradeCount, err := s.dashboardRepo.GetActiveTradeCount()
	if err != nil {
		s.logger.Error("Error getting active trade count: %v", err)
		activeTradeCount = 0
	}
	dashboard.ActiveTradeCount = activeTradeCount

	// Get average hold time
	avgHoldTime, err := s.dashboardRepo.GetAverageHoldTime()
	if err != nil {
		s.logger.Error("Error getting average hold time: %v", err)
		avgHoldTime = "0m 0s"
	}
	dashboard.AvgHoldTime = avgHoldTime

	// Get top performing strategy
	strategy, roi, tradeCount, err := s.dashboardRepo.GetTopPerformingStrategy()
	if err != nil {
		s.logger.Error("Error getting top performing strategy: %v", err)
	}

	if strategy != nil {
		dashboard.TopPerformer = models.TopPerformer{
			ID:     strategy.ID,
			Name:   strategy.Name,
			ROI:    roi,
			Trades: tradeCount,
		}
	} else {
		dashboard.TopPerformer = models.TopPerformer{
			ID:     0,
			Name:   "No strategies yet",
			ROI:    0,
			Trades: 0,
		}
	}

	// Get market conditions
	marketStatus, volatilityIndex, err := s.dashboardRepo.GetMarketConditions()
	if err != nil {
		s.logger.Error("Error getting market conditions: %v", err)
		marketStatus, volatilityIndex = "neutral", 50
	}
	dashboard.MarketStatus = marketStatus
	dashboard.VolatilityIndex = volatilityIndex

	return dashboard, nil
}

// GetDashboardCharts retrieves chart data for the dashboard
func (s *DashboardService) GetDashboardCharts(timeframe string) (*models.Dashboard, error) {
	s.logger.Info("Getting dashboard charts for timeframe: %s", timeframe)

	dashboard := &models.Dashboard{
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	// Determine days based on timeframe
	var days int
	switch timeframe {
	case "24h":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	default:
		days = 7 // Default to one week
	}

	// Get performance history
	performanceData, err := s.dashboardRepo.GetPerformanceHistory(days)
	if err != nil {
		s.logger.Error("Error getting performance history: %v", err)
		performanceData = []models.PerformanceDataPoint{}
	}
	dashboard.PerformanceData = performanceData

	// Get strategy distribution
	strategyData, err := s.dashboardRepo.GetStrategyDistribution(5)
	if err != nil {
		s.logger.Error("Error getting strategy distribution: %v", err)
		strategyData = []models.StrategyDistribution{}
	}
	dashboard.StrategyData = strategyData

	// Get recent performance stats
	recentStats, err := s.dashboardRepo.GetRecentPerformance()
	if err != nil {
		s.logger.Error("Error getting recent performance: %v", err)
		recentStats = []models.RecentPerformance{}
	}
	dashboard.RecentStats = recentStats

	return dashboard, nil
}

// GetCompleteDashboard retrieves all dashboard data
func (s *DashboardService) GetCompleteDashboard(timeframe string) (*models.Dashboard, error) {
	s.logger.Info("Getting complete dashboard for timeframe: %s", timeframe)

	// Get basic stats
	dashboard, err := s.GetDashboardStats(timeframe)
	if err != nil {
		return nil, err
	}

	// Get chart data
	chartDashboard, err := s.GetDashboardCharts(timeframe)
	if err != nil {
		s.logger.Error("Error getting dashboard charts: %v", err)
	} else {
		dashboard.PerformanceData = chartDashboard.PerformanceData
		dashboard.StrategyData = chartDashboard.StrategyData
		dashboard.RecentStats = chartDashboard.RecentStats
	}

	return dashboard, nil
}
