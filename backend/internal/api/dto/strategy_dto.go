package dto

import (
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

// StrategyResponseDto
type StrategyResponseDto struct {
	ID              int64       `json:"id"`
	Name            string      `json:"name"`
	Description     string      `json:"description,omitempty"`
	Config          interface{} `json:"config"`
	UserID          int64       `json:"user_id"`
	IsPublic        bool        `json:"is_public"`
	VoteCount       int         `json:"vote_count"`
	WinCount        int         `json:"win_count"`
	LastWinTime     *string     `json:"last_win_time,omitempty"`
	Tags            []string    `json:"tags,omitempty"`
	ComplexityScore int         `json:"complexity_score"`
	RiskScore       int         `json:"risk_score"`
	AIEnhanced      bool        `json:"ai_enhanced"`
}

func NewStrategyResponseDto(strategy *models.Strategy) StrategyResponseDto {
	var lastWinTimePtr *string
	if !strategy.LastWinTime.IsZero() {
		lastWinTime := strategy.LastWinTime.Format(time.RFC3339)
		lastWinTimePtr = &lastWinTime
	}
	return StrategyResponseDto{
		ID:              strategy.ID,
		Name:            strategy.Name,
		Description:     strategy.Description,
		Config:          strategy.Config,
		UserID:          strategy.UserID,
		IsPublic:        strategy.IsPublic,
		VoteCount:       strategy.VoteCount,
		WinCount:        strategy.WinCount,
		LastWinTime:     lastWinTimePtr,
		Tags:            strategy.Tags,
		ComplexityScore: strategy.ComplexityScore,
		RiskScore:       strategy.RiskScore,
		AIEnhanced:      strategy.AIEnhanced,
	}
}

func NewStrategyResponseDtoList(strategies []models.Strategy) []StrategyResponseDto {
	strategyDTOs := make([]StrategyResponseDto, len(strategies))
	for i, strategy := range strategies {
		strategyDTOs[i] = NewStrategyResponseDto(&strategy)
	}
	return strategyDTOs
}

func (dto StrategyCreateDto) ToModel(userID int64) *models.Strategy {
	return &models.Strategy{
		Name:        dto.Name,
		Description: dto.Description,
		Config: models.JSONB{
			"rules":      dto.Config.Rules,
			"risk_level": dto.Config.RiskLevel,
		},
		UserID:     userID,
		IsPublic:   dto.IsPublic,
		Tags:       dto.Tags,
		AIEnhanced: dto.AIEnhanced,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
