// internal/service/data_service.go
package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
)

// DataService handles data processing and storage
type DataService struct {
	tokenRepo repository.TokenRepositoryInterface
	tradeRepo repository.TradeRepositoryInterface
	logger    *logger.Logger
}

// NewDataService creates a new data service
func NewDataService(db *sql.DB, logger *logger.Logger) *DataService {
	return &DataService{
		tokenRepo: repository.NewTokenRepository(db),
		tradeRepo: repository.NewTradeRepository(db),
		logger:    logger,
	}
}

// ProcessTokenData processes token data from WebSocket
func (s *DataService) ProcessTokenData(data map[string]interface{}) error {
	s.logger.Debug("Processing token data")

	// Extract required fields
	mintAddress, ok := data["mint"].(string)
	if !ok || mintAddress == "" {
		return fmt.Errorf("invalid mint address")
	}

	creatorAddress, ok := data["creator"].(string)
	if !ok || creatorAddress == "" {
		return fmt.Errorf("invalid creator address")
	}

	// Extract optional fields with defaults
	name := ""
	if nameVal, ok := data["name"].(string); ok {
		name = nameVal
	}

	symbol := ""
	if symbolVal, ok := data["symbol"].(string); ok {
		symbol = symbolVal
	}

	imageUrl := ""
	if val, ok := data["image_uri"].(string); ok {
		imageUrl = val
	}

	twitterUrl := ""
	if val, ok := data["twitter"].(string); ok {
		twitterUrl = val
	}

	websiteUrl := ""
	if val, ok := data["website"].(string); ok {
		websiteUrl = val
	}

	telegramUrl := ""
	if val, ok := data["telegram"].(string); ok {
		telegramUrl = val
	}

	metadataUrl := ""
	if val, ok := data["metadata_uri"].(string); ok {
		metadataUrl = val
	}

	var createdTimestamp int64
	if tsVal, ok := data["created_timestamp"].(float64); ok {
		createdTimestamp = int64(tsVal)
	}

	var marketCap float64
	if mcVal, ok := data["market_cap"].(float64); ok {
		marketCap = mcVal
	}

	var usdMarketCap float64
	if usdMcVal, ok := data["usd_market_cap"].(float64); ok {
		usdMarketCap = usdMcVal
	}

	var completed bool
	if val, ok := data["complete"].(bool); ok {
		completed = val
	}

	var kingOfTheHillTimestamp int64
	if val, ok := data["king_of_the_hill_timestamp"].(float64); ok {
		kingOfTheHillTimestamp = int64(val)
	}

	// Create token model
	token := &models.Token{
		MintAddress:            mintAddress,
		CreatorAddress:         creatorAddress,
		Name:                   name,
		Symbol:                 symbol,
		ImageUrl:               imageUrl,
		TwitterUrl:             twitterUrl,
		WebsiteUrl:             websiteUrl,
		TelegramUrl:            telegramUrl,
		MetadataUrl:            metadataUrl,
		CreatedTimestamp:       createdTimestamp,
		MarketCap:              marketCap,
		UsdMarketCap:           usdMarketCap,
		Completed:              completed,
		KingOfTheHillTimeStamp: kingOfTheHillTimestamp,
	}

	// Save token to database
	id, err := s.tokenRepo.Save(token)
	if err != nil {
		return fmt.Errorf("error saving token: %v", err)
	}

	s.logger.Info("Saved token: %s (ID: %d)", token.Name, id)
	return nil
}

// ProcessTradeData processes trade data from WebSocket
func (s *DataService) ProcessTradeData(data map[string]interface{}) error {
	s.logger.Debug("Processing trade data")

	// Extract required fields
	mintAddress, ok := data["mint"].(string)
	if !ok || mintAddress == "" {
		return fmt.Errorf("invalid mint address")
	}

	signature, ok := data["signature"].(string)
	if !ok || signature == "" {
		return fmt.Errorf("invalid signature")
	}

	// Get the token first to get its ID
	token, err := s.tokenRepo.GetByMintAddress(mintAddress)
	if err != nil {
		return fmt.Errorf("error getting token: %v", err)
	}

	if token == nil {
		// If token doesn't exist, try to process it
		if err := s.ProcessTokenData(data); err != nil {
			return fmt.Errorf("error processing token data: %v", err)
		}

		// Try to get the token again
		token, err = s.tokenRepo.GetByMintAddress(mintAddress)
		if err != nil || token == nil {
			return fmt.Errorf("token still not found after processing")
		}
	} else {
		// Check if we have market cap updates in the trade data
		var marketCap float64
		if mcVal, ok := data["market_cap"].(float64); ok {
			marketCap = mcVal
		}

		var usdMarketCap float64
		if usdMcVal, ok := data["usd_market_cap"].(float64); ok {
			usdMarketCap = usdMcVal
		}

		// If we have updated market cap info, update the token
		if marketCap > 0 || usdMarketCap > 0 {
			needsUpdate := false

			if marketCap > 0 && token.MarketCap != marketCap {
				token.MarketCap = marketCap
				needsUpdate = true
			}

			if usdMarketCap > 0 && token.UsdMarketCap != usdMarketCap {
				token.UsdMarketCap = usdMarketCap
				needsUpdate = true
			}

			if needsUpdate {
				// Update token with new market cap info
				_, err := s.tokenRepo.Save(token)
				if err != nil {
					s.logger.Error("Error updating token market cap: %v", err)
					// Continue processing trade even if token update fails
				} else {
					s.logger.Info("Updated token market cap for %s, USD: %f, SOL: %f",
						token.Symbol, token.UsdMarketCap, token.MarketCap)
				}
			}
		}
	}

	// Extract other trade fields
	var solAmount float64
	if solVal, ok := data["sol_amount"].(float64); ok {
		solAmount = solVal
	} else if solStr, ok := data["sol_amount"].(string); ok {
		solAmount, _ = strconv.ParseFloat(solStr, 64)
	} else {
		// Try to get from lamports by dividing by 10^9
		if lamports, ok := data["sol_amount"].(float64); ok {
			solAmount = lamports / 1000000000
		}
	}

	var tokenAmount float64
	if tokVal, ok := data["token_amount"].(float64); ok {
		tokenAmount = tokVal
	} else if tokStr, ok := data["token_amount"].(string); ok {
		tokenAmount, _ = strconv.ParseFloat(tokStr, 64)
	}

	var isBuy bool
	if buyVal, ok := data["is_buy"].(bool); ok {
		isBuy = buyVal
	}

	var userAddress string
	if userVal, ok := data["user"].(string); ok {
		userAddress = userVal
	}

	var timestamp int64
	if tsVal, ok := data["timestamp"].(float64); ok {
		timestamp = int64(tsVal)
	}

	// Create trade model
	trade := &models.Trade{
		TokenID:     token.ID,
		MintAddress: mintAddress,
		Signature:   signature,
		SolAmount:   solAmount,
		TokenAmount: tokenAmount,
		IsBuy:       isBuy,
		UserAddress: userAddress,
		Timestamp:   timestamp,
	}

	// Save trade to database
	id, err := s.tradeRepo.Save(trade)
	if err != nil {
		return fmt.Errorf("error saving trade: %v", err)
	}

	if id > 0 {
		s.logger.Info("Saved trade: %s (ID: %d)", signature, id)
	} else {
		s.logger.Debug("Trade already exists: %s", signature)
	}

	return nil
}
