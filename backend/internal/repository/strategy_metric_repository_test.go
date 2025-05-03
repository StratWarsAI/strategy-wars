// internal/repository/strategy_metric_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestStrategyMetricRepositorySave(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %v", err)
		}
	}()

	// Create test metric
	now := time.Now()
	simulationRunID := int64(5)
	metric := &models.StrategyMetric{
		StrategyID:       1,
		SimulationRunID:  &simulationRunID,
		WinRate:          0.75,
		AvgProfit:        100.50,
		AvgLoss:          50.25,
		MaxDrawdown:      200.0,
		TotalTrades:      20,
		SuccessfulTrades: 15,
		RiskScore:        1,
		CreatedAt:        now,
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO strategy_metrics`).
		WithArgs(
			metric.StrategyID,
			sqlmock.AnyArg(), // simulation_run_id (NullInt64)
			metric.WinRate,
			metric.AvgProfit,
			metric.AvgLoss,
			metric.MaxDrawdown,
			metric.TotalTrades,
			metric.SuccessfulTrades,
			metric.RiskScore,
			metric.CreatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewStrategyMetricRepository(db)

	// Execute test
	id, err := repo.Save(metric)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyMetricRepositoryGetByID(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %v", err)
		}
	}()

	// Test data
	metricID := int64(1)
	now := time.Now()
	simulationRunID := int64(5)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "strategy_id", "simulation_run_id", "win_rate", "avg_profit", "avg_loss",
		"max_drawdown", "total_trades", "successful_trades", "risk_score", "created_at",
	}).
		AddRow(metricID, 1, simulationRunID, 0.75, 100.50, 50.25, 200.0, 20, 15, 1, now)

	mock.ExpectQuery(`SELECT (.+) FROM strategy_metrics WHERE id = \$1`).
		WithArgs(metricID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyMetricRepository(db)

	// Execute test
	metric, err := repo.GetByID(metricID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, metricID, metric.ID)
	assert.Equal(t, int64(1), metric.StrategyID)
	assert.Equal(t, &simulationRunID, metric.SimulationRunID)
	assert.Equal(t, float64(0.75), metric.WinRate)
	assert.Equal(t, float64(100.50), metric.AvgProfit)
	assert.Equal(t, float64(50.25), metric.AvgLoss)
	assert.Equal(t, float64(200.0), metric.MaxDrawdown)
	assert.Equal(t, 20, metric.TotalTrades)
	assert.Equal(t, 15, metric.SuccessfulTrades)
	assert.Equal(t, 1, metric.RiskScore)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyMetricRepositoryGetByIDNotFound(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %v", err)
		}
	}()

	// Test data
	metricID := int64(999)

	// Setup expected query with no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM strategy_metrics WHERE id = \$1`).
		WithArgs(metricID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "strategy_id", "simulation_run_id", "win_rate", "avg_profit", "avg_loss",
			"max_drawdown", "total_trades", "successful_trades", "risk_score", "created_at",
		}))

	// Create repository with mock DB
	repo := NewStrategyMetricRepository(db)

	// Execute test
	metric, err := repo.GetByID(metricID)

	// Assert results
	assert.NoError(t, err)
	assert.Nil(t, metric)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyMetricRepositoryGetByStrategy(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %v", err)
		}
	}()

	// Test data
	strategyID := int64(1)
	now := time.Now()
	simulationRunID1 := int64(5)
	simulationRunID2 := int64(6)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "strategy_id", "simulation_run_id", "win_rate", "avg_profit", "avg_loss",
		"max_drawdown", "total_trades", "successful_trades", "risk_score", "created_at",
	}).
		AddRow(1, strategyID, simulationRunID1, 0.75, 100.50, 50.25, 200.0, 20, 15, 1, now).
		AddRow(2, strategyID, simulationRunID2, 0.65, 90.25, 55.75, 180.0, 18, 12, 2, now.Add(1*time.Minute))

	mock.ExpectQuery(`SELECT (.+) FROM strategy_metrics WHERE strategy_id = \$1 ORDER BY created_at DESC`).
		WithArgs(strategyID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyMetricRepository(db)

	// Execute test
	metrics, err := repo.GetByStrategy(strategyID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 2, len(metrics))
	assert.Equal(t, int64(1), metrics[0].ID)
	assert.Equal(t, int64(2), metrics[1].ID)
	assert.Equal(t, &simulationRunID1, metrics[0].SimulationRunID)
	assert.Equal(t, &simulationRunID2, metrics[1].SimulationRunID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyMetricRepositoryGetBySimulationRun(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %v", err)
		}
	}()

	// Test data
	simulationRunID := int64(5)
	now := time.Now()
	strategyID1 := int64(1)
	strategyID2 := int64(2)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "strategy_id", "simulation_run_id", "win_rate", "avg_profit", "avg_loss",
		"max_drawdown", "total_trades", "successful_trades", "risk_score", "created_at",
	}).
		AddRow(1, strategyID1, simulationRunID, 0.75, 100.50, 50.25, 200.0, 20, 15, 1, now).
		AddRow(2, strategyID2, simulationRunID, 0.65, 90.25, 55.75, 180.0, 18, 12, 2, now.Add(1*time.Minute))

	mock.ExpectQuery(`SELECT (.+) FROM strategy_metrics WHERE simulation_run_id = \$1 ORDER BY created_at DESC`).
		WithArgs(simulationRunID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyMetricRepository(db)

	// Execute test
	metrics, err := repo.GetBySimulationRun(simulationRunID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 2, len(metrics))
	assert.Equal(t, int64(1), metrics[0].ID)
	assert.Equal(t, int64(2), metrics[1].ID)
	assert.Equal(t, strategyID1, metrics[0].StrategyID)
	assert.Equal(t, strategyID2, metrics[1].StrategyID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyMetricRepositoryGetLatestByStrategy(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %v", err)
		}
	}()

	// Test data
	strategyID := int64(1)
	now := time.Now()
	simulationRunID := int64(5)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "strategy_id", "simulation_run_id", "win_rate", "avg_profit", "avg_loss",
		"max_drawdown", "total_trades", "successful_trades", "risk_score", "created_at",
	}).
		AddRow(1, strategyID, simulationRunID, 0.75, 100.50, 50.25, 200.0, 20, 15, 1, now)

	mock.ExpectQuery(`SELECT (.+) FROM strategy_metrics WHERE strategy_id = \$1 ORDER BY created_at DESC LIMIT 1`).
		WithArgs(strategyID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyMetricRepository(db)

	// Execute test
	metric, err := repo.GetLatestByStrategy(strategyID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, int64(1), metric.ID)
	assert.Equal(t, strategyID, metric.StrategyID)
	assert.Equal(t, &simulationRunID, metric.SimulationRunID)
	assert.Equal(t, float64(0.75), metric.WinRate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyMetricRepositoryGetLatestByStrategyNotFound(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %v", err)
		}
	}()

	// Test data
	strategyID := int64(999)

	// Setup expected query with no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM strategy_metrics WHERE strategy_id = \$1 ORDER BY created_at DESC LIMIT 1`).
		WithArgs(strategyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "strategy_id", "simulation_run_id", "win_rate", "avg_profit", "avg_loss",
			"max_drawdown", "total_trades", "successful_trades", "risk_score", "created_at",
		}))

	// Create repository with mock DB
	repo := NewStrategyMetricRepository(db)

	// Execute test
	metric, err := repo.GetLatestByStrategy(strategyID)

	// Assert results
	assert.NoError(t, err)
	assert.Nil(t, metric)
	assert.NoError(t, mock.ExpectationsWereMet())
}
