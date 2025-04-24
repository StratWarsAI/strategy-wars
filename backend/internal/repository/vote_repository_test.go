// internal/repository/vote_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestVoteRepositorySave(t *testing.T) {
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
	// Create test vote
	now := time.Now()
	vote := &models.Vote{
		DuelID:     1,
		StrategyID: 5,
		UserID:     10,
		CreatedAt:  now,
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO votes`).
		WithArgs(
			vote.DuelID,
			vote.StrategyID,
			vote.UserID,
			vote.CreatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewVoteRepository(db)

	// Execute test
	id, err := repo.Save(vote)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVoteRepositoryGetByUserAndDuel(t *testing.T) {
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
	userID := int64(10)
	duelID := int64(1)
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "duel_id", "strategy_id", "user_id", "created_at",
	}).
		AddRow(1, duelID, 5, userID, now)

	mock.ExpectQuery(`SELECT (.+) FROM votes WHERE user_id = \$1 AND duel_id = \$2`).
		WithArgs(userID, duelID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewVoteRepository(db)

	// Execute test
	vote, err := repo.GetByUserAndDuel(userID, duelID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, vote)
	assert.Equal(t, int64(1), vote.ID)
	assert.Equal(t, duelID, vote.DuelID)
	assert.Equal(t, int64(5), vote.StrategyID)
	assert.Equal(t, userID, vote.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVoteRepositoryGetByUserAndDuelNotFound(t *testing.T) {
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
	userID := int64(10)
	duelID := int64(1)

	// Setup expected query with no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM votes WHERE user_id = \$1 AND duel_id = \$2`).
		WithArgs(userID, duelID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "duel_id", "strategy_id", "user_id", "created_at",
		}))

	// Create repository with mock DB
	repo := NewVoteRepository(db)

	// Execute test
	vote, err := repo.GetByUserAndDuel(userID, duelID)

	// Assert results
	assert.NoError(t, err)
	assert.Nil(t, vote)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVoteRepositoryGetByDuel(t *testing.T) {
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
	duelID := int64(1)
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "duel_id", "strategy_id", "user_id", "created_at",
	}).
		AddRow(1, duelID, 5, 10, now).
		AddRow(2, duelID, 7, 11, now.Add(1*time.Minute))

	mock.ExpectQuery(`SELECT (.+) FROM votes WHERE duel_id = \$1 ORDER BY created_at DESC`).
		WithArgs(duelID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewVoteRepository(db)

	// Execute test
	votes, err := repo.GetByDuel(duelID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, votes)
	assert.Equal(t, 2, len(votes))
	assert.Equal(t, int64(1), votes[0].ID)
	assert.Equal(t, int64(2), votes[1].ID)
	assert.Equal(t, int64(5), votes[0].StrategyID)
	assert.Equal(t, int64(7), votes[1].StrategyID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVoteRepositoryGetVoteCounts(t *testing.T) {
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
	duelID := int64(1)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"strategy_id", "vote_count",
	}).
		AddRow(5, 3).
		AddRow(7, 2)

	mock.ExpectQuery(`SELECT strategy_id, COUNT\(\*\) as vote_count FROM votes WHERE duel_id = \$1 GROUP BY strategy_id ORDER BY vote_count DESC`).
		WithArgs(duelID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewVoteRepository(db)

	// Execute test
	voteCounts, err := repo.GetVoteCounts(duelID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, voteCounts)
	assert.Equal(t, 2, len(voteCounts))
	assert.Equal(t, 3, voteCounts[5])
	assert.Equal(t, 2, voteCounts[7])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVoteRepositoryGetVoteCountsForStrategy(t *testing.T) {
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
	strategyID := int64(5)

	// Setup expected query and result
	mock.ExpectQuery(`SELECT COUNT\(\*\) as total_votes FROM votes WHERE strategy_id = \$1`).
		WithArgs(strategyID).
		WillReturnRows(sqlmock.NewRows([]string{"total_votes"}).AddRow(10))

	// Create repository with mock DB
	repo := NewVoteRepository(db)

	// Execute test
	totalVotes, err := repo.GetVoteCountsForStrategy(strategyID)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, 10, totalVotes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVoteRepositoryDeleteByUserAndDuel(t *testing.T) {
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
	userID := int64(10)
	duelID := int64(1)

	// Setup expected query and result
	mock.ExpectExec(`DELETE FROM votes WHERE user_id = \$1 AND duel_id = \$2`).
		WithArgs(userID, duelID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewVoteRepository(db)

	// Execute test
	err = repo.DeleteByUserAndDuel(userID, duelID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVoteRepositoryDeleteByUserAndDuelNotFound(t *testing.T) {
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
	userID := int64(10)
	duelID := int64(1)

	// Setup expected query and result (no rows affected)
	mock.ExpectExec(`DELETE FROM votes WHERE user_id = \$1 AND duel_id = \$2`).
		WithArgs(userID, duelID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Create repository with mock DB
	repo := NewVoteRepository(db)

	// Execute test
	err = repo.DeleteByUserAndDuel(userID, duelID)

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vote not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}
