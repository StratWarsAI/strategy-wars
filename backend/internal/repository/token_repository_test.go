// internal/repository/token_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestTokenRepositorySave(t *testing.T) {
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

	// Create test token
	token := &models.Token{
		MintAddress:            "test-mint-address",
		CreatorAddress:         "test-creator-address",
		Name:                   "Test Token",
		Symbol:                 "TEST",
		ImageUrl:               "http://example.com/image.png",
		TwitterUrl:             "http://twitter.com/test",
		WebsiteUrl:             "http://example.com",
		TelegramUrl:            "http://t.me/test",
		MetadataUrl:            "http://example.com/metadata.json",
		CreatedTimestamp:       time.Now().Unix(),
		MarketCap:              1000.0,
		UsdMarketCap:           50000.0,
		Completed:              false,
		KingOfTheHillTimeStamp: 0,
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO tokens`).
		WithArgs(
			token.MintAddress,
			token.CreatorAddress,
			token.Name,
			token.Symbol,
			token.ImageUrl,
			token.TwitterUrl,
			token.WebsiteUrl,
			token.TelegramUrl,
			token.MetadataUrl,
			token.CreatedTimestamp,
			token.MarketCap,
			token.UsdMarketCap,
			token.Completed,
			token.KingOfTheHillTimeStamp,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewTokenRepository(db)

	// Execute test
	id, err := repo.Save(token)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepositoryGetByMintAddress(t *testing.T) {
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
	mintAddress := "test-mint-address"
	createdTime := time.Now()
	timestamp := time.Now().Unix()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "mint_address", "creator_address", "name", "symbol", "image_url",
		"twitter_url", "website_url", "telegram_url", "metadata_url", "created_timestamp",
		"market_cap", "usd_market_cap", "completed", "king_of_the_hill_timestamp", "created_at",
	}).
		AddRow(
			1, mintAddress, "test-creator-address", "Test Token", "TEST", "http://example.com/image.png",
			"http://twitter.com/test", "http://example.com", "http://t.me/test", "http://example.com/metadata.json",
			timestamp, 1000.0, 50000.0, false, 0, createdTime,
		)

	mock.ExpectQuery(`SELECT (.+) FROM tokens WHERE mint_address = \$1`).
		WithArgs(mintAddress).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewTokenRepository(db)

	// Execute test
	token, err := repo.GetByMintAddress(mintAddress)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, int64(1), token.ID)
	assert.Equal(t, mintAddress, token.MintAddress)
	assert.Equal(t, "Test Token", token.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepositoryGetRecentTokens(t *testing.T) {
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
	createdTime := time.Now()
	timestamp := time.Now().Unix()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "mint_address", "creator_address", "name", "symbol", "image_url",
		"twitter_url", "website_url", "telegram_url", "metadata_url", "created_timestamp",
		"market_cap", "usd_market_cap", "completed", "king_of_the_hill_timestamp", "created_at",
	}).
		AddRow(
			1, "mint-address-1", "creator-address-1", "Token 1", "TOK1", "http://example.com/image1.png",
			"http://twitter.com/test1", "http://example1.com", "http://t.me/test1", "http://example.com/metadata1.json",
			timestamp, 1000.0, 50000.0, false, 0, createdTime,
		).
		AddRow(
			2, "mint-address-2", "creator-address-2", "Token 2", "TOK2", "http://example.com/image2.png",
			"http://twitter.com/test2", "http://example2.com", "http://t.me/test2", "http://example.com/metadata2.json",
			timestamp-100, 2000.0, 100000.0, false, 0, createdTime.Add(-1*time.Hour),
		)

	mock.ExpectQuery(`SELECT (.+) FROM tokens ORDER BY created_timestamp DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewTokenRepository(db)

	// Execute test
	tokens, err := repo.GetRecentTokens(limit)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.Equal(t, 2, len(tokens))
	assert.Equal(t, "Token 1", tokens[0].Name)
	assert.Equal(t, "Token 2", tokens[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
