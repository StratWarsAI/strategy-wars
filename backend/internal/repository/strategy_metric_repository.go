// internal/repository/strategy_metric_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// StrategyMetricRepository handles database operations for strategy metrics
type StrategyMetricRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewStrategyMetricRepository creates a new strategy metric repository
func NewStrategyMetricRepository(db *sql.DB) *StrategyMetricRepository {
	return &StrategyMetricRepository{db: db}
}

// Save inserts a strategy metric into the database
func (r *StrategyMetricRepository) Save(metric *models.StrategyMetric) (int64, error) {
	query := `
		INSERT INTO strategy_metrics 
			(strategy_id, duel_id, win_rate, avg_profit, avg_loss, max_drawdown, 
			total_trades, successful_trades, risk_score, created_at) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		RETURNING id
	`

	now := time.Now()
	if metric.CreatedAt.IsZero() {
		metric.CreatedAt = now
	}

	var id int64
	var duelID sql.NullInt64
	if metric.DuelID != nil {
		duelID.Int64 = *metric.DuelID
		duelID.Valid = true
	}

	err := r.db.QueryRow(
		query,
		metric.StrategyID,
		duelID,
		metric.WinRate,
		metric.AvgProfit,
		metric.AvgLoss,
		metric.MaxDrawdown,
		metric.TotalTrades,
		metric.SuccessfulTrades,
		metric.RiskScore,
		metric.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving strategy metric: %v", err)
	}

	return id, nil
}

// GetByID retrieves a strategy metric by ID
func (r *StrategyMetricRepository) GetByID(id int64) (*models.StrategyMetric, error) {
	query := `
		SELECT id, strategy_id, duel_id, win_rate, avg_profit, avg_loss, max_drawdown, 
			total_trades, successful_trades, risk_score, created_at
		FROM strategy_metrics 
		WHERE id = $1
	`

	var metric models.StrategyMetric
	var duelID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&metric.ID,
		&metric.StrategyID,
		&duelID,
		&metric.WinRate,
		&metric.AvgProfit,
		&metric.AvgLoss,
		&metric.MaxDrawdown,
		&metric.TotalTrades,
		&metric.SuccessfulTrades,
		&metric.RiskScore,
		&metric.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No metric found
		}
		return nil, fmt.Errorf("error getting strategy metric: %v", err)
	}

	if duelID.Valid {
		metric.DuelID = &duelID.Int64
	}

	return &metric, nil
}

// GetByStrategy retrieves metrics for a strategy
func (r *StrategyMetricRepository) GetByStrategy(strategyID int64) ([]*models.StrategyMetric, error) {
	query := `
		SELECT id, strategy_id, duel_id, win_rate, avg_profit, avg_loss, max_drawdown, 
			total_trades, successful_trades, risk_score, created_at
		FROM strategy_metrics 
		WHERE strategy_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, strategyID)
	if err != nil {
		return nil, fmt.Errorf("error getting strategy metrics: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanMetricRows(rows)
}

// GetByDuel retrieves metrics for a duel
func (r *StrategyMetricRepository) GetByDuel(duelID int64) ([]*models.StrategyMetric, error) {
	query := `
		SELECT id, strategy_id, duel_id, win_rate, avg_profit, avg_loss, max_drawdown, 
			total_trades, successful_trades, risk_score, created_at
		FROM strategy_metrics 
		WHERE duel_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, duelID)
	if err != nil {
		return nil, fmt.Errorf("error getting duel metrics: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	return r.scanMetricRows(rows)
}

// GetLatestByStrategy retrieves the latest metric for a strategy
func (r *StrategyMetricRepository) GetLatestByStrategy(strategyID int64) (*models.StrategyMetric, error) {
	query := `
		SELECT id, strategy_id, duel_id, win_rate, avg_profit, avg_loss, max_drawdown, 
			total_trades, successful_trades, risk_score, created_at
		FROM strategy_metrics 
		WHERE strategy_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var metric models.StrategyMetric
	var duelID sql.NullInt64

	err := r.db.QueryRow(query, strategyID).Scan(
		&metric.ID,
		&metric.StrategyID,
		&duelID,
		&metric.WinRate,
		&metric.AvgProfit,
		&metric.AvgLoss,
		&metric.MaxDrawdown,
		&metric.TotalTrades,
		&metric.SuccessfulTrades,
		&metric.RiskScore,
		&metric.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No metric found
		}
		return nil, fmt.Errorf("error getting latest strategy metric: %v", err)
	}

	if duelID.Valid {
		metric.DuelID = &duelID.Int64
	}

	return &metric, nil
}

// scanMetricRows is a helper function to scan multiple metric rows
func (r *StrategyMetricRepository) scanMetricRows(rows *sql.Rows) ([]*models.StrategyMetric, error) {
	var metrics []*models.StrategyMetric

	for rows.Next() {
		var metric models.StrategyMetric
		var duelID sql.NullInt64

		if err := rows.Scan(
			&metric.ID,
			&metric.StrategyID,
			&duelID,
			&metric.WinRate,
			&metric.AvgProfit,
			&metric.AvgLoss,
			&metric.MaxDrawdown,
			&metric.TotalTrades,
			&metric.SuccessfulTrades,
			&metric.RiskScore,
			&metric.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning strategy metric row: %v", err)
		}

		if duelID.Valid {
			metric.DuelID = &duelID.Int64
		}

		metrics = append(metrics, &metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating strategy metric rows: %v", err)
	}

	return metrics, nil
}
