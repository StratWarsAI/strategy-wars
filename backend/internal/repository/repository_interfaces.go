// internal/repository/repository_interfaces.go
package repository

import (
	"context"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
)

// TokenRepositoryInterface defines the interface for token repository operations
type TokenRepositoryInterface interface {
	Save(token *models.Token) (int64, error)
	GetByMintAddress(mintAddress string) (*models.Token, error)
	GetRecentTokens(limit int) ([]*models.Token, error)
	GetFilteredTokens(minMarketCapUSD float64, maxAgeSeconds int64, limit int) ([]*models.Token, error)
	GetByID(tokenID int64) (*models.Token, error)
}

// TradeRepositoryInterface defines the interface for trade repository operations
type TradeRepositoryInterface interface {
	Save(trade *models.Trade) (int64, error)
	GetTradesByTokenID(tokenID int64, limit int) ([]*models.Trade, error)
	GetTradesByTokenIDWithContext(ctx context.Context, tokenID int64, limit int) ([]*models.Trade, error)
	GetTradesBySignature(signature string) (*models.Trade, error)
}

// StrategyRepositoryInterface defines the interface for strategy repository operations
type StrategyRepositoryInterface interface {
	Save(strategy *models.Strategy) (int64, error)
	GetByID(id int64) (*models.Strategy, error)
	ListPublic(limit, offset int) ([]*models.Strategy, error)
	Update(strategy *models.Strategy) error
	Delete(id int64) error
	IncrementVoteCount(id int64) error
	IncrementWinCount(id int64, winTime time.Time) error
	SearchByTags(tags []string, limit int) ([]*models.Strategy, error)
	GetTopVoted(limit int) ([]*models.Strategy, error)
	GetTopWinners(limit int) ([]*models.Strategy, error)
}

// StrategyMetricRepositoryInterface defines the interface for strategy metric repository operations
type StrategyMetricRepositoryInterface interface {
	Save(metric *models.StrategyMetric) (int64, error)
	GetByID(id int64) (*models.StrategyMetric, error)
	GetByStrategy(strategyID int64) ([]*models.StrategyMetric, error)
	GetBySimulationRun(simulationRunID int64) ([]*models.StrategyMetric, error)
	GetLatestByStrategy(strategyID int64) (*models.StrategyMetric, error)
	GetLatestByStrategyAndSimulation(strategyID int64, simulationRunID *int64) (*models.StrategyMetric, error)
	UpdateLatestByStrategy(metric *models.StrategyMetric) error
}

// SimulationRunRepositoryInterface for managing simulation runs
type SimulationRunRepositoryInterface interface {
	Save(run *models.SimulationRun) (int64, error)
	GetByID(id int64) (*models.SimulationRun, error)
	GetCurrent() (*models.SimulationRun, error)
	GetByStatus(status string, limit int) ([]*models.SimulationRun, error)
	GetByTimeRange(start, end time.Time) ([]*models.SimulationRun, error)
	UpdateStatus(id int64, status string) error
	UpdateWinner(id int64, strategyID int64) error
}

// SimulationResultRepositoryInterface for managing simulation results
type SimulationResultRepositoryInterface interface {
	Save(result *models.SimulationResult) (int64, error)
	GetByID(id int64) (*models.SimulationResult, error)
	GetBySimulationRun(simulationRunID int64) ([]*models.SimulationResult, error)
	GetTopPerformers(simulationRunID int64, limit int) ([]*models.SimulationResult, error)
	GetByStrategy(strategyID int64, limit int) ([]*models.SimulationResult, error)
}

// StrategyGenerationRepositoryInterface for managing strategy generations
type StrategyGenerationRepositoryInterface interface {
	Save(generation *models.StrategyGeneration) (int64, error)
	GetByID(id int64) (*models.StrategyGeneration, error)
	GetByParentStrategy(parentStrategyID int64) ([]*models.StrategyGeneration, error)
	GetByChildStrategy(childStrategyID int64) ([]*models.StrategyGeneration, error)
	GetByGenerationNumber(generationNumber int, limit, offset int) ([]*models.StrategyGeneration, error)
	GetLatestGeneration() (int, error)
}

// SimulatedTradeRepositoryInterface defines the interface for simulated trade repository operations
type SimulatedTradeRepositoryInterface interface {
	Save(trade *models.SimulatedTrade) (int64, error)
	SaveWithContext(ctx context.Context, trade *models.SimulatedTrade) (int64, error)
	Update(trade *models.SimulatedTrade) error
	UpdateWithContext(ctx context.Context, trade *models.SimulatedTrade) error
	GetByStrategyID(strategyID int64) ([]*models.SimulatedTrade, error)
	GetByStrategyIDWithContext(ctx context.Context, strategyID int64) ([]*models.SimulatedTrade, error)
	GetActiveByStrategyID(strategyID int64) ([]*models.SimulatedTrade, error)
	GetActiveByStrategyIDWithContext(ctx context.Context, strategyID int64) ([]*models.SimulatedTrade, error)
	GetSummaryByStrategyID(strategyID int64) (map[string]interface{}, error)
	GetSummaryByStrategyIDWithContext(ctx context.Context, strategyID int64) (map[string]interface{}, error)
	DeleteByStrategyID(strategyID int64) error
	DeleteByStrategyIDWithContext(ctx context.Context, strategyID int64) error
	GetTradesByTokenID(tokenID int64, limit int) ([]*models.SimulatedTrade, error)
	GetTradesByTokenIDWithContext(ctx context.Context, tokenID int64, limit int) ([]*models.SimulatedTrade, error)
	GetBySimulationRun(simulationRunID int64) ([]*models.SimulatedTrade, error)
	GetBySimulationRunWithContext(ctx context.Context, simulationRunID int64) ([]*models.SimulatedTrade, error)
	ExistsByStrategyIDAndTokenID(strategyID int64, tokenID int64) (bool, error)
	ExistsByStrategyIDAndTokenIDWithContext(ctx context.Context, strategyID int64, tokenID int64) (bool, error)
}

// SimulationEventRepositoryInterface for managing simulation events
type SimulationEventRepositoryInterface interface {
	Save(event *models.SimulationEvent) (int64, error)
	GetByID(id int64) (*models.SimulationEvent, error)
	GetByStrategyID(strategyID int64, limit, offset int) ([]*models.SimulationEvent, error)
	GetBySimulationRunID(simulationRunID int64, limit, offset int) ([]*models.SimulationEvent, error)
	GetLatestByStrategyID(strategyID int64, limit int) ([]*models.SimulationEvent, error)
}
