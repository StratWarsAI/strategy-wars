// internal/repository/simulated_trade_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// SimulatedTradeRepository handles database operations for simulated trades
type SimulatedTradeRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewSimulatedTradeRepository creates a new simulated trade repository
func NewSimulatedTradeRepository(db *sql.DB) *SimulatedTradeRepository {
	return &SimulatedTradeRepository{
		db:     db,
		logger: logger.New("simulated-trade-repository"),
	}
}

// Save inserts a new simulated trade into the database
func (r *SimulatedTradeRepository) Save(trade *models.SimulatedTrade) (int64, error) {
	// Use the context-based version with a default timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.SaveWithContext(ctx, trade)
}

// SaveWithContext inserts a new simulated trade with context for timeout control
func (r *SimulatedTradeRepository) SaveWithContext(ctx context.Context, trade *models.SimulatedTrade) (int64, error) {
	query := `
		INSERT INTO simulated_trades 
		(strategy_id, token_id, simulation_run_id, entry_price, exit_price, entry_timestamp, exit_timestamp, 
		position_size, profit_loss, status, exit_reason, entry_usd_market_cap, exit_usd_market_cap, created_at, updated_at) 
		VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`

	now := time.Now()
	var id int64

	// Handle simulation_run_id as a nullable field
	var simulationRunID sql.NullInt64
	if trade.SimulationRunID != nil {
		simulationRunID.Int64 = *trade.SimulationRunID
		simulationRunID.Valid = true
	}

	// Handle nullable exit_price
	var exitPrice sql.NullFloat64
	if trade.ExitPrice != nil {
		exitPrice.Float64 = *trade.ExitPrice
		exitPrice.Valid = true
	}

	// Handle nullable exit_timestamp
	var exitTimestamp sql.NullInt64
	if trade.ExitTimestamp != nil {
		exitTimestamp.Int64 = *trade.ExitTimestamp
		exitTimestamp.Valid = true
	}

	// Handle nullable profit_loss
	var profitLoss sql.NullFloat64
	if trade.ProfitLoss != nil {
		profitLoss.Float64 = *trade.ProfitLoss
		profitLoss.Valid = true
	}

	// Handle nullable exit_reason
	var exitReason sql.NullString
	if trade.ExitReason != nil {
		exitReason.String = *trade.ExitReason
		exitReason.Valid = true
	}

	// Handle nullable exit_usd_market_cap
	var exitUsdMarketCap sql.NullFloat64
	if trade.ExitUsdMarketCap != nil {
		exitUsdMarketCap.Float64 = *trade.ExitUsdMarketCap
		exitUsdMarketCap.Valid = true
	}

	// Use QueryRowContext with the provided context
	err := r.db.QueryRowContext(
		ctx,
		query,
		trade.StrategyID,
		trade.TokenID,
		simulationRunID,
		trade.EntryPrice,
		exitPrice,
		trade.EntryTimestamp,
		exitTimestamp,
		trade.PositionSize,
		profitLoss,
		trade.Status,
		exitReason,
		trade.EntryUsdMarketCap,
		exitUsdMarketCap,
		now,
		now,
	).Scan(&id)

	if err != nil {
		// More detailed error reporting
		if ctx.Err() != nil {
			r.logger.Error("Context deadline exceeded while saving simulated trade: %v", err)
			return 0, fmt.Errorf("context deadline exceeded while saving simulated trade: %v", err)
		}
		r.logger.Error("Error saving simulated trade: %v", err)
		return 0, fmt.Errorf("error saving simulated trade: %v", err)
	}

	r.logger.Info("Saved simulated trade ID %d for strategy %d", id, trade.StrategyID)
	return id, nil
}

// Update updates an existing simulated trade
func (r *SimulatedTradeRepository) Update(trade *models.SimulatedTrade) error {
	// Use the context-based version with a default timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.UpdateWithContext(ctx, trade)
}

