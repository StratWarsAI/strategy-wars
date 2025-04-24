// internal/repository/strategy_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestStrategyRepositorySave(t *testing.T) {
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

	// Create test strategy
	strategy := &models.Strategy{
		Name:            "Test Strategy",
		Description:     "A test strategy",
		Config:          models.JSONB{"key": "value"},
		UserID:          1,
		IsPublic:        true,
		VoteCount:       0,
		WinCount:        0,
		Tags:            []string{"test", "strategy"},
		ComplexityScore: 5,
		RiskScore:       3,
		AIEnhanced:      false,
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO strategies`).
		WithArgs(
			strategy.Name,
			strategy.Description,
			strategy.Config,
			strategy.UserID,
			strategy.IsPublic,
			strategy.VoteCount,
			strategy.WinCount,
			sqlmock.AnyArg(), // LastWinTime as AnyArg
			pq.Array(strategy.Tags),
			strategy.ComplexityScore,
			strategy.RiskScore,
			strategy.AIEnhanced,
			sqlmock.AnyArg(), // CreatedAt as AnyArg
			sqlmock.AnyArg(), // UpdatedAt as AnyArg
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	id, err := repo.Save(strategy)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryGetByID(t *testing.T) {
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
	strategyID := int64(1)
	now := time.Now()
	lastWinTime := now.Add(-24 * time.Hour)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "config", "user_id", "is_public", "vote_count", "win_count",
		"last_win_time", "tags", "complexity_score", "risk_score", "ai_enhanced", "created_at", "updated_at",
	}).
		AddRow(
			strategyID, "Test Strategy", "A test strategy", `{"key":"value"}`, 1, true, 10, 5,
			lastWinTime, pq.Array([]string{"test", "strategy"}), 5, 3, false, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM strategies WHERE id = \$1`).
		WithArgs(strategyID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	strategy, err := repo.GetByID(strategyID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, strategy)
	assert.Equal(t, strategyID, strategy.ID)
	assert.Equal(t, "Test Strategy", strategy.Name)
	assert.Equal(t, int64(1), strategy.UserID)
	assert.Equal(t, 10, strategy.VoteCount)
	assert.Equal(t, 5, strategy.WinCount)
	assert.Equal(t, 2, len(strategy.Tags))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryListByUser(t *testing.T) {
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
	userID := int64(1)
	includePrivate := true
	limit := 10
	offset := 0
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "config", "user_id", "is_public", "vote_count", "win_count",
		"last_win_time", "tags", "complexity_score", "risk_score", "ai_enhanced", "created_at", "updated_at",
	}).
		AddRow(
			1, "Strategy 1", "Public strategy", `{"key":"value1"}`, userID, true, 5, 2,
			now, pq.Array([]string{"public"}), 5, 3, false, now, now,
		).
		AddRow(
			2, "Strategy 2", "Private strategy", `{"key":"value2"}`, userID, false, 0, 0,
			time.Time{}, pq.Array([]string{"private"}), 7, 8, true, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM strategies WHERE user_id = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	strategies, err := repo.ListByUser(userID, includePrivate, limit, offset)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, strategies)
	assert.Equal(t, 2, len(strategies))
	assert.Equal(t, "Strategy 1", strategies[0].Name)
	assert.Equal(t, "Strategy 2", strategies[1].Name)
	assert.True(t, strategies[0].IsPublic)
	assert.False(t, strategies[1].IsPublic)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryListPublic(t *testing.T) {
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
	limit := 10
	offset := 0
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "config", "user_id", "is_public", "vote_count", "win_count",
		"last_win_time", "tags", "complexity_score", "risk_score", "ai_enhanced", "created_at", "updated_at",
	}).
		AddRow(
			1, "Public Strategy 1", "First public strategy", `{"key":"value1"}`, 1, true, 5, 2,
			now, pq.Array([]string{"public", "first"}), 5, 3, false, now, now,
		).
		AddRow(
			2, "Public Strategy 2", "Second public strategy", `{"key":"value2"}`, 2, true, 10, 4,
			now, pq.Array([]string{"public", "second"}), 7, 8, true, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM strategies WHERE is_public = true ORDER BY updated_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(limit, offset).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	strategies, err := repo.ListPublic(limit, offset)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, strategies)
	assert.Equal(t, 2, len(strategies))
	assert.Equal(t, "Public Strategy 1", strategies[0].Name)
	assert.Equal(t, "Public Strategy 2", strategies[1].Name)
	assert.True(t, strategies[0].IsPublic)
	assert.True(t, strategies[1].IsPublic)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryUpdate(t *testing.T) {
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
	now := time.Now()
	strategy := &models.Strategy{
		ID:              1,
		Name:            "Updated Strategy",
		Description:     "Updated description",
		Config:          models.JSONB{"key": "updated"},
		UserID:          1,
		IsPublic:        true,
		VoteCount:       5,
		WinCount:        2,
		LastWinTime:     now,
		Tags:            []string{"updated", "tags"},
		ComplexityScore: 6,
		RiskScore:       4,
		AIEnhanced:      true,
	}

	// Setup expected query and result
	mock.ExpectExec(`UPDATE strategies SET`).
		WithArgs(
			strategy.Name,
			strategy.Description,
			strategy.Config,
			strategy.UserID,
			strategy.IsPublic,
			strategy.VoteCount,
			strategy.WinCount,
			strategy.LastWinTime,
			pq.Array(strategy.Tags),
			strategy.ComplexityScore,
			strategy.RiskScore,
			strategy.AIEnhanced,
			sqlmock.AnyArg(), // updated_at changes
			strategy.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	err = repo.Update(strategy)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryDelete(t *testing.T) {
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
	strategyID := int64(1)

	// Setup expected query and result
	mock.ExpectExec(`DELETE FROM strategies WHERE id = \$1`).
		WithArgs(strategyID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	err = repo.Delete(strategyID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryIncrementVoteCount(t *testing.T) {
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
	strategyID := int64(1)

	// Setup expected query and result
	mock.ExpectExec(`UPDATE strategies SET vote_count = vote_count \+ 1, updated_at = \$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), strategyID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	err = repo.IncrementVoteCount(strategyID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryIncrementWinCount(t *testing.T) {
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
	strategyID := int64(1)
	winTime := time.Now()

	// Setup expected query and result
	mock.ExpectExec(`UPDATE strategies SET win_count = win_count \+ 1, last_win_time = \$1, updated_at = \$2 WHERE id = \$3`).
		WithArgs(winTime, sqlmock.AnyArg(), strategyID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	err = repo.IncrementWinCount(strategyID, winTime)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositorySearchByTags(t *testing.T) {
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
	tags := []string{"ai", "trading"}
	limit := 5
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "config", "user_id", "is_public", "vote_count", "win_count",
		"last_win_time", "tags", "complexity_score", "risk_score", "ai_enhanced", "created_at", "updated_at",
	}).
		AddRow(
			1, "AI Trading Strategy", "Uses AI", `{"key":"value1"}`, 1, true, 5, 2,
			now, pq.Array([]string{"ai", "trading"}), 5, 3, true, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM strategies WHERE is_public = true AND tags && \$1 ORDER BY updated_at DESC LIMIT \$2`).
		WithArgs(pq.Array(tags), limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	strategies, err := repo.SearchByTags(tags, limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, strategies)
	assert.Equal(t, 1, len(strategies))
	assert.Equal(t, "AI Trading Strategy", strategies[0].Name)
	assert.Contains(t, strategies[0].Tags, "ai")
	assert.Contains(t, strategies[0].Tags, "trading")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryGetTopVoted(t *testing.T) {
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
	limit := 2
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "config", "user_id", "is_public", "vote_count", "win_count",
		"last_win_time", "tags", "complexity_score", "risk_score", "ai_enhanced", "created_at", "updated_at",
	}).
		AddRow(
			1, "Popular Strategy", "Most votes", `{"key":"value1"}`, 1, true, 50, 5,
			now, pq.Array([]string{"popular"}), 5, 3, false, now, now,
		).
		AddRow(
			2, "Second Popular", "Second most votes", `{"key":"value2"}`, 2, true, 30, 3,
			now, pq.Array([]string{"popular"}), 4, 5, false, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM strategies WHERE is_public = true ORDER BY vote_count DESC, updated_at DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	strategies, err := repo.GetTopVoted(limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, strategies)
	assert.Equal(t, 2, len(strategies))
	assert.Equal(t, "Popular Strategy", strategies[0].Name)
	assert.Equal(t, "Second Popular", strategies[1].Name)
	assert.Equal(t, 50, strategies[0].VoteCount)
	assert.Equal(t, 30, strategies[1].VoteCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStrategyRepositoryGetTopWinners(t *testing.T) {
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
	limit := 2
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "config", "user_id", "is_public", "vote_count", "win_count",
		"last_win_time", "tags", "complexity_score", "risk_score", "ai_enhanced", "created_at", "updated_at",
	}).
		AddRow(
			1, "Winning Strategy", "Most wins", `{"key":"value1"}`, 1, true, 20, 15,
			now, pq.Array([]string{"winning"}), 5, 3, false, now, now,
		).
		AddRow(
			2, "Second Winner", "Second most wins", `{"key":"value2"}`, 2, true, 15, 10,
			now, pq.Array([]string{"winning"}), 4, 5, false, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM strategies WHERE is_public = true ORDER BY win_count DESC, updated_at DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewStrategyRepository(db)

	// Execute test
	strategies, err := repo.GetTopWinners(limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, strategies)
	assert.Equal(t, 2, len(strategies))
	assert.Equal(t, "Winning Strategy", strategies[0].Name)
	assert.Equal(t, "Second Winner", strategies[1].Name)
	assert.Equal(t, 15, strategies[0].WinCount)
	assert.Equal(t, 10, strategies[1].WinCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}
