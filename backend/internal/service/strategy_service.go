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
	strategyRepo       repository.StrategyRepositoryInterface
	strategyMetricRepo repository.StrategyMetricRepositoryInterface
	logger             *logger.Logger
}

// NewStrategyService creates a new strategy service
func NewStrategyService(
	strategyRepo repository.StrategyRepositoryInterface,
	strategyMetricRepo repository.StrategyMetricRepositoryInterface,
	logger *logger.Logger,
) StrategyServiceInterface {
	return &StrategyService{
		strategyRepo:       strategyRepo,
		strategyMetricRepo: strategyMetricRepo,
		logger:             logger,
	}
}

// CreateStrategy creates a new strategy
func (s *StrategyService) CreateStrategy(strategy *models.Strategy) (int64, error) {
	// Validate the strategy
	if err := s.validateStrategy(strategy); err != nil {
		return 0, err
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

	s.logger.Info("Created strategy with ID %d", id)
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

	s.logger.Info("Updated strategy with ID %d", strategy.ID)
	return nil
}

// DeleteStrategy deletes a strategy
func (s *StrategyService) DeleteStrategy(id int64) error {
	// Validate the strategy exists
	strategy, err := s.strategyRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("error checking strategy: %v", err)
	}

	if strategy == nil {
		return fmt.Errorf("strategy not found: %d", id)
	}

	// Delete strategy
	if err := s.strategyRepo.Delete(id); err != nil {
		return fmt.Errorf("error deleting strategy: %v", err)
	}

	s.logger.Info("Deleted strategy with ID %d", id)
	return nil
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
	case "performance":
		// New criteria for simulation performance
		return s.getTopPerformingStrategies(limit)
	default:
		return nil, fmt.Errorf("invalid criteria: %s", criteria)
	}
}

// getTopPerformingStrategies returns strategies with the best simulation performance
func (s *StrategyService) getTopPerformingStrategies(limit int) ([]*models.Strategy, error) {
	// Get all strategy metrics
	metrics, err := s.getTopPerformingMetrics(limit)
	if err != nil {
		return nil, fmt.Errorf("error getting top performing metrics: %v", err)
	}

	if len(metrics) == 0 {
		return []*models.Strategy{}, nil
	}

	// Get unique strategy IDs
	strategyIDs := make([]int64, 0, len(metrics))
	seen := make(map[int64]bool)
	for _, metric := range metrics {
		if !seen[metric.StrategyID] {
			strategyIDs = append(strategyIDs, metric.StrategyID)
			seen[metric.StrategyID] = true
		}
	}

	// Get strategies by IDs
	var strategies []*models.Strategy
	for _, id := range strategyIDs {
		strategy, err := s.strategyRepo.GetByID(id)
		if err != nil {
			s.logger.Error("Error getting strategy %d: %v", id, err)
			continue
		}
		if strategy != nil {
			strategies = append(strategies, strategy)
		}
		if len(strategies) >= limit {
			break
		}
	}

	return strategies, nil
}

// getTopPerformingMetrics returns metrics with the best performance
func (s *StrategyService) getTopPerformingMetrics(limit int) ([]*models.StrategyMetric, error) {
	// Get all public strategies
	strategies, err := s.strategyRepo.ListPublic(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("error listing strategies: %v", err)
	}

	// Get latest metrics for each strategy
	var allMetrics []*models.StrategyMetric
	for _, strategy := range strategies {
		metric, err := s.strategyMetricRepo.GetLatestByStrategy(strategy.ID)
		if err != nil {
			s.logger.Error("Error getting metrics for strategy %d: %v", strategy.ID, err)
			continue
		}
		if metric != nil {
			allMetrics = append(allMetrics, metric)
		}
	}

	// Sort metrics by performance (win rate in this case)
	sortMetricsByWinRate(allMetrics)

	// Limit the number of results
	if len(allMetrics) > limit {
		allMetrics = allMetrics[:limit]
	}

	return allMetrics, nil
}

// sortMetricsByWinRate sorts metrics by win rate in descending order
func sortMetricsByWinRate(metrics []*models.StrategyMetric) {
	for i := 0; i < len(metrics); i++ {
		for j := i + 1; j < len(metrics); j++ {
			if metrics[i].WinRate < metrics[j].WinRate {
				metrics[i], metrics[j] = metrics[j], metrics[i]
			}
		}
	}
}

// SearchStrategiesByTags searches strategies by tags
func (s *StrategyService) SearchStrategiesByTags(tags []string, limit int) ([]*models.Strategy, error) {
	return s.strategyRepo.SearchByTags(tags, limit)
}

// RecordWin records a win for a strategy in a simulation
func (s *StrategyService) RecordWin(strategyID int64, simulationID int64, winTime time.Time) error {
	// Update the strategy's win count and last win time
	if err := s.strategyRepo.IncrementWinCount(strategyID, winTime); err != nil {
		return fmt.Errorf("error incrementing win count: %v", err)
	}

	// Record metrics for the win
	metric := &models.StrategyMetric{
		StrategyID:      strategyID,
		SimulationRunID: &simulationID,
		WinRate:         1.0, // 100% win rate for this simulation
		CreatedAt:       time.Now(),
	}

	if _, err := s.strategyMetricRepo.Save(metric); err != nil {
		return fmt.Errorf("error saving strategy metric: %v", err)
	}

	return nil
}

// Private helper methods

// validateStrategy validates a strategy
func (s *StrategyService) validateStrategy(strategy *models.Strategy) error {
	if strategy.Name == "" {
		return fmt.Errorf("strategy name is required")
	}

	if len(strategy.Config) == 0 {
		return fmt.Errorf("strategy configuration is required")
	}

	// Validate config rules
	if err := s.validateStrategyConfig(strategy.Config); err != nil {
		return err
	}

	return nil
}