// UpdateWithContext updates an existing simulated trade with context for timeout control
func (r *SimulatedTradeRepository) UpdateWithContext(ctx context.Context, trade *models.SimulatedTrade) error {
	query := `
		UPDATE simulated_trades
		SET exit_price = $1, 
			exit_timestamp = $2, 
			profit_loss = $3, 
			status = $4, 
			exit_reason = $5,
			exit_usd_market_cap = $6,
			updated_at = $7
		WHERE id = $8
	`

	now := time.Now()

	// Handle nullable fields
	var exitPrice sql.NullFloat64
	if trade.ExitPrice != nil {
		exitPrice.Float64 = *trade.ExitPrice
		exitPrice.Valid = true
	}

	var exitTimestamp sql.NullInt64
	if trade.ExitTimestamp != nil {
		exitTimestamp.Int64 = *trade.ExitTimestamp
		exitTimestamp.Valid = true
	}

	var profitLoss sql.NullFloat64
	if trade.ProfitLoss != nil {
		profitLoss.Float64 = *trade.ProfitLoss
		profitLoss.Valid = true
	}

	var exitReason sql.NullString
	if trade.ExitReason != nil {
		exitReason.String = *trade.ExitReason
		exitReason.Valid = true
	}

	var exitUsdMarketCap sql.NullFloat64
	if trade.ExitUsdMarketCap != nil {
		exitUsdMarketCap.Float64 = *trade.ExitUsdMarketCap
		exitUsdMarketCap.Valid = true
	}

	// Use ExecContext with the provided context
	result, err := r.db.ExecContext(
		ctx,
		query,
		exitPrice,
		exitTimestamp,
		profitLoss,
		trade.Status,
		exitReason,
		exitUsdMarketCap,
		now,
		trade.ID,
	)

	if err != nil {
		if ctx.Err() != nil {
			r.logger.Error("Context deadline exceeded while updating simulated trade: %v", err)
			return fmt.Errorf("context deadline exceeded while updating simulated trade: %v", err)
		}
		r.logger.Error("Error updating simulated trade: %v", err)
		return fmt.Errorf("error updating simulated trade: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error getting rows affected: %v", err)
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("No trade found with ID %d", trade.ID)
		return fmt.Errorf("no trade found with ID %d", trade.ID)
	}

	r.logger.Info("Updated simulated trade ID %d for strategy %d", trade.ID, trade.StrategyID)
	return nil
}

// scanTrade is a helper function to scan a trade row
func (r *SimulatedTradeRepository) scanTrade(rows *sql.Rows) (*models.SimulatedTrade, error) {
	var trade models.SimulatedTrade

	// For nullable fields
	var simulationRunID sql.NullInt64
	var exitPrice sql.NullFloat64
	var exitTimestamp sql.NullInt64
	var profitLoss sql.NullFloat64
	var exitReason sql.NullString
	var exitUsdMarketCap sql.NullFloat64

	err := rows.Scan(
		&trade.ID,
		&trade.StrategyID,
		&trade.TokenID,
		&simulationRunID,
		&trade.EntryPrice,
		&exitPrice,
		&trade.EntryTimestamp,
		&exitTimestamp,
		&trade.PositionSize,
		&profitLoss,
		&trade.Status,
		&exitReason,
		&trade.EntryUsdMarketCap,
		&exitUsdMarketCap,
		&trade.CreatedAt,
		&trade.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error scanning simulated trade: %v", err)
	}

	// Convert nullable fields to pointers
	if simulationRunID.Valid {
		trade.SimulationRunID = &simulationRunID.Int64
	}

	if exitPrice.Valid {
		trade.ExitPrice = &exitPrice.Float64
	}

	if exitTimestamp.Valid {
		trade.ExitTimestamp = &exitTimestamp.Int64
	}

	if profitLoss.Valid {
		trade.ProfitLoss = &profitLoss.Float64
	}

	if exitReason.Valid {
		trade.ExitReason = &exitReason.String
	}

	if exitUsdMarketCap.Valid {
		trade.ExitUsdMarketCap = &exitUsdMarketCap.Float64
	}

	return &trade, nil
}

// GetByStrategyID retrieves all simulated trades for a specific strategy
func (r *SimulatedTradeRepository) GetByStrategyID(strategyID int64) ([]*models.SimulatedTrade, error) {
	// Use the context-based version with a default timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return r.GetByStrategyIDWithContext(ctx, strategyID)
}

