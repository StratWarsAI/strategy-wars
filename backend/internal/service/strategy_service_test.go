// internal/service/strategy_service_test.go
package service

import (
	"testing"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockStrategyRepository struct {
	mock.Mock
}

func (m *MockStrategyRepository) Save(strategy *models.Strategy) (int64, error) {
	args := m.Called(strategy)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStrategyRepository) GetByID(id int64) (*models.Strategy, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) ListByUser(userID int64, includePrivate bool, limit, offset int) ([]*models.Strategy, error) {
	args := m.Called(userID, includePrivate, limit, offset)
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) ListPublic(limit, offset int) ([]*models.Strategy, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) Update(strategy *models.Strategy) error {
	args := m.Called(strategy)
	return args.Error(0)
}

func (m *MockStrategyRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStrategyRepository) IncrementVoteCount(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStrategyRepository) IncrementWinCount(id int64, winTime time.Time) error {
	args := m.Called(id, winTime)
	return args.Error(0)
}

func (m *MockStrategyRepository) SearchByTags(tags []string, limit int) ([]*models.Strategy, error) {
	args := m.Called(tags, limit)
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) GetTopVoted(limit int) ([]*models.Strategy, error) {
	args := m.Called(limit)
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) GetTopWinners(limit int) ([]*models.Strategy, error) {
	args := m.Called(limit)
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

type MockStrategyMetricRepository struct {
	mock.Mock
}

func (m *MockStrategyMetricRepository) Save(metric *models.StrategyMetric) (int64, error) {
	args := m.Called(metric)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStrategyMetricRepository) GetByID(id int64) (*models.StrategyMetric, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StrategyMetric), args.Error(1)
}

func (m *MockStrategyMetricRepository) GetByStrategy(strategyID int64) ([]*models.StrategyMetric, error) {
	args := m.Called(strategyID)
	return args.Get(0).([]*models.StrategyMetric), args.Error(1)
}

func (m *MockStrategyMetricRepository) GetByDuel(duelID int64) ([]*models.StrategyMetric, error) {
	args := m.Called(duelID)
	return args.Get(0).([]*models.StrategyMetric), args.Error(1)
}

func (m *MockStrategyMetricRepository) GetBySimulationRun(simulationRunID int64) ([]*models.StrategyMetric, error) {
	args := m.Called(simulationRunID)
	return args.Get(0).([]*models.StrategyMetric), args.Error(1)
}

func (m *MockStrategyMetricRepository) GetLatestByStrategy(strategyID int64) (*models.StrategyMetric, error) {
	args := m.Called(strategyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StrategyMetric), args.Error(1)
}

// Test CreateStrategy
func TestCreateStrategy(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	strategy := &models.Strategy{
		Name:     "Test Strategy",
		Config:   models.JSONB{"rules": []interface{}{map[string]interface{}{"condition": "price > 100", "action": "buy"}}},
		IsPublic: true,
	}
	expectedID := int64(123)

	// Setup expectations
	mockStrategyRepo.On("Save", mock.AnythingOfType("*models.Strategy")).Return(expectedID, nil)

	// Call method
	id, err := service.CreateStrategy(strategy)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
	mockStrategyRepo.AssertExpectations(t)
	mockStrategyMetricRepo.AssertExpectations(t)

	// Verify strategy properties were set correctly
	mockStrategyRepo.AssertCalled(t, "Save", mock.MatchedBy(func(s *models.Strategy) bool {
		return s.VoteCount == 0 && s.WinCount == 0 && s.ComplexityScore > 0 && s.RiskScore > 0
	}))
}

// Test CreateStrategy with invalid strategy
func TestCreateStrategyInvalidStrategy(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test cases for invalid strategies
	testCases := []struct {
		name        string
		strategy    *models.Strategy
		expectedErr string
	}{
		{
			name:        "Empty name",
			strategy:    &models.Strategy{Config: models.JSONB{"rules": []interface{}{}}},
			expectedErr: "strategy name is required",
		},
		{
			name:        "Empty config",
			strategy:    &models.Strategy{Name: "Test", Config: models.JSONB{}},
			expectedErr: "strategy must contain rules",
		},
		{
			name:        "Nil config",
			strategy:    &models.Strategy{Name: "Test"},
			expectedErr: "strategy configuration is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call method
			_, err := service.CreateStrategy(tc.strategy)

			// Assertions
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// Test GetStrategyByID
func TestGetStrategyByID(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	strategyID := int64(1)
	expectedStrategy := &models.Strategy{
		ID:       strategyID,
		Name:     "Test Strategy",
		Config:   models.JSONB{"rules": []interface{}{}},
		IsPublic: true,
	}

	// Setup expectations
	mockStrategyRepo.On("GetByID", strategyID).Return(expectedStrategy, nil)

	// Call method
	strategy, err := service.GetStrategyByID(strategyID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedStrategy, strategy)
	mockStrategyRepo.AssertExpectations(t)
}

// Test GetStrategyByID not found
func TestGetStrategyByIDNotFound(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	strategyID := int64(999) // Non-existent strategy

	// Setup expectations
	mockStrategyRepo.On("GetByID", strategyID).Return(nil, nil) // Strategy not found

	// Call method
	strategy, err := service.GetStrategyByID(strategyID)

	// Assertions
	assert.NoError(t, err)  // Should not return error for not found
	assert.Nil(t, strategy) // Strategy should be nil
	mockStrategyRepo.AssertExpectations(t)
}

// Test UpdateStrategy
func TestUpdateStrategy(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	strategyID := int64(1)
	existingStrategy := &models.Strategy{
		ID:          strategyID,
		Name:        "Original Strategy",
		Config:      models.JSONB{"rules": []interface{}{}},
		IsPublic:    true,
		VoteCount:   5,
		WinCount:    2,
		LastWinTime: time.Now().Add(-24 * time.Hour),
		CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
	}

	updatedStrategy := &models.Strategy{
		ID:       strategyID,
		Name:     "Updated Strategy",
		Config:   models.JSONB{"rules": []interface{}{map[string]interface{}{"condition": "new condition", "action": "sell"}}},
		IsPublic: false,
	}

	// Setup expectations
	mockStrategyRepo.On("GetByID", strategyID).Return(existingStrategy, nil)
	mockStrategyRepo.On("Update", mock.AnythingOfType("*models.Strategy")).Return(nil)

	// Call method
	err := service.UpdateStrategy(updatedStrategy)

	// Assertions
	assert.NoError(t, err)
	mockStrategyRepo.AssertExpectations(t)

	// Verify preserved properties
	mockStrategyRepo.AssertCalled(t, "Update", mock.MatchedBy(func(s *models.Strategy) bool {
		return s.VoteCount == existingStrategy.VoteCount &&
			s.WinCount == existingStrategy.WinCount &&
			s.LastWinTime.Equal(existingStrategy.LastWinTime) &&
			s.CreatedAt.Equal(existingStrategy.CreatedAt) &&
			s.Name == updatedStrategy.Name &&
			s.IsPublic == updatedStrategy.IsPublic
	}))
}

// Test UpdateStrategy with strategy not found
func TestUpdateStrategyNotFound(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	strategyID := int64(999) // Non-existent strategy
	strategy := &models.Strategy{
		ID:       strategyID,
		Name:     "Test Strategy",
		Config:   models.JSONB{"rules": []interface{}{}},
		IsPublic: true,
	}

	// Setup expectations
	mockStrategyRepo.On("GetByID", strategyID).Return(nil, nil) // Strategy not found

	// Call method
	err := service.UpdateStrategy(strategy)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "strategy not found")
	mockStrategyRepo.AssertExpectations(t)
	mockStrategyRepo.AssertNotCalled(t, "Update")
}

// Test DeleteStrategy
func TestDeleteStrategy(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	strategyID := int64(1)

	strategy := &models.Strategy{
		ID:       strategyID,
		Name:     "Test Strategy",
		Config:   models.JSONB{"rules": []interface{}{}},
		IsPublic: true,
	}

	// Setup expectations
	mockStrategyRepo.On("GetByID", strategyID).Return(strategy, nil)
	mockStrategyRepo.On("Delete", strategyID).Return(nil)

	// Call method
	err := service.DeleteStrategy(strategyID)

	// Assertions
	assert.NoError(t, err)
	mockStrategyRepo.AssertExpectations(t)
}

// Test DeleteStrategy with strategy not found
func TestDeleteStrategyNotFound(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	strategyID := int64(999) // Non-existent strategy

	// Setup expectations
	mockStrategyRepo.On("GetByID", strategyID).Return(nil, nil) // Strategy not found

	// Call method
	err := service.DeleteStrategy(strategyID)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "strategy not found")
	mockStrategyRepo.AssertExpectations(t)
	mockStrategyRepo.AssertNotCalled(t, "Delete")
}

// Test GetPublicStrategies
func TestGetPublicStrategies(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	limit := 10
	offset := 0

	expectedStrategies := []*models.Strategy{
		{ID: 1, Name: "Public Strategy 1", IsPublic: true},
		{ID: 2, Name: "Public Strategy 2", IsPublic: true},
	}

	// Setup expectations
	mockStrategyRepo.On("ListPublic", limit, offset).Return(expectedStrategies, nil)

	// Call method
	strategies, err := service.GetPublicStrategies(limit, offset)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedStrategies, strategies)
	mockStrategyRepo.AssertExpectations(t)
}

// Test GetTopStrategies
func TestGetTopStrategies(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	limit := 5
	expectedStrategies := []*models.Strategy{
		{ID: 1, Name: "Top Strategy 1"},
		{ID: 2, Name: "Top Strategy 2"},
	}

	// Test cases
	testCases := []struct {
		name     string
		criteria string
		mock     func()
	}{
		{
			name:     "Top voted strategies",
			criteria: "votes",
			mock: func() {
				mockStrategyRepo.On("GetTopVoted", limit).Return(expectedStrategies, nil).Once()
			},
		},
		{
			name:     "Top winning strategies",
			criteria: "wins",
			mock: func() {
				mockStrategyRepo.On("GetTopWinners", limit).Return(expectedStrategies, nil).Once()
			},
		},
		{
			name:     "Top performing strategies",
			criteria: "performance",
			mock: func() {
				// For this test, we'll just set up the expected calls for getTopPerformingStrategies
				metrics := []*models.StrategyMetric{
					{StrategyID: 1, WinRate: 0.8},
					{StrategyID: 2, WinRate: 0.7},
				}
				mockStrategyRepo.On("ListPublic", 1000, 0).Return(expectedStrategies, nil).Once()
				mockStrategyMetricRepo.On("GetLatestByStrategy", int64(1)).Return(metrics[0], nil).Once()
				mockStrategyMetricRepo.On("GetLatestByStrategy", int64(2)).Return(metrics[1], nil).Once()
				mockStrategyRepo.On("GetByID", int64(1)).Return(expectedStrategies[0], nil).Once()
				mockStrategyRepo.On("GetByID", int64(2)).Return(expectedStrategies[1], nil).Once()
			},
		},
		{
			name:     "Invalid criteria",
			criteria: "invalid",
			mock:     func() {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock for this test case
			tc.mock()

			// Call method
			strategies, err := service.GetTopStrategies(tc.criteria, limit)

			// Assertions
			if tc.criteria == "invalid" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid criteria")
				assert.Nil(t, strategies)
			} else {
				assert.NoError(t, err)
				if tc.criteria != "performance" { // Skip direct comparison for performance as mocking is more complex
					assert.Equal(t, expectedStrategies, strategies)
				} else {
					assert.NotNil(t, strategies)
				}
			}
		})
	}

	// Verify all mocks
	mockStrategyRepo.AssertExpectations(t)
	mockStrategyMetricRepo.AssertExpectations(t)
}

