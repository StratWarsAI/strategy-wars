// internal/repository/notification_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// NotificationRepository handles database operations for notifications
type NotificationRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Save inserts a notification into the database
func (r *NotificationRepository) Save(notification *models.Notification) (int64, error) {
	query := `
		INSERT INTO notifications 
			(user_id, type, content, is_read, related_id, created_at) 
		VALUES 
			($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`

	now := time.Now()
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = now
	}

	var id int64
	var relatedID sql.NullInt64
	if notification.RelatedID != nil {
		relatedID.Int64 = *notification.RelatedID
		relatedID.Valid = true
	}

	err := r.db.QueryRow(
		query,
		notification.UserID,
		notification.Type,
		notification.Content,
		notification.IsRead,
		relatedID,
		notification.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving notification: %v", err)
	}

	return id, nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(id int64) (*models.Notification, error) {
	query := `
		SELECT id, user_id, type, content, is_read, related_id, created_at
		FROM notifications 
		WHERE id = $1
	`

	var notification models.Notification
	var relatedID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Content,
		&notification.IsRead,
		&relatedID,
		&notification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No notification found
		}
		return nil, fmt.Errorf("error getting notification: %v", err)
	}

	if relatedID.Valid {
		notification.RelatedID = &relatedID.Int64
	}

	return &notification, nil
}

// GetByUser retrieves notifications for a user
func (r *NotificationRepository) GetByUser(userID int64, limit, offset int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, type, content, is_read, related_id, created_at
		FROM notifications 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting notifications for user: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanNotificationRows(rows)
}

// GetUnreadByUser retrieves unread notifications for a user
func (r *NotificationRepository) GetUnreadByUser(userID int64) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, type, content, is_read, related_id, created_at
		FROM notifications 
		WHERE user_id = $1 AND is_read = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting unread notifications for user: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanNotificationRows(rows)
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(id int64) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error marking notification as read: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %d", id)
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (r *NotificationRepository) MarkAllAsRead(userID int64) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND is_read = false
	`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error marking all notifications as read: %v", err)
	}

	return nil
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(id int64) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting notification: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %d", id)
	}

	return nil
}

// scanNotificationRows is a helper function to scan multiple notification rows
func (r *NotificationRepository) scanNotificationRows(rows *sql.Rows) ([]*models.Notification, error) {
	var notifications []*models.Notification

	for rows.Next() {
		var notification models.Notification
		var relatedID sql.NullInt64

		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Content,
			&notification.IsRead,
			&relatedID,
			&notification.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning notification row: %v", err)
		}

		if relatedID.Valid {
			notification.RelatedID = &relatedID.Int64
		}

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification rows: %v", err)
	}

	return notifications, nil
}
