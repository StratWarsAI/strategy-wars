// internal/service/service_interfaces.go
package service

import (
	"github.com/StratWarsAI/strategy-wars/internal/models"
)

// StrategyServiceInterface defines the interface for strategy service
type StrategyServiceInterface interface {
	CreateStrategy(strategy *models.Strategy) (int64, error)
	GetStrategyByID(id int64) (*models.Strategy, error)
	UpdateStrategy(strategy *models.Strategy) error
	DeleteStrategy(id int64, userID int64) error
	GetUserStrategies(userID int64, includePrivate bool, limit, offset int) ([]*models.Strategy, error)
	GetPublicStrategies(limit, offset int) ([]*models.Strategy, error)
	GetTopStrategies(criteria string, limit int) ([]*models.Strategy, error)
	SearchStrategiesByTags(tags []string, limit int) ([]*models.Strategy, error)
}
