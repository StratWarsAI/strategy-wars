// internal/service/ai_service.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
	"github.com/StratWarsAI/strategy-wars/pkg/utils"
)

// AIService handles integration with AI API for strategy generation
type AIService struct {
	apiKey          string
	apiURL          string
	httpClient      *http.Client
	logger          *logger.Logger
	strategyRepo    repository.StrategyRepositoryInterface
	autoGenInterval time.Duration // Interval between automatic strategy generation
	lastGenTime     time.Time
}

// AIRequest represents a request to the AI API
type AIRequest struct {
	Model       string      `json:"model"`
	Messages    []AIMessage `json:"messages"`
	Temperature float64     `json:"temperature"`
	MaxTokens   int         `json:"max_tokens"`
}

// AIMessage represents a message in the AI conversation
type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse represents a response from the AI API
type AIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// NewAIService creates a new AI service
func NewAIService(
	apiKey string,
	apiURL string,
	strategyRepo repository.StrategyRepositoryInterface,
	logger *logger.Logger,
) *AIService {
	return &AIService{
		apiKey: apiKey,
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger:          logger,
		strategyRepo:    strategyRepo,
		autoGenInterval: 1 * time.Hour, // Generate strategies every hour
		lastGenTime:     time.Now(),
	}
}

// StartAutoGeneration starts the automatic strategy generation process
func (s *AIService) StartAutoGeneration(ctx context.Context) {
	s.logger.Info("Starting automatic strategy generation service")

	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes if it's time to generate
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping automatic strategy generation service")
			return
		case <-ticker.C:
			// Check if it's time to generate new strategies
			if time.Since(s.lastGenTime) >= s.autoGenInterval {
				s.logger.Info("Generating new automatic strategies")

				// Number of strategies to generate in each cycle
				strategiesPerCycle := 2

				for i := 0; i < strategiesPerCycle; i++ {
					// Get top performing strategies to learn from
					topStrategies, err := s.GetTopPerformingStrategies()
					if err != nil {
						s.logger.Error("Error getting top strategies: %v", err)
						continue
					}

					// Create metadata with top strategies
					metadata := map[string]interface{}{
						"top_strategies": topStrategies,
					}

					// Generate strategy prompt based on index
					var prompt string
					if i == 0 {
						prompt = "Generate a profitable trading strategy for cryptocurrency tokens focused on momentum and quick profitability"
					} else {
						prompt = "Generate a defensive trading strategy for cryptocurrency tokens with strong risk management"
					}

					// Generate new strategy
					strategy, err := s.GenerateStrategy(prompt, metadata)
					if err != nil {
						s.logger.Error("Error generating strategy %d: %v", i+1, err)
						continue
					}

					// Add timestamp to make name unique
					strategy.Name = fmt.Sprintf("%s (%s)", strategy.Name, time.Now().Format("20060102-1504"))

					// Save the strategy
					id, err := s.strategyRepo.Save(strategy)
					if err != nil {
						s.logger.Error("Error saving generated strategy: %v", err)
						continue
					}

					s.logger.Info("Successfully generated and saved new strategy %d with ID: %d", i+1, id)
				}

				s.lastGenTime = time.Now()
			}
		}
	}
}

