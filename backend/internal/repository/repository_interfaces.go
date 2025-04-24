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

// DuelRepositoryInterface defines the interface for duel repository operations
type DuelRepositoryInterface interface {
	Save(duel *models.Duel) (int64, error)
	GetByID(id int64) (*models.Duel, error)
	GetCurrent() (*models.Duel, error)
	GetUpcoming(limit int) ([]*models.Duel, error)
	GetPast(limit int) ([]*models.Duel, error)
	GetByStatus(status string, limit int) ([]*models.Duel, error)
	GetByTimeRange(start, end time.Time) ([]*models.Duel, error)
	UpdateStatus(id int64, status string) error
	UpdateWinner(id int64, strategyID int64) error
}

// VoteRepositoryInterface defines the interface for vote repository operations
type VoteRepositoryInterface interface {
	Save(vote *models.Vote) (int64, error)
	GetByUserAndDuel(userID, duelID int64) (*models.Vote, error)
	GetByDuel(duelID int64) ([]*models.Vote, error)
	GetVoteCounts(duelID int64) (map[int64]int, error)
	GetVoteCountsForStrategy(strategyID int64) (int, error)
	DeleteByUserAndDuel(userID, duelID int64) error
}

// UserScoreRepositoryInterface defines the interface for user score repository operations
type UserScoreRepositoryInterface interface {
	GetByUserID(userID int64) (*models.UserScore, error)
	GetTopUsers(limit int) ([]*models.UserScore, error)
	IncrementPoints(userID int64, points int) error
	IncrementWins(userID int64) error
	IncrementStrategies(userID int64) error
	IncrementVotes(userID int64) error
	UpdateLastUpdated(userID int64) error
}

// NotificationRepositoryInterface defines the interface for notification repository operations
type NotificationRepositoryInterface interface {
	Save(notification *models.Notification) (int64, error)
	GetByID(id int64) (*models.Notification, error)
	GetByUser(userID int64, limit, offset int) ([]*models.Notification, error)
	GetUnreadByUser(userID int64) ([]*models.Notification, error)
	MarkAsRead(id int64) error
	MarkAllAsRead(userID int64) error
	Delete(id int64) error
}
