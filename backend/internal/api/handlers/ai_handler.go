// internal/api/handlers/ai_handler.go
package handlers

import (
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/api/dto"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/service"
	"github.com/gofiber/fiber/v2"
)

// AIHandler handles AI-related requests
type AIHandler struct {
	aiService           *service.AIService
	performanceAnalyzer *service.AIPerformanceAnalyzer
	logger              *logger.Logger
}

// NewAIHandler creates a new AI handler
func NewAIHandler(
	aiService *service.AIService,
	performanceAnalyzer *service.AIPerformanceAnalyzer,
	logger *logger.Logger,
) *AIHandler {
	return &AIHandler{
		aiService:           aiService,
		performanceAnalyzer: performanceAnalyzer,
		logger:              logger,
	}
}

// GetAIAnalysis returns AI analysis for a strategy
func (h *AIHandler) GetAIAnalysis(c *fiber.Ctx) error {
	h.logger.Info("Getting AI analysis for strategy")

	// Get strategy ID from params
	strategyID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	// Call the performance analyzer to get the analysis
	report, err := h.performanceAnalyzer.AnalyzeStrategyPerformance(int64(strategyID))
	if err != nil {
		h.logger.Error("Error generating AI analysis: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error generating AI analysis: %v", err),
		})
	}

	// Create the response
	response := fiber.Map{
		"strategy_id":   report.StrategyID,
		"strategy_name": report.StrategyName,
		"analysis":      report.Analysis,
		"rating":        report.Rating,
		"metrics": fiber.Map{
			"roi":              report.ROI,
			"win_rate":         report.WinRate,
			"total_trades":     report.TotalTrades,
			"max_drawdown":     report.MaxDrawdown,
			"net_pnl":          report.NetPnL,
			"avg_trade_profit": report.AvgTradeProfit,
		},
		"generated_at": report.GeneratedAt,
	}

	// Create a WebSocket event for real-time updates
	aiEvent := dto.AIAnalysisEvent{
		BaseEventDTO: dto.BaseEventDTO{
			Type:       "ai_analysis",
			StrategyID: report.StrategyID,
			Timestamp:  time.Now().Unix(),
		},
		StrategyName:   report.StrategyName,
		Analysis:       report.Analysis,
		Rating:         report.Rating,
		ROI:            report.ROI,
		WinRate:        report.WinRate,
		TotalTrades:    report.TotalTrades,
		MaxDrawdown:    report.MaxDrawdown,
		NetPnL:         report.NetPnL,
		AvgTradeProfit: report.AvgTradeProfit,
	}

	// Save the analysis to the database
	err = h.performanceAnalyzer.SaveAnalysisReport(report, &aiEvent)
	if err != nil {
		h.logger.Error("Error saving analysis report: %v", err)
		// Continue even if save fails - we still want to return the analysis
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// RegisterRoutes registers all AI routes
func (h *AIHandler) RegisterRoutes(app fiber.Router) {
	ai := app.Group("/ai")
	ai.Get("/analysis/:id", h.GetAIAnalysis)
}