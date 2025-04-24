// internal/repository/user_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestUserRepositorySave(t *testing.T) {
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

	// Create test user
	user := &models.User{
		Username:      "testuser",
		Email:         "test@example.com",
		PasswordHash:  "hashedpassword",
		WalletAddress: "wallet123",
		AvatarURL:     "https://example.com/avatar.jpg",
		IsActive:      true,
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(
			user.Username,
			user.Email,
			user.PasswordHash,
			user.WalletAddress,
			user.AvatarURL,
			user.IsActive,
			sqlmock.AnyArg(), // created_at - using AnyArg for time values
			sqlmock.AnyArg(), // updated_at - using AnyArg for time values
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewUserRepository(db)

	// Execute test
	id, err := repo.Save(user)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryGetByID(t *testing.T) {
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
	userID := int64(1)
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "wallet_address", "avatar_url", "is_active", "created_at", "updated_at",
	}).
		AddRow(
			userID, "testuser", "test@example.com", "hashedpassword", "wallet123", "https://example.com/avatar.jpg", true, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewUserRepository(db)

	// Execute test
	user, err := repo.GetByID(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryGetByUsername(t *testing.T) {
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
	username := "testuser"
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "wallet_address", "avatar_url", "is_active", "created_at", "updated_at",
	}).
		AddRow(
			1, username, "test@example.com", "hashedpassword", "wallet123", "https://example.com/avatar.jpg", true, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE username = \$1`).
		WithArgs(username).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewUserRepository(db)

	// Execute test
	user, err := repo.GetByUsername(username)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryGetByWalletAddress(t *testing.T) {
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
	walletAddress := "wallet123"
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "wallet_address", "avatar_url", "is_active", "created_at", "updated_at",
	}).
		AddRow(
			1, "testuser", "test@example.com", "hashedpassword", walletAddress, "https://example.com/avatar.jpg", true, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE wallet_address = \$1`).
		WithArgs(walletAddress).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewUserRepository(db)

	// Execute test
	user, err := repo.GetByWalletAddress(walletAddress)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, walletAddress, user.WalletAddress)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryUpdate(t *testing.T) {
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
	user := &models.User{
		ID:            1,
		Username:      "updateduser",
		Email:         "updated@example.com",
		PasswordHash:  "newhashedpassword",
		WalletAddress: "newwallet123",
		AvatarURL:     "https://example.com/newavatar.jpg",
		IsActive:      true,
	}

	// Setup expected query and result
	mock.ExpectExec(`UPDATE users SET`).
		WithArgs(
			user.Username,
			user.Email,
			user.PasswordHash,
			user.WalletAddress,
			user.AvatarURL,
			user.IsActive,
			sqlmock.AnyArg(), // updated_at - using AnyArg for time values
			user.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserRepository(db)

	// Execute test
	err = repo.Update(user)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryList(t *testing.T) {
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
	limit := 10
	offset := 0
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "wallet_address", "avatar_url", "is_active", "created_at", "updated_at",
	}).
		AddRow(
			1, "user1", "user1@example.com", "hash1", "wallet1", "https://example.com/avatar1.jpg", true, now, now,
		).
		AddRow(
			2, "user2", "user2@example.com", "hash2", "wallet2", "https://example.com/avatar2.jpg", true, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM users ORDER BY username LIMIT \$1 OFFSET \$2`).
		WithArgs(limit, offset).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewUserRepository(db)

	// Execute test
	users, err := repo.List(limit, offset)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Equal(t, 2, len(users))
	assert.Equal(t, "user1", users[0].Username)
	assert.Equal(t, "user2", users[1].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryDelete(t *testing.T) {
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
	userID := int64(1)

	// Setup expected query and result
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewUserRepository(db)

	// Execute test
	err = repo.Delete(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryGetByIDNotFound(t *testing.T) {
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
	userID := int64(999) // Non-existent ID

	// Setup expected query with no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "wallet_address", "avatar_url", "is_active", "created_at", "updated_at",
		}))

	// Create repository with mock DB
	repo := NewUserRepository(db)

	// Execute test
	user, err := repo.GetByID(userID)

	// Assert results
	assert.NoError(t, err) // No error, just nil result
	assert.Nil(t, user)    // User should be nil
	assert.NoError(t, mock.ExpectationsWereMet())
}
