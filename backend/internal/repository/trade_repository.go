// internal/repository/trade_repository.go
package repository

import (
	"database/sql"
	"fmt"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// TradeRepository handles database operations for trades
type TradeRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewTradeRepository creates a new trade repository
func NewTradeRepository(db *sql.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

// Save inserts a trade into the database
func (r *TradeRepository) Save(trade *models.Trade) (int64, error) {
	query := `
		INSERT INTO trades 
		    (token_id, signature, sol_amount, token_amount, is_buy, user_address, timestamp) 
		VALUES 
		    ($1, $2, $3, $4, $5, $6, $7) 
		ON CONFLICT (signature) DO NOTHING
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		trade.TokenID,
		trade.Signature,
		trade.SolAmount,
		trade.TokenAmount,
		trade.IsBuy,
		trade.UserAddress,
		trade.Timestamp,
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			// This happens when ON CONFLICT DO NOTHING is triggered
			return 0, nil
		}
		return 0, fmt.Errorf("error saving trade: %v", err)
	}

	return id, nil
}

// GetTradesByTokenID retrieves trades for a specific token
func (r *TradeRepository) GetTradesByTokenID(tokenID int64, limit int) ([]*models.Trade, error) {
	query := `
		SELECT id, token_id, signature, sol_amount, token_amount, is_buy, user_address, timestamp
		FROM trades 
		WHERE token_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2
	`

	rows, err := r.db.Query(query, tokenID, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting trades: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	var trades []*models.Trade
	for rows.Next() {
		var trade models.Trade
		if err := rows.Scan(
			&trade.ID,
			&trade.TokenID,
			&trade.Signature,
			&trade.SolAmount,
			&trade.TokenAmount,
			&trade.IsBuy,
			&trade.UserAddress,
			&trade.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("error scanning trade row: %v", err)
		}
		trades = append(trades, &trade)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trade rows: %v", err)
	}

	return trades, nil
}

// GetTradesBySignature retrieves a trade by its signature
func (r *TradeRepository) GetTradesBySignature(signature string) (*models.Trade, error) {
	query := `
		SELECT id, token_id, signature, sol_amount, token_amount, is_buy, user_address, timestamp
		FROM trades 
		WHERE signature = $1
	`

	var trade models.Trade
	err := r.db.QueryRow(query, signature).Scan(
		&trade.ID,
		&trade.TokenID,
		&trade.Signature,
		&trade.SolAmount,
		&trade.TokenAmount,
		&trade.IsBuy,
		&trade.UserAddress,
		&trade.Timestamp,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No trade found
		}
		return nil, fmt.Errorf("error getting trade: %v", err)
	}

	return &trade, nil
}
