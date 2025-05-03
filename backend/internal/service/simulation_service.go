// internal/service/simulation_service.go
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
	"github.com/StratWarsAI/strategy-wars/internal/websocket"
)

// SimulationService handles strategy simulations
type SimulationService struct {
	db                  *sql.DB
	strategyRepo        repository.StrategyRepositoryInterface
	tokenRepo           repository.TokenRepositoryInterface
	tradeRepo           repository.TradeRepositoryInterface
	simulatedTradeRepo  repository.SimulatedTradeRepositoryInterface
	strategyMetricRepo  repository.StrategyMetricRepositoryInterface
	simulationRunRepo   repository.SimulationRunRepositoryInterface
	simulationEventRepo repository.SimulationEventRepositoryInterface
	logger              *logger.Logger
	wsHub               *websocket.WSHub
	activeSimsMu        sync.RWMutex
	activeSims          map[int64]*SimulationContext
	simulationDone      chan int64
	workerPool          chan struct{} // Limit concurrent token evaluations
	shutdownCh          chan struct{} // Channel for graceful shutdown
}

// SimulationContext holds the context for an active simulation
type SimulationContext struct {
	StrategyID      int64
	Strategy        *models.Strategy
	Config          StrategyConfig
	StartTime       time.Time
	Trades          []*models.SimulatedTrade
	IsRunning       bool
	StopRequested   bool
	CurrentBalance  float64
	InitialBalance  float64
	SimulationRunID int64              // ID of the database record for this simulation run
	mu              sync.RWMutex       // For thread-safe access to context data
	tokensMu        sync.RWMutex       // For thread-safe access to trades slice
	wg              sync.WaitGroup     // To wait for all goroutines to finish
	ctx             context.Context    // Context for cancellation
	cancel          context.CancelFunc // Function to cancel goroutines
}

// IsActive returns whether the simulation is currently running
func (s *SimulationContext) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IsRunning
}

// StrategyConfig defines the configuration for a trading strategy
type StrategyConfig struct {
	MarketCapThreshold   float64 `json:"marketCapThreshold"`
	MinBuysForEntry      int     `json:"minBuysForEntry"`
	EntryTimeWindowSec   int     `json:"entryTimeWindowSec"`
	TakeProfitPct        float64 `json:"takeProfitPct"`
	StopLossPct          float64 `json:"stopLossPct"`
	MaxHoldTimeSec       int     `json:"maxHoldTimeSec"`
	FixedPositionSizeSol float64 `json:"fixedPositionSizeSol"`
	InitialBalance       float64 `json:"initialBalance"`
}

// NewSimulationService creates a new simulation service
func NewSimulationService(
	db *sql.DB,
	strategyRepo repository.StrategyRepositoryInterface,
	tokenRepo repository.TokenRepositoryInterface,
	tradeRepo repository.TradeRepositoryInterface,
	simulatedTradeRepo repository.SimulatedTradeRepositoryInterface,
	strategyMetricRepo repository.StrategyMetricRepositoryInterface,
	simulationRunRepo repository.SimulationRunRepositoryInterface,
	wsHub *websocket.WSHub,
	logger *logger.Logger,
) *SimulationService {
	const maxConcurrentWorkers = 50 // Increased limit for concurrent goroutines

	simulationEventRepo := repository.NewSimulationEventRepository(db)
	if simulationEventRepo == nil {
		logger.Error("Failed to create simulation event repository")
	} else {
		logger.Info("Successfully created simulation event repository")
	}

	service := &SimulationService{
		db:                  db,
		strategyRepo:        strategyRepo,
		tokenRepo:           tokenRepo,
		tradeRepo:           tradeRepo,
		simulatedTradeRepo:  simulatedTradeRepo,
		strategyMetricRepo:  strategyMetricRepo,
		simulationRunRepo:   simulationRunRepo,
		simulationEventRepo: simulationEventRepo,
		logger:              logger,
		wsHub:               wsHub,
		activeSims:          make(map[int64]*SimulationContext),
		simulationDone:      make(chan int64, 10),
		workerPool:          make(chan struct{}, maxConcurrentWorkers), // Worker pool for limiting goroutines
		shutdownCh:          make(chan struct{}),
	}

	// Start background monitoring of simulations
	go service.monitorSimulations()

	return service
}

// Shutdown gracefully shuts down the service
func (s *SimulationService) Shutdown() {
	s.logger.Info("Shutting down simulation service...")

	// Signal all simulations to stop
	s.activeSimsMu.Lock()
	for _, sim := range s.activeSims {
		sim.StopRequested = true
		if sim.cancel != nil {
			sim.cancel()
		}
	}
	s.activeSimsMu.Unlock()

	// Close shutdown channel
	close(s.shutdownCh)

	// Wait for monitor goroutine to exit
	s.logger.Info("Shutdown complete")
}

// monitorSimulations cleans up completed simulations
func (s *SimulationService) monitorSimulations() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdownCh:
			return
		case strategyID := <-s.simulationDone:
			s.logger.Info("Simulation for strategy %d completed", strategyID)
			s.cleanupSimulation(strategyID)
		case <-ticker.C:
			// Periodically check for stalled simulations
			s.checkStalledSimulations()
		}
	}
}

