package dto

import (
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
)

type StrategyConfig struct {
	Rules     []StrategyRule `json:"rules" validate:"required,dive,required"`
	RiskLevel string         `json:"risk_level,omitempty"` // "low", "medium", "high"
}

type StrategyRule struct {
	Condition string `json:"condition" validate:"required"`
	Action    string `json:"action" validate:"required"`
	Priority  int    `json:"priority,omitempty"`
}

type StrategyCreateDto struct {
	Name        string         `json:"name" validate:"required,min=3,max=100"`
	Description string         `json:"description" validate:"omitempty,max=1000"`
	Config      StrategyConfig `json:"config" validate:"required"`
	IsPublic    bool           `json:"is_public"`
	Tags        []string       `json:"tags,omitempty" validate:"omitempty,dive,min=1,max=30"`
	AIEnhanced  bool           `json:"ai_enhanced"`
}

// StrategyMetricsDto represents metrics about a strategy's performance
type StrategyMetricsDto struct {
	TotalTrades      int     `json:"totalTrades"`
	WinningTrades    int     `json:"winningTrades"`
	LosingTrades     int     `json:"losingTrades"`
	WinRate          float64 `json:"winRate"`
	AverageProfitPct float64 `json:"averageProfitPct"`
	AverageLossPct   float64 `json:"averageLossPct"`
	LargestWinPct    float64 `json:"largestWinPct"`
	LargestLossPct   float64 `json:"largestLossPct"`
	Balance          float64 `json:"balance"`
	InitialBalance   float64 `json:"initialBalance"`
	ROI              float64 `json:"roi"`
	ProfitFactor     float64 `json:"profitFactor"`
	SharpeRatio      float64 `json:"sharpeRatio,omitempty"`
}

// StrategyResponseDto
type StrategyResponseDto struct {
	ID              int64            `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description,omitempty"`
	Config          interface{}      `json:"config"`
	UserID          int64            `json:"user_id"`
	IsPublic        bool             `json:"is_public"`
	VoteCount       int              `json:"vote_count"`
	WinCount        int              `json:"win_count"`
	LastWinTime     *string          `json:"last_win_time,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	ComplexityScore int              `json:"complexity_score"`
	RiskScore       int              `json:"risk_score"`
	AIEnhanced      bool             `json:"ai_enhanced"`
	Metrics         *StrategyMetricsDto `json:"metrics,omitempty"`
}

func NewStrategyResponseDto(strategy *models.Strategy) StrategyResponseDto {
	var lastWinTimePtr *string
	if !strategy.LastWinTime.IsZero() {
		lastWinTime := strategy.LastWinTime.Format(time.RFC3339)
		lastWinTimePtr = &lastWinTime
	}

	// Create response DTO
	dto := StrategyResponseDto{
		ID:              strategy.ID,
		Name:            strategy.Name,
		Description:     strategy.Description,
		Config:          strategy.Config,
		IsPublic:        strategy.IsPublic,
		VoteCount:       strategy.VoteCount,
		WinCount:        strategy.WinCount,
		LastWinTime:     lastWinTimePtr,
		Tags:            strategy.Tags,
		ComplexityScore: strategy.ComplexityScore,
		RiskScore:       strategy.RiskScore,
		AIEnhanced:      strategy.AIEnhanced,
	}

	// Check if there are any metrics for this strategy in the config
	if metricsData, ok := strategy.Config["metrics"].(map[string]interface{}); ok {
		metrics := &StrategyMetricsDto{}
		
		// Extract metrics data
		if totalTrades, ok := metricsData["total_trades"].(float64); ok {
			metrics.TotalTrades = int(totalTrades)
		}
		if winningTrades, ok := metricsData["winning_trades"].(float64); ok {
			metrics.WinningTrades = int(winningTrades)
		}
		if losingTrades, ok := metricsData["losing_trades"].(float64); ok {
			metrics.LosingTrades = int(losingTrades)
		}
		if winRate, ok := metricsData["win_rate"].(float64); ok {
			metrics.WinRate = winRate
		}
		if avgProfit, ok := metricsData["avg_profit_pct"].(float64); ok {
			metrics.AverageProfitPct = avgProfit
		}
		if avgLoss, ok := metricsData["avg_loss_pct"].(float64); ok {
			metrics.AverageLossPct = avgLoss
		}
		if largestWin, ok := metricsData["largest_win_pct"].(float64); ok {
			metrics.LargestWinPct = largestWin
		}
		if largestLoss, ok := metricsData["largest_loss_pct"].(float64); ok {
			metrics.LargestLossPct = largestLoss
		}
		if balance, ok := metricsData["balance"].(float64); ok {
			metrics.Balance = balance
		}
		if initialBalance, ok := metricsData["initial_balance"].(float64); ok {
			metrics.InitialBalance = initialBalance
		}
		if roi, ok := metricsData["roi"].(float64); ok {
			metrics.ROI = roi
		}
		if profitFactor, ok := metricsData["profit_factor"].(float64); ok {
			metrics.ProfitFactor = profitFactor
		}
		if sharpeRatio, ok := metricsData["sharpe_ratio"].(float64); ok {
			metrics.SharpeRatio = sharpeRatio
		}

		dto.Metrics = metrics
	}

	return dto
}

func NewStrategyResponseDtoList(strategies []models.Strategy) []StrategyResponseDto {
	strategyDTOs := make([]StrategyResponseDto, len(strategies))
	for i, strategy := range strategies {
		strategyDTOs[i] = NewStrategyResponseDto(&strategy)
	}
	return strategyDTOs
}

func (dto StrategyCreateDto) ToModel(userID int64) *models.Strategy {
	// Create a deep copy of rules as a slice of maps
	rules := make([]map[string]interface{}, len(dto.Config.Rules))
	for i, rule := range dto.Config.Rules {
		rules[i] = map[string]interface{}{
			"condition": rule.Condition,
			"action":    rule.Action,
			"priority":  rule.Priority,
		}
	}

	// Create the strategy with the converted rules
	strategy := &models.Strategy{
		Name:        dto.Name,
		Description: dto.Description,
		Config: models.JSONB{
			"rules":      rules, // Use the converted rules slice
			"risk_level": dto.Config.RiskLevel,
		},
		IsPublic:   dto.IsPublic,
		Tags:       dto.Tags,
		AIEnhanced: dto.AIEnhanced,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Add extra debug log to console
	fmt.Printf("Config structure in DTO.ToModel: %+v\n", strategy.Config)

	return strategy
}
