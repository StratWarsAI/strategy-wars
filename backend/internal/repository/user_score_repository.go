// internal/repository/user_score_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// UserScoreRepository handles database operations for user scores
type UserScoreRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewUserScoreRepository creates a new user score repository
func NewUserScoreRepository(db *sql.DB) *UserScoreRepository {
	return &UserScoreRepository{db: db}
}

// GetByUserID retrieves a user score by user ID
func (r *UserScoreRepository) GetByUserID(userID int64) (*models.UserScore, error) {
	query := `
		SELECT user_id, total_points, win_count, strategy_count, vote_count, last_updated
		FROM user_scores 
		WHERE user_id = $1
	`

	var userScore models.UserScore
	err := r.db.QueryRow(query, userID).Scan(
		&userScore.UserID,
		&userScore.TotalPoints,
		&userScore.WinCount,
		&userScore.StrategyCount,
		&userScore.VoteCount,
		&userScore.LastUpdated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Initialize a new user score entry if it doesn't exist
			if err := r.createUserScore(userID); err != nil {
				return nil, fmt.Errorf("error creating user score: %v", err)
			}
			// Return default values for new user
			return &models.UserScore{
				UserID:        userID,
				TotalPoints:   0,
				WinCount:      0,
				StrategyCount: 0,
				VoteCount:     0,
				LastUpdated:   time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("error getting user score: %v", err)
	}

	return &userScore, nil
}

// GetTopUsers retrieves top users by total points
func (r *UserScoreRepository) GetTopUsers(limit int) ([]*models.UserScore, error) {
	query := `
		SELECT user_id, total_points, win_count, strategy_count, vote_count, last_updated
		FROM user_scores 
		ORDER BY total_points DESC, win_count DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting top users: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	var userScores []*models.UserScore
	for rows.Next() {
		var userScore models.UserScore
		if err := rows.Scan(
			&userScore.UserID,
			&userScore.TotalPoints,
			&userScore.WinCount,
			&userScore.StrategyCount,
			&userScore.VoteCount,
			&userScore.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("error scanning user score row: %v", err)
		}
		userScores = append(userScores, &userScore)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user score rows: %v", err)
	}

	return userScores, nil
}

// IncrementPoints increments the points for a user
func (r *UserScoreRepository) IncrementPoints(userID int64, points int) error {
	// Check if user score exists
	exists, err := r.checkUserScoreExists(userID)
	if err != nil {
		return err
	}

	// Create user score if it doesn't exist
	if !exists {
		if err := r.createUserScore(userID); err != nil {
			return err
		}
	}

	// Increment points
	query := `
		UPDATE user_scores
		SET total_points = total_points + $1, last_updated = $2
		WHERE user_id = $3
	`

	now := time.Now()
	result, err := r.db.Exec(query, points, now, userID)
	if err != nil {
		return fmt.Errorf("error incrementing points: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user score found for user ID: %d", userID)
	}

	return nil
}

// IncrementWins increments the win count for a user
func (r *UserScoreRepository) IncrementWins(userID int64) error {
	// Check if user score exists
	exists, err := r.checkUserScoreExists(userID)
	if err != nil {
		return err
	}

	// Create user score if it doesn't exist
	if !exists {
		if err := r.createUserScore(userID); err != nil {
			return err
		}
	}

	// Increment wins
	query := `
		UPDATE user_scores
		SET win_count = win_count + 1, last_updated = $1
		WHERE user_id = $2
	`

	now := time.Now()
	result, err := r.db.Exec(query, now, userID)
	if err != nil {
		return fmt.Errorf("error incrementing wins: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user score found for user ID: %d", userID)
	}

	return nil
}

// IncrementStrategies increments the strategy count for a user
func (r *UserScoreRepository) IncrementStrategies(userID int64) error {
	// Check if user score exists
	exists, err := r.checkUserScoreExists(userID)
	if err != nil {
		return err
	}

	// Create user score if it doesn't exist
	if !exists {
		if err := r.createUserScore(userID); err != nil {
			return err
		}
	}

	// Increment strategies
	query := `
		UPDATE user_scores
		SET strategy_count = strategy_count + 1, last_updated = $1
		WHERE user_id = $2
	`

	now := time.Now()
	result, err := r.db.Exec(query, now, userID)
	if err != nil {
		return fmt.Errorf("error incrementing strategies: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user score found for user ID: %d", userID)
	}

	return nil
}

// IncrementVotes increments the vote count for a user
func (r *UserScoreRepository) IncrementVotes(userID int64) error {
	// Check if user score exists
	exists, err := r.checkUserScoreExists(userID)
	if err != nil {
		return err
	}

	// Create user score if it doesn't exist
	if !exists {
		if err := r.createUserScore(userID); err != nil {
			return err
		}
	}

	// Increment votes
	query := `
		UPDATE user_scores
		SET vote_count = vote_count + 1, last_updated = $1
		WHERE user_id = $2
	`

	now := time.Now()
	result, err := r.db.Exec(query, now, userID)
	if err != nil {
		return fmt.Errorf("error incrementing votes: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user score found for user ID: %d", userID)
	}

	return nil
}

// UpdateLastUpdated updates the last updated timestamp for a user
func (r *UserScoreRepository) UpdateLastUpdated(userID int64) error {
	// Check if user score exists
	exists, err := r.checkUserScoreExists(userID)
	if err != nil {
		return err
	}

	// Create user score if it doesn't exist
	if !exists {
		if err := r.createUserScore(userID); err != nil {
			return err
		}
	}

	// Update last updated timestamp
	query := `
		UPDATE user_scores
		SET last_updated = $1
		WHERE user_id = $2
	`

	now := time.Now()
	result, err := r.db.Exec(query, now, userID)
	if err != nil {
		return fmt.Errorf("error updating last updated timestamp: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user score found for user ID: %d", userID)
	}

	return nil
}

// Helper methods

// checkUserScoreExists checks if a user score entry exists
func (r *UserScoreRepository) checkUserScoreExists(userID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_scores WHERE user_id = $1)`

	var exists bool
	err := r.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if user score exists: %v", err)
	}

	return exists, nil
}

// createUserScore creates a new user score entry
func (r *UserScoreRepository) createUserScore(userID int64) error {
	query := `
		INSERT INTO user_scores 
		(user_id, total_points, win_count, strategy_count, vote_count, last_updated) 
		VALUES 
		($1, 0, 0, 0, 0, $2)
	`

	now := time.Now()
	_, err := r.db.Exec(query, userID, now)
	if err != nil {
		return fmt.Errorf("error creating user score: %v", err)
	}

	return nil
}