// cleanupSimulation removes a simulation from the active simulations map
func (s *SimulationService) cleanupSimulation(strategyID int64) {
	s.activeSimsMu.Lock()
	defer s.activeSimsMu.Unlock()

	sim, exists := s.activeSims[strategyID]
	if exists {
		// Mark as not running before removing
		sim.mu.Lock()
		sim.IsRunning = false
		sim.mu.Unlock()

		// Wait for all goroutines to finish before removing
		if sim.cancel != nil {
			sim.cancel() // Cancel all goroutines
		}
		sim.wg.Wait() // Wait for all goroutines to finish

		// Remove from active simulations map
		delete(s.activeSims, strategyID)
		s.logger.Info("Simulation for strategy %d cleaned up", strategyID)
	} else {
		s.logger.Warn("Attempted to cleanup simulation for strategy %d but it wasn't in active simulations map", strategyID)
	}
}

// checkStalledSimulations checks for simulations that may have stalled
func (s *SimulationService) checkStalledSimulations() {
	s.activeSimsMu.Lock()
	defer s.activeSimsMu.Unlock()

	now := time.Now()
	for strategyID, sim := range s.activeSims {
		sim.mu.RLock()
		isRunning := sim.IsRunning
		startTime := sim.StartTime
		sim.mu.RUnlock()

		// If simulation has been running for more than 1 hour, mark it for cleanup
		if isRunning && now.Sub(startTime) > 1*time.Hour {
			s.logger.Warn("Simulation for strategy %d has been running for over 1 hour, marking for cleanup", strategyID)
			sim.mu.Lock()
			sim.StopRequested = true
			sim.mu.Unlock()

			if sim.cancel != nil {
				sim.cancel() // Cancel all goroutines
			}
		}
	}
}

// ForceCleanSimulation force-removes a simulation from the simulation map
func (s *SimulationService) ForceCleanSimulation(strategyID int64) {
	s.activeSimsMu.Lock()
	defer s.activeSimsMu.Unlock()

	// Simply remove from the map without waiting for anything
	delete(s.activeSims, strategyID)
	s.logger.Info("Forced cleanup of simulation for strategy %d", strategyID)
}

// StartSimulation starts a simulation for a strategy
func (s *SimulationService) StartSimulation(strategyID int64) error {
	// Force clean any existing simulation first to avoid "already running" errors
	s.ForceCleanSimulation(strategyID)

	// Check if the strategy exists
	strategy, err := s.strategyRepo.GetByID(strategyID)
	if err != nil {
		return fmt.Errorf("error fetching strategy: %v", err)
	}
	if strategy == nil {
		return fmt.Errorf("strategy not found: %d", strategyID)
	}

	// Parse the strategy configuration
	var config StrategyConfig
	configData, err := json.Marshal(strategy.Config)
	if err != nil {
		return fmt.Errorf("error marshaling strategy config: %v", err)
	}

	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("invalid strategy configuration: %v", err)
	}

	// Validate configuration
	if err := validateStrategyConfig(&config); err != nil {
		return fmt.Errorf("invalid strategy configuration: %v", err)
	}

	// Check if there's already an active simulation for this strategy
	s.activeSimsMu.Lock()
	defer s.activeSimsMu.Unlock()

	// Create a simulation run record in the database
	simulationRun := &models.SimulationRun{
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour), // Default to 1 hour max runtime
		Status:    "running",
		SimulationParameters: models.JSONB{
			"strategyID":     strategyID,
			"initialBalance": config.InitialBalance,
			"positionSize":   config.FixedPositionSizeSol,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	simulationRunID, err := s.simulationRunRepo.Save(simulationRun)
	if err != nil {
		return fmt.Errorf("error creating simulation run record: %v", err)
	}

	// Create cancellation context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new simulation context
	context := &SimulationContext{
		StrategyID:      strategyID,
		Strategy:        strategy,
		Config:          config,
		StartTime:       time.Now(),
		Trades:          make([]*models.SimulatedTrade, 0),
		IsRunning:       true,
		CurrentBalance:  config.InitialBalance,
		InitialBalance:  config.InitialBalance,
		SimulationRunID: simulationRunID,
		ctx:             ctx,
		cancel:          cancel,
	}
	s.activeSims[strategyID] = context

	// Start the simulation in a goroutine
	go s.runSimulation(context)

	return nil
}

// validateStrategyConfig validates the strategy configuration
func validateStrategyConfig(config *StrategyConfig) error {
	if config.InitialBalance <= 0 {
		return fmt.Errorf("initial balance must be positive")
	}
	if config.FixedPositionSizeSol <= 0 {
		return fmt.Errorf("position size must be positive")
	}
	if config.MaxHoldTimeSec <= 0 {
		return fmt.Errorf("max hold time must be positive")
	}
	return nil
}

