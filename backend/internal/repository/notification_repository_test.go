// internal/repository/notification_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNotificationRepositorySave(t *testing.T) {
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

	// Create test notification
	now := time.Now()
	relatedID := int64(5)
	notification := &models.Notification{
		UserID:    10,
		Type:      "comment",
		Content:   "New comment on your strategy",
		IsRead:    false,
		RelatedID: &relatedID,
		CreatedAt: now,
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO notifications`).
		WithArgs(
			notification.UserID,
			notification.Type,
			notification.Content,
			notification.IsRead,
			sqlmock.AnyArg(), // related_id (NullInt64)
			notification.CreatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	id, err := repo.Save(notification)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryGetByID(t *testing.T) {
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
	notificationID := int64(1)
	now := time.Now()
	relatedID := int64(5)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "type", "content", "is_read", "related_id", "created_at",
	}).
		AddRow(notificationID, 10, "comment", "New comment", false, relatedID, now)

	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	notification, err := repo.GetByID(notificationID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, notification)
	assert.Equal(t, notificationID, notification.ID)
	assert.Equal(t, int64(10), notification.UserID)
	assert.Equal(t, "comment", notification.Type)
	assert.Equal(t, "New comment", notification.Content)
	assert.False(t, notification.IsRead)
	assert.Equal(t, &relatedID, notification.RelatedID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryGetByIDNotFound(t *testing.T) {
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
	notificationID := int64(999)

	// Setup expected query with no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "type", "content", "is_read", "related_id", "created_at",
		}))

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	notification, err := repo.GetByID(notificationID)

	// Assert results
	assert.NoError(t, err)
	assert.Nil(t, notification)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryGetByUser(t *testing.T) {
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
	limit := 10
	offset := 0
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "type", "content", "is_read", "related_id", "created_at",
	}).
		AddRow(1, userID, "comment", "Notification 1", false, int64(5), now).
		AddRow(2, userID, "vote", "Notification 2", true, int64(6), now.Add(1*time.Minute))

	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE user_id = \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	notifications, err := repo.GetByUser(userID, limit, offset)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, notifications)
	assert.Equal(t, 2, len(notifications))
	assert.Equal(t, int64(1), notifications[0].ID)
	assert.Equal(t, int64(2), notifications[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryGetUnreadByUser(t *testing.T) {
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
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "type", "content", "is_read", "related_id", "created_at",
	}).
		AddRow(1, userID, "comment", "Unread Notification 1", false, int64(5), now).
		AddRow(2, userID, "vote", "Unread Notification 2", false, int64(6), now.Add(1*time.Minute))

	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE user_id = \$1 AND is_read = false ORDER BY created_at DESC`).
		WithArgs(userID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	notifications, err := repo.GetUnreadByUser(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, notifications)
	assert.Equal(t, 2, len(notifications))
	assert.Equal(t, int64(1), notifications[0].ID)
	assert.Equal(t, int64(2), notifications[1].ID)
	assert.False(t, notifications[0].IsRead)
	assert.False(t, notifications[1].IsRead)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryMarkAsRead(t *testing.T) {
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
	notificationID := int64(1)

	// Setup expected query and result
	mock.ExpectExec(`UPDATE notifications SET is_read = true WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	err = repo.MarkAsRead(notificationID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryMarkAsReadNotFound(t *testing.T) {
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
	notificationID := int64(999)

	// Setup expected query and result (no rows affected)
	mock.ExpectExec(`UPDATE notifications SET is_read = true WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	err = repo.MarkAsRead(notificationID)

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryMarkAllAsRead(t *testing.T) {
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

	// Setup expected query and result
	mock.ExpectExec(`UPDATE notifications SET is_read = true WHERE user_id = \$1 AND is_read = false`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 2)) // Simulating 2 rows updated

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	err = repo.MarkAllAsRead(userID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryDelete(t *testing.T) {
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
	notificationID := int64(1)

	// Setup expected query and result
	mock.ExpectExec(`DELETE FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	err = repo.Delete(notificationID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepositoryDeleteNotFound(t *testing.T) {
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
	notificationID := int64(999)

	// Setup expected query and result (no rows affected)
	mock.ExpectExec(`DELETE FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Create repository with mock DB
	repo := NewNotificationRepository(db)

	// Execute test
	err = repo.Delete(notificationID)

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}
