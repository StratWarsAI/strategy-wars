// internal/repository/strategy_metric_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// StrategyMetricRepository handles database operations for strategy metrics
type StrategyMetricRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewStrategyMetricRepository creates a new strategy metric repository
func NewStrategyMetricRepository(db *sql.DB) *StrategyMetricRepository {
	return &StrategyMetricRepository{db: db}
}

// Save inserts a strategy metric into the database
func (r *StrategyMetricRepository) Save(metric *models.StrategyMetric) (int64, error) {
	query := `
        INSERT INTO strategy_metrics 
            (strategy_id, simulation_run_id, win_rate, avg_profit, avg_loss, max_drawdown, 
            total_trades, successful_trades, risk_score, roi, current_balance, initial_balance, created_at) 
        VALUES 
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
        RETURNING id
    `

	now := time.Now()
	if metric.CreatedAt.IsZero() {
		metric.CreatedAt = now
	}

	var id int64
	var simulationRunID sql.NullInt64
	if metric.SimulationRunID != nil {
		simulationRunID.Int64 = *metric.SimulationRunID
		simulationRunID.Valid = true
	}

	err := r.db.QueryRow(
		query,
		metric.StrategyID,
		simulationRunID,
		metric.WinRate,
		metric.AvgProfit,
		metric.AvgLoss,
		metric.MaxDrawdown,
		metric.TotalTrades,
		metric.SuccessfulTrades,
		metric.RiskScore,
		metric.ROI,
		metric.CurrentBalance,
		metric.InitialBalance,
		metric.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving strategy metric: %v", err)
	}

	return id, nil
}

// GetByID retrieves a strategy metric by ID
func (r *StrategyMetricRepository) GetByID(id int64) (*models.StrategyMetric, error) {
	query := `
        SELECT id, strategy_id, simulation_run_id, win_rate, avg_profit, avg_loss, max_drawdown, 
            total_trades, successful_trades, risk_score, roi, current_balance, initial_balance, created_at
        FROM strategy_metrics 
        WHERE id = $1
    `

	var metric models.StrategyMetric
	var simulationRunID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&metric.ID,
		&metric.StrategyID,
		&simulationRunID,
		&metric.WinRate,
		&metric.AvgProfit,
		&metric.AvgLoss,
		&metric.MaxDrawdown,
		&metric.TotalTrades,
		&metric.SuccessfulTrades,
		&metric.RiskScore,
		&metric.ROI,
		&metric.CurrentBalance,
		&metric.InitialBalance,
		&metric.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No metric found
		}
		return nil, fmt.Errorf("error getting strategy metric: %v", err)
	}

	if simulationRunID.Valid {
		metric.SimulationRunID = &simulationRunID.Int64
	}

	return &metric, nil
}

// GetByStrategy retrieves metrics for a strategy
func (r *StrategyMetricRepository) GetByStrategy(strategyID int64) ([]*models.StrategyMetric, error) {
	query := `
        SELECT id, strategy_id, simulation_run_id, win_rate, avg_profit, avg_loss, max_drawdown, 
            total_trades, successful_trades, risk_score, roi, current_balance, initial_balance, created_at
        FROM strategy_metrics 
        WHERE strategy_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query, strategyID)
	if err != nil {
		return nil, fmt.Errorf("error getting strategy metrics: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanMetricRows(rows)
}

// GetBySimulationRun retrieves metrics for a simulation run
func (r *StrategyMetricRepository) GetBySimulationRun(simulationRunID int64) ([]*models.StrategyMetric, error) {
	query := `
        SELECT id, strategy_id, simulation_run_id, win_rate, avg_profit, avg_loss, max_drawdown, 
            total_trades, successful_trades, risk_score, roi, current_balance, initial_balance, created_at
        FROM strategy_metrics 
        WHERE simulation_run_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query, simulationRunID)
	if err != nil {
		return nil, fmt.Errorf("error getting simulation run metrics: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanMetricRows(rows)
}

// GetLatestByStrategy retrieves the latest metric for a strategy
func (r *StrategyMetricRepository) GetLatestByStrategy(strategyID int64) (*models.StrategyMetric, error) {
	return r.GetLatestByStrategyAndSimulation(strategyID, nil)
}

// GetLatestByStrategyAndSimulation retrieves the latest metric for a strategy and simulation run
func (r *StrategyMetricRepository) GetLatestByStrategyAndSimulation(strategyID int64, simulationRunID *int64) (*models.StrategyMetric, error) {
	var query string
	var args []interface{}

	if simulationRunID != nil {
		// If simulation run ID is provided, filter by both strategy and simulation run
		query = `
			SELECT id, strategy_id, simulation_run_id, win_rate, avg_profit, avg_loss, max_drawdown, 
				total_trades, successful_trades, risk_score, roi, current_balance, initial_balance, created_at
			FROM strategy_metrics 
			WHERE strategy_id = $1 AND simulation_run_id = $2
			ORDER BY created_at DESC
			LIMIT 1
		`
		args = []interface{}{strategyID, *simulationRunID}
	} else {
		// If only strategy ID is provided, get the latest for the strategy
		query = `
			SELECT id, strategy_id, simulation_run_id, win_rate, avg_profit, avg_loss, max_drawdown, 
				total_trades, successful_trades, risk_score, roi, current_balance, initial_balance, created_at
			FROM strategy_metrics 
			WHERE strategy_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`
		args = []interface{}{strategyID}
	}

	var metric models.StrategyMetric
	var nullableSimRunID sql.NullInt64

	err := r.db.QueryRow(query, args...).Scan(
		&metric.ID,
		&metric.StrategyID,
		&nullableSimRunID,
		&metric.WinRate,
		&metric.AvgProfit,
		&metric.AvgLoss,
		&metric.MaxDrawdown,
		&metric.TotalTrades,
		&metric.SuccessfulTrades,
		&metric.RiskScore,
		&metric.ROI,
		&metric.CurrentBalance,
		&metric.InitialBalance,
		&metric.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No metric found
		}
		return nil, fmt.Errorf("error getting latest strategy metric: %v", err)
	}

	if nullableSimRunID.Valid {
		metric.SimulationRunID = &nullableSimRunID.Int64
	}

	return &metric, nil
}

