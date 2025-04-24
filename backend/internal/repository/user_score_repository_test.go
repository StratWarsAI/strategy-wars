// internal/repository/user_score_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserScoreRepositoryGetByUserID(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()

	// Test data
	userID := int64(10)
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"user_id", "total_points", "win_count", "strategy_count", "vote_count", "last_updated",
	}).
		AddRow(userID, 100, 5, 3, 20, now)

	mock.ExpectQuery(`SELECT (.+) FROM user_scores WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	userScore, err := repo.GetByUserID(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, userScore)
	assert.Equal(t, userID, userScore.UserID)
	assert.Equal(t, 100, userScore.TotalPoints)
	assert.Equal(t, 5, userScore.WinCount)
	assert.Equal(t, 3, userScore.StrategyCount)
	assert.Equal(t, 20, userScore.VoteCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryGetByUserIDNotFound(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()

	// Test data
	userID := int64(10)

	// Expect initial query to return no rows
	mock.ExpectQuery(`SELECT (.+) FROM user_scores WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "total_points", "win_count", "strategy_count", "vote_count", "last_updated",
		}))

	// Expect insert for new user score
	mock.ExpectExec(`INSERT INTO user_scores`).
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	userScore, err := repo.GetByUserID(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, userScore)
	assert.Equal(t, userID, userScore.UserID)
	assert.Equal(t, 0, userScore.TotalPoints)
	assert.Equal(t, 0, userScore.WinCount)
	assert.Equal(t, 0, userScore.StrategyCount)
	assert.Equal(t, 0, userScore.VoteCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryGetTopUsers(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()

	// Test data
	limit := 2
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"user_id", "total_points", "win_count", "strategy_count", "vote_count", "last_updated",
	}).
		AddRow(1, 200, 10, 5, 30, now).
		AddRow(2, 150, 8, 4, 25, now)

	mock.ExpectQuery(`SELECT (.+) FROM user_scores ORDER BY total_points DESC, win_count DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	topUsers, err := repo.GetTopUsers(limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, topUsers)
	assert.Equal(t, 2, len(topUsers))
	assert.Equal(t, int64(1), topUsers[0].UserID)
	assert.Equal(t, 200, topUsers[0].TotalPoints)
	assert.Equal(t, int64(2), topUsers[1].UserID)
	assert.Equal(t, 150, topUsers[1].TotalPoints)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryIncrementPoints(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()

	// Test data
	userID := int64(10)
	points := 50

	// Expect check for existing user score
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_scores WHERE user_id = \$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Expect update points
	mock.ExpectExec(`UPDATE user_scores SET total_points = total_points \+ \$1, last_updated = \$2 WHERE user_id = \$3`).
		WithArgs(points, sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	err = repo.IncrementPoints(userID, points)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryIncrementPointsNewUser(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()

	// Test data
	userID := int64(10)
	points := 50

	// Expect check for existing user score (not exists)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_scores WHERE user_id = \$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Expect insert for new user score
	mock.ExpectExec(`INSERT INTO user_scores`).
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect update points
	mock.ExpectExec(`UPDATE user_scores SET total_points = total_points \+ \$1, last_updated = \$2 WHERE user_id = \$3`).
		WithArgs(points, sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	err = repo.IncrementPoints(userID, points)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryIncrementWins(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()

	// Test data
	userID := int64(10)

	// Expect check for existing user score
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_scores WHERE user_id = \$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Expect update wins
	mock.ExpectExec(`UPDATE user_scores SET win_count = win_count \+ 1, last_updated = \$1 WHERE user_id = \$2`).
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	err = repo.IncrementWins(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryIncrementStrategies(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()

	// Test data
	userID := int64(10)

	// Expect check for existing user score
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_scores WHERE user_id = \$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Expect update strategies
	mock.ExpectExec(`UPDATE user_scores SET strategy_count = strategy_count \+ 1, last_updated = \$1 WHERE user_id = \$2`).
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	err = repo.IncrementStrategies(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryIncrementVotes(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()
	// Test data
	userID := int64(10)

	// Expect check for existing user score
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_scores WHERE user_id = \$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Expect update votes
	mock.ExpectExec(`UPDATE user_scores SET vote_count = vote_count \+ 1, last_updated = \$1 WHERE user_id = \$2`).
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	err = repo.IncrementVotes(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryUpdateLastUpdated(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()

	// Test data
	userID := int64(10)

	// Expect check for existing user score
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_scores WHERE user_id = \$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Expect update last updated
	mock.ExpectExec(`UPDATE user_scores SET last_updated = \$1 WHERE user_id = \$2`).
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	err = repo.UpdateLastUpdated(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserScoreRepositoryUpdateLastUpdatedNewUser(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing db: %v", err)
		}
	}()
	// Test data
	userID := int64(10)

	// Expect check for existing user score (not exists)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_scores WHERE user_id = \$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Expect insert for new user score
	mock.ExpectExec(`INSERT INTO user_scores`).
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect update last updated
	mock.ExpectExec(`UPDATE user_scores SET last_updated = \$1 WHERE user_id = \$2`).
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserScoreRepository(db)

	// Execute test
	err = repo.UpdateLastUpdated(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
