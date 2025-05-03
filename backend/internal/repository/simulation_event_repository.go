// repository/simulation_event_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// SimulationEventRepository handles database operations for simulation events
type SimulationEventRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewSimulationEventRepository creates a new simulation event repository
func NewSimulationEventRepository(db *sql.DB) *SimulationEventRepository {
	return &SimulationEventRepository{db: db}
}

// Save inserts a simulation event into the database
func (r *SimulationEventRepository) Save(event *models.SimulationEvent) (int64, error) {
	query := `
        INSERT INTO simulation_events 
            (strategy_id, simulation_run_id, event_type, event_data, timestamp, created_at) 
        VALUES 
            ($1, $2, $3, $4, $5, $6) 
        RETURNING id
    `

	now := time.Now()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = now
	}

	var id int64
	err := r.db.QueryRow(
		query,
		event.StrategyID,
		event.SimulationRunID,
		event.EventType,
		event.EventData,
		event.Timestamp,
		event.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving simulation event: %v", err)
	}

	return id, nil
}

// GetByID retrieves a simulation event by ID
func (r *SimulationEventRepository) GetByID(id int64) (*models.SimulationEvent, error) {
	query := `
        SELECT id, strategy_id, simulation_run_id, event_type, event_data, timestamp, created_at
        FROM simulation_events 
        WHERE id = $1
    `

	var event models.SimulationEvent
	err := r.db.QueryRow(query, id).Scan(
		&event.ID,
		&event.StrategyID,
		&event.SimulationRunID,
		&event.EventType,
		&event.EventData,
		&event.Timestamp,
		&event.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No event found
		}
		return nil, fmt.Errorf("error getting simulation event: %v", err)
	}

	return &event, nil
}

// GetByStrategyID retrieves simulation events by strategy ID
func (r *SimulationEventRepository) GetByStrategyID(strategyID int64, limit, offset int) ([]*models.SimulationEvent, error) {
	query := `
        SELECT id, strategy_id, simulation_run_id, event_type, event_data, timestamp, created_at
        FROM simulation_events 
        WHERE strategy_id = $1
        ORDER BY timestamp DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.Query(query, strategyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting simulation events: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanEventRows(rows)
}

// GetBySimulationRunID retrieves simulation events by simulation run ID
func (r *SimulationEventRepository) GetBySimulationRunID(simulationRunID int64, limit, offset int) ([]*models.SimulationEvent, error) {
	query := `
        SELECT id, strategy_id, simulation_run_id, event_type, event_data, timestamp, created_at
        FROM simulation_events 
        WHERE simulation_run_id = $1
        ORDER BY timestamp ASC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.Query(query, simulationRunID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting simulation events: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanEventRows(rows)
}

// GetLatestByStrategyID retrieves the latest simulation events for a strategy
func (r *SimulationEventRepository) GetLatestByStrategyID(strategyID int64, limit int) ([]*models.SimulationEvent, error) {
	query := `
        SELECT id, strategy_id, simulation_run_id, event_type, event_data, timestamp, created_at
        FROM simulation_events 
        WHERE strategy_id = $1
        ORDER BY timestamp DESC
        LIMIT $2
    `

	rows, err := r.db.Query(query, strategyID, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting latest simulation events: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanEventRows(rows)
}

// scanEventRows is a helper function to scan multiple event rows
func (r *SimulationEventRepository) scanEventRows(rows *sql.Rows) ([]*models.SimulationEvent, error) {
	var events []*models.SimulationEvent

	for rows.Next() {
		var event models.SimulationEvent

		if err := rows.Scan(
			&event.ID,
			&event.StrategyID,
			&event.SimulationRunID,
			&event.EventType,
			&event.EventData,
			&event.Timestamp,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning simulation event row: %v", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating simulation event rows: %v", err)
	}

	return events, nil
}