// GetTopPerformingStrategies gets the top performing strategies to learn from
func (s *AIService) GetTopPerformingStrategies() ([]map[string]interface{}, error) {
	// Get top strategies by win count
	strategies, err := s.strategyRepo.GetTopWinners(5)
	if err != nil {
		return nil, fmt.Errorf("error getting top strategies: %v", err)
	}

	var topStrategies []map[string]interface{}

	for _, strategy := range strategies {
		// Extract parameters from config
		config := make(map[string]interface{})

		// Get basic strategy info
		strategyInfo := map[string]interface{}{
			"id":         strategy.ID,
			"name":       strategy.Name,
			"win_count":  strategy.WinCount,
			"win_rate":   float64(strategy.WinCount) / float64(max(strategy.WinCount+strategy.VoteCount, 1)) * 100,
			"parameters": config,
		}

		// Add strategy parameters if they exist in the config
		if marketCap, ok := strategy.Config["marketCapThreshold"].(float64); ok {
			config["marketCapThreshold"] = marketCap
		}

		if minBuys, ok := strategy.Config["minBuysForEntry"].(float64); ok {
			config["minBuysForEntry"] = minBuys
		}

		if timeWindow, ok := strategy.Config["entryTimeWindowSec"].(float64); ok {
			config["entryTimeWindowSec"] = timeWindow
		}

		if takeProfit, ok := strategy.Config["takeProfitPct"].(float64); ok {
			config["takeProfitPct"] = takeProfit
		}

		if stopLoss, ok := strategy.Config["stopLossPct"].(float64); ok {
			config["stopLossPct"] = stopLoss
		}

		if maxHold, ok := strategy.Config["maxHoldTimeSec"].(float64); ok {
			config["maxHoldTimeSec"] = maxHold
		}

		if positionSize, ok := strategy.Config["fixedPositionSizeSol"].(float64); ok {
			config["fixedPositionSizeSol"] = positionSize
		}

		if initialBalance, ok := strategy.Config["initialBalance"].(float64); ok {
			config["initialBalance"] = initialBalance
		}

		topStrategies = append(topStrategies, strategyInfo)
	}

	return topStrategies, nil
}

