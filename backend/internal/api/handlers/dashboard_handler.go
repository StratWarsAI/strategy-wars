package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/StratWarsAI/strategy-wars/internal/api/dto"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/service"
)

// DashboardHandler handles dashboard endpoints
type DashboardHandler struct {
	service service.DashboardServiceInterface
	logger  *logger.Logger
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(
	dashboardService service.DashboardServiceInterface,
	logger *logger.Logger,
) *DashboardHandler {
	return &DashboardHandler{
		service: dashboardService,
		logger:  logger,
	}
}

// GetDashboardStats retrieves basic dashboard statistics
func (h *DashboardHandler) GetDashboardStats(c *fiber.Ctx) error {
	// Get timeframe parameter, default to "24h"
	timeframe := c.Query("timeframe", "24h")

	h.logger.Info("Getting dashboard stats with timeframe: %s", timeframe)

	// Get dashboard stats from service
	dashboard, err := h.service.GetDashboardStats(timeframe)
	if err != nil {
		h.logger.Error("Error getting dashboard stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve dashboard statistics",
		})
	}

	// Convert to response DTO
	response := dto.DashboardStatsResponseDto{
		TotalBalance:         dashboard.TotalBalance,
		BalanceChange:        dashboard.BalanceChange,
		BalanceChangePercent: dashboard.BalanceChangePercent,
		TotalProfits:         dashboard.TotalProfits,
		TotalTrades:          dashboard.TotalTrades,
		WinningTrades:        dashboard.WinningTrades,
		LosingTrades:         dashboard.LosingTrades,
		WinRate:              dashboard.WinRate,
		ActiveTradeCount:     dashboard.ActiveTradeCount,
		AvgHoldTime:          dashboard.AvgHoldTime,
		TopPerformer: struct {
			ID    int64   `json:"id"`
			Name  string  `json:"name"`
			ROI   float64 `json:"roi"`
			Trades int    `json:"trades"`
		}{
			ID:    dashboard.TopPerformer.ID,
			Name:  dashboard.TopPerformer.Name,
			ROI:   dashboard.TopPerformer.ROI,
			Trades: dashboard.TopPerformer.Trades,
		},
		MarketStatus:    dashboard.MarketStatus,
		VolatilityIndex: dashboard.VolatilityIndex,
		LastUpdated:     dashboard.LastUpdated,
	}

	return c.JSON(response)
}

// GetDashboardCharts retrieves dashboard chart data
func (h *DashboardHandler) GetDashboardCharts(c *fiber.Ctx) error {
	// Get timeframe parameter, default to "7d"
	timeframe := c.Query("timeframe", "7d")

	h.logger.Info("Getting dashboard charts with timeframe: %s", timeframe)

	// Get dashboard charts from service
	dashboard, err := h.service.GetDashboardCharts(timeframe)
	if err != nil {
		h.logger.Error("Error getting dashboard charts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve dashboard chart data",
		})
	}

	// Convert performance data points
	performanceData := make([]dto.PerformanceDataPoint, len(dashboard.PerformanceData))
	for i, dp := range dashboard.PerformanceData {
		performanceData[i] = dto.PerformanceDataPoint{
			Date:    dp.Date,
			Balance: dp.Balance,
		}
	}

	// Convert strategy distribution data
	strategyData := make([]dto.StrategyDistributionItem, len(dashboard.StrategyData))
	for i, s := range dashboard.StrategyData {
		strategyData[i] = dto.StrategyDistributionItem{
			ID:     s.ID,
			Name:   s.Name,
			Trades: s.Trades,
			Profit: s.Profit,
			Color:  s.Color,
		}
	}

	// Convert recent stats
	recentStats := make([]dto.RecentPerformanceStats, len(dashboard.RecentStats))
	for i, s := range dashboard.RecentStats {
		recentStats[i] = dto.RecentPerformanceStats{
			Period:    s.Period,
			Trades:    s.Trades,
			Profit:    s.Profit,
			WinRate:   s.WinRate,
			BestTrade: s.BestTrade,
		}
	}

	// Create response DTO
	response := dto.DashboardChartsResponseDto{
		PerformanceData: performanceData,
		StrategyData:    strategyData,
		RecentStats:     recentStats,
	}

	return c.JSON(response)
}

