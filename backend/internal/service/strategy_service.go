// internal/service/strategy_service.go
package service

import (
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
)

var _ StrategyServiceInterface = (*StrategyService)(nil)

// StrategyService handles business logic for strategies
type StrategyService struct {
	strategyRepo  repository.StrategyRepositoryInterface
	userRepo      repository.UserRepositoryInterface
	userScoreRepo repository.UserScoreRepositoryInterface
	logger        *logger.Logger
}

// NewStrategyService creates a new strategy service
func NewStrategyService(
	strategyRepo repository.StrategyRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	userScoreRepo repository.UserScoreRepositoryInterface,
	logger *logger.Logger,
) StrategyServiceInterface {
	return &StrategyService{
		strategyRepo:  strategyRepo,
		userRepo:      userRepo,
		userScoreRepo: userScoreRepo,
		logger:        logger,
	}
}

// CreateStrategy creates a new strategy
func (s *StrategyService) CreateStrategy(strategy *models.Strategy) (int64, error) {
	// Validate the strategy
	if err := s.validateStrategy(strategy); err != nil {
		return 0, err
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(strategy.UserID)
	if err != nil {
		return 0, fmt.Errorf("error checking user: %v", err)
	}
	if user == nil {
		return 0, fmt.Errorf("user not found: %d", strategy.UserID)
	}

	// Set initial values
	strategy.VoteCount = 0
	strategy.WinCount = 0
	strategy.IsPublic = true
	strategy.CreatedAt = time.Now()

	// Calculate complexity and risk scores if not provided
	if strategy.ComplexityScore == 0 {
		strategy.ComplexityScore = s.calculateComplexityScore(strategy)
	}

	if strategy.RiskScore == 0 {
		strategy.RiskScore = s.calculateRiskScore(strategy)
	}

	// Save strategy
	id, err := s.strategyRepo.Save(strategy)
	if err != nil {
		return 0, fmt.Errorf("error saving strategy: %v", err)
	}

	// Update user score (increment strategy count)
	if err := s.userScoreRepo.IncrementStrategies(strategy.UserID); err != nil {
		s.logger.Error("Error updating user score: %v", err)
		// Continue despite error in updating score
	}

	s.logger.Info("Created strategy with ID %d for user %d", id, strategy.UserID)
	return id, nil
}

// GetStrategyByID retrieves a strategy by ID
func (s *StrategyService) GetStrategyByID(id int64) (*models.Strategy, error) {
	strategy, err := s.strategyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("error getting strategy: %v", err)
	}

	if strategy == nil {
		return nil, nil // Not found
	}

	return strategy, nil
}

// UpdateStrategy updates an existing strategy
func (s *StrategyService) UpdateStrategy(strategy *models.Strategy) error {
	// Validate the strategy exists
	existingStrategy, err := s.strategyRepo.GetByID(strategy.ID)
	if err != nil {
		return fmt.Errorf("error checking strategy: %v", err)
	}

	if existingStrategy == nil {
		return fmt.Errorf("strategy not found: %d", strategy.ID)
	}

	// Ensure user owns the strategy
	if existingStrategy.UserID != strategy.UserID {
		return fmt.Errorf("user does not own this strategy")
	}

	// Validate the strategy
	if err := s.validateStrategy(strategy); err != nil {
		return err
	}

	// Preserve values that shouldn't be updated
	strategy.VoteCount = existingStrategy.VoteCount
	strategy.WinCount = existingStrategy.WinCount
	strategy.LastWinTime = existingStrategy.LastWinTime
	strategy.CreatedAt = existingStrategy.CreatedAt

	// Recalculate complexity and risk scores
	strategy.ComplexityScore = s.calculateComplexityScore(strategy)
	strategy.RiskScore = s.calculateRiskScore(strategy)

	// Update strategy
	if err := s.strategyRepo.Update(strategy); err != nil {
		return fmt.Errorf("error updating strategy: %v", err)
	}

	s.logger.Info("Updated strategy with ID %d for user %d", strategy.ID, strategy.UserID)
	return nil
}

