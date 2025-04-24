// internal/repository/trade_repository_test.go
package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestTradeRepositorySave(t *testing.T) {
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
	// Create test trade
	trade := &models.Trade{
		TokenID:     1,
		Signature:   "test-signature-12345",
		SolAmount:   0.5,
		TokenAmount: 1000,
		IsBuy:       true,
		UserAddress: "user-wallet-address",
		Timestamp:   time.Now().Unix(),
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO trades`).
		WithArgs(
			trade.TokenID,
			trade.Signature,
			trade.SolAmount,
			trade.TokenAmount,
			trade.IsBuy,
			trade.UserAddress,
			trade.Timestamp,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewTradeRepository(db)

	// Execute test
	id, err := repo.Save(trade)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTradeRepositoryGetTradesByTokenID(t *testing.T) {
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
	tokenID := int64(1)
	limit := 10
	timestamp := time.Now().Unix()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "token_id", "signature", "sol_amount", "token_amount", "is_buy", "user_address", "timestamp",
	}).
		AddRow(
			1, tokenID, "signature-1", 0.5, 1000, true, "user-address-1", timestamp,
		).
		AddRow(
			2, tokenID, "signature-2", 0.3, 500, false, "user-address-2", timestamp-10,
		)

	mock.ExpectQuery(`SELECT (.+) FROM trades WHERE token_id = \$1 ORDER BY timestamp DESC LIMIT \$2`).
		WithArgs(tokenID, limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewTradeRepository(db)

	// Execute test
	trades, err := repo.GetTradesByTokenID(tokenID, limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, trades)
	assert.Equal(t, 2, len(trades))
	assert.Equal(t, "signature-1", trades[0].Signature)
	assert.Equal(t, "signature-2", trades[1].Signature)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTradeRepositoryGetTradesBySignature(t *testing.T) {
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
	signature := "test-signature-12345"
	tokenID := int64(1)
	timestamp := time.Now().Unix()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "token_id", "signature", "sol_amount", "token_amount", "is_buy", "user_address", "timestamp",
	}).
		AddRow(
			1, tokenID, signature, 0.5, 1000, true, "user-address-1", timestamp,
		)

	mock.ExpectQuery(`SELECT (.+) FROM trades WHERE signature = \$1`).
		WithArgs(signature).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewTradeRepository(db)

	// Execute test
	trade, err := repo.GetTradesBySignature(signature)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, trade)
	assert.Equal(t, signature, trade.Signature)
	assert.Equal(t, tokenID, trade.TokenID)
	assert.Equal(t, true, trade.IsBuy)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTradeRepositorySaveConflict(t *testing.T) {
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

	// Create test trade
	trade := &models.Trade{
		TokenID:     1,
		Signature:   "existing-signature",
		SolAmount:   0.5,
		TokenAmount: 1000,
		IsBuy:       true,
		UserAddress: "user-wallet-address",
		Timestamp:   time.Now().Unix(),
	}

	// Setup expected query with no rows returned (ON CONFLICT DO NOTHING case)
	mock.ExpectQuery(`INSERT INTO trades`).
		WithArgs(
			trade.TokenID,
			trade.Signature,
			trade.SolAmount,
			trade.TokenAmount,
			trade.IsBuy,
			trade.UserAddress,
			trade.Timestamp,
		).
		WillReturnError(sql.ErrNoRows)

	// Create repository with mock DB
	repo := NewTradeRepository(db)

	// Execute test
	id, err := repo.Save(trade)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(0), id) // Should return 0 for ID when ON CONFLICT DO NOTHING
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTradeRepositoryGetTradesBySignatureNotFound(t *testing.T) {
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
	signature := "non-existing-signature"

	// Setup expected query with no rows
	mock.ExpectQuery(`SELECT (.+) FROM trades WHERE signature = \$1`).
		WithArgs(signature).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "token_id", "signature", "sol_amount", "token_amount", "is_buy", "user_address", "timestamp",
		}))

	// Create repository with mock DB
	repo := NewTradeRepository(db)

	// Execute test
	trade, err := repo.GetTradesBySignature(signature)

	// Assert results
	assert.NoError(t, err)
	assert.Nil(t, trade) // Should return nil for non-existing trade
	assert.NoError(t, mock.ExpectationsWereMet())
}