// GetCompleteDashboard retrieves all dashboard data
func (h *DashboardHandler) GetCompleteDashboard(c *fiber.Ctx) error {
	// Get timeframe parameter, default to "24h"
	timeframe := c.Query("timeframe", "24h")

	h.logger.Info("Getting complete dashboard with timeframe: %s", timeframe)

	// Get complete dashboard from service
	dashboard, err := h.service.GetCompleteDashboard(timeframe)
	if err != nil {
		h.logger.Error("Error getting complete dashboard: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve dashboard data",
		})
	}

	// Convert performance data points
	performanceData := make([]dto.PerformanceDataPoint, len(dashboard.PerformanceData))
	for i, dp := range dashboard.PerformanceData {
		performanceData[i] = dto.PerformanceDataPoint{
			Date:    dp.Date,
			Balance: dp.Balance,
		}
	}

	// Convert strategy distribution data
	strategyData := make([]dto.StrategyDistributionItem, len(dashboard.StrategyData))
	for i, s := range dashboard.StrategyData {
		strategyData[i] = dto.StrategyDistributionItem{
			ID:     s.ID,
			Name:   s.Name,
			Trades: s.Trades,
			Profit: s.Profit,
			Color:  s.Color,
		}
	}

	// Convert recent stats
	recentStats := make([]dto.RecentPerformanceStats, len(dashboard.RecentStats))
	for i, s := range dashboard.RecentStats {
		recentStats[i] = dto.RecentPerformanceStats{
			Period:    s.Period,
			Trades:    s.Trades,
			Profit:    s.Profit,
			WinRate:   s.WinRate,
			BestTrade: s.BestTrade,
		}
	}

	// Create response DTOs
	statsDto := dto.DashboardStatsResponseDto{
		TotalBalance:         dashboard.TotalBalance,
		BalanceChange:        dashboard.BalanceChange,
		BalanceChangePercent: dashboard.BalanceChangePercent,
		TotalProfits:         dashboard.TotalProfits,
		TotalTrades:          dashboard.TotalTrades,
		WinningTrades:        dashboard.WinningTrades,
		LosingTrades:         dashboard.LosingTrades,
		WinRate:              dashboard.WinRate,
		ActiveTradeCount:     dashboard.ActiveTradeCount,
		AvgHoldTime:          dashboard.AvgHoldTime,
		TopPerformer: struct {
			ID    int64   `json:"id"`
			Name  string  `json:"name"`
			ROI   float64 `json:"roi"`
			Trades int    `json:"trades"`
		}{
			ID:    dashboard.TopPerformer.ID,
			Name:  dashboard.TopPerformer.Name,
			ROI:   dashboard.TopPerformer.ROI,
			Trades: dashboard.TopPerformer.Trades,
		},
		MarketStatus:    dashboard.MarketStatus,
		VolatilityIndex: dashboard.VolatilityIndex,
		LastUpdated:     dashboard.LastUpdated,
	}

	chartsDto := dto.DashboardChartsResponseDto{
		PerformanceData: performanceData,
		StrategyData:    strategyData,
		RecentStats:     recentStats,
	}

	response := dto.CompleteDashboardResponseDto{
		Stats:  statsDto,
		Charts: chartsDto,
	}

	return c.JSON(response)
}

// RegisterRoutes registers the dashboard routes
func (h *DashboardHandler) RegisterRoutes(app fiber.Router) {
	route := app.Group("/dashboard")
	route.Get("/stats", h.GetDashboardStats)
	route.Get("/charts", h.GetDashboardCharts)
	route.Get("/complete", h.GetCompleteDashboard)
}