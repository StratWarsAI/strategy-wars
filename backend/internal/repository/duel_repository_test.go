// internal/repository/duel_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDuelRepositorySave(t *testing.T) {
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

	// Create test duel
	now := time.Now()
	startTime := now.Add(1 * time.Hour)
	endTime := now.Add(2 * time.Hour)
	votingEndTime := now.Add(90 * time.Minute)
	winnerStrategyID := int64(5)

	duel := &models.Duel{
		StartTime:        startTime,
		EndTime:          endTime,
		VotingEndTime:    votingEndTime,
		WinnerStrategyID: winnerStrategyID,
		Status:           "voting",
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO duels`).
		WithArgs(
			duel.StartTime,
			duel.EndTime,
			duel.VotingEndTime,
			sqlmock.AnyArg(), // winner_strategy_id - using AnyArg
			duel.Status,
			sqlmock.AnyArg(), // created_at - using AnyArg
			sqlmock.AnyArg(), // updated_at - using AnyArg
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	id, err := repo.Save(duel)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryGetByID(t *testing.T) {
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
	startTime := now.Add(1 * time.Hour)
	endTime := now.Add(2 * time.Hour)
	votingEndTime := now.Add(90 * time.Minute)
	winnerStrategyID := int64(5)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "start_time", "end_time", "voting_end_time", "winner_strategy_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			duelID, startTime, endTime, votingEndTime, winnerStrategyID, "voting", now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM duels WHERE id = \$1`).
		WithArgs(duelID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	duel, err := repo.GetByID(duelID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, duel)
	assert.Equal(t, duelID, duel.ID)
	assert.Equal(t, startTime, duel.StartTime)
	assert.Equal(t, endTime, duel.EndTime)
	assert.Equal(t, votingEndTime, duel.VotingEndTime)
	assert.Equal(t, winnerStrategyID, duel.WinnerStrategyID)
	assert.Equal(t, "voting", duel.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryGetCurrent(t *testing.T) {
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
	startTime := now.Add(-30 * time.Minute)
	endTime := now.Add(30 * time.Minute)
	votingEndTime := now.Add(15 * time.Minute)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "start_time", "end_time", "voting_end_time", "winner_strategy_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			duelID, startTime, endTime, votingEndTime, nil, "voting", now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM duels WHERE status = 'voting' OR status = 'simulating' ORDER BY start_time ASC LIMIT 1`).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	duel, err := repo.GetCurrent()

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, duel)
	assert.Equal(t, duelID, duel.ID)
	assert.Equal(t, "voting", duel.Status)
	assert.Equal(t, int64(0), duel.WinnerStrategyID) // No winner yet
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryGetUpcoming(t *testing.T) {
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
	future1 := now.Add(1 * time.Hour)
	end1 := future1.Add(10 * time.Minute)
	vote1 := future1.Add(5 * time.Minute)

	future2 := now.Add(2 * time.Hour)
	end2 := future2.Add(10 * time.Minute)
	vote2 := future2.Add(5 * time.Minute)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "start_time", "end_time", "voting_end_time", "winner_strategy_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			1, future1, end1, vote1, nil, "pending", now, now,
		).
		AddRow(
			2, future2, end2, vote2, nil, "pending", now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM duels WHERE status = 'pending' AND start_time > NOW\(\) ORDER BY start_time ASC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	duels, err := repo.GetUpcoming(limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, duels)
	assert.Equal(t, 2, len(duels))
	assert.Equal(t, future1, duels[0].StartTime)
	assert.Equal(t, future2, duels[1].StartTime)
	assert.Equal(t, "pending", duels[0].Status)
	assert.Equal(t, "pending", duels[1].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryGetPast(t *testing.T) {
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
	past1 := now.Add(-2 * time.Hour)
	end1 := past1.Add(10 * time.Minute)
	vote1 := past1.Add(5 * time.Minute)
	winner1 := int64(5)

	past2 := now.Add(-1 * time.Hour)
	end2 := past2.Add(10 * time.Minute)
	vote2 := past2.Add(5 * time.Minute)
	winner2 := int64(8)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "start_time", "end_time", "voting_end_time", "winner_strategy_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			1, past1, end1, vote1, winner1, "completed", now, now,
		).
		AddRow(
			2, past2, end2, vote2, winner2, "completed", now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM duels WHERE status = 'completed' ORDER BY end_time DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	duels, err := repo.GetPast(limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, duels)
	assert.Equal(t, 2, len(duels))
	assert.Equal(t, winner1, duels[0].WinnerStrategyID)
	assert.Equal(t, winner2, duels[1].WinnerStrategyID)
	assert.Equal(t, "completed", duels[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryGetByStatus(t *testing.T) {
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
	status := "voting"
	limit := 2
	now := time.Now()
	time1 := now.Add(10 * time.Minute)
	end1 := time1.Add(10 * time.Minute)
	vote1 := time1.Add(5 * time.Minute)

	time2 := now.Add(30 * time.Minute)
	end2 := time2.Add(10 * time.Minute)
	vote2 := time2.Add(5 * time.Minute)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "start_time", "end_time", "voting_end_time", "winner_strategy_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			1, time1, end1, vote1, nil, status, now, now,
		).
		AddRow(
			2, time2, end2, vote2, nil, status, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM duels WHERE status = \$1 ORDER BY start_time ASC LIMIT \$2`).
		WithArgs(status, limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	duels, err := repo.GetByStatus(status, limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, duels)
	assert.Equal(t, 2, len(duels))
	assert.Equal(t, status, duels[0].Status)
	assert.Equal(t, status, duels[1].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryGetByTimeRange(t *testing.T) {
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
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)

	duel1Start := now.Add(-30 * time.Minute)
	duel1End := duel1Start.Add(10 * time.Minute)
	duel1Vote := duel1Start.Add(5 * time.Minute)

	duel2Start := now.Add(30 * time.Minute)
	duel2End := duel2Start.Add(10 * time.Minute)
	duel2Vote := duel2Start.Add(5 * time.Minute)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "start_time", "end_time", "voting_end_time", "winner_strategy_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			1, duel1Start, duel1End, duel1Vote, nil, "completed", now, now,
		).
		AddRow(
			2, duel2Start, duel2End, duel2Vote, nil, "pending", now, now,
		)

	mock.ExpectQuery(
		`SELECT id, start_time, end_time, voting_end_time, winner_strategy_id, status, created_at, updated_at FROM duels WHERE \(start_time BETWEEN \$1 AND \$2\) OR \(end_time BETWEEN \$1 AND \$2\) ORDER BY start_time ASC`,
	).
		WithArgs(start, end).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	duels, err := repo.GetByTimeRange(start, end)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, duels)
	assert.Equal(t, 2, len(duels))
	assert.Equal(t, duel1Start, duels[0].StartTime)
	assert.Equal(t, duel2Start, duels[1].StartTime)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryUpdateStatus(t *testing.T) {
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
	newStatus := "simulating"

	// Setup expected query and result
	mock.ExpectExec(`UPDATE duels SET status = \$1, updated_at = \$2 WHERE id = \$3`).
		WithArgs(newStatus, sqlmock.AnyArg(), duelID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	err = repo.UpdateStatus(duelID, newStatus)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryUpdateWinner(t *testing.T) {
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
	strategyID := int64(5)

	// Setup expected query and result
	mock.ExpectExec(`UPDATE duels SET winner_strategy_id = \$1, status = 'completed', updated_at = \$2 WHERE id = \$3`).
		WithArgs(strategyID, sqlmock.AnyArg(), duelID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	err = repo.UpdateWinner(duelID, strategyID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryGetByIDNotFound(t *testing.T) {
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
	duelID := int64(999) // Non-existent ID

	// Setup expected query with no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM duels WHERE id = \$1`).
		WithArgs(duelID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "start_time", "end_time", "voting_end_time", "winner_strategy_id", "status", "created_at", "updated_at",
		}))

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	duel, err := repo.GetByID(duelID)

	// Assert results
	assert.NoError(t, err) // No error, just nil result
	assert.Nil(t, duel)    // Duel should be nil
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuelRepositoryGetCurrentWhenNoneActive(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Error closing database: %v", err)
		}
	}()

	// Setup expected query with no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM duels WHERE status = 'voting' OR status = 'simulating' ORDER BY start_time ASC LIMIT 1`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "start_time", "end_time", "voting_end_time", "winner_strategy_id", "status", "created_at", "updated_at",
		}))

	// Create repository with mock DB
	repo := NewDuelRepository(db)

	// Execute test
	duel, err := repo.GetCurrent()

	// Assert results
	assert.NoError(t, err) // No error, just nil result
	assert.Nil(t, duel)    // Duel should be nil when no active duel
	assert.NoError(t, mock.ExpectationsWereMet())
}