// GetByStrategyIDWithContext retrieves all simulated trades for a strategy with context
func (r *SimulatedTradeRepository) GetByStrategyIDWithContext(ctx context.Context, strategyID int64) ([]*models.SimulatedTrade, error) {
	query := `
		SELECT id, strategy_id, token_id, simulation_run_id, entry_price, exit_price, entry_timestamp, 
			exit_timestamp, position_size, profit_loss, status, exit_reason, 
			entry_usd_market_cap, exit_usd_market_cap, created_at, updated_at
		FROM simulated_trades
		WHERE strategy_id = $1
		ORDER BY entry_timestamp DESC
	`

	r.logger.Debug("Retrieving simulated trades for strategy %d", strategyID)

	// Use QueryContext with the provided context
	rows, err := r.db.QueryContext(ctx, query, strategyID)
	if err != nil {
		if ctx.Err() != nil {
			r.logger.Error("Context deadline exceeded while querying simulated trades: %v", err)
			return nil, fmt.Errorf("context deadline exceeded while querying simulated trades: %v", err)
		}
		r.logger.Error("Error querying simulated trades: %v", err)
		return nil, fmt.Errorf("error querying simulated trades: %v", err)
	}
	defer rows.Close()

	var trades []*models.SimulatedTrade
	for rows.Next() {
		trade, err := r.scanTrade(rows)
		if err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating through rows: %v", err)
		return nil, fmt.Errorf("error iterating through rows: %v", err)
	}

	r.logger.Info("Retrieved %d simulated trades for strategy %d", len(trades), strategyID)
	return trades, nil
}

// GetActiveByStrategyID retrieves active simulated trades for a strategy
func (r *SimulatedTradeRepository) GetActiveByStrategyID(strategyID int64) ([]*models.SimulatedTrade, error) {
	// Use the context-based version with a default timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.GetActiveByStrategyIDWithContext(ctx, strategyID)
}

// GetActiveByStrategyIDWithContext retrieves active simulated trades for a strategy with context
func (r *SimulatedTradeRepository) GetActiveByStrategyIDWithContext(ctx context.Context, strategyID int64) ([]*models.SimulatedTrade, error) {
	query := `
		SELECT id, strategy_id, token_id, simulation_run_id, entry_price, exit_price, entry_timestamp, 
			exit_timestamp, position_size, profit_loss, status, exit_reason, 
			entry_usd_market_cap, exit_usd_market_cap, created_at, updated_at
		FROM simulated_trades
		WHERE strategy_id = $1 AND status = 'active'
		ORDER BY entry_timestamp DESC
	`

	r.logger.Debug("Retrieving active simulated trades for strategy %d", strategyID)

	// Use QueryContext with the provided context
	rows, err := r.db.QueryContext(ctx, query, strategyID)
	if err != nil {
		if ctx.Err() != nil {
			r.logger.Error("Context deadline exceeded while querying active simulated trades: %v", err)
			return nil, fmt.Errorf("context deadline exceeded while querying active simulated trades: %v", err)
		}
		r.logger.Error("Error querying active simulated trades: %v", err)
		return nil, fmt.Errorf("error querying active simulated trades: %v", err)
	}
	defer rows.Close()

	var trades []*models.SimulatedTrade
	for rows.Next() {
		trade, err := r.scanTrade(rows)
		if err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating through rows: %v", err)
		return nil, fmt.Errorf("error iterating through rows: %v", err)
	}

	r.logger.Info("Retrieved %d active simulated trades for strategy %d", len(trades), strategyID)
	return trades, nil
}

// GetSummaryByStrategyID calculates summary statistics for a strategy
func (r *SimulatedTradeRepository) GetSummaryByStrategyID(strategyID int64) (map[string]interface{}, error) {
	// Use the context-based version with a default timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return r.GetSummaryByStrategyIDWithContext(ctx, strategyID)
}

