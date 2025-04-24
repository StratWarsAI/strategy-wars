// internal/repository/comment_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
)

// CommentRepository handles database operations for comments
type CommentRepository struct {
	db *sql.DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Save inserts a comment into the database
func (r *CommentRepository) Save(comment *models.Comment) (int64, error) {
	query := `
		INSERT INTO comments 
			(strategy_id, user_id, parent_id, content, created_at, updated_at) 
		VALUES 
			($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`

	now := time.Now()
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = now
	}
	comment.UpdatedAt = now

	var id int64
	var parentID sql.NullInt64
	if comment.ParentID != nil {
		parentID.Int64 = *comment.ParentID
		parentID.Valid = true
	}

	err := r.db.QueryRow(
		query,
		comment.StrategyID,
		comment.UserID,
		parentID,
		comment.Content,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving comment: %v", err)
	}

	return id, nil
}

// GetByID retrieves a comment by ID
func (r *CommentRepository) GetByID(id int64) (*models.Comment, error) {
	query := `
		SELECT id, strategy_id, user_id, parent_id, content, created_at, updated_at
		FROM comments 
		WHERE id = $1
	`

	var comment models.Comment
	var parentID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&comment.ID,
		&comment.StrategyID,
		&comment.UserID,
		&parentID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No comment found
		}
		return nil, fmt.Errorf("error getting comment: %v", err)
	}

	if parentID.Valid {
		comment.ParentID = &parentID.Int64
	}

	return &comment, nil
}

// GetByStrategy retrieves comments for a strategy
func (r *CommentRepository) GetByStrategy(strategyID int64, limit, offset int) ([]*models.Comment, error) {
	query := `
		SELECT id, strategy_id, user_id, parent_id, content, created_at, updated_at
		FROM comments 
		WHERE strategy_id = $1 AND parent_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, strategyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting comments for strategy: %v", err)
	}
	defer rows.Close()

	return r.scanCommentRows(rows)
}

// GetByUser retrieves comments for a user
func (r *CommentRepository) GetByUser(userID int64, limit, offset int) ([]*models.Comment, error) {
	query := `
		SELECT id, strategy_id, user_id, parent_id, content, created_at, updated_at
		FROM comments 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting comments for user: %v", err)
	}
	defer rows.Close()

	return r.scanCommentRows(rows)
}

// GetReplies retrieves replies to a comment
func (r *CommentRepository) GetReplies(parentID int64, limit, offset int) ([]*models.Comment, error) {
	query := `
		SELECT id, strategy_id, user_id, parent_id, content, created_at, updated_at
		FROM comments 
		WHERE parent_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, parentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting replies for comment: %v", err)
	}
	defer rows.Close()

	return r.scanCommentRows(rows)
}

// Update updates a comment
func (r *CommentRepository) Update(comment *models.Comment) error {
	query := `
		UPDATE comments
		SET content = $1, updated_at = $2
		WHERE id = $3
	`

	comment.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		comment.Content,
		comment.UpdatedAt,
		comment.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating comment: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found: %d", comment.ID)
	}

	return nil
}

// Delete deletes a comment
func (r *CommentRepository) Delete(id int64) error {
	query := `DELETE FROM comments WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting comment: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found: %d", id)
	}

	return nil
}

// scanCommentRows is a helper function to scan multiple comment rows
func (r *CommentRepository) scanCommentRows(rows *sql.Rows) ([]*models.Comment, error) {
	var comments []*models.Comment

	for rows.Next() {
		var comment models.Comment
		var parentID sql.NullInt64

		if err := rows.Scan(
			&comment.ID,
			&comment.StrategyID,
			&comment.UserID,
			&parentID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning comment row: %v", err)
		}

		if parentID.Valid {
			comment.ParentID = &parentID.Int64
		}

		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comment rows: %v", err)
	}

	return comments, nil
}