// StopSimulation stops an active simulation
func (s *SimulationService) StopSimulation(strategyID int64) error {
	s.activeSimsMu.RLock()
	sim, exists := s.activeSims[strategyID]
	isRunning := false
	if exists {
		sim.mu.RLock()
		isRunning = sim.IsRunning
		sim.mu.RUnlock()
	}
	s.activeSimsMu.RUnlock()

	if !exists || !isRunning {
		s.cleanupSimulation(strategyID)
		return fmt.Errorf("no active simulation found for strategy %d", strategyID)
	}

	// Mark for stopping
	sim.mu.Lock()
	sim.StopRequested = true
	sim.IsRunning = false
	sim.mu.Unlock()

	// Cancel all goroutines
	if sim.cancel != nil {
		sim.cancel()
	}

	go func() {
		time.Sleep(1 * time.Second)
		s.simulationDone <- strategyID
	}()

	s.logger.Info("Simulation for strategy %d marked for stopping", strategyID)
	return nil
}

// GetSimulationStatus returns the status of a simulation
func (s *SimulationService) GetSimulationStatus(strategyID int64) (*SimulationContext, error) {
	s.activeSimsMu.RLock()
	defer s.activeSimsMu.RUnlock()

	sim, exists := s.activeSims[strategyID]
	if !exists {
		return nil, fmt.Errorf("no simulation found for strategy %d", strategyID)
	}

	return sim, nil
}

// runSimulation runs the actual simulation
func (s *SimulationService) runSimulation(ctx *SimulationContext) {
	s.logger.Info("Starting simulation for strategy %d: %s", ctx.StrategyID, ctx.Strategy.Name)

	// Notify about simulation start
	s.sendSimulationEvent(ctx, "simulation_started", nil)

	// Define simulation parameters - increasing interval to avoid too frequent evaluations
	// which could lead to excessive database load and rapidly depleting the available token pool
	iterationInterval := 10 * time.Second // Increased from 3 seconds to 10 seconds for more sustainable simulation

	// Create a ticker for iteration intervals
	ticker := time.NewTicker(iterationInterval)
	defer ticker.Stop()

	// Create a done channel for this simulation
	done := make(chan bool)
	iteration := 0

	// Run the simulation loop
	go func() {
		defer close(done)

		for {
			// Check if simulation stop was requested
			ctx.mu.RLock()
			if ctx.StopRequested || !ctx.IsRunning {
				ctx.mu.RUnlock()
				s.logger.Info("Simulation stop requested for strategy %d", ctx.StrategyID)
				break
			}
			ctx.mu.RUnlock()

			// Run one iteration of the simulation
			if err := s.runSimulationIteration(ctx); err != nil {
				s.logger.Error("Error in simulation iteration: %v", err)
				break
			}

			iteration++
			s.logger.Info("Completed iteration %d for strategy %d", iteration, ctx.StrategyID)

			// Sleep for the iteration interval
			select {
			case <-ticker.C:
				// Continue to next iteration
			case <-ctx.ctx.Done():
				// Context cancelled, exit loop
				return
			}
		}

		// Mark simulation as complete when stopped
		ctx.mu.Lock()
		ctx.IsRunning = false
		ctx.StopRequested = true // Make sure it's marked as requested to stop
		ctx.mu.Unlock()

		// Update simulation run status in database
		if err := s.simulationRunRepo.UpdateStatus(ctx.SimulationRunID, "completed"); err != nil {
			s.logger.Error("Error updating simulation run status: %v", err)
		}

		// Calculate and save simulation metrics
		if err := s.saveSimulationMetrics(ctx); err != nil {
			s.logger.Error("Error saving simulation metrics: %v", err)
		}

		// Send simulation completed event
		s.sendSimulationEvent(ctx, "simulation_completed", map[string]interface{}{
			"total_iterations":   iteration,
			"execution_time_sec": time.Since(ctx.StartTime).Seconds(),
		})

		s.logger.Info("Simulation stopped for strategy %d: %s", ctx.StrategyID, ctx.Strategy.Name)

		// Wait for all goroutines to finish
		ctx.wg.Wait()

		// Signal that simulation is done - make sure this happens by adding timeout
		select {
		case s.simulationDone <- ctx.StrategyID:
			s.logger.Info("Sent simulation done signal for strategy %d", ctx.StrategyID)
		case <-time.After(5 * time.Second):
			s.logger.Error("Timed out trying to send simulation done signal for strategy %d, cleaning up directly", ctx.StrategyID)
			// Force cleanup if channel is blocked
			go s.cleanupSimulation(ctx.StrategyID)
		}
	}()
}