// UpdateLatestByStrategy updates the latest metric for a strategy
func (r *StrategyMetricRepository) UpdateLatestByStrategy(metric *models.StrategyMetric) error {
	// First check if a metric exists for this specific simulation run
	var latestMetric *models.StrategyMetric
	var err error
	
	if metric.SimulationRunID != nil {
		// Find latest metric for this strategy and simulation run
		latestMetric, err = r.GetLatestByStrategyAndSimulation(metric.StrategyID, metric.SimulationRunID)
		if err != nil {
			return fmt.Errorf("error checking for existing metric by simulation: %v", err)
		}
	} else {
		// Fall back to just strategy-based lookup if no simulation run
		latestMetric, err = r.GetLatestByStrategy(metric.StrategyID)
		if err != nil {
			return fmt.Errorf("error checking for existing metric: %v", err)
		}
	}

	// If a metric exists with matching simulation run ID (or both are nil), update it
	if latestMetric != nil &&
		((metric.SimulationRunID != nil && latestMetric.SimulationRunID != nil &&
			*metric.SimulationRunID == *latestMetric.SimulationRunID) ||
			(metric.SimulationRunID == nil && latestMetric.SimulationRunID == nil)) {

		// Update the existing metric
		query := `
			UPDATE strategy_metrics 
			SET win_rate = $1, avg_profit = $2, avg_loss = $3, max_drawdown = $4, 
				total_trades = $5, successful_trades = $6, risk_score = $7,
				roi = $8, current_balance = $9, initial_balance = $10
			WHERE id = $11
		`

		_, err := r.db.Exec(
			query,
			metric.WinRate,
			metric.AvgProfit,
			metric.AvgLoss,
			metric.MaxDrawdown,
			metric.TotalTrades,
			metric.SuccessfulTrades,
			metric.RiskScore,
			metric.ROI,
			metric.CurrentBalance,
			metric.InitialBalance,
			latestMetric.ID,
		)

		if err != nil {
			return fmt.Errorf("error updating strategy metric: %v", err)
		}

		return nil
	}

	// Otherwise create a new metric
	// If we're here, either:
	// 1. No metrics exist for this strategy, or
	// 2. No metrics exist for this specific simulation run
	_, err = r.Save(metric)
	if err != nil {
		return fmt.Errorf("error creating new strategy metric: %v", err)
	}

	return nil
}

// scanMetricRows is a helper function to scan multiple metric rows
func (r *StrategyMetricRepository) scanMetricRows(rows *sql.Rows) ([]*models.StrategyMetric, error) {
	var metrics []*models.StrategyMetric

	for rows.Next() {
		var metric models.StrategyMetric
		var simulationRunID sql.NullInt64

		if err := rows.Scan(
			&metric.ID,
			&metric.StrategyID,
			&simulationRunID,
			&metric.WinRate,
			&metric.AvgProfit,
			&metric.AvgLoss,
			&metric.MaxDrawdown,
			&metric.TotalTrades,
			&metric.SuccessfulTrades,
			&metric.RiskScore,
			&metric.ROI,
			&metric.CurrentBalance,
			&metric.InitialBalance,
			&metric.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning strategy metric row: %v", err)
		}

		if simulationRunID.Valid {
			metric.SimulationRunID = &simulationRunID.Int64
		}

		metrics = append(metrics, &metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating strategy metric rows: %v", err)
	}

	return metrics, nil
}