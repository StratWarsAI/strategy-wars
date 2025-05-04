// internal/repository/token_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
)

// TokenRepository handles database operations for tokens
type TokenRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// Save inserts or updates a token in the database
func (r *TokenRepository) Save(token *models.Token) (int64, error) {
	query := `
		INSERT INTO tokens 
			(mint_address, creator_address, name, symbol, image_url, twitter_url, website_url, telegram_url, metadata_url, created_timestamp, market_cap, usd_market_cap, completed, king_of_the_hill_timestamp) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) 
		ON CONFLICT (mint_address) 
		DO UPDATE SET 
			creator_address = $2, 
			name = $3, 
			symbol = $4,
			image_url = $5,
			twitter_url = $6,
			website_url = $7,
			telegram_url = $8,
			metadata_url = $9,
			created_timestamp = $10, 
			market_cap = $11,
			usd_market_cap = $12,
			completed = $13,
			king_of_the_hill_timestamp = $14
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
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
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error saving token: %v", err)
	}

	return id, nil
}

// GetByMintAddress retrieves a token by its mint address
func (r *TokenRepository) GetByMintAddress(mintAddress string) (*models.Token, error) {
	query := `
		SELECT id, mint_address, creator_address, name, symbol, image_url, twitter_url, website_url, telegram_url, metadata_url, created_timestamp, market_cap, usd_market_cap, completed, king_of_the_hill_timestamp, created_at
		FROM tokens 
		WHERE mint_address = $1
	`

	var token models.Token
	err := r.db.QueryRow(query, mintAddress).Scan(
		&token.ID,
		&token.MintAddress,
		&token.CreatorAddress,
		&token.Name,
		&token.Symbol,
		&token.ImageUrl,
		&token.TwitterUrl,
		&token.WebsiteUrl,
		&token.TelegramUrl,
		&token.MetadataUrl,
		&token.CreatedTimestamp,
		&token.MarketCap,
		&token.UsdMarketCap,
		&token.Completed,
		&token.KingOfTheHillTimeStamp,
		&token.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No token found
		}
		return nil, fmt.Errorf("error getting token: %v", err)
	}

	return &token, nil
}

// GetRecentTokens retrieves the most recent tokens
func (r *TokenRepository) GetRecentTokens(limit int) ([]*models.Token, error) {
	query := `
		SELECT id, mint_address, creator_address, name, symbol, image_url, twitter_url, website_url, telegram_url, metadata_url, 
			created_timestamp, market_cap, usd_market_cap, completed, king_of_the_hill_timestamp, created_at
		FROM tokens 
		ORDER BY created_timestamp DESC 
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting recent tokens: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	var tokens []*models.Token
	for rows.Next() {
		var token models.Token
		if err := rows.Scan(
			&token.ID,
			&token.MintAddress,
			&token.CreatorAddress,
			&token.Name,
			&token.Symbol,
			&token.ImageUrl,
			&token.TwitterUrl,
			&token.WebsiteUrl,
			&token.TelegramUrl,
			&token.MetadataUrl,
			&token.CreatedTimestamp,
			&token.MarketCap,
			&token.UsdMarketCap,
			&token.Completed,
			&token.KingOfTheHillTimeStamp,
			&token.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning token row: %v", err)
		}
		tokens = append(tokens, &token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating token rows: %v", err)
	}

	return tokens, nil
}

// GetFilteredTokens retrieves tokens based on market cap and time criteria
func (r *TokenRepository) GetFilteredTokens(minMarketCapUSD float64, maxAgeSeconds int64, limit int) ([]*models.Token, error) {
	// Calculate the minimum timestamp
	minTimestamp := (time.Now().Unix() - maxAgeSeconds) * 1000

	query := `
        SELECT id, mint_address, creator_address, name, symbol, image_url, twitter_url, website_url, telegram_url, metadata_url, 
            created_timestamp, market_cap, usd_market_cap, completed, king_of_the_hill_timestamp, created_at
        FROM tokens 
        WHERE usd_market_cap >= $1
        AND created_timestamp >= $2
        ORDER BY created_timestamp DESC 
        LIMIT $3
    `

	rows, err := r.db.Query(query, minMarketCapUSD, minTimestamp, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting filtered tokens: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	var tokens []*models.Token
	for rows.Next() {
		var token models.Token
		if err := rows.Scan(
			&token.ID,
			&token.MintAddress,
			&token.CreatorAddress,
			&token.Name,
			&token.Symbol,
			&token.ImageUrl,
			&token.TwitterUrl,
			&token.WebsiteUrl,
			&token.TelegramUrl,
			&token.MetadataUrl,
			&token.CreatedTimestamp,
			&token.MarketCap,
			&token.UsdMarketCap,
			&token.Completed,
			&token.KingOfTheHillTimeStamp,
			&token.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning token row: %v", err)
		}
		tokens = append(tokens, &token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating token rows: %v", err)
	}

	return tokens, nil
}

func (r *TokenRepository) GetByID(tokenID int64) (*models.Token, error) {
	query := `
		SELECT id, mint_address, creator_address, name, symbol, image_url, twitter_url, website_url, telegram_url, metadata_url, 
			created_timestamp, market_cap, usd_market_cap, completed, king_of_the_hill_timestamp, created_at
		FROM tokens 
		WHERE id = $1
	`

	rows, err := r.db.Query(query, tokenID)
	if err != nil {
		return nil, fmt.Errorf("error getting token by id: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Error closing rows: %v", err)
		}
	}()

	var token models.Token
	err = r.db.QueryRow(query, tokenID).Scan(
		&token.ID,
		&token.MintAddress,
		&token.CreatorAddress,
		&token.Name,
		&token.Symbol,
		&token.ImageUrl,
		&token.TwitterUrl,
		&token.WebsiteUrl,
		&token.TelegramUrl,
		&token.MetadataUrl,
		&token.CreatedTimestamp,
		&token.MarketCap,
		&token.UsdMarketCap,
		&token.Completed,
		&token.KingOfTheHillTimeStamp,
		&token.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting token by id: %v", err)
	}

	return &token, nil
}