// validateStrategyConfig validates the strategy configuration
func (s *StrategyService) validateStrategyConfig(config models.JSONB) error {
	// Log the entire config for debugging
	s.logger.Info("Validating strategy config: %+v", config)

	// Check if required fields exist
	if _, ok := config["rules"]; !ok {
		s.logger.Error("Rules field missing in config")
		return fmt.Errorf("strategy must contain rules")
	}

	// Validate rules structure
	rulesInterface, ok := config["rules"]
	if !ok {
		s.logger.Error("Rules field missing in config (second check)")
		return fmt.Errorf("strategy must contain rules")
	}

	// Log the type of the rules field
	s.logger.Info("Rules field is of type: %T, value: %+v", rulesInterface, rulesInterface)

	// Try to handle both potential types: []interface{} and []map[string]interface{}
	var rules []interface{}
	var rulesLength int

	// First try as []interface{}
	if rulesAsInterface, ok := rulesInterface.([]interface{}); ok {
		rules = rulesAsInterface
		rulesLength = len(rulesAsInterface)
		s.logger.Info("Rules parsed as []interface{}, length: %d", rulesLength)
	} else if rulesAsMapSlice, ok := rulesInterface.([]map[string]interface{}); ok {
		// Convert []map[string]interface{} to []interface{} for compatibility
		rulesLength = len(rulesAsMapSlice)
		rules = make([]interface{}, rulesLength)
		for i, r := range rulesAsMapSlice {
			rules[i] = r
		}
		s.logger.Info("Rules parsed as []map[string]interface{}, length: %d", rulesLength)
	} else {
		s.logger.Error("Rules is not an array, type is: %T", rulesInterface)
		return fmt.Errorf("rules must be an array")
	}

	if len(rules) == 0 {
		s.logger.Error("Rules array is empty")
		return fmt.Errorf("strategy must contain at least one rule")
	}

	s.logger.Info("Rules array contains %d elements", len(rules))

	// Validate each rule
	for i, ruleInterface := range rules {
		s.logger.Info("Validating rule %d: %+v", i, ruleInterface)

		var rule map[string]interface{}

		// Try to get the rule as map[string]interface{} directly
		if r, ok := ruleInterface.(map[string]interface{}); ok {
			rule = r
		} else {
			s.logger.Error("Rule %d is not a valid object, type: %T", i, ruleInterface)
			return fmt.Errorf("rule %d is not a valid object", i)
		}

		// Check required fields
		condition, conditionOk := rule["condition"].(string)
		action, actionOk := rule["action"].(string)

		s.logger.Info("Rule %d - condition: '%v' (%t), action: '%v' (%t)",
			i, rule["condition"], conditionOk, rule["action"], actionOk)

		if !conditionOk || condition == "" {
			s.logger.Error("Rule %d has invalid condition: %v", i, rule["condition"])
			return fmt.Errorf("rule %d must have a valid condition", i)
		}

		if !actionOk || action == "" {
			s.logger.Error("Rule %d has invalid action: %v", i, rule["action"])
			return fmt.Errorf("rule %d must have a valid action", i)
		}
	}

	// Validate SimulationService required parameters
	if marketCapThreshold, ok := config["marketCapThreshold"].(float64); !ok || marketCapThreshold <= 0 {
		s.logger.Warn("marketCapThreshold missing or invalid, using default 5000")
		config["marketCapThreshold"] = 5000.0
	}

	if minBuysForEntry, ok := config["minBuysForEntry"].(float64); !ok || minBuysForEntry <= 0 {
		s.logger.Warn("minBuysForEntry missing or invalid, using default 3")
		config["minBuysForEntry"] = 3.0
	}

	if entryTimeWindowSec, ok := config["entryTimeWindowSec"].(float64); !ok || entryTimeWindowSec <= 0 {
		s.logger.Warn("entryTimeWindowSec missing or invalid, using default 300")
		config["entryTimeWindowSec"] = 300.0
	}

	if takeProfitPct, ok := config["takeProfitPct"].(float64); !ok || takeProfitPct <= 0 {
		s.logger.Warn("takeProfitPct missing or invalid, using default 30")
		config["takeProfitPct"] = 30.0
	}

	if stopLossPct, ok := config["stopLossPct"].(float64); !ok || stopLossPct <= 0 {
		s.logger.Warn("stopLossPct missing or invalid, using default 15")
		config["stopLossPct"] = 15.0
	}

	if maxHoldTimeSec, ok := config["maxHoldTimeSec"].(float64); !ok || maxHoldTimeSec <= 0 {
		s.logger.Warn("maxHoldTimeSec missing or invalid, using default 1800")
		config["maxHoldTimeSec"] = 1800.0
	}

	if fixedPositionSizeSol, ok := config["fixedPositionSizeSol"].(float64); !ok || fixedPositionSizeSol <= 0 {
		s.logger.Warn("fixedPositionSizeSol missing or invalid, using default 0.5")
		config["fixedPositionSizeSol"] = 0.5
	}

	if initialBalance, ok := config["initialBalance"].(float64); !ok || initialBalance <= 0 {
		s.logger.Warn("initialBalance missing or invalid, using default 10")
		config["initialBalance"] = 10.0
	}

	s.logger.Info("Strategy config validation successful")
	return nil
}

// calculateComplexityScore calculates a complexity score for the strategy
func (s *StrategyService) calculateComplexityScore(strategy *models.Strategy) int {
	// This would contain logic for calculating complexity
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
	// This would contain logic for calculating risk
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