// GenerateStrategy generates a new trading strategy using AI
func (s *AIService) GenerateStrategy(basePrompt string, metaData map[string]interface{}) (*models.Strategy, error) {
	// Format prompt with metadata
	prompt := s.formatPrompt(basePrompt, metaData)

	// Create request
	req := AIRequest{
		Model: "gpt-4",
		Messages: []AIMessage{
			{
				Role:    "system",
				Content: "You are a professional algorithmic trader specializing in creating strategies for cryptocurrency trading.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	// Execute request
	response, err := s.executeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error executing ChatGPT request: %v", err)
	}

	// Parse response into strategy
	strategy, err := s.parseStrategyResponse(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing ChatGPT response: %v", err)
	}

	return strategy, nil
}

// formatPrompt creates a detailed prompt with metadata
func (s *AIService) formatPrompt(basePrompt string, metaData map[string]interface{}) string {
	// Format metadata into a structured prompt context
	var contextBuilder strings.Builder

	contextBuilder.WriteString(basePrompt)
	contextBuilder.WriteString("\n\n")
	contextBuilder.WriteString("You are tasked with creating a trading strategy for cryptocurrency tokens on the Strategy Wars platform. ")
	contextBuilder.WriteString("The strategy should be designed to identify and trade tokens with high potential for price increase in short timeframes.\n\n")

	contextBuilder.WriteString("Your strategy must include the following required parameters as a JSON object:\n\n")
	contextBuilder.WriteString("1. name: A catchy name for your strategy (string)\n")
	contextBuilder.WriteString("2. description: A brief description of how your strategy works (string)\n")
	contextBuilder.WriteString("3. marketCapThreshold: The minimum market cap in USD for tokens to consider (number, typically between 5000-50000)\n")
	contextBuilder.WriteString("4. minBuysForEntry: Minimum number of buy transactions required in the time window to trigger entry (number, typically 2-10)\n")
	contextBuilder.WriteString("5. entryTimeWindowSec: Time window in seconds to count transactions for entry signal (number, typically 60-600)\n")
	contextBuilder.WriteString("6. takeProfitPct: Percentage gain to trigger take profit (number, typically 10-100)\n")
	contextBuilder.WriteString("7. stopLossPct: Percentage loss to trigger stop loss (number, typically 5-50)\n")
	contextBuilder.WriteString("8. maxHoldTimeSec: Maximum time to hold a position in seconds (number, typically 30-3600)\n")
	contextBuilder.WriteString("9. fixedPositionSizeSol: Fixed position size in SOL for each trade (number, typically 0.1-2)\n")
	contextBuilder.WriteString("10. initialBalance: Starting balance in SOL (number, typically 10-100)\n\n")

	contextBuilder.WriteString("Example strategy format:\n")
	contextBuilder.WriteString(`{
		"name": "Momentum Chaser",
		"description": "This strategy looks for tokens with rapid buy activity as an indicator of positive momentum",
		"marketCapThreshold": 7000,
		"minBuysForEntry": 3,
		"entryTimeWindowSec": 300,
		"takeProfitPct": 50,
		"stopLossPct": 30,
		"maxHoldTimeSec": 60,
		"fixedPositionSizeSol": 0.5,
		"initialBalance": 10
		}`)
	contextBuilder.WriteString("\n\n")

	// Add information about top performing strategies if available
	if topStrategies, ok := metaData["top_strategies"].([]map[string]interface{}); ok && len(topStrategies) > 0 {
		contextBuilder.WriteString("Here are some of our top performing strategies you can learn from:\n\n")

		for i, strategy := range topStrategies {
			if i >= 3 { // Limit to top 3
				break
			}

			contextBuilder.WriteString(fmt.Sprintf("Strategy: %s\n", strategy["name"]))
			contextBuilder.WriteString(fmt.Sprintf("Win Rate: %.2f%%\n", strategy["win_rate"]))

			if params, ok := strategy["parameters"].(map[string]interface{}); ok {
				contextBuilder.WriteString("Parameters:\n")

				if val, ok := params["marketCapThreshold"].(float64); ok {
					contextBuilder.WriteString(fmt.Sprintf("- Market Cap Threshold: $%.2f\n", val))
				}

				if val, ok := params["takeProfitPct"].(float64); ok {
					contextBuilder.WriteString(fmt.Sprintf("- Take Profit: %.2f%%\n", val))
				}

				if val, ok := params["stopLossPct"].(float64); ok {
					contextBuilder.WriteString(fmt.Sprintf("- Stop Loss: %.2f%%\n", val))
				}
			}

			contextBuilder.WriteString("\n")
		}
	} else {
		// If no top strategies, provide guidelines for creating a good strategy
		contextBuilder.WriteString("Guidelines for creating an effective strategy:\n\n")
		contextBuilder.WriteString("1. Balance risk and reward: Higher take profit levels should be paired with appropriate stop losses\n")
		contextBuilder.WriteString("2. Consider market cap: Lower market cap tokens may have higher volatility but also higher potential returns\n")
		contextBuilder.WriteString("3. Entry timing: The entry time window and minimum buys should capture meaningful momentum\n")
		contextBuilder.WriteString("4. Position sizing: Smaller position sizes allow for more trades but may limit profits\n")
		contextBuilder.WriteString("5. Hold time: Shorter hold times reduce exposure to downside risks but may limit profit potential\n\n")
	}

	contextBuilder.WriteString("Generate a complete strategy that you believe would be profitable. Be creative and innovative while considering risk management. Return ONLY the JSON object for your strategy.\n")

	return contextBuilder.String()
}

// executeRequest sends the request to the AI API
func (s *AIService) executeRequest(req AIRequest) (*AIResponse, error) {
	// Convert request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	// Execute request
	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}
	defer httpResp.Body.Close()

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		// Read error response
		var errResp map[string]interface{}
		if err := json.NewDecoder(httpResp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("error decoding error response: %v", err)
		}
		return nil, fmt.Errorf("error from ChatGPT API: %v", errResp)
	}

	// Decode response
	var resp AIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &resp, nil
}

// parseStrategyResponse parses the ChatGPT response into a strategy
func (s *AIService) parseStrategyResponse(response *AIResponse) (*models.Strategy, error) {
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := response.Choices[0].Message.Content

	// Extract JSON configuration from response
	strategy, err := s.extractStrategyConfig(content)
	if err != nil {
		return nil, fmt.Errorf("error extracting strategy config: %v", err)
	}

	return strategy, nil
}