// runSimulationIteration runs a single iteration of the simulation
func (s *SimulationService) runSimulationIteration(ctx *SimulationContext) error {
	// Fetch tokens to evaluate - using a much longer age window to get more tokens
	// Increased from 5 minutes to 24 hours to have more tokens for trading
	maxAgeSec := int64(86400)                                                                   // 24 hours instead of 5 minutes
	tokens, err := s.tokenRepo.GetFilteredTokens(ctx.Config.MarketCapThreshold, maxAgeSec, 200) // Increased limit from 100 to 200
	if err != nil {
		return fmt.Errorf("error fetching tokens for simulation: %v", err)
	}

	s.logger.Info("Found %d tokens to evaluate for strategy %d", len(tokens), ctx.StrategyID)

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].CreatedTimestamp > tokens[j].CreatedTimestamp
	})

	// Process tokens in parallel with a worker pool
	var processWg sync.WaitGroup

	// Check if we already processed this token ID in this iteration
	processedTokens := make(map[int64]bool)

	for i, token := range tokens {
		tokenID := token.ID

		// Skip tokens we already processed in this iteration
		if _, alreadyProcessed := processedTokens[tokenID]; alreadyProcessed {
			continue
		}
		processedTokens[tokenID] = true

		// Check if simulation stop was requested
		ctx.mu.RLock()
		if ctx.StopRequested || !ctx.IsRunning {
			ctx.mu.RUnlock()
			s.logger.Info("Simulation stop requested during token evaluation")
			break
		}
		ctx.mu.RUnlock()

		// Log progress periodically
		if i > 0 && i%10 == 0 {
			s.logger.Info("Simulation progress: %d/%d tokens evaluated", i, len(tokens))
		}

		// Check if we already have a trade for this token
		if s.hasExistingTrade(ctx, token.ID) {
			continue // Skip tokens we've already evaluated
		}

		// Add to wait group and start evaluation in worker pool
		processWg.Add(1)
		ctx.wg.Add(1) // Add to context wait group for overall tracking

		// Use worker pool to limit concurrency
		go func(token *models.Token) {
			// Acquire worker slot
			s.workerPool <- struct{}{}
			defer func() {
				// Release worker slot
				<-s.workerPool
				processWg.Done()
				ctx.wg.Done()
			}()

			// Evaluate token with strategy
			if err := s.evaluateToken(ctx, token); err != nil {
				s.logger.Error("Error evaluating token %s: %v", token.MintAddress, err)
			}
		}(token)
	}

	// Wait for all token evaluations to complete
	processWg.Wait()

	return nil
}

// hasExistingTrade checks if we already have any trade (active or completed) for this token
func (s *SimulationService) hasExistingTrade(ctx *SimulationContext, tokenID int64) bool {
	ctx.tokensMu.RLock()
	defer ctx.tokensMu.RUnlock()

	for _, existingTrade := range ctx.Trades {
		if existingTrade.TokenID == tokenID {
			return true
		}
	}
	return false
}

// hasActiveTradeForToken checks if we already have an ACTIVE trade for this token
func (s *SimulationService) hasActiveTradeForToken(ctx *SimulationContext, tokenID int64) bool {
	ctx.tokensMu.RLock()
	defer ctx.tokensMu.RUnlock()

	for _, existingTrade := range ctx.Trades {
		if existingTrade.TokenID == tokenID && existingTrade.Status == "active" {
			return true
		}
	}
	return false
}

