// internal/repository/vote_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
)

// VoteRepository handles database operations for votes
type VoteRepository struct {
	db *sql.DB
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository(db *sql.DB) *VoteRepository {
	return &VoteRepository{db: db}
}

// Save inserts a vote into the database
func (r *VoteRepository) Save(vote *models.Vote) (int64, error) {
	query := `
		INSERT INTO votes 
			(duel_id, strategy_id, user_id, created_at) 
		VALUES 
			($1, $2, $3, $4) 
		ON CONFLICT (duel_id, user_id) 
		DO UPDATE SET 
			strategy_id = $2,
			created_at = $4
		RETURNING id
	`

	now := time.Now()
	if vote.CreatedAt.IsZero() {
		vote.CreatedAt = now
	}

	var id int64
	err := r.db.QueryRow(
		query,
		vote.DuelID,
		vote.StrategyID,
		vote.UserID,
		vote.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving vote: %v", err)
	}

	return id, nil
}

// GetByUserAndDuel retrieves a vote by user and duel IDs
func (r *VoteRepository) GetByUserAndDuel(userID, duelID int64) (*models.Vote, error) {
	query := `
		SELECT id, duel_id, strategy_id, user_id, created_at
		FROM votes 
		WHERE user_id = $1 AND duel_id = $2
	`

	var vote models.Vote
	err := r.db.QueryRow(query, userID, duelID).Scan(
		&vote.ID,
		&vote.DuelID,
		&vote.StrategyID,
		&vote.UserID,
		&vote.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No vote found
		}
		return nil, fmt.Errorf("error getting vote: %v", err)
	}

	return &vote, nil
}

// GetByDuel retrieves all votes for a duel
func (r *VoteRepository) GetByDuel(duelID int64) ([]*models.Vote, error) {
	query := `
		SELECT id, duel_id, strategy_id, user_id, created_at
		FROM votes 
		WHERE duel_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, duelID)
	if err != nil {
		return nil, fmt.Errorf("error getting votes for duel: %v", err)
	}
	defer rows.Close()

	var votes []*models.Vote
	for rows.Next() {
		var vote models.Vote
		if err := rows.Scan(
			&vote.ID,
			&vote.DuelID,
			&vote.StrategyID,
			&vote.UserID,
			&vote.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning vote row: %v", err)
		}
		votes = append(votes, &vote)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating vote rows: %v", err)
	}

	return votes, nil
}

// GetVoteCounts retrieves vote counts for each strategy in a duel
func (r *VoteRepository) GetVoteCounts(duelID int64) (map[int64]int, error) {
	query := `
		SELECT strategy_id, COUNT(*) as vote_count
		FROM votes 
		WHERE duel_id = $1
		GROUP BY strategy_id
		ORDER BY vote_count DESC
	`

	rows, err := r.db.Query(query, duelID)
	if err != nil {
		return nil, fmt.Errorf("error getting vote counts: %v", err)
	}
	defer rows.Close()

	voteCounts := make(map[int64]int)
	for rows.Next() {
		var strategyID int64
		var count int
		if err := rows.Scan(&strategyID, &count); err != nil {
			return nil, fmt.Errorf("error scanning vote count row: %v", err)
		}
		voteCounts[strategyID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating vote count rows: %v", err)
	}

	return voteCounts, nil
}

// GetVoteCountsForStrategy retrieves the total votes for a strategy across all duels
func (r *VoteRepository) GetVoteCountsForStrategy(strategyID int64) (int, error) {
	query := `
		SELECT COUNT(*) as total_votes
		FROM votes 
		WHERE strategy_id = $1
	`

	var totalVotes int
	err := r.db.QueryRow(query, strategyID).Scan(&totalVotes)
	if err != nil {
		return 0, fmt.Errorf("error getting total votes for strategy: %v", err)
	}

	return totalVotes, nil
}

// DeleteByUserAndDuel deletes a vote by user and duel IDs
func (r *VoteRepository) DeleteByUserAndDuel(userID, duelID int64) error {
	query := `
		DELETE FROM votes
		WHERE user_id = $1 AND duel_id = $2
	`

	result, err := r.db.Exec(query, userID, duelID)
	if err != nil {
		return fmt.Errorf("error deleting vote: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vote not found for user %d and duel %d", userID, duelID)
	}

	return nil
}
