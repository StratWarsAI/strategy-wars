// internal/repository/repository_interfaces.go
package repository

import (
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
)

// TokenRepositoryInterface defines the interface for token repository operations
type TokenRepositoryInterface interface {
	Save(token *models.Token) (int64, error)
	GetByMintAddress(mintAddress string) (*models.Token, error)
	GetRecentTokens(limit int) ([]*models.Token, error)
	GetFilteredTokens(minMarketCapUSD float64, maxAgeSeconds int64, limit int) ([]*models.Token, error)
	GetByID(tokenID int64) (*models.Token, error)
}

// TradeRepositoryInterface defines the interface for trade repository operations
type TradeRepositoryInterface interface {
	Save(trade *models.Trade) (int64, error)
	GetTradesByTokenID(tokenID int64, limit int) ([]*models.Trade, error)
	GetTradesBySignature(signature string) (*models.Trade, error)
}

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	Save(user *models.User) (int64, error)
	GetByID(id int64) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByWalletAddress(walletAddress string) (*models.User, error)
	Update(user *models.User) error
	List(limit, offset int) ([]*models.User, error)
	Delete(id int64) error
}

// StrategyRepositoryInterface defines the interface for strategy repository operations
type StrategyRepositoryInterface interface {
	Save(strategy *models.Strategy) (int64, error)
	GetByID(id int64) (*models.Strategy, error)
	ListByUser(userID int64, includePrivate bool, limit, offset int) ([]*models.Strategy, error)
	ListPublic(limit, offset int) ([]*models.Strategy, error)
	Update(strategy *models.Strategy) error
	Delete(id int64) error
	IncrementVoteCount(id int64) error
	IncrementWinCount(id int64, winTime time.Time) error
	SearchByTags(tags []string, limit int) ([]*models.Strategy, error)
	GetTopVoted(limit int) ([]*models.Strategy, error)
	GetTopWinners(limit int) ([]*models.Strategy, error)
}
