// internal/repository/strategy_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/lib/pq"
)

// StrategyRepository handles database operations for strategies
type StrategyRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewStrategyRepository creates a new strategy repository
func NewStrategyRepository(db *sql.DB) *StrategyRepository {
	return &StrategyRepository{db: db}
}

// Save inserts a strategy into the database
func (r *StrategyRepository) Save(strategy *models.Strategy) (int64, error) {
	query := `
		INSERT INTO strategies 
			(name, description, config, is_public, vote_count, win_count, 
			last_win_time, tags, complexity_score, risk_score, ai_enhanced, created_at, updated_at) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	now := time.Now()
	if strategy.CreatedAt.IsZero() {
		strategy.CreatedAt = now
	}
	strategy.UpdatedAt = now

	var id int64
	err := r.db.QueryRow(
		query,
		strategy.Name,
		strategy.Description,
		strategy.Config,
		strategy.IsPublic,
		strategy.VoteCount,
		strategy.WinCount,
		strategy.LastWinTime,
		pq.Array(strategy.Tags),
		strategy.ComplexityScore,
		strategy.RiskScore,
		strategy.AIEnhanced,
		strategy.CreatedAt,
		strategy.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving strategy: %v", err)
	}

	return id, nil
}

// GetByID retrieves a strategy by its ID
func (r *StrategyRepository) GetByID(id int64) (*models.Strategy, error) {
	query := `
		SELECT id, name, description, config, is_public, vote_count, win_count, 
			last_win_time, tags, complexity_score, risk_score, ai_enhanced, created_at, updated_at
		FROM strategies 
		WHERE id = $1
	`

	var strategy models.Strategy
	var lastWinTime sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&strategy.ID,
		&strategy.Name,
		&strategy.Description,
		&strategy.Config,
		&strategy.IsPublic,
		&strategy.VoteCount,
		&strategy.WinCount,
		&lastWinTime,
		pq.Array(&strategy.Tags),
		&strategy.ComplexityScore,
		&strategy.RiskScore,
		&strategy.AIEnhanced,
		&strategy.CreatedAt,
		&strategy.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No strategy found
		}
		return nil, fmt.Errorf("error getting strategy: %v", err)
	}

	if lastWinTime.Valid {
		strategy.LastWinTime = lastWinTime.Time
	}

	return &strategy, nil
}

// ListPublic retrieves public strategies
func (r *StrategyRepository) ListPublic(limit, offset int) ([]*models.Strategy, error) {
	query := `
		SELECT id, name, description, config, is_public, vote_count, win_count, 
			last_win_time, tags, complexity_score, risk_score, ai_enhanced, created_at, updated_at
		FROM strategies 
		WHERE is_public = true
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing public strategies: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanStrategyRows(rows)
}

// Update updates an existing strategy
func (r *StrategyRepository) Update(strategy *models.Strategy) error {
	query := `
		UPDATE strategies 
        SET name = $1, description = $2, config = $3, is_public = $4, 
            vote_count = $5, win_count = $6, last_win_time = $7, tags = $8, 
            complexity_score = $9, risk_score = $10, ai_enhanced = $11, updated_at = $12
        WHERE id = $13
	`

	strategy.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		strategy.Name,
		strategy.Description,
		strategy.Config,
		strategy.IsPublic,
		strategy.VoteCount,
		strategy.WinCount,
		strategy.LastWinTime,
		pq.Array(strategy.Tags),
		strategy.ComplexityScore,
		strategy.RiskScore,
		strategy.AIEnhanced,
		strategy.UpdatedAt,
		strategy.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating strategy: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("strategy not found: %d", strategy.ID)
	}

	return nil
}

// Delete deletes a strategy
func (r *StrategyRepository) Delete(id int64) error {
	query := `DELETE FROM strategies WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting strategy: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("strategy not found: %d", id)
	}

	return nil
}

// IncrementVoteCount increments the vote count for a strategy
func (r *StrategyRepository) IncrementVoteCount(id int64) error {
	query := `
		UPDATE strategies
		SET vote_count = vote_count + 1, updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error incrementing vote count: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("strategy not found: %d", id)
	}

	return nil
}

// IncrementWinCount increments the win count for a strategy and updates the last win time
func (r *StrategyRepository) IncrementWinCount(id int64, winTime time.Time) error {
	query := `
		UPDATE strategies
		SET win_count = win_count + 1, last_win_time = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, winTime, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error incrementing win count: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("strategy not found: %d", id)
	}

	return nil
}

// SearchByTags searches strategies by tags
func (r *StrategyRepository) SearchByTags(tags []string, limit int) ([]*models.Strategy, error) {
	query := `
		SELECT id, name, description, config, is_public, vote_count, win_count, 
			last_win_time, tags, complexity_score, risk_score, ai_enhanced, created_at, updated_at
		FROM strategies 
		WHERE is_public = true AND tags && $1
		ORDER BY updated_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, pq.Array(tags), limit)
	if err != nil {
		return nil, fmt.Errorf("error searching strategies by tags: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanStrategyRows(rows)
}

// GetTopVoted retrieves top voted strategies
func (r *StrategyRepository) GetTopVoted(limit int) ([]*models.Strategy, error) {
	query := `
		SELECT id, name, description, config, is_public, vote_count, win_count, 
			last_win_time, tags, complexity_score, risk_score, ai_enhanced, created_at, updated_at
		FROM strategies 
		WHERE is_public = true
		ORDER BY vote_count DESC, updated_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting top voted strategies: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanStrategyRows(rows)
}

// GetTopWinners retrieves top winning strategies
func (r *StrategyRepository) GetTopWinners(limit int) ([]*models.Strategy, error) {
	query := `
		SELECT id, name, description, config, is_public, vote_count, win_count, 
			last_win_time, tags, complexity_score, risk_score, ai_enhanced, created_at, updated_at
		FROM strategies 
		WHERE is_public = true
		ORDER BY win_count DESC, updated_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting top winning strategies: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanStrategyRows(rows)
}

// scanStrategyRows is a helper function to scan multiple strategy rows
func (r *StrategyRepository) scanStrategyRows(rows *sql.Rows) ([]*models.Strategy, error) {
	var strategies []*models.Strategy

	for rows.Next() {
		var strategy models.Strategy
		var lastWinTime sql.NullTime

		if err := rows.Scan(
			&strategy.ID,
			&strategy.Name,
			&strategy.Description,
			&strategy.Config,
			&strategy.IsPublic,
			&strategy.VoteCount,
			&strategy.WinCount,
			&lastWinTime,
			pq.Array(&strategy.Tags),
			&strategy.ComplexityScore,
			&strategy.RiskScore,
			&strategy.AIEnhanced,
			&strategy.CreatedAt,
			&strategy.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning strategy row: %v", err)
		}

		if lastWinTime.Valid {
			strategy.LastWinTime = lastWinTime.Time
		}

		strategies = append(strategies, &strategy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating strategy rows: %v", err)
	}

	return strategies, nil
}
