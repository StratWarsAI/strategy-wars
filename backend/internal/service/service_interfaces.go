// internal/service/service_interfaces.go
package service

import (
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
)

// StrategyServiceInterface defines the interface for strategy service
type StrategyServiceInterface interface {
	CreateStrategy(strategy *models.Strategy) (int64, error)
	GetStrategyByID(id int64) (*models.Strategy, error)
	UpdateStrategy(strategy *models.Strategy) error
	DeleteStrategy(id int64) error
	GetPublicStrategies(limit, offset int) ([]*models.Strategy, error)
	GetTopStrategies(criteria string, limit int) ([]*models.Strategy, error)
	SearchStrategiesByTags(tags []string, limit int) ([]*models.Strategy, error)
	RecordWin(strategyID int64, simulationID int64, winTime time.Time) error
}
