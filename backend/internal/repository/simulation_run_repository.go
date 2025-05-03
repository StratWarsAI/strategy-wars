// internal/repository/simulation_run_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// SimulationRunRepository handles database operations for simulation runs
type SimulationRunRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewSimulationRunRepository creates a new simulation run repository
func NewSimulationRunRepository(db *sql.DB) *SimulationRunRepository {
	return &SimulationRunRepository{
		db: db,
	}
}

// SetLogger sets the logger for this repository
func (r *SimulationRunRepository) SetLogger(logger *logger.Logger) {
	r.logger = logger
}

// Save inserts a simulation run into the database
func (r *SimulationRunRepository) Save(run *models.SimulationRun) (int64, error) {
	query := `
		INSERT INTO simulation_runs 
			(start_time, end_time, winner_strategy_id, status, simulation_parameters, created_at, updated_at) 
		VALUES 
			($1, $2, NULL, $3, $4, $5, $6)
		RETURNING id
	`

	now := time.Now()
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	run.UpdatedAt = now

	var id int64
	err := r.db.QueryRow(
		query,
		run.StartTime,
		run.EndTime,
		run.Status,
		run.SimulationParameters,
		run.CreatedAt,
		run.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving simulation run: %v", err)
	}

	return id, nil
}

// GetByID retrieves a simulation run by its ID
func (r *SimulationRunRepository) GetByID(id int64) (*models.SimulationRun, error) {
	query := `
		SELECT id, start_time, end_time, winner_strategy_id, status, simulation_parameters, created_at, updated_at
		FROM simulation_runs 
		WHERE id = $1
	`

	var run models.SimulationRun
	var winnerStrategyID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&run.ID,
		&run.StartTime,
		&run.EndTime,
		&winnerStrategyID,
		&run.Status,
		&run.SimulationParameters,
		&run.CreatedAt,
		&run.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No simulation run found
		}
		return nil, fmt.Errorf("error getting simulation run: %v", err)
	}

	if winnerStrategyID.Valid {
		run.WinnerStrategyID = winnerStrategyID.Int64
	}

	return &run, nil
}

// GetCurrent retrieves the most recent active simulation run
func (r *SimulationRunRepository) GetCurrent() (*models.SimulationRun, error) {
	query := `
		SELECT id, start_time, end_time, winner_strategy_id, status, simulation_parameters, created_at, updated_at
		FROM simulation_runs 
		WHERE status = 'running' OR status = 'preparing'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var run models.SimulationRun
	var winnerStrategyID sql.NullInt64

	err := r.db.QueryRow(query).Scan(
		&run.ID,
		&run.StartTime,
		&run.EndTime,
		&winnerStrategyID,
		&run.Status,
		&run.SimulationParameters,
		&run.CreatedAt,
		&run.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No current simulation run
		}
		return nil, fmt.Errorf("error getting current simulation run: %v", err)
	}

	if winnerStrategyID.Valid {
		run.WinnerStrategyID = winnerStrategyID.Int64
	}

	return &run, nil
}

// GetByStatus retrieves simulation runs by status
func (r *SimulationRunRepository) GetByStatus(status string, limit int) ([]*models.SimulationRun, error) {
	query := `
		SELECT id, start_time, end_time, winner_strategy_id, status, simulation_parameters, created_at, updated_at
		FROM simulation_runs 
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying simulation runs by status: %v", err)
	}
	defer rows.Close()

	return r.scanSimulationRunRows(rows)
}

// GetByTimeRange retrieves simulation runs within a time range
func (r *SimulationRunRepository) GetByTimeRange(start, end time.Time) ([]*models.SimulationRun, error) {
	query := `
		SELECT id, start_time, end_time, winner_strategy_id, status, simulation_parameters, created_at, updated_at
		FROM simulation_runs 
		WHERE start_time >= $1 AND end_time <= $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("error querying simulation runs by time range: %v", err)
	}
	defer rows.Close()

	return r.scanSimulationRunRows(rows)
}

// UpdateStatus updates the status of a simulation run
func (r *SimulationRunRepository) UpdateStatus(id int64, status string) error {
	query := `
		UPDATE simulation_runs
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error updating simulation run status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("simulation run not found: %d", id)
	}

	return nil
}

// UpdateWinner updates the winner strategy of a simulation run
func (r *SimulationRunRepository) UpdateWinner(id int64, strategyID int64) error {
	query := `
		UPDATE simulation_runs
		SET winner_strategy_id = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, strategyID, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error updating simulation run winner: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("simulation run not found: %d", id)
	}

	return nil
}

// scanSimulationRunRows is a helper function to scan multiple simulation run rows
func (r *SimulationRunRepository) scanSimulationRunRows(rows *sql.Rows) ([]*models.SimulationRun, error) {
	var runs []*models.SimulationRun

	for rows.Next() {
		var run models.SimulationRun
		var winnerStrategyID sql.NullInt64

		if err := rows.Scan(
			&run.ID,
			&run.StartTime,
			&run.EndTime,
			&winnerStrategyID,
			&run.Status,
			&run.SimulationParameters,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning simulation run row: %v", err)
		}

		if winnerStrategyID.Valid {
			run.WinnerStrategyID = winnerStrategyID.Int64
		}

		runs = append(runs, &run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating simulation run rows: %v", err)
	}

	return runs, nil
}