// GetSummaryByStrategyIDWithContext calculates summary statistics with context
func (r *SimulatedTradeRepository) GetSummaryByStrategyIDWithContext(ctx context.Context, strategyID int64) (map[string]interface{}, error) {
	// Get all trades for this strategy
	trades, err := r.GetByStrategyIDWithContext(ctx, strategyID)
	if err != nil {
		return nil, err
	}

	if len(trades) == 0 {
		r.logger.Info("No simulated trades found for strategy %d", strategyID)
		return map[string]interface{}{
			"strategy_id":       strategyID,
			"total_trades":      0,
			"profitable_trades": 0,
			"losing_trades":     0,
			"win_rate":          0.0,
			"total_profit":      0.0,
			"total_loss":        0.0,
			"net_pnl":           0.0,
			"message":           "No simulated trades found for this strategy",
		}, nil
	}

	// Calculate summary statistics
	var totalTrades, profitableTrades, losingTrades int
	var totalProfit, totalLoss, totalInvestment float64
	var initialTimestamp, lastTimestamp int64

	for _, trade := range trades {
		// Track earliest and latest timestamps for duration calculation
		if initialTimestamp == 0 || trade.EntryTimestamp < initialTimestamp {
			initialTimestamp = trade.EntryTimestamp
		}

		var exitTimestamp int64
		if trade.ExitTimestamp != nil {
			exitTimestamp = *trade.ExitTimestamp
		} else {
			exitTimestamp = time.Now().Unix() // Use current time for active trades
		}

		if exitTimestamp > lastTimestamp {
			lastTimestamp = exitTimestamp
		}

		// Only count completed trades for performance metrics
		if trade.Status == "completed" || trade.Status == "closed" {
			totalTrades++
			totalInvestment += trade.PositionSize

			if trade.ProfitLoss != nil && *trade.ProfitLoss > 0 {
				profitableTrades++
				totalProfit += *trade.ProfitLoss
			} else if trade.ProfitLoss != nil {
				losingTrades++
				totalLoss += *trade.ProfitLoss // This will be negative
			}
		}
	}

	// Calculate derived metrics
	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(profitableTrades) / float64(totalTrades) * 100
	}

	roi := 0.0
	if totalInvestment > 0 {
		roi = (totalProfit + totalLoss) / totalInvestment * 100
	}

	// Calculate simulation duration in seconds
	var durationSec int64 = 0
	if initialTimestamp > 0 && lastTimestamp > 0 {
		durationSec = lastTimestamp - initialTimestamp
	}

	// Get strategy name with context
	var strategyName string
	err = r.db.QueryRowContext(ctx, "SELECT name FROM strategies WHERE id = $1", strategyID).Scan(&strategyName)
	if err != nil {
		strategyName = fmt.Sprintf("Strategy %d", strategyID)
		r.logger.Warn("Could not retrieve strategy name for ID %d: %v", strategyID, err)
	}

	// Calculate average holding time
	var totalHoldingTime int64
	completedTradeCount := 0

	for _, trade := range trades {
		if trade.Status == "completed" && trade.ExitTimestamp != nil {
			holdingTime := *trade.ExitTimestamp - trade.EntryTimestamp
			totalHoldingTime += holdingTime
			completedTradeCount++
		}
	}

	avgHoldingTimeSec := int64(0)
	if completedTradeCount > 0 {
		avgHoldingTimeSec = totalHoldingTime / int64(completedTradeCount)
	}

	r.logger.Info("Generated performance summary for strategy %d (win rate: %.2f%%)", strategyID, winRate)

	// Return enriched summary
	return map[string]interface{}{
		"strategy_id":                   strategyID,
		"strategy_name":                 strategyName,
		"total_trades":                  totalTrades,
		"profitable_trades":             profitableTrades,
		"losing_trades":                 losingTrades,
		"win_rate":                      winRate,
		"total_profit":                  totalProfit,
		"total_loss":                    totalLoss,
		"net_pnl":                       totalProfit + totalLoss,
		"roi":                           roi,
		"total_investment":              totalInvestment,
		"simulation_duration":           durationSec,
		"avg_holding_time_sec":          avgHoldingTimeSec,
		"total_trades_including_active": len(trades),
	}, nil
}

// DeleteByStrategyID deletes all simulated trades for a strategy
func (r *SimulatedTradeRepository) DeleteByStrategyID(strategyID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return r.DeleteByStrategyIDWithContext(ctx, strategyID)
}

