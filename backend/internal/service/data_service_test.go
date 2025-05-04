// internal/service/data_service_test.go
package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTokenRepository is a mock implementation of token repository
type MockTokenRepository struct {
	mock.Mock
}

// Ensure MockTokenRepository implements the TokenRepositoryInterface
var _ repository.TokenRepositoryInterface = (*MockTokenRepository)(nil)

func (m *MockTokenRepository) Save(token *models.Token) (int64, error) {
	args := m.Called(token)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenRepository) GetByMintAddress(mintAddress string) (*models.Token, error) {
	args := m.Called(mintAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Token), args.Error(1)
}

func (m *MockTokenRepository) GetRecentTokens(limit int) ([]*models.Token, error) {
	args := m.Called(limit)
	return args.Get(0).([]*models.Token), args.Error(1)
}

func (m *MockTokenRepository) GetFilteredTokens(minMarketCapUSD float64, maxAgeSeconds int64, limit int) ([]*models.Token, error) {
	args := m.Called(minMarketCapUSD, maxAgeSeconds, limit)
	return args.Get(0).([]*models.Token), args.Error(1)
}

func (m *MockTokenRepository) GetByID(tokenID int64) (*models.Token, error) {
	args := m.Called(tokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Token), args.Error(1)
}

// MockTradeRepository is a mock implementation of trade repository
type MockTradeRepository struct {
	mock.Mock
}

// Ensure MockTradeRepository implements the TradeRepositoryInterface
var _ repository.TradeRepositoryInterface = (*MockTradeRepository)(nil)

func (m *MockTradeRepository) Save(trade *models.Trade) (int64, error) {
	args := m.Called(trade)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTradeRepository) GetTradesByTokenID(tokenID int64, limit int) ([]*models.Trade, error) {
	args := m.Called(tokenID, limit)
	return args.Get(0).([]*models.Trade), args.Error(1)
}

// GetTradesByTokenIDWithContext implements repository.TradeRepositoryInterface.
func (m *MockTradeRepository) GetTradesByTokenIDWithContext(ctx context.Context, tokenID int64, limit int) ([]*models.Trade, error) {
	args := m.Called(ctx, tokenID, limit)
	return args.Get(0).([]*models.Trade), args.Error(1)
}

func (m *MockTradeRepository) GetTradesBySignature(signature string) (*models.Trade, error) {
	args := m.Called(signature)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trade), args.Error(1)
}

type DataServiceWithMocks struct {
	tokenRepo repository.TokenRepositoryInterface
	tradeRepo repository.TradeRepositoryInterface
	logger    *logger.Logger
}

// Implement DataService methods for our mock service
func (s *DataServiceWithMocks) ProcessTokenData(data map[string]interface{}) error {
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
	_, err := s.tokenRepo.Save(token)
	if err != nil {
		return fmt.Errorf("error saving token: %v", err)
	}

	return nil
}

func (s *DataServiceWithMocks) ProcessTradeData(data map[string]interface{}) error {
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
					return fmt.Errorf("error updating token market cap: %v", err)
				}
			}
		}
	}

	// Extract other trade fields
	var solAmount float64
	if solVal, ok := data["sol_amount"].(float64); ok {
		solAmount = solVal
	}

	var tokenAmount float64
	if tokVal, ok := data["token_amount"].(float64); ok {
		tokenAmount = tokVal
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
	_, err = s.tradeRepo.Save(trade)
	if err != nil {
		return fmt.Errorf("error saving trade: %v", err)
	}

	return nil
}

// createMockDataService creates a service with mocked repositories for testing
func createMockDataService() (*DataServiceWithMocks, *MockTokenRepository, *MockTradeRepository) {
	mockTokenRepo := new(MockTokenRepository)
	mockTradeRepo := new(MockTradeRepository)

	service := &DataServiceWithMocks{
		tokenRepo: mockTokenRepo,
		tradeRepo: mockTradeRepo,
		logger:    logger.New("test"),
	}

	return service, mockTokenRepo, mockTradeRepo
}