// DeleteStrategy deletes a strategy
func (s *StrategyService) DeleteStrategy(id int64, userID int64) error {
	// Validate the strategy exists and belongs to user
	strategy, err := s.strategyRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("error checking strategy: %v", err)
	}

	if strategy == nil {
		return fmt.Errorf("strategy not found: %d", id)
	}

	// Ensure user owns the strategy
	if strategy.UserID != userID {
		return fmt.Errorf("user does not own this strategy")
	}

	// Delete strategy
	if err := s.strategyRepo.Delete(id); err != nil {
		return fmt.Errorf("error deleting strategy: %v", err)
	}

	s.logger.Info("Deleted strategy with ID %d for user %d", id, userID)
	return nil
}

// GetUserStrategies gets all strategies for a user
func (s *StrategyService) GetUserStrategies(userID int64, includePrivate bool, limit, offset int) ([]*models.Strategy, error) {
	return s.strategyRepo.ListByUser(userID, includePrivate, limit, offset)
}

// GetPublicStrategies gets public strategies
func (s *StrategyService) GetPublicStrategies(limit, offset int) ([]*models.Strategy, error) {
	return s.strategyRepo.ListPublic(limit, offset)
}

// GetTopStrategies gets top strategies by criteria
func (s *StrategyService) GetTopStrategies(criteria string, limit int) ([]*models.Strategy, error) {
	switch criteria {
	case "votes":
		return s.strategyRepo.GetTopVoted(limit)
	case "wins":
		return s.strategyRepo.GetTopWinners(limit)
	default:
		return nil, fmt.Errorf("invalid criteria: %s", criteria)
	}
}

// SearchStrategiesByTags searches strategies by tags
func (s *StrategyService) SearchStrategiesByTags(tags []string, limit int) ([]*models.Strategy, error) {
	return s.strategyRepo.SearchByTags(tags, limit)
}

// Private helper methods

// validateStrategy validates a strategy
func (s *StrategyService) validateStrategy(strategy *models.Strategy) error {
	if strategy.Name == "" {
		return fmt.Errorf("strategy name is required")
	}

	if strategy.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}

	if len(strategy.Config) == 0 {
		return fmt.Errorf("strategy configuration is required")
	}

	// Validate config rules (this would depend on your specific requirements)
	if err := s.validateStrategyConfig(strategy.Config); err != nil {
		return err
	}

	return nil
}

// validateStrategyConfig validates the strategy configuration
func (s *StrategyService) validateStrategyConfig(config models.JSONB) error {
	// This would contain your specific logic for validating strategy rules
	// For example:

	// Check if required fields exist
	if _, ok := config["rules"]; !ok {
		return fmt.Errorf("strategy must contain rules")
	}

	// Further validation based on your requirements

	return nil
}

// calculateComplexityScore calculates a complexity score for the strategy
func (s *StrategyService) calculateComplexityScore(strategy *models.Strategy) int {
	// This would contain your logic for calculating complexity
	// For example, based on number of rules, conditions, etc.

	// Placeholder implementation
	rulesInterface, ok := strategy.Config["rules"]
	if !ok {
		return 1 // Minimum score
	}

	rules, ok := rulesInterface.([]interface{})
	if !ok {
		return 1
	}

	// Base score on number of rules (simplified example)
	score := len(rules) + 1
	if score > 10 {
		score = 10 // Cap at 10
	}

	return score
}

// calculateRiskScore calculates a risk score for the strategy
func (s *StrategyService) calculateRiskScore(strategy *models.Strategy) int {
	// This would contain your logic for calculating risk
	// For example, based on stop loss settings, position size, etc.

	// Placeholder implementation
	riskLevel, ok := strategy.Config["risk_level"]
	if !ok {
		return 5 // Default medium risk
	}

	// Parse risk level string
	riskStr, ok := riskLevel.(string)
	if !ok {
		return 5
	}

	// Assign score based on risk level
	switch riskStr {
	case "low":
		return 2
	case "medium":
		return 5
	case "high":
		return 8
	default:
		return 5
	}
}
