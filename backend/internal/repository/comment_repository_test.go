// internal/repository/comment_repository_test.go
package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCommentRepositorySave(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Create test comment
	now := time.Now()
	parentID := int64(5)
	comment := &models.Comment{
		StrategyID: 1,
		UserID:     10,
		ParentID:   &parentID,
		Content:    "Test comment",
		CreatedAt:  now,
	}

	// Setup expected query and result
	mock.ExpectQuery(`INSERT INTO comments`).
		WithArgs(
			comment.StrategyID,
			comment.UserID,
			sqlmock.AnyArg(), // parent_id (NullInt64)
			comment.Content,
			comment.CreatedAt,
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	id, err := repo.Save(comment)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryGetByID(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	commentID := int64(1)
	now := time.Now()
	parentID := int64(5)

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "strategy_id", "user_id", "parent_id", "content", "created_at", "updated_at",
	}).
		AddRow(commentID, 1, 10, parentID, "Test comment", now, now)

	mock.ExpectQuery(`SELECT (.+) FROM comments WHERE id = \$1`).
		WithArgs(commentID).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	comment, err := repo.GetByID(commentID)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, commentID, comment.ID)
	assert.Equal(t, int64(1), comment.StrategyID)
	assert.Equal(t, int64(10), comment.UserID)
	assert.Equal(t, &parentID, comment.ParentID)
	assert.Equal(t, "Test comment", comment.Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryGetByIDNotFound(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	commentID := int64(999)

	// Setup expected query with no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM comments WHERE id = \$1`).
		WithArgs(commentID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "strategy_id", "user_id", "parent_id", "content", "created_at", "updated_at",
		}))

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	comment, err := repo.GetByID(commentID)

	// Assert results
	assert.NoError(t, err)
	assert.Nil(t, comment)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryGetByStrategy(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	strategyID := int64(1)
	limit := 10
	offset := 0
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "strategy_id", "user_id", "parent_id", "content", "created_at", "updated_at",
	}).
		AddRow(1, strategyID, 10, nil, "Comment 1", now, now).
		AddRow(2, strategyID, 11, nil, "Comment 2", now.Add(1*time.Minute), now.Add(1*time.Minute))

	mock.ExpectQuery(`SELECT (.+) FROM comments WHERE strategy_id = \$1 AND parent_id IS NULL ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(strategyID, limit, offset).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	comments, err := repo.GetByStrategy(strategyID, limit, offset)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, comments)
	assert.Equal(t, 2, len(comments))
	assert.Equal(t, int64(1), comments[0].ID)
	assert.Equal(t, int64(2), comments[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryGetByUser(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	userID := int64(10)
	limit := 10
	offset := 0
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "strategy_id", "user_id", "parent_id", "content", "created_at", "updated_at",
	}).
		AddRow(1, 1, userID, nil, "Comment 1", now, now).
		AddRow(2, 2, userID, nil, "Comment 2", now.Add(1*time.Minute), now.Add(1*time.Minute))

	mock.ExpectQuery(`SELECT (.+) FROM comments WHERE user_id = \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	comments, err := repo.GetByUser(userID, limit, offset)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, comments)
	assert.Equal(t, 2, len(comments))
	assert.Equal(t, int64(1), comments[0].ID)
	assert.Equal(t, int64(2), comments[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryGetReplies(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	parentID := int64(1)
	limit := 10
	offset := 0
	now := time.Now()

	// Setup expected query and result
	rows := sqlmock.NewRows([]string{
		"id", "strategy_id", "user_id", "parent_id", "content", "created_at", "updated_at",
	}).
		AddRow(2, 1, 10, parentID, "Reply 1", now, now).
		AddRow(3, 1, 11, parentID, "Reply 2", now.Add(1*time.Minute), now.Add(1*time.Minute))

	mock.ExpectQuery(`SELECT (.+) FROM comments WHERE parent_id = \$1 ORDER BY created_at ASC LIMIT \$2 OFFSET \$3`).
		WithArgs(parentID, limit, offset).
		WillReturnRows(rows)

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	replies, err := repo.GetReplies(parentID, limit, offset)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, replies)
	assert.Equal(t, 2, len(replies))
	assert.Equal(t, int64(2), replies[0].ID)
	assert.Equal(t, int64(3), replies[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryUpdate(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	commentID := int64(1)
	comment := &models.Comment{
		ID:      commentID,
		Content: "Updated comment",
	}

	// Setup expected query and result
	mock.ExpectExec(`UPDATE comments SET content = \$1, updated_at = \$2 WHERE id = \$3`).
		WithArgs(
			comment.Content,
			sqlmock.AnyArg(), // updated_at
			commentID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	err = repo.Update(comment)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryUpdateNotFound(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	commentID := int64(999)
	comment := &models.Comment{
		ID:      commentID,
		Content: "Updated comment",
	}

	// Setup expected query and result (no rows affected)
	mock.ExpectExec(`UPDATE comments SET content = \$1, updated_at = \$2 WHERE id = \$3`).
		WithArgs(
			comment.Content,
			sqlmock.AnyArg(), // updated_at
			commentID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	err = repo.Update(comment)

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryDelete(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	commentID := int64(1)

	// Setup expected query and result
	mock.ExpectExec(`DELETE FROM comments WHERE id = \$1`).
		WithArgs(commentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	err = repo.Delete(commentID)

	// Assert results
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentRepositoryDeleteNotFound(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing rows: %v", err)
		}
	}()

	// Test data
	commentID := int64(999)

	// Setup expected query and result (no rows affected)
	mock.ExpectExec(`DELETE FROM comments WHERE id = \$1`).
		WithArgs(commentID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Create repository with mock DB
	repo := NewCommentRepository(db)

	// Execute test
	err = repo.Delete(commentID)

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}