func TestProcessTokenData(t *testing.T) {
	// Create mock service and repositories
	service, mockTokenRepo, _ := createMockDataService()

	// Setup test data
	tokenData := map[string]interface{}{
		"mint":                       "test-mint-address",
		"creator":                    "test-creator-address",
		"name":                       "Test Token",
		"symbol":                     "TEST",
		"image_uri":                  "http://example.com/image.png",
		"twitter":                    "http://twitter.com/test",
		"website":                    "http://example.com",
		"telegram":                   "http://t.me/test",
		"metadata_uri":               "http://example.com/metadata.json",
		"created_timestamp":          float64(time.Now().Unix()),
		"market_cap":                 float64(1000.0),
		"usd_market_cap":             float64(50000.0),
		"complete":                   false,
		"king_of_the_hill_timestamp": float64(0),
	}

	// Setup expectations
	mockTokenRepo.On("Save", mock.AnythingOfType("*models.Token")).Return(int64(1), nil)

	// Call the method under test
	err := service.ProcessTokenData(tokenData)

	// Assertions
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)

	// Verify the arguments passed to Save
	mockTokenRepo.AssertCalled(t, "Save", mock.MatchedBy(func(token *models.Token) bool {
		return token.MintAddress == "test-mint-address" &&
			token.Name == "Test Token" &&
			token.Symbol == "TEST" &&
			token.MarketCap == 1000.0 &&
			token.UsdMarketCap == 50000.0
	}))
}

func TestProcessTradeData_ExistingToken(t *testing.T) {
	// Create mock service and repositories
	service, mockTokenRepo, mockTradeRepo := createMockDataService()

	// Create existing token
	existingToken := &models.Token{
		ID:             1,
		MintAddress:    "test-mint-address",
		CreatorAddress: "test-creator-address",
		Name:           "Test Token",
		Symbol:         "TEST",
		MarketCap:      1000.0,
		UsdMarketCap:   50000.0,
	}

	// Setup trade data
	tradeData := map[string]interface{}{
		"mint":         "test-mint-address",
		"signature":    "test-signature",
		"sol_amount":   float64(0.5),
		"token_amount": float64(1000),
		"is_buy":       true,
		"user":         "test-user-address",
		"timestamp":    float64(time.Now().Unix()),
	}

	// Setup expectations
	mockTokenRepo.On("GetByMintAddress", "test-mint-address").Return(existingToken, nil)
	mockTradeRepo.On("Save", mock.AnythingOfType("*models.Trade")).Return(int64(1), nil)

	// Call the method under test
	err := service.ProcessTradeData(tradeData)

	// Assertions
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
	mockTradeRepo.AssertExpectations(t)

	// Verify the arguments passed to Save
	mockTradeRepo.AssertCalled(t, "Save", mock.MatchedBy(func(trade *models.Trade) bool {
		return trade.TokenID == existingToken.ID &&
			trade.MintAddress == "test-mint-address" &&
			trade.Signature == "test-signature" &&
			trade.SolAmount == 0.5 &&
			trade.TokenAmount == 1000 &&
			trade.IsBuy == true
	}))
}

func TestProcessTradeData_NewToken(t *testing.T) {
	// Create mock service and repositories
	service, mockTokenRepo, mockTradeRepo := createMockDataService()

	// Setup the newly created token
	newToken := &models.Token{
		ID:             1,
		MintAddress:    "test-mint-address",
		CreatorAddress: "test-creator-address",
		Name:           "Test Token",
		Symbol:         "TEST",
		MarketCap:      1000.0,
		UsdMarketCap:   50000.0,
	}

	// Setup trade data - this will cause a new token to be created
	tradeData := map[string]interface{}{
		"mint":           "test-mint-address",
		"creator":        "test-creator-address",
		"name":           "Test Token",
		"symbol":         "TEST",
		"signature":      "test-signature",
		"sol_amount":     float64(0.5),
		"token_amount":   float64(1000),
		"is_buy":         true,
		"user":           "test-user-address",
		"timestamp":      float64(time.Now().Unix()),
		"market_cap":     float64(1000.0),
		"usd_market_cap": float64(50000.0),
	}

	// Setup expectations - first call returns nil (token not found), second call (after creation) returns the new token
	mockTokenRepo.On("GetByMintAddress", "test-mint-address").Return(nil, nil).Once()
	mockTokenRepo.On("Save", mock.AnythingOfType("*models.Token")).Return(int64(1), nil)
	mockTokenRepo.On("GetByMintAddress", "test-mint-address").Return(newToken, nil).Once()
	mockTradeRepo.On("Save", mock.AnythingOfType("*models.Trade")).Return(int64(1), nil)

	// Call the method under test
	err := service.ProcessTradeData(tradeData)

	// Assertions
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
	mockTradeRepo.AssertExpectations(t)

	// Verify the calls
	mockTokenRepo.AssertNumberOfCalls(t, "GetByMintAddress", 2)
	mockTokenRepo.AssertNumberOfCalls(t, "Save", 1)
	mockTradeRepo.AssertNumberOfCalls(t, "Save", 1)
}