// DeleteByStrategyIDWithContext deletes all simulated trades for a strategy with context
func (r *SimulatedTradeRepository) DeleteByStrategyIDWithContext(ctx context.Context, strategyID int64) error {
	query := `DELETE FROM simulated_trades WHERE strategy_id = $1`

	r.logger.Info("Deleting all simulated trades for strategy %d", strategyID)

	result, err := r.db.ExecContext(ctx, query, strategyID)
	if err != nil {
		if ctx.Err() != nil {
			r.logger.Error("Context deadline exceeded while deleting simulated trades: %v", err)
			return fmt.Errorf("context deadline exceeded while deleting simulated trades: %v", err)
		}
		r.logger.Error("Error deleting simulated trades: %v", err)
		return fmt.Errorf("error deleting simulated trades: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error getting rows affected: %v", err)
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("No simulated trades found for strategy %d", strategyID)
		return fmt.Errorf("no simulated trades found for strategy %d", strategyID)
	}

	r.logger.Info("Deleted %d simulated trades for strategy %d", rowsAffected, strategyID)
	return nil
}

// GetTradesByTokenID gets trades for a token
func (r *SimulatedTradeRepository) GetTradesByTokenID(tokenID int64, limit int) ([]*models.SimulatedTrade, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.GetTradesByTokenIDWithContext(ctx, tokenID, limit)
}

// GetTradesByTokenIDWithContext gets trades for a token with context
func (r *SimulatedTradeRepository) GetTradesByTokenIDWithContext(ctx context.Context, tokenID int64, limit int) ([]*models.SimulatedTrade, error) {
	query := `
		SELECT id, strategy_id, token_id, simulation_run_id, entry_price, exit_price, entry_timestamp, 
			exit_timestamp, position_size, profit_loss, status, exit_reason, 
			entry_usd_market_cap, exit_usd_market_cap, created_at, updated_at
		FROM simulated_trades
		WHERE token_id = $1
		ORDER BY entry_timestamp DESC
		LIMIT $2
	`

	r.logger.Debug("Retrieving simulated trades for token %d (limit: %d)", tokenID, limit)

	rows, err := r.db.QueryContext(ctx, query, tokenID, limit)
	if err != nil {
		if ctx.Err() != nil {
			r.logger.Error("Context deadline exceeded while querying token trades: %v", err)
			return nil, fmt.Errorf("context deadline exceeded while querying token trades: %v", err)
		}
		r.logger.Error("Error querying token trades: %v", err)
		return nil, fmt.Errorf("error querying token trades: %v", err)
	}
	defer rows.Close()

	var trades []*models.SimulatedTrade
	for rows.Next() {
		trade, err := r.scanTrade(rows)
		if err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating through rows: %v", err)
		return nil, fmt.Errorf("error iterating through rows: %v", err)
	}

	r.logger.Info("Retrieved %d simulated trades for token %d", len(trades), tokenID)
	return trades, nil
}

// GetBySimulationRun retrieves all simulated trades for a specific simulation run
func (r *SimulatedTradeRepository) GetBySimulationRun(simulationRunID int64) ([]*models.SimulatedTrade, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return r.GetBySimulationRunWithContext(ctx, simulationRunID)
}

// GetBySimulationRunWithContext retrieves all simulated trades for a simulation run with context
func (r *SimulatedTradeRepository) GetBySimulationRunWithContext(ctx context.Context, simulationRunID int64) ([]*models.SimulatedTrade, error) {
	query := `
		SELECT id, strategy_id, token_id, simulation_run_id, entry_price, exit_price, entry_timestamp, 
			exit_timestamp, position_size, profit_loss, status, exit_reason, 
			entry_usd_market_cap, exit_usd_market_cap, created_at, updated_at
		FROM simulated_trades
		WHERE simulation_run_id = $1
		ORDER BY strategy_id, entry_timestamp DESC
	`

	r.logger.Debug("Retrieving simulated trades for simulation run %d", simulationRunID)

	rows, err := r.db.QueryContext(ctx, query, simulationRunID)
	if err != nil {
		if ctx.Err() != nil {
			r.logger.Error("Context deadline exceeded while querying simulation trades: %v", err)
			return nil, fmt.Errorf("context deadline exceeded while querying simulation trades: %v", err)
		}
		r.logger.Error("Error querying simulation trades: %v", err)
		return nil, fmt.Errorf("error querying simulation trades: %v", err)
	}
	defer rows.Close()

	var trades []*models.SimulatedTrade
	for rows.Next() {
		trade, err := r.scanTrade(rows)
		if err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating through rows: %v", err)
		return nil, fmt.Errorf("error iterating through rows: %v", err)
	}

	r.logger.Info("Retrieved %d simulated trades for simulation run %d", len(trades), simulationRunID)
	return trades, nil
}