// extractStrategyConfig extracts strategy configuration from the response content
func (s *AIService) extractStrategyConfig(content string) (*models.Strategy, error) {
	// Look for JSON blocks in the content
	jsonPattern := regexp.MustCompile(`(?s)\{.*\}`)
	jsonMatches := jsonPattern.FindAllString(content, -1)

	if len(jsonMatches) == 0 {
		return nil, fmt.Errorf("no JSON configuration found in response")
	}

	// Try each JSON block until we find a valid one
	var strategy models.Strategy
	var validJSON bool

	for _, jsonStr := range jsonMatches {
		// Try to unmarshal into a config
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
			continue
		}

		// Check if it has required fields
		if _, hasName := config["name"]; !hasName {
			continue
		}

		// Create strategy with the required parameters
		strategyConfig := map[string]interface{}{
			"marketCapThreshold":   utils.GetFloat64OrDefault(config, "marketCapThreshold", 7000.0),
			"minBuysForEntry":      utils.GetIntOrDefault(config, "minBuysForEntry", 3),
			"entryTimeWindowSec":   utils.GetIntOrDefault(config, "entryTimeWindowSec", 300),
			"takeProfitPct":        utils.GetFloat64OrDefault(config, "takeProfitPct", 50.0),
			"stopLossPct":          utils.GetFloat64OrDefault(config, "stopLossPct", 30.0),
			"maxHoldTimeSec":       utils.GetIntOrDefault(config, "maxHoldTimeSec", 1800),
			"fixedPositionSizeSol": utils.GetFloat64OrDefault(config, "fixedPositionSizeSol", 0.5),
			"initialBalance":       utils.GetFloat64OrDefault(config, "initialBalance", 10.0),
		}

		strategy = models.Strategy{
			Name:        utils.GetStringOrDefault(config, "name", "AI Generated Strategy"),
			Description: utils.GetStringOrDefault(config, "description", fmt.Sprintf("AI-generated strategy on %s", time.Now().Format("2006-01-02 15:04:05"))),
			Config:      models.JSONB(strategyConfig),
			IsPublic:    true,
			AIEnhanced:  true,
			Tags:        []string{"ai-generated", "auto-optimized"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		validJSON = true
		break
	}

	if !validJSON {
		return nil, fmt.Errorf("no valid strategy configuration found in response")
	}

	// Calculate complexity score based on parameters
	strategy.ComplexityScore = s.calculateComplexityScore(strategy.Config)

	// Estimate risk score based on configuration parameters
	strategy.RiskScore = s.calculateRiskScore(strategy.Config)

	return &strategy, nil
}

// calculateComplexityScore calculates the complexity score of a strategy (1-10)
func (s *AIService) calculateComplexityScore(config models.JSONB) int {
	// Base complexity
	complexity := 3

	// Higher complexity for more aggressive strategies
	if takeProfitPct, ok := config["takeProfitPct"].(float64); ok {
		if takeProfitPct > 70 {
			complexity += 2 // Very aggressive
		} else if takeProfitPct > 40 {
			complexity += 1 // Moderately aggressive
		}
	}

	// Higher complexity for tighter entry criteria
	if minBuys, ok := config["minBuysForEntry"].(float64); ok {
		if minBuys > 5 {
			complexity += 1 // More selective
		}
	}

	// Higher complexity for shorter timeframes
	if timeWindow, ok := config["entryTimeWindowSec"].(float64); ok {
		if timeWindow < 120 {
			complexity += 1 // Very short timeframe
		}
	}

	// Higher complexity for precise hold times
	if maxHold, ok := config["maxHoldTimeSec"].(float64); ok {
		if maxHold < 600 {
			complexity += 1 // Short hold time
		}
	}

	// Add some randomness to make strategies more unique
	complexity += rand.Intn(3) // Add 0-2 random complexity

	// Ensure complexity is within 1-10 range
	if complexity < 1 {
		complexity = 1
	} else if complexity > 10 {
		complexity = 10
	}

	return complexity
}

// calculateRiskScore calculates the risk score of a strategy (1-10)
func (s *AIService) calculateRiskScore(config models.JSONB) int {
	// Default risk
	risk := 5

	// Adjust based on take profit
	if takeProfitPct, ok := config["takeProfitPct"].(float64); ok {
		// Higher take profit usually means higher risk
		if takeProfitPct > 70 {
			risk += 3 // Very high risk
		} else if takeProfitPct > 40 {
			risk += 2 // High risk
		} else if takeProfitPct < 20 {
			risk -= 1 // Conservative
		}
	}

	// Adjust based on stop loss
	if stopLossPct, ok := config["stopLossPct"].(float64); ok {
		// Larger stop loss usually means higher risk
		if stopLossPct > 40 {
			risk += 2 // Very risky
		} else if stopLossPct > 20 {
			risk += 1 // Somewhat risky
		} else if stopLossPct < 10 {
			risk -= 2 // Conservative
		}
	}

	// Adjust based on market cap threshold
	if marketCap, ok := config["marketCapThreshold"].(float64); ok {
		// Lower market cap usually means higher risk
		if marketCap < 5000 {
			risk += 2 // Very risky
		} else if marketCap < 7000 {
			risk += 1 // Somewhat risky
		} else if marketCap > 15000 {
			risk -= 1 // Lower risk
		}
	}

	// Adjust based on position size
	if posSize, ok := config["fixedPositionSizeSol"].(float64); ok {
		if posSize > 1 {
			risk += 1 // Larger positions
		} else if posSize < 0.2 {
			risk -= 1 // Smaller positions
		}
	}

	// Ensure risk score is within bounds
	if risk < 1 {
		risk = 1
	} else if risk > 10 {
		risk = 10
	}

	return risk
}

// SaveStrategy saves a strategy to the database
func (s *AIService) SaveStrategy(strategy *models.Strategy) (int64, error) {
	return s.strategyRepo.Save(strategy)
}

// GenerateAnalysis generates an AI-powered analysis of a strategy's performance
func (s *AIService) GenerateAnalysis(strategyName string, metrics map[string]interface{}, hasActiveTrades bool, isActiveSimulation bool) (string, error) {
	s.logger.Info("Generating AI analysis for strategy: %s", strategyName)

	// Convert metrics to a description format for the AI
	var metricsDescription strings.Builder
	metricsDescription.WriteString(fmt.Sprintf("Strategy Name: %s\n", strategyName))

	if totalTrades, ok := metrics["total_trades"].(int); ok {
		metricsDescription.WriteString(fmt.Sprintf("Total Completed Trades: %d\n", totalTrades))
	}

	if hasActiveTrades {
		if activeTrades, ok := metrics["active_trades"].(int); ok {
			metricsDescription.WriteString(fmt.Sprintf("Active Positions: %d\n", activeTrades))
		}
	}

	if winRate, ok := metrics["win_rate"].(float64); ok {
		metricsDescription.WriteString(fmt.Sprintf("Win Rate: %.2f%%\n", winRate))
	}

	if roi, ok := metrics["roi"].(float64); ok {
		metricsDescription.WriteString(fmt.Sprintf("Return on Investment (ROI): %.2f%%\n", roi))
	}

	if maxDrawdown, ok := metrics["max_drawdown"].(float64); ok {
		metricsDescription.WriteString(fmt.Sprintf("Maximum Drawdown: %.2f%%\n", maxDrawdown))
	}

	if netPnL, ok := metrics["net_pnl"].(float64); ok {
		metricsDescription.WriteString(fmt.Sprintf("Net Profit/Loss: %.4f\n", netPnL))
	}

	if avgProfit, ok := metrics["avg_profit"].(float64); ok {
		metricsDescription.WriteString(fmt.Sprintf("Average Profit per Trade: %.4f\n", avgProfit))
	}

	if avgLoss, ok := metrics["avg_loss"].(float64); ok {
		metricsDescription.WriteString(fmt.Sprintf("Average Loss per Trade: %.4f\n", avgLoss))
	}

	// Additional context about the current state
	if isActiveSimulation {
		metricsDescription.WriteString("\nThis strategy is currently being simulated. The metrics may change as trades complete.\n")
	}

	if hasActiveTrades {
		metricsDescription.WriteString("\nThe strategy has active open positions that are not reflected in the completed trade metrics.\n")
	}

	// Create the prompt for the AI
	prompt := fmt.Sprintf(`Analyze the trading performance of a cryptocurrency strategy based on the provided metrics and generate an insightful performance analysis. Your analysis should:

			1. Objectively assess the strategy's performance based on the metrics
			2. Highlight strengths and weaknesses
			3. Assess the level of risk based on drawdown and other metrics
			4. Provide a clear assessment (excellent, good, average, poor, or very poor)
			5. Suggest potential areas of improvement if applicable

			Here are the performance metrics:

		%s

		Write a concise, clear analysis of approximately 3-5 sentences that a trader would find valuable.`, metricsDescription.String())

	// Create request using the same model and API configuration already set in the AIService
	req := AIRequest{
		Model: "gpt-4",
		Messages: []AIMessage{
			{
				Role:    "system",
				Content: "You are an expert trading strategy analyst specializing in cryptocurrency trading. You provide concise, insightful analysis of trading strategy performance.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   500,
	}

	// Execute request using the existing API configuration
	response, err := s.executeRequest(req)
	if err != nil {
		return "", fmt.Errorf("error executing analysis generation request: %v", err)
	}

	// Check for valid response
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no analysis content in AI response")
	}

	analysis := response.Choices[0].Message.Content

	// Clean the analysis text (remove quotes, etc. if needed)
	analysis = strings.TrimSpace(analysis)

	s.logger.Info("Generated AI analysis for strategy %s successfully", strategyName)

	return analysis, nil
}

// GenerateEvolutionaryStrategy creates a new strategy based on existing successful ones
func (s *AIService) GenerateEvolutionaryStrategy() (*models.Strategy, error) {
	// Get top strategies to base the new one on
	topStrategies, err := s.GetTopPerformingStrategies()
	if err != nil {
		return nil, fmt.Errorf("error getting top strategies: %v", err)
	}

	if len(topStrategies) == 0 {
		return nil, fmt.Errorf("no existing strategies to evolve from")
	}

	// Create metadata with top strategies
	metadata := map[string]interface{}{
		"top_strategies": topStrategies,
		"evolution":      true,
	}

	// Generate evolutionary prompt
	prompt := "Create an evolved trading strategy that improves upon our best performing strategies. " +
		"Analyze the winning strategies provided and create a new strategy that combines their strengths " +
		"while addressing their weaknesses."

	// Generate the evolved strategy
	return s.GenerateStrategy(prompt, metadata)
}

// GenerateOptimizedStrategy creates an optimized version of an existing strategy
func (s *AIService) GenerateOptimizedStrategy(baseStrategyID int64) (*models.Strategy, error) {
	// Get the base strategy
	baseStrategy, err := s.strategyRepo.GetByID(baseStrategyID)
	if err != nil {
		return nil, fmt.Errorf("error getting base strategy: %v", err)
	}

	if baseStrategy == nil {
		return nil, fmt.Errorf("base strategy not found: %d", baseStrategyID)
	}

	// Create metadata with the base strategy
	baseStrategyInfo := map[string]interface{}{
		"id":          baseStrategy.ID,
		"name":        baseStrategy.Name,
		"description": baseStrategy.Description,
		"config":      baseStrategy.Config,
		"win_count":   baseStrategy.WinCount,
		"vote_count":  baseStrategy.VoteCount,
	}

	metadata := map[string]interface{}{
		"base_strategy": baseStrategyInfo,
		"optimization":  true,
	}

	// Generate optimization prompt
	prompt := fmt.Sprintf("Optimize the trading strategy named '%s'. "+
		"The strategy has had %d wins. "+
		"Create an improved version that keeps its core strengths but fine-tunes the parameters "+
		"for better performance.",
		baseStrategy.Name, baseStrategy.WinCount)

	// Generate the optimized strategy
	optimizedStrategy, err := s.GenerateStrategy(prompt, metadata)
	if err != nil {
		return nil, fmt.Errorf("error generating optimized strategy: %v", err)
	}

	// Update name to indicate it's an optimized version
	optimizedStrategy.Name = fmt.Sprintf("%s (Optimized v%d)", baseStrategy.Name, rand.Intn(9)+1)
	optimizedStrategy.Description = fmt.Sprintf("AI-optimized version of strategy #%d (%s)",
		baseStrategy.ID, baseStrategy.Name)

	// Add "optimized" tag
	optimizedStrategy.Tags = append(optimizedStrategy.Tags, "optimized")

	return optimizedStrategy, nil
}