func TestProcessTradeData_UpdateTokenMarketCap(t *testing.T) {
	// Create mock service and repositories
	service, mockTokenRepo, mockTradeRepo := createMockDataService()

	// Create existing token with old market cap values
	existingToken := &models.Token{
		ID:             1,
		MintAddress:    "test-mint-address",
		CreatorAddress: "test-creator-address",
		Name:           "Test Token",
		Symbol:         "TEST",
		MarketCap:      1000.0,
		UsdMarketCap:   50000.0,
	}

	// Setup trade data with updated market cap
	tradeData := map[string]interface{}{
		"mint":           "test-mint-address",
		"signature":      "test-signature",
		"sol_amount":     float64(0.5),
		"token_amount":   float64(1000),
		"is_buy":         true,
		"user":           "test-user-address",
		"timestamp":      float64(time.Now().Unix()),
		"market_cap":     float64(1500.0),  // Updated market cap
		"usd_market_cap": float64(75000.0), // Updated USD market cap
	}

	// Setup expectations
	mockTokenRepo.On("GetByMintAddress", "test-mint-address").Return(existingToken, nil)
	// Token should be saved with updated market cap
	mockTokenRepo.On("Save", mock.MatchedBy(func(token *models.Token) bool {
		return token.MarketCap == 1500.0 && token.UsdMarketCap == 75000.0
	})).Return(int64(1), nil)
	mockTradeRepo.On("Save", mock.AnythingOfType("*models.Trade")).Return(int64(1), nil)

	// Call the method under test
	err := service.ProcessTradeData(tradeData)

	// Assertions
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
	mockTradeRepo.AssertExpectations(t)

	// Verify token was updated with new market cap values
	mockTokenRepo.AssertCalled(t, "Save", mock.MatchedBy(func(token *models.Token) bool {
		return token.ID == 1 && token.MarketCap == 1500.0 && token.UsdMarketCap == 75000.0
	}))
}

func TestProcessTradeData_InvalidData(t *testing.T) {
	// Create mock service
	service, _, _ := createMockDataService()

	// Test cases for invalid data
	testCases := []struct {
		name          string
		tradeData     map[string]interface{}
		expectedError string
	}{
		{
			name: "Missing mint address",
			tradeData: map[string]interface{}{
				"signature":  "test-signature",
				"sol_amount": float64(0.5),
			},
			expectedError: "invalid mint address",
		},
		{
			name: "Empty mint address",
			tradeData: map[string]interface{}{
				"mint":      "",
				"signature": "test-signature",
			},
			expectedError: "invalid mint address",
		},
		{
			name: "Missing signature",
			tradeData: map[string]interface{}{
				"mint":       "test-mint-address",
				"sol_amount": float64(0.5),
			},
			expectedError: "invalid signature",
		},
		{
			name: "Empty signature",
			tradeData: map[string]interface{}{
				"mint":      "test-mint-address",
				"signature": "",
			},
			expectedError: "invalid signature",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ProcessTradeData(tc.tradeData)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestProcessTokenData_InvalidData(t *testing.T) {
	// Create mock service
	service, _, _ := createMockDataService()

	// Test cases for invalid data
	testCases := []struct {
		name          string
		tokenData     map[string]interface{}
		expectedError string
	}{
		{
			name: "Missing mint address",
			tokenData: map[string]interface{}{
				"creator": "test-creator-address",
				"name":    "Test Token",
			},
			expectedError: "invalid mint address",
		},
		{
			name: "Empty mint address",
			tokenData: map[string]interface{}{
				"mint":    "",
				"creator": "test-creator-address",
			},
			expectedError: "invalid mint address",
		},
		{
			name: "Missing creator address",
			tokenData: map[string]interface{}{
				"mint": "test-mint-address",
				"name": "Test Token",
			},
			expectedError: "invalid creator address",
		},
		{
			name: "Empty creator address",
			tokenData: map[string]interface{}{
				"mint":    "test-mint-address",
				"creator": "",
			},
			expectedError: "invalid creator address",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ProcessTokenData(tc.tokenData)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}