// evaluateToken evaluates a token against a strategy
func (s *SimulationService) evaluateToken(ctx *SimulationContext, token *models.Token) error {
	// Check if context is cancelled
	select {
	case <-ctx.ctx.Done():
		return fmt.Errorf("evaluation cancelled")
	default:
		// Continue execution
	}

	// Check if we have enough balance for a trade
	ctx.mu.RLock()
	currentBalance := ctx.CurrentBalance
	positionSize := ctx.Config.FixedPositionSizeSol
	ctx.mu.RUnlock()

	if currentBalance < positionSize {
		s.logger.Info("Insufficient balance (%.6f SOL) for position size (%.6f SOL), skipping token %s",
			currentBalance, positionSize, token.Symbol)

		// Send event for balance depletion
		ctx.mu.RLock()
		isRunning := ctx.IsRunning
		ctx.mu.RUnlock()

		if isRunning {
			ctx.mu.Lock()
			ctx.StopRequested = true
			ctx.mu.Unlock()

			s.sendSimulationEvent(ctx, "simulation_balance_depleted", map[string]interface{}{
				"remaining_balance": currentBalance,
				"position_size":     positionSize,
				"timestamp":         time.Now().Unix(),
			})
		}
		return nil
	}

	// We've removed the token age restriction here that was previously limiting evaluation
	// to only tokens less than 3 minutes old. This restriction was causing simulations
	// to end prematurely after only 1-2 trades because there weren't enough qualifying tokens.
	// Now we evaluate all tokens regardless of age, allowing for longer-running simulations
	// with more trading opportunities.
	/*
		now := time.Now().Unix()
		if now-token.CreatedTimestamp > 180 { // 3 min = 180 seconds
			return nil // Skip older tokens
		}
	*/

	// Check if token meets basic criteria like market cap threshold
	const marketCapLowerTolerance = -10.0
	const marketCapUpperTolerance = 40.0
	marketCapLowerLimit := ctx.Config.MarketCapThreshold * (1.0 + marketCapLowerTolerance/100.0)
	marketCapUpperLimit := ctx.Config.MarketCapThreshold * (1.0 + marketCapUpperTolerance/100.0)

	if token.UsdMarketCap < marketCapLowerLimit {
		return nil // Skip tokens below the lower limit
	}

	if token.UsdMarketCap > marketCapUpperLimit {
		s.logger.Debug("Token %s (%s) exceeds market cap upper limit: $%.2f > $%.2f",
			token.Symbol, token.Name, token.UsdMarketCap, marketCapUpperLimit)
		return nil // Skip tokens above the upper tolerance limit
	}

	// Get recent trades for this token
	trades, err := s.tradeRepo.GetTradesByTokenID(token.ID, 50)
	if err != nil {
		return fmt.Errorf("error fetching trades: %v", err)
	}

	if len(trades) == 0 {
		return nil // Skip tokens with no trades
	}

	// Analyze trades based on strategy
	entrySignal, entrySignalData := s.analyzeEntrySignal(ctx, token, trades)
	if !entrySignal {
		return nil // No entry signal detected
	}

	// Add random chance (40%) to skip some qualifying tokens
	// This helps pace the simulation and avoid depleting all trading opportunities too quickly
	// Creating a more natural trading pattern and extending simulation runtime
	if rand.Float64() > 0.6 {
		return nil // Randomly skip this trade opportunity
	}

	entryPrice, err := s.calculateConsistentPrice(token.ID)
	if err != nil {
		return fmt.Errorf("cannot calculate entry price: %v", err)
	}

	// Create a simulated trade for this token
	simTrade := &models.SimulatedTrade{
		StrategyID:        ctx.StrategyID,
		TokenID:           token.ID,
		EntryPrice:        entryPrice,
		EntryTimestamp:    time.Now().Unix(),
		EntryUsdMarketCap: token.UsdMarketCap,
		PositionSize:      positionSize,
		Status:            "active",
		SimulationRunID:   &ctx.SimulationRunID,
	}

	// Update balance (with mutex protection)
	ctx.mu.Lock()
	ctx.CurrentBalance -= positionSize
	ctx.mu.Unlock()

	// Save to database
	tradeID, err := s.simulatedTradeRepo.Save(simTrade)
	if err != nil {
		// Revert balance deduction
		ctx.mu.Lock()
		ctx.CurrentBalance += positionSize
		ctx.mu.Unlock()

		s.logger.Error("Error saving simulated trade: %v", err)
		// Check for critical database errors
		if strings.Contains(err.Error(), "connection") {
			s.logger.Error("Database connection issue, might need to stop simulation")
		}
		return fmt.Errorf("error saving simulated trade to database: %v", err)
	}
	simTrade.ID = tradeID

	// Add to in-memory list of trades (with mutex protection)
	ctx.tokensMu.Lock()
	ctx.Trades = append(ctx.Trades, simTrade)
	ctx.tokensMu.Unlock()

	// Send trade event
	s.sendSimulationEvent(ctx, "trade_executed", map[string]interface{}{
		"token_id":         token.ID,
		"token_symbol":     token.Symbol,
		"token_name":       token.Name,
		"token_mint":       token.MintAddress,
		"image_url":        token.ImageUrl,
		"twitter_url":      token.TwitterUrl,
		"website_url":      token.WebsiteUrl,
		"action":           "buy",
		"price":            entryPrice,
		"amount":           positionSize,
		"timestamp":        time.Now().Unix(),
		"entry_market_cap": token.UsdMarketCap,
		"usd_market_cap":   token.UsdMarketCap,
		"current_balance":  ctx.CurrentBalance,
		"signal_data":      entrySignalData,
	})

	s.logger.Info("Trade opened for %s: Entry Price: %.6f, Balance remaining: %.6f SOL",
		token.Symbol, entryPrice, ctx.CurrentBalance)

	// Start trade monitoring in a separate goroutine
	ctx.wg.Add(1)
	go s.monitorTrade(ctx, simTrade, token)

	return nil
}

// monitorTrade monitors an active trade for exit conditions
func (s *SimulationService) monitorTrade(ctx *SimulationContext, trade *models.SimulatedTrade, token *models.Token) {
	defer ctx.wg.Done()

	// Define max hold time from strategy
	maxHoldTime := time.Duration(ctx.Config.MaxHoldTimeSec) * time.Second

	// Define take profit and stop loss levels
	entryPrice := trade.EntryPrice
	takeProfitLevel := entryPrice * (1 + ctx.Config.TakeProfitPct/100)
	stopLossLevel := entryPrice * (1 - ctx.Config.StopLossPct/100)

	s.logger.Info("Trade opened for %s: Entry Price: %.6f, Take Profit: %.6f (%.1f%%), Stop Loss: %.6f (%.1f%%)",
		token.Symbol, entryPrice, takeProfitLevel, ctx.Config.TakeProfitPct,
		stopLossLevel, ctx.Config.StopLossPct)

	// Loop to check prices at regular intervals
	ticker := time.NewTicker(10 * time.Second) // Increased from 3 seconds to 10 seconds for consistency with simulation interval
	defer ticker.Stop()

	entryTime := time.Unix(trade.EntryTimestamp, 0)
	deadline := entryTime.Add(maxHoldTime)

	// Track last checked price
	var lastCheckedPrice float64

	for {
		select {
		case <-ticker.C:
			// Check if simulation is still running
			ctx.mu.RLock()
			stopRequested := ctx.StopRequested
			isRunning := ctx.IsRunning
			ctx.mu.RUnlock()

			if stopRequested || !isRunning {
				s.closeTradeWithReason(trade, token, entryPrice, "simulation_stopped", ctx)
				return
			}

			// Check if max hold time has elapsed
			if time.Now().After(deadline) {
				currentPrice, err := s.calculatePriceWithFallback(token.ID, entryPrice)
				if err != nil {
					s.logger.Error("Error calculating exit price: %v, using entry price", err)
					currentPrice = entryPrice
				}
				s.closeTradeWithReason(trade, token, currentPrice, "max_hold_time", ctx)
				return
			}

			// Get latest token data from database
			latestToken, err := s.tokenRepo.GetByID(token.ID)
			if err != nil {
				s.logger.Error("Error getting latest token data: %v", err)
				continue
			}

			if latestToken == nil {
				s.logger.Error("Token not found in database: %d", token.ID)
				continue
			}

			// Calculate current price based on latest trades
			currentPrice, err := s.calculatePriceWithRefresh(token.ID)
			if err != nil {
				s.logger.Debug("Error calculating price for %s: %v", token.Symbol, err)
				continue // Skip this check
			}

			// Log if price has changed since last check
			s.logger.Debug("Price check for %s: current=%.6f, previous=%.6f, TP=%.6f, SL=%.6f",
				token.Symbol, currentPrice, lastCheckedPrice, takeProfitLevel, stopLossLevel)

			// Update last checked price
			lastCheckedPrice = currentPrice

			// Check exit conditions
			exitReason := ""

			// Take profit condition
			if currentPrice >= takeProfitLevel {
				exitReason = "take_profit"
			} else if currentPrice <= stopLossLevel { // Stop loss condition
				exitReason = "stop_loss"
			}

			// If we have an exit reason, close the trade
			if exitReason != "" {
				s.closeTradeWithReason(trade, token, currentPrice, exitReason, ctx)
				return
			}

		case <-ctx.ctx.Done():
			// Context canceled, close trade
			s.closeTradeWithReason(trade, token, entryPrice, "simulation_stopped", ctx)
			return
		}
	}
}

// calculatePriceWithFallback calculates price with fallback to entry price
func (s *SimulationService) calculatePriceWithFallback(tokenID int64, fallbackPrice float64) (float64, error) {
	price, err := s.calculateConsistentPrice(tokenID)
	if err != nil {
		return fallbackPrice, err
	}
	return price, nil
}

// closeTradeWithReason closes a trade with the specified reason
func (s *SimulationService) closeTradeWithReason(trade *models.SimulatedTrade, token *models.Token, currentPrice float64, exitReason string, ctx *SimulationContext) {
	entryPrice := trade.EntryPrice

	exitPrice := currentPrice
	exitTime := time.Now().Unix()

	// Update the trade with exit information
	trade.ExitPrice = &exitPrice
	trade.ExitTimestamp = &exitTime
	trade.Status = "completed"
	trade.ExitReason = &exitReason

	// Calculate PnL
	pnlAmount := (currentPrice/entryPrice - 1.0) * trade.PositionSize
	trade.ProfitLoss = &pnlAmount

	// Get latest market cap
	latestToken, err := s.tokenRepo.GetByID(token.ID)
	if err == nil && latestToken != nil {
		trade.ExitUsdMarketCap = &latestToken.UsdMarketCap
		token = latestToken
	} else {
		exitMarketCap := token.UsdMarketCap
		trade.ExitUsdMarketCap = &exitMarketCap
		s.logger.Warn("Couldn't get latest token data, using existing market cap")
	}

	// Update balance (safely)
	ctx.mu.Lock()
	// Ensure we don't go negative by capping the loss
	returnAmount := trade.PositionSize + pnlAmount
	if returnAmount < 0 {
		s.logger.Warn("Trade resulted in complete loss, capping at position size")
		returnAmount = 0 // Cap at zero to avoid negative balance
	}
	ctx.CurrentBalance += returnAmount
	ctx.mu.Unlock()

	// Calculate PnL percentage
	profitLossPct := (currentPrice/entryPrice - 1.0) * 100

	s.logger.Info("Closing trade for %s: Reason: %s, PnL: %.2f%%, New Balance: %.6f SOL",
		token.Symbol, exitReason, profitLossPct, ctx.CurrentBalance)

	// Update in database
	if err := s.simulatedTradeRepo.Update(trade); err != nil {
		s.logger.Error("Error updating simulated trade: %v", err)
	}

	// Send trade exit event
	s.sendSimulationEvent(ctx, "trade_closed", map[string]interface{}{
		"token_id":         token.ID,
		"token_symbol":     token.Symbol,
		"token_name":       token.Name,
		"token_mint":       token.MintAddress,
		"image_url":        token.ImageUrl,
		"twitter_url":      token.TwitterUrl,
		"website_url":      token.WebsiteUrl,
		"action":           "sell",
		"entry_price":      entryPrice,
		"exit_price":       currentPrice,
		"profit_loss":      pnlAmount,
		"profit_loss_pct":  profitLossPct,
		"exit_reason":      exitReason,
		"timestamp":        time.Now().Unix(),
		"entry_market_cap": trade.EntryUsdMarketCap,
		"exit_market_cap":  token.UsdMarketCap,
		"usd_market_cap":   token.UsdMarketCap,
	})
}

// analyzeEntrySignal determines if a token should be bought based on strategy rules
func (s *SimulationService) analyzeEntrySignal(ctx *SimulationContext, token *models.Token, trades []*models.Trade) (bool, map[string]interface{}) {
	// Count buy transactions in the time window
	buyCount := 0
	var latestPrice float64

	// Track additional signal data for logging/debugging
	signalData := make(map[string]interface{})

	// Get current time
	now := time.Now().Unix()

	// Calculate lookback window
	lookbackTime := now - int64(ctx.Config.EntryTimeWindowSec)

	// Analyze trades
	for _, trade := range trades {
		// Only look at recent trades within our time window
		if trade.Timestamp < lookbackTime {
			continue
		}

		// Count buys
		if trade.IsBuy {
			buyCount++
		}

		// Track latest price (assuming last trade price is representative)
		if trade.TokenAmount > 0 {
			latestPrice = trade.SolAmount / trade.TokenAmount
		}
	}

	// Save signal analysis data
	signalData["buy_count"] = buyCount
	signalData["min_buys_required"] = ctx.Config.MinBuysForEntry
	signalData["lookback_window_sec"] = ctx.Config.EntryTimeWindowSec
	signalData["latest_price"] = latestPrice

	// Check if we meet the minimum buys threshold
	return buyCount >= ctx.Config.MinBuysForEntry, signalData
}

// sendSimulationEvent sends a simulation event via WebSocket
func (s *SimulationService) sendSimulationEvent(ctx *SimulationContext, eventType string, data map[string]interface{}) {
	if s.simulationEventRepo == nil {
		s.logger.Error("simulationEventRepo is nil, cannot save event")
		return
	}

	if s.wsHub == nil {
		s.logger.Error("wsHub is nil, cannot broadcast event")
		return
	}

	event := map[string]interface{}{
		"type":        eventType,
		"strategy_id": ctx.StrategyID,
		"timestamp":   time.Now().Unix(),
	}

	// Add additional data if provided
	if data != nil {
		for k, v := range data {
			event[k] = v
		}
	}

	// Create event model
	simulationEvent := &models.SimulationEvent{
		StrategyID:      ctx.StrategyID,
		SimulationRunID: ctx.SimulationRunID,
		EventType:       eventType,
		EventData:       models.JSONB(event),
		Timestamp:       time.Now(),
		CreatedAt:       time.Now(),
	}

	// Save event to database
	_, err := s.simulationEventRepo.Save(simulationEvent)
	if err != nil {
		s.logger.Error("Error saving simulation event to database: %v", err)
	}

	// Broadcast the event
	s.wsHub.BroadcastJSON(event)
}

// GetSimulationSummary returns a summary of all simulated trades for a strategy
func (s *SimulationService) GetSimulationSummary(strategyID int64) (map[string]interface{}, error) {
	// First check if there's an active simulation
	s.activeSimsMu.RLock()
	sim, exists := s.activeSims[strategyID]
	s.activeSimsMu.RUnlock()

	if exists && sim.IsRunning {
		// If simulation is running, calculate summary from memory
		return s.calculateInMemorySummary(sim), nil
	}

	// Otherwise, get summary from database
	return s.simulatedTradeRepo.GetSummaryByStrategyID(strategyID)
}

// calculateInMemorySummary calculates summary from in-memory simulation context
func (s *SimulationService) calculateInMemorySummary(sim *SimulationContext) map[string]interface{} {
	// Lock for reading trades
	sim.tokensMu.RLock()
	defer sim.tokensMu.RUnlock()

	// Lock for reading balance
	sim.mu.RLock()
	currentBalance := sim.CurrentBalance
	initialBalance := sim.InitialBalance
	isRunning := sim.IsRunning
	startTime := sim.StartTime
	sim.mu.RUnlock()

	var totalTrades, profitableTrades, lossTrades int
	var totalProfit, totalLoss, totalInvestment float64
	var maxDrawdown float64

	// Calculate running balance for drawdown tracking
	runningBalance := initialBalance
	peakBalance := initialBalance

	// Track trade PnLs for performance metrics
	tradePnLs := make([]float64, 0)

	for _, trade := range sim.Trades {
		if trade.Status == "completed" || trade.Status == "closed" {
			totalTrades++
			totalInvestment += trade.PositionSize

			if trade.ProfitLoss != nil {
				tradePnLs = append(tradePnLs, *trade.ProfitLoss)

				// Update running balance
				runningBalance += *trade.ProfitLoss

				// Track peak balance for drawdown calculation
				if runningBalance > peakBalance {
					peakBalance = runningBalance
				}

				// Calculate drawdown
				currentDrawdown := (peakBalance - runningBalance) / peakBalance * 100
				if currentDrawdown > maxDrawdown {
					maxDrawdown = currentDrawdown
				}

				if *trade.ProfitLoss > 0 {
					profitableTrades++
					totalProfit += *trade.ProfitLoss
				} else {
					lossTrades++
					totalLoss += *trade.ProfitLoss
				}
			}
		}
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(profitableTrades) / float64(totalTrades) * 100
	}

	roi := 0.0
	if initialBalance > 0 {
		roi = ((currentBalance / initialBalance) - 1) * 100
	}

	// Calculate average profit per trade
	avgProfit := 0.0
	if profitableTrades > 0 {
		avgProfit = totalProfit / float64(profitableTrades)
	}

	// Calculate average loss per trade
	avgLoss := 0.0
	if lossTrades > 0 {
		avgLoss = totalLoss / float64(lossTrades)
	}

	return map[string]interface{}{
		"strategy_id":       sim.StrategyID,
		"strategy_name":     sim.Strategy.Name,
		"is_running":        isRunning,
		"start_time":        startTime.Unix(),
		"execution_time":    time.Since(startTime).Seconds(),
		"total_trades":      totalTrades,
		"profitable_trades": profitableTrades,
		"losing_trades":     lossTrades,
		"win_rate":          winRate,
		"total_profit":      totalProfit,
		"total_loss":        totalLoss,
		"avg_profit":        avgProfit,
		"avg_loss":          avgLoss,
		"max_drawdown":      maxDrawdown,
		"net_pnl":           totalProfit + totalLoss,
		"initial_balance":   initialBalance,
		"current_balance":   currentBalance,
		"roi":               roi,
	}
}

// calculatePriceWithRefresh calculates current price based on latest trades
func (s *SimulationService) calculatePriceWithRefresh(tokenID int64) (float64, error) {
	// Get recent trades
	latestTrades, err := s.tradeRepo.GetTradesByTokenID(tokenID, 10)
	if err != nil {
		return 0, fmt.Errorf("error fetching trades: %v", err)
	}

	if len(latestTrades) == 0 {
		return 0, fmt.Errorf("no trades found for token %d", tokenID)
	}

	// Apply weighted average calculation
	var weightedSum, totalVolume float64
	tradeCount := 0

	// Add random variability to simulate market movement (for simulation purposes)
	randomFactor := 1.0 + (rand.Float64()*0.02 - 0.01) // Random variation between -1% and +1%

	for _, t := range latestTrades {
		if t.TokenAmount > 0 {
			price := t.SolAmount / t.TokenAmount
			volume := t.SolAmount

			// Validate price is reasonable (filters out extreme values)
			if price > 0 && price < 1e10 { // Reasonable price range check
				weightedSum += price * volume
				totalVolume += volume
				tradeCount++
			}
		}
	}

	if totalVolume > 0 && tradeCount >= 3 {
		return (weightedSum / totalVolume) * randomFactor, nil
	}

	// Fallback to simple average if not enough volume-weighted data
	if tradeCount > 0 {
		var sum float64
		count := 0

		for _, t := range latestTrades {
			if t.TokenAmount > 0 {
				price := t.SolAmount / t.TokenAmount
				if price > 0 && price < 1e10 {
					sum += price
					count++
				}
			}
		}

		if count > 0 {
			return (sum / float64(count)) * randomFactor, nil
		}
	}

	return 0, fmt.Errorf("insufficient valid trade data for price calculation")
}

// calculateConsistentPrice calculates a consistent price based on recent trades
func (s *SimulationService) calculateConsistentPrice(tokenID int64) (float64, error) {
	return s.calculatePriceWithRefresh(tokenID)
}

// saveSimulationMetrics saves the performance metrics for a completed simulation
func (s *SimulationService) saveSimulationMetrics(ctx *SimulationContext) error {
	// Calculate metrics
	metrics := s.calculateInMemorySummary(ctx)

	// Create strategy metric model
	strategyMetric := &models.StrategyMetric{
		StrategyID:       ctx.StrategyID,
		SimulationRunID:  &ctx.SimulationRunID,
		WinRate:          metrics["win_rate"].(float64),
		AvgProfit:        metrics["avg_profit"].(float64),
		AvgLoss:          metrics["avg_loss"].(float64),
		MaxDrawdown:      metrics["max_drawdown"].(float64),
		TotalTrades:      metrics["total_trades"].(int),
		SuccessfulTrades: metrics["profitable_trades"].(int),
		RiskScore:        ctx.Strategy.RiskScore,
		CreatedAt:        time.Now(),
	}

	// Save to database
	metricID, err := s.strategyMetricRepo.Save(strategyMetric)
	if err != nil {
		return fmt.Errorf("error saving strategy metrics: %v", err)
	}

	s.logger.Info("Saved strategy metrics with ID %d for strategy %d", metricID, ctx.StrategyID)

	// If this is a winning strategy and ROI is positive, update the simulation run with this strategy as winner
	roi := metrics["roi"].(float64)
	if roi > 0 {
		// Update simulation run with the winning strategy
		if err := s.simulationRunRepo.UpdateWinner(ctx.SimulationRunID, ctx.StrategyID); err != nil {
			s.logger.Error("Error updating simulation run winner: %v", err)
		} else {
			s.logger.Info("Strategy %d set as winner for simulation run %d with ROI: %.2f%%",
				ctx.StrategyID, ctx.SimulationRunID, roi)
		}
	}

	return nil
}