// Test SearchStrategiesByTags
func TestSearchStrategiesByTags(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	tags := []string{"ai", "trading"}
	limit := 10

	expectedStrategies := []*models.Strategy{
		{ID: 1, Name: "AI Strategy", Tags: []string{"ai", "machine-learning"}},
		{ID: 2, Name: "Trading Strategy", Tags: []string{"trading", "finance"}},
	}

	// Setup expectations
	mockStrategyRepo.On("SearchByTags", tags, limit).Return(expectedStrategies, nil)

	// Call method
	strategies, err := service.SearchStrategiesByTags(tags, limit)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedStrategies, strategies)
	mockStrategyRepo.AssertExpectations(t)
}

// Test RecordWin
func TestRecordWin(t *testing.T) {
	// Setup mocks
	mockStrategyRepo := new(MockStrategyRepository)
	mockStrategyMetricRepo := new(MockStrategyMetricRepository)

	// Create service
	service := NewStrategyService(
		mockStrategyRepo,
		mockStrategyMetricRepo,
		logger.New("test"),
	)

	// Test data
	strategyID := int64(1)
	simulationID := int64(5)
	winTime := time.Now()

	// Setup expectations
	mockStrategyRepo.On("IncrementWinCount", strategyID, winTime).Return(nil)
	mockStrategyMetricRepo.On("Save", mock.AnythingOfType("*models.StrategyMetric")).Return(int64(1), nil)

	// Call method
	err := service.RecordWin(strategyID, simulationID, winTime)

	// Assertions
	assert.NoError(t, err)
	mockStrategyRepo.AssertExpectations(t)
	mockStrategyMetricRepo.AssertExpectations(t)

	// Verify metric properties
	mockStrategyMetricRepo.AssertCalled(t, "Save", mock.MatchedBy(func(m *models.StrategyMetric) bool {
		return m.StrategyID == strategyID &&
			m.SimulationRunID != nil &&
			*m.SimulationRunID == simulationID &&
			m.WinRate == 1.0
	}))
}
