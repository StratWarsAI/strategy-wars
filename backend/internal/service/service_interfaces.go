// internal/service/service_interfaces.go
package service

import (
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
)

// DashboardServiceInterface defines methods for dashboard service
type DashboardServiceInterface interface {
	GetDashboardStats(timeframe string) (*models.Dashboard, error)
	GetDashboardCharts(timeframe string) (*models.Dashboard, error)
	GetCompleteDashboard(timeframe string) (*models.Dashboard, error)
}

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
