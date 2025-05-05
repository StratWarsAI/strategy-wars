// repository/simulation_result_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// SimulationResultRepository handles database operations for simulation results
type SimulationResultRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewSimulationResultRepository creates a new simulation result repository
func NewSimulationResultRepository(db *sql.DB) *SimulationResultRepository {
	return &SimulationResultRepository{db: db}
}

// Save inserts a simulation result into the database
func (r *SimulationResultRepository) Save(result *models.SimulationResult) (int64, error) {
	query := `
        INSERT INTO simulation_results
            (simulation_run_id, strategy_id, roi, trade_count, win_rate, max_drawdown, performance_rating, analysis, rank, created_at) 
        VALUES 
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
        RETURNING id
    `

	now := time.Now()
	if result.CreatedAt.IsZero() {
		result.CreatedAt = now
	}

	var id int64
	err := r.db.QueryRow(
		query,
		result.SimulationRunID,
		result.StrategyID,
		result.ROI,
		result.TradeCount,
		result.WinRate,
		result.MaxDrawdown,
		result.PerformanceRating,
		result.Analysis,
		result.Rank,
		result.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving simulation result: %v", err)
	}

	return id, nil
}

// GetByID retrieves a simulation result by ID
func (r *SimulationResultRepository) GetByID(id int64) (*models.SimulationResult, error) {
	query := `
        SELECT id, simulation_run_id, strategy_id, roi, trade_count, win_rate, max_drawdown, performance_rating, analysis, rank, created_at
        FROM simulation_results 
        WHERE id = $1
    `

	var result models.SimulationResult
	err := r.db.QueryRow(query, id).Scan(
		&result.ID,
		&result.SimulationRunID,
		&result.StrategyID,
		&result.ROI,
		&result.TradeCount,
		&result.WinRate,
		&result.MaxDrawdown,
		&result.PerformanceRating,
		&result.Analysis,
		&result.Rank,
		&result.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No result found
		}
		return nil, fmt.Errorf("error getting simulation result: %v", err)
	}

	return &result, nil
}

// GetBySimulationRun retrieves simulation results by simulation run ID
func (r *SimulationResultRepository) GetBySimulationRun(simulationRunID int64) ([]*models.SimulationResult, error) {
	query := `
        SELECT id, simulation_run_id, strategy_id, roi, trade_count, win_rate, max_drawdown, performance_rating, analysis, rank, created_at
        FROM simulation_results 
        WHERE simulation_run_id = $1
        ORDER BY rank ASC
    `

	rows, err := r.db.Query(query, simulationRunID)
	if err != nil {
		return nil, fmt.Errorf("error getting simulation results: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanResultRows(rows)
}

// GetTopPerformers retrieves top performing strategies in a simulation run
func (r *SimulationResultRepository) GetTopPerformers(simulationRunID int64, limit int) ([]*models.SimulationResult, error) {
	query := `
        SELECT id, simulation_run_id, strategy_id, roi, trade_count, win_rate, max_drawdown, performance_rating, analysis, rank, created_at
        FROM simulation_results 
        WHERE simulation_run_id = $1
        ORDER BY roi DESC
        LIMIT $2
    `

	rows, err := r.db.Query(query, simulationRunID, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting top performers: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanResultRows(rows)
}

// GetByStrategy retrieves simulation results for a specific strategy
func (r *SimulationResultRepository) GetByStrategy(strategyID int64, limit int) ([]*models.SimulationResult, error) {
	query := `
        SELECT id, simulation_run_id, strategy_id, roi, trade_count, win_rate, max_drawdown, performance_rating, analysis, rank, created_at
        FROM simulation_results 
        WHERE strategy_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `

	rows, err := r.db.Query(query, strategyID, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting strategy results: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanResultRows(rows)
}

// scanResultRows is a helper function to scan multiple result rows
func (r *SimulationResultRepository) scanResultRows(rows *sql.Rows) ([]*models.SimulationResult, error) {
	var results []*models.SimulationResult

	for rows.Next() {
		var result models.SimulationResult

		if err := rows.Scan(
			&result.ID,
			&result.SimulationRunID,
			&result.StrategyID,
			&result.ROI,
			&result.TradeCount,
			&result.WinRate,
			&result.MaxDrawdown,
			&result.PerformanceRating,
			&result.Analysis,
			&result.Rank,
			&result.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning simulation result row: %v", err)
		}

		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating simulation result rows: %v", err)
	}

	return results, nil
}