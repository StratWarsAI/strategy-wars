// internal/repository/duel_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// DuelRepository handles database operations for duels
type DuelRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewDuelRepository creates a new duel repository
func NewDuelRepository(db *sql.DB) *DuelRepository {
	return &DuelRepository{db: db}
}

// Save inserts or updates a duel in the database
func (r *DuelRepository) Save(duel *models.Duel) (int64, error) {
	query := `
		INSERT INTO duels 
			(start_time, end_time, voting_end_time, winner_strategy_id, status, created_at, updated_at) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id
	`

	now := time.Now()
	if duel.CreatedAt.IsZero() {
		duel.CreatedAt = now
	}
	duel.UpdatedAt = now

	var id int64
	var winnerStrategyID sql.NullInt64
	if duel.WinnerStrategyID != 0 {
		winnerStrategyID.Int64 = duel.WinnerStrategyID
		winnerStrategyID.Valid = true
	}

	err := r.db.QueryRow(
		query,
		duel.StartTime,
		duel.EndTime,
		duel.VotingEndTime,
		winnerStrategyID,
		duel.Status,
		duel.CreatedAt,
		duel.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving duel: %v", err)
	}

	return id, nil
}

// GetByID retrieves a duel by its ID
func (r *DuelRepository) GetByID(id int64) (*models.Duel, error) {
	query := `
		SELECT id, start_time, end_time, voting_end_time, winner_strategy_id, status, created_at, updated_at
		FROM duels 
		WHERE id = $1
	`

	var duel models.Duel
	var winnerStrategyID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&duel.ID,
		&duel.StartTime,
		&duel.EndTime,
		&duel.VotingEndTime,
		&winnerStrategyID,
		&duel.Status,
		&duel.CreatedAt,
		&duel.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No duel found
		}
		return nil, fmt.Errorf("error getting duel: %v", err)
	}

	if winnerStrategyID.Valid {
		duel.WinnerStrategyID = winnerStrategyID.Int64
	}

	return &duel, nil
}

// GetCurrent retrieves the current active duel
func (r *DuelRepository) GetCurrent() (*models.Duel, error) {
	query := `
		SELECT id, start_time, end_time, voting_end_time, winner_strategy_id, status, created_at, updated_at
		FROM duels 
		WHERE status = 'voting' OR status = 'simulating'
		ORDER BY start_time ASC
		LIMIT 1
	`

	var duel models.Duel
	var winnerStrategyID sql.NullInt64

	err := r.db.QueryRow(query).Scan(
		&duel.ID,
		&duel.StartTime,
		&duel.EndTime,
		&duel.VotingEndTime,
		&winnerStrategyID,
		&duel.Status,
		&duel.CreatedAt,
		&duel.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No current duel
		}
		return nil, fmt.Errorf("error getting current duel: %v", err)
	}

	if winnerStrategyID.Valid {
		duel.WinnerStrategyID = winnerStrategyID.Int64
	}

	return &duel, nil
}

// GetUpcoming retrieves upcoming duels
func (r *DuelRepository) GetUpcoming(limit int) ([]*models.Duel, error) {
	query := `
		SELECT id, start_time, end_time, voting_end_time, winner_strategy_id, status, created_at, updated_at
		FROM duels 
		WHERE status = 'pending' AND start_time > NOW()
		ORDER BY start_time ASC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting upcoming duels: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanDuelRows(rows)
}

// GetPast retrieves past duels
func (r *DuelRepository) GetPast(limit int) ([]*models.Duel, error) {
	query := `
		SELECT id, start_time, end_time, voting_end_time, winner_strategy_id, status, created_at, updated_at
		FROM duels 
		WHERE status = 'completed'
		ORDER BY end_time DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting past duels: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanDuelRows(rows)
}

// GetByStatus retrieves duels by status
func (r *DuelRepository) GetByStatus(status string, limit int) ([]*models.Duel, error) {
	query := `
		SELECT id, start_time, end_time, voting_end_time, winner_strategy_id, status, created_at, updated_at
		FROM duels 
		WHERE status = $1
		ORDER BY start_time ASC
		LIMIT $2
	`

	rows, err := r.db.Query(query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting duels by status: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanDuelRows(rows)
}

// GetByTimeRange retrieves duels within a time range
func (r *DuelRepository) GetByTimeRange(start, end time.Time) ([]*models.Duel, error) {
	query := `
		SELECT id, start_time, end_time, voting_end_time, winner_strategy_id, status, created_at, updated_at
		FROM duels 
		WHERE (start_time BETWEEN $1 AND $2) OR (end_time BETWEEN $1 AND $2)
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("error getting duels by time range: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanDuelRows(rows)
}

// UpdateStatus updates the status of a duel
func (r *DuelRepository) UpdateStatus(id int64, status string) error {
	query := `
		UPDATE duels
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error updating duel status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("duel with ID %d not found", id)
	}

	return nil
}

// UpdateWinner updates the winning strategy of a duel
func (r *DuelRepository) UpdateWinner(id int64, strategyID int64) error {
	query := `
		UPDATE duels
		SET winner_strategy_id = $1, status = 'completed', updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, strategyID, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error updating duel winner: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("duel with ID %d not found", id)
	}

	return nil
}

// scanDuelRows is a helper function to scan multiple duel rows
func (r *DuelRepository) scanDuelRows(rows *sql.Rows) ([]*models.Duel, error) {
	var duels []*models.Duel

	for rows.Next() {
		var duel models.Duel
		var winnerStrategyID sql.NullInt64

		if err := rows.Scan(
			&duel.ID,
			&duel.StartTime,
			&duel.EndTime,
			&duel.VotingEndTime,
			&winnerStrategyID,
			&duel.Status,
			&duel.CreatedAt,
			&duel.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning duel row: %v", err)
		}

		if winnerStrategyID.Valid {
			duel.WinnerStrategyID = winnerStrategyID.Int64
		}

		duels = append(duels, &duel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating duel rows: %v", err)
	}

	return duels, nil
}
