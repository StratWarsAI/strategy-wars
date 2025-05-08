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

	"github.com/StratWarsAI/strategy-wars/internal/api/dto"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
	"github.com/StratWarsAI/strategy-wars/internal/websocket"
)

// SimulationService handles strategy simulations
type SimulationService struct {
	db                   *sql.DB
	strategyRepo         repository.StrategyRepositoryInterface
	tokenRepo            repository.TokenRepositoryInterface
	tradeRepo            repository.TradeRepositoryInterface
	simulatedTradeRepo   repository.SimulatedTradeRepositoryInterface
	strategyMetricRepo   repository.StrategyMetricRepositoryInterface
	simulationRunRepo    repository.SimulationRunRepositoryInterface
	simulationEventRepo  repository.SimulationEventRepositoryInterface
	simulationResultRepo repository.SimulationResultRepositoryInterface
	logger               *logger.Logger
	wsHub                *websocket.WSHub
	activeSimsMu         sync.RWMutex
	activeSims           map[int64]*SimulationContext
	simulationDone       chan int64
	workerPool           chan struct{} // Limit concurrent token evaluations
	shutdownCh           chan struct{} // Channel for graceful shutdown
}

// SimulationContext holds the context for an active simulation
type SimulationContext struct {
	StrategyID      int64
	Strategy        *models.Strategy
	Config          models.StrategyConfig
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

	simulationResultRepo := repository.NewSimulationResultRepository(db)
	if simulationResultRepo == nil {
		logger.Error("Failed to create simulation result repository")
	} else {
		logger.Info("Successfully created simulation result repository")
	}

	service := &SimulationService{
		db:                   db,
		strategyRepo:         strategyRepo,
		tokenRepo:            tokenRepo,
		tradeRepo:            tradeRepo,
		simulatedTradeRepo:   simulatedTradeRepo,
		strategyMetricRepo:   strategyMetricRepo,
		simulationRunRepo:    simulationRunRepo,
		simulationEventRepo:  simulationEventRepo,
		simulationResultRepo: simulationResultRepo,
		logger:               logger,
		wsHub:                wsHub,
		activeSims:           make(map[int64]*SimulationContext),
		simulationDone:       make(chan int64, 10),
		workerPool:           make(chan struct{}, maxConcurrentWorkers), // Worker pool for limiting goroutines
		shutdownCh:           make(chan struct{}),
	}

	// Reset any simulations that were left in "running" state
	if err := service.ResetStuckSimulations(); err != nil {
		logger.Error("Failed to reset stuck simulations: %v", err)
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

		// Update status in database
		if err := s.simulationRunRepo.UpdateStatus(sim.SimulationRunID, "stopped"); err != nil {
			s.logger.Error("Error updating simulation run status during shutdown: %v", err)
		}
	}
	s.activeSimsMu.Unlock()

	// Close shutdown channel
	close(s.shutdownCh)

	// Wait for monitor goroutine to exit
	s.logger.Info("Shutdown complete")
}

func (s *SimulationService) ResetStuckSimulations() error {
	s.logger.Info("Checking for stuck simulations...")

	// Get all simulations that are still marked as "running"
	runningSimulations, err := s.simulationRunRepo.GetByStatus("running", 10)
	if err != nil {
		return fmt.Errorf("error fetching running simulations: %v", err)
	}

	for _, run := range runningSimulations {
		s.logger.Info("Resetting stuck simulation: %d", run.ID)
		if err := s.simulationRunRepo.UpdateStatus(run.ID, "stopped"); err != nil {
			s.logger.Error("Error updating stuck simulation status: %v", err)
		}
	}

	s.logger.Info("Reset %d stuck simulations", len(runningSimulations))
	return nil
}

// monitorSimulations cleans up completed simulations and periodically updates metrics
func (s *SimulationService) monitorSimulations() {
	// Ticker for checking stalled simulations (every 30 seconds)
	stalledCheckTicker := time.NewTicker(30 * time.Second)
	// Ticker for periodically updating metrics (every 1 minute)
	metricsUpdateTicker := time.NewTicker(1 * time.Minute)

	defer stalledCheckTicker.Stop()
	defer metricsUpdateTicker.Stop()

	for {
		select {
		case <-s.shutdownCh:
			return
		case strategyID := <-s.simulationDone:
			s.logger.Info("Simulation for strategy %d completed", strategyID)
			s.cleanupSimulation(strategyID)
		case <-stalledCheckTicker.C:
			// Periodically check for stalled simulations
			s.checkStalledSimulations()
		case <-metricsUpdateTicker.C:
			// Periodically update metrics for all running simulations
			s.updateAllSimulationMetrics()
		}
	}
}

// updateAllSimulationMetrics updates metrics for all running simulations
func (s *SimulationService) updateAllSimulationMetrics() {
	s.activeSimsMu.RLock()
	defer s.activeSimsMu.RUnlock()

	for _, sim := range s.activeSims {
		sim.mu.RLock()
		isRunning := sim.IsRunning
		sim.mu.RUnlock()

		if isRunning {
			// Update metrics for this running simulation
			s.sendSimulationStatusUpdate(sim)
		}
	}
}

// cleanupSimulation removes a simulation from the active simulations map
func (s *SimulationService) cleanupSimulation(strategyID int64) {
	s.activeSimsMu.Lock()
	defer s.activeSimsMu.Unlock()

	sim, exists := s.activeSims[strategyID]
	if exists {
		// Wait for all goroutines to finish before removing
		if sim.cancel != nil {
			sim.cancel() // Cancel all goroutines
		}
		sim.wg.Wait() // Wait for all goroutines to finish
		delete(s.activeSims, strategyID)
		s.logger.Info("Simulation for strategy %d cleaned up", strategyID)
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
	var config models.StrategyConfig
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
func validateStrategyConfig(config *models.StrategyConfig) error {
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
	var simulationRunID int64
	if exists {
		sim.mu.RLock()
		isRunning = sim.IsRunning
		simulationRunID = sim.SimulationRunID
		sim.mu.RUnlock()
	}
	s.activeSimsMu.RUnlock()

	if !exists || !isRunning {
		s.cleanupSimulation(strategyID)

		// Even if not found in memory, try to update any database records that might be stuck
		runningSimulations, err := s.simulationRunRepo.GetByStatus("running", 10)
		if err == nil {
			for _, run := range runningSimulations {
				// Check if this simulation belongs to the current strategy
				params := run.SimulationParameters
				if strategyIDParam, ok := params["strategyID"]; ok {
					if int64(strategyIDParam.(float64)) == strategyID {
						// Found a database record for this strategy, update it
						if err := s.simulationRunRepo.UpdateStatus(run.ID, "completed"); err != nil {
							s.logger.Error("Error updating stuck simulation status for run %d: %v", run.ID, err)
						} else {
							s.logger.Info("Updated status of simulation run %d to completed", run.ID)
						}
					}
				}
			}
		}

		return fmt.Errorf("no active simulation found for strategy %d", strategyID)
	}

	// Mark for stopping
	sim.mu.Lock()
	sim.StopRequested = true
	sim.IsRunning = false
	sim.mu.Unlock()

	// Update simulation status in database immediately
	if err := s.simulationRunRepo.UpdateStatus(simulationRunID, "completed"); err != nil {
		s.logger.Error("Error updating simulation run status: %v", err)
	} else {
		s.logger.Info("Updated status of simulation run %d to completed", simulationRunID)
	}

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

	// Define simulation parameters
	// Use shorter interval for more frequent evaluations
	iterationInterval := 3 * time.Second // Use 3 seconds interval for more frequent token evaluation

	// Create a ticker for iteration intervals
	ticker := time.NewTicker(iterationInterval)

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
		ctx.mu.Unlock()

		// Update simulation run status in database
		if err := s.simulationRunRepo.UpdateStatus(ctx.SimulationRunID, "completed"); err != nil {
			s.logger.Error("Error updating simulation run status: %v", err)
		} else {
			s.logger.Info("Updated status of simulation run %d to completed", ctx.SimulationRunID)
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
	// Fetch tokens to evaluate - using a shorter age window to focus on fresh tokens
	// Changed back from 24 hours to 5 minutes to focus on newest tokens
	maxAgeSec := int64(300)
	tokens, err := s.tokenRepo.GetFilteredTokens(ctx.Config.MarketCapThreshold, maxAgeSec, 100)
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

		// Check if we already have any trade for this token
		if s.hasExistingTrade(ctx, token.ID) {
			continue // Skip tokens we've already traded
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

	// Update and save metrics after each iteration to track real-time performance
	s.sendSimulationStatusUpdate(ctx)

	return nil
}

// hasExistingTrade checks if we already have any trade (active or completed) for this token
func (s *SimulationService) hasExistingTrade(ctx *SimulationContext, tokenID int64) bool {
	// First check in-memory trades
	ctx.tokensMu.RLock()
	for _, existingTrade := range ctx.Trades {
		if existingTrade.TokenID == tokenID {
			ctx.tokensMu.RUnlock()
			return true
		}
	}
	ctx.tokensMu.RUnlock()

	// Then check database for previous trades with this token and strategy
	tradeExists, err := s.simulatedTradeRepo.ExistsByStrategyIDAndTokenID(ctx.StrategyID, tokenID)
	if err != nil {
		s.logger.Error("Error checking for existing trade in database: %v", err)
		return false // Default to false if there's an error
	}

	return tradeExists
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

	// We're adding back the token age restriction to focus on fresh tokens
	// This ensures we're looking at the newest tokens on the market
	now := time.Now().Unix()
	if now-token.CreatedTimestamp/1000 > 180 { // 3 min = 180 seconds
		return nil // Skip older tokens
	}

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

	// We've removed the random skip to ensure we evaluate all qualifying tokens
	// This ensures maximum trading opportunities are captured
	// Original code had a 40% chance to skip tokens

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

	s.sendSimulationStatusUpdate(ctx)

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
	ticker := time.NewTicker(3 * time.Second) // Changed back to 3 seconds for more frequent price checks
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

	// Update simulation status and save metrics after every trade close
	s.sendSimulationStatusUpdate(ctx)
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

	// Log time information for debugging
	s.logger.Info("Current time: %d, Lookback window: %d (%d seconds ago)",
		now, lookbackTime, ctx.Config.EntryTimeWindowSec)

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

	// Create a timestamp once to ensure consistency
	now := time.Now()
	unixTimestamp := now.Unix()

	var eventObject dto.WebSocketMessage

	// Create the appropriate event type based on event type
	switch eventType {
	case "simulation_started":
		eventObject = &dto.SimulationStartedEvent{
			BaseEventDTO: dto.BaseEventDTO{
				Type:       eventType,
				StrategyID: ctx.StrategyID,
				Timestamp:  unixTimestamp,
			},
		}

	case "simulation_completed":
		eventObject = &dto.SimulationCompletedEvent{
			BaseEventDTO: dto.BaseEventDTO{
				Type:       eventType,
				StrategyID: ctx.StrategyID,
				Timestamp:  unixTimestamp,
			},
			TotalIterations:  data["total_iterations"].(int),
			ExecutionTimeSec: data["execution_time_sec"].(float64),
		}

	case "simulation_balance_depleted":
		eventObject = &dto.SimulationBalanceDepletedEvent{
			BaseEventDTO: dto.BaseEventDTO{
				Type:       eventType,
				StrategyID: ctx.StrategyID,
				Timestamp:  unixTimestamp,
			},
			RemainingBalance: data["remaining_balance"].(float64),
			PositionSize:     data["position_size"].(float64),
		}

	case "simulation_status":
		eventObject = &dto.SimulationStatusEvent{
			BaseEventDTO: dto.BaseEventDTO{
				Type:       eventType,
				StrategyID: ctx.StrategyID,
				Timestamp:  unixTimestamp,
			},
			TotalTrades:      data["total_trades"].(int),
			ActiveTrades:     data["active_trades"].(int),
			ProfitableTrades: data["profitable_trades"].(int),
			LosingTrades:     data["losing_trades"].(int),
			WinRate:          data["win_rate"].(float64),
			ROI:              data["roi"].(float64),
			CurrentBalance:   data["current_balance"].(float64),
			InitialBalance:   data["initial_balance"].(float64),
		}

	case "trade_executed":
		eventObject = &dto.TradeExecutedEvent{
			BaseEventDTO: dto.BaseEventDTO{
				Type:       eventType,
				StrategyID: ctx.StrategyID,
				Timestamp:  unixTimestamp,
			},
			TokenID:        data["token_id"].(int64),
			TokenSymbol:    data["token_symbol"].(string),
			TokenName:      data["token_name"].(string),
			TokenMint:      data["token_mint"].(string),
			ImageUrl:       data["image_url"].(string),
			TwitterUrl:     data["twitter_url"].(string),
			WebsiteUrl:     data["website_url"].(string),
			Action:         data["action"].(string),
			Price:          data["price"].(float64),
			Amount:         data["amount"].(float64),
			EntryMarketCap: data["entry_market_cap"].(float64),
			UsdMarketCap:   data["usd_market_cap"].(float64),
			CurrentBalance: data["current_balance"].(float64),
		}

		// Conditional signal data
		if signalData, ok := data["signal_data"].(map[string]interface{}); ok {
			tradeEvent := eventObject.(*dto.TradeExecutedEvent)
			tradeEvent.SignalData = signalData
		}

	case "trade_closed":
		eventObject = &dto.TradeClosedEvent{
			BaseEventDTO: dto.BaseEventDTO{
				Type:       eventType,
				StrategyID: ctx.StrategyID,
				Timestamp:  unixTimestamp,
			},
			TokenID:        data["token_id"].(int64),
			TokenSymbol:    data["token_symbol"].(string),
			TokenName:      data["token_name"].(string),
			TokenMint:      data["token_mint"].(string),
			ImageUrl:       data["image_url"].(string),
			TwitterUrl:     data["twitter_url"].(string),
			WebsiteUrl:     data["website_url"].(string),
			Action:         data["action"].(string),
			EntryPrice:     data["entry_price"].(float64),
			ExitPrice:      data["exit_price"].(float64),
			ProfitLoss:     data["profit_loss"].(float64),
			ProfitLossPct:  data["profit_loss_pct"].(float64),
			ExitReason:     data["exit_reason"].(string),
			EntryMarketCap: data["entry_market_cap"].(float64),
			ExitMarketCap:  data["exit_market_cap"].(float64),
			UsdMarketCap:   data["usd_market_cap"].(float64),
		}

	default:
		// For unknown event types, create a generic map
		genericEvent := make(map[string]interface{})
		genericEvent["type"] = eventType
		genericEvent["strategy_id"] = ctx.StrategyID
		genericEvent["timestamp"] = unixTimestamp

		// Add all additional data
		for k, v := range data {
			genericEvent[k] = v
		}

		// Create event model
		simulationEvent := &models.SimulationEvent{
			StrategyID:      ctx.StrategyID,
			SimulationRunID: ctx.SimulationRunID,
			EventType:       eventType,
			EventData:       models.JSONB(genericEvent),
			Timestamp:       now,
			CreatedAt:       now,
		}

		// Save event to database
		_, err := s.simulationEventRepo.Save(simulationEvent)
		if err != nil {
			s.logger.Error("Error saving simulation event to database: %v", err)
			return
		}

		// Broadcast the event
		s.wsHub.BroadcastJSON(genericEvent)
		return
	}

	// Create event model for typed events
	eventData, err := json.Marshal(eventObject)
	if err != nil {
		s.logger.Error("Error marshaling event data: %v", err)
		return
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(eventData, &jsonData); err != nil {
		s.logger.Error("Error unmarshaling event data: %v", err)
		return
	}

	simulationEvent := &models.SimulationEvent{
		StrategyID:      ctx.StrategyID,
		SimulationRunID: ctx.SimulationRunID,
		EventType:       eventType,
		EventData:       models.JSONB(jsonData),
		Timestamp:       now,
		CreatedAt:       now,
	}

	// Save event to database
	_, err = s.simulationEventRepo.Save(simulationEvent)
	if err != nil {
		s.logger.Error("Error saving simulation event to database: %v", err)
		return
	}

	// Broadcast the event via WebSocket
	s.wsHub.BroadcastJSON(eventObject)
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
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get recent trades with context - force latest data with no cache
	latestTrades, err := s.tradeRepo.GetTradesByTokenIDWithContext(ctx, tokenID, 10)
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

// calculatePerformanceRating calculates a performance rating based on ROI and win rate
func (s *SimulationService) calculatePerformanceRating(roi float64, winRate float64) string {
	// Performance rating thresholds
	if roi >= 30 && winRate >= 70 {
		return "excellent"
	} else if roi >= 15 && winRate >= 60 {
		return "good"
	} else if roi >= 5 && winRate >= 50 {
		return "average"
	} else if roi > 0 {
		return "poor"
	} else {
		return "very_poor"
	}
}

// saveSimulationResult saves a simulation result record based on the metrics
func (s *SimulationService) saveSimulationResult(strategyMetric *models.StrategyMetric, simulationRunID int64, strategyID int64) (int64, error) {
	// Calculate performance rating based on metrics
	performanceRating := s.calculatePerformanceRating(strategyMetric.ROI, strategyMetric.WinRate)

	roi := strategyMetric.ROI

	winRate := strategyMetric.WinRate

	maxDrawdown := strategyMetric.MaxDrawdown

	// Create simulation result model
	simulationResult := &models.SimulationResult{
		SimulationRunID:   simulationRunID,
		StrategyID:        strategyID,
		ROI:               roi,
		TradeCount:        strategyMetric.TotalTrades,
		WinRate:           winRate,
		MaxDrawdown:       maxDrawdown,
		PerformanceRating: performanceRating,
		Analysis:          "", // Can be filled later with AI analysis
		Rank:              0,  // Will be updated later based on comparison with other strategies
		CreatedAt:         time.Now(),
	}

	// Save the simulation result
	resultID, err := s.simulationResultRepo.Save(simulationResult)
	if err != nil {
		return 0, fmt.Errorf("error saving simulation result: %v", err)
	}

	s.logger.Info("Saved simulation result with ID %d for strategy %d with rating %s",
		resultID, strategyID, performanceRating)

	// Update the ranks for this simulation run
	if err := s.updateRanksForSimulationRun(simulationRunID); err != nil {
		s.logger.Error("Error updating ranks for simulation run %d: %v", simulationRunID, err)
		// Continue without error since this is supplementary info
	}

	return resultID, nil
}

// updateRanksForSimulationRun updates the ranks of all results for a given simulation run
// based on their ROI performance
func (s *SimulationService) updateRanksForSimulationRun(simulationRunID int64) error {
	// Get all results for this simulation run
	results, err := s.simulationResultRepo.GetBySimulationRun(simulationRunID)
	if err != nil {
		return fmt.Errorf("error getting results for simulation run %d: %v", simulationRunID, err)
	}

	if len(results) == 0 {
		return nil // Nothing to rank
	}

	// Sort results by ROI in descending order (highest ROI first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ROI > results[j].ROI
	})

	// Update ranks in database using a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction for rank updates: %v", err)
	}

	// Define SQL query for updating ranks
	query := `UPDATE simulation_results SET rank = $1 WHERE id = $2`

	// Update ranks for all results
	for i, result := range results {
		rank := i + 1 // Ranks start at 1

		// Update the rank for this result
		_, err := tx.Exec(query, rank, result.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error updating rank for result %d: %v", result.ID, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing rank updates: %v", err)
	}

	s.logger.Info("Updated ranks for %d results in simulation run %d", len(results), simulationRunID)
	return nil
}

// saveSimulationMetrics saves the performance metrics for a completed simulation
func (s *SimulationService) saveSimulationMetrics(ctx *SimulationContext) error {
	// Calculate metrics
	metrics := s.calculateInMemorySummary(ctx)

	// Create strategy metric model for final metrics (at simulation completion)
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
		ROI:              metrics["roi"].(float64),
		CurrentBalance:   ctx.CurrentBalance,
		InitialBalance:   ctx.InitialBalance,
		CreatedAt:        time.Now(),
	}

	// For final metrics, we create a new record instead of updating
	// This is to maintain a history of final metrics for each simulation run
	metricID, err := s.strategyMetricRepo.Save(strategyMetric)
	if err != nil {
		return fmt.Errorf("error saving final strategy metrics: %v", err)
	}

	s.logger.Info("Saved final strategy metrics with ID %d for strategy %d", metricID, ctx.StrategyID)

	// Save corresponding simulation result
	_, err = s.saveSimulationResult(strategyMetric, ctx.SimulationRunID, ctx.StrategyID)
	if err != nil {
		s.logger.Error("Error saving simulation result: %v", err)
		// Continue execution - the metrics were saved successfully
	}

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

// GetRunningSimulations returns all currently running simulations
func (s *SimulationService) GetRunningSimulations() []*dto.SimulationStatusDTO {
	s.activeSimsMu.RLock()
	defer s.activeSimsMu.RUnlock()
	runningSimulations := make([]*dto.SimulationStatusDTO, 0, len(s.activeSims))

	for strategyID, sim := range s.activeSims {
		sim.mu.RLock()
		isRunning := sim.IsRunning
		sim.mu.RUnlock()

		s.logger.Info("Found simulation for strategy ID=%d, isRunning=%v", strategyID, isRunning)

		// Include all active simulations regardless of IsRunning flag
		{
			// Get active trades count
			sim.tokensMu.RLock()
			activeTrades := 0
			completedTrades := 0
			for _, trade := range sim.Trades {
				if trade.Status == "active" {
					activeTrades++
				} else if trade.Status == "completed" || trade.Status == "closed" {
					completedTrades++
				}
			}
			sim.tokensMu.RUnlock()

			// Calculate metrics
			summary := s.calculateInMemorySummary(sim)

			// Create DTO with configuration details
			simDTO := &dto.SimulationStatusDTO{
				StrategyID:       strategyID,
				StrategyName:     sim.Strategy.Name,
				IsRunning:        isRunning,
				StartTime:        sim.StartTime.Unix(),
				ExecutionTimeSec: time.Since(sim.StartTime).Seconds(),
				TotalTrades:      summary["total_trades"].(int),
				ActiveTrades:     activeTrades,
				ProfitableTrades: summary["profitable_trades"].(int),
				LosingTrades:     summary["losing_trades"].(int),
				WinRate:          summary["win_rate"].(float64),
				TotalProfit:      summary["total_profit"].(float64),
				TotalLoss:        summary["total_loss"].(float64),
				AvgProfit:        summary["avg_profit"].(float64),
				AvgLoss:          summary["avg_loss"].(float64),
				MaxDrawdown:      summary["max_drawdown"].(float64),
				NetPnL:           summary["total_profit"].(float64) + summary["total_loss"].(float64),
				InitialBalance:   sim.InitialBalance,
				CurrentBalance:   sim.CurrentBalance,
				ROI:              summary["roi"].(float64),
				SimConfig: &dto.SimConfigDTO{
					InitialBalance:       sim.Config.InitialBalance,
					FixedPositionSizeSol: sim.Config.FixedPositionSizeSol,
					MarketCapThreshold:   sim.Config.MarketCapThreshold,
					TakeProfitPct:        sim.Config.TakeProfitPct,
					StopLossPct:          sim.Config.StopLossPct,
					MaxHoldTimeSec:       sim.Config.MaxHoldTimeSec,
					EntryTimeWindowSec:   sim.Config.EntryTimeWindowSec,
					MinBuysForEntry:      sim.Config.MinBuysForEntry,
				},
			}

			runningSimulations = append(runningSimulations, simDTO)
		}
	}

	// Now add all simulations that are in 'running' state in the database as well
	runningSimulationsDB, err := s.simulationRunRepo.GetByStatus("running", 10)
	if err != nil {
		s.logger.Error("Error fetching running simulations from database: %v", err)
	} else {
		s.logger.Info("Found %d running simulations in database", len(runningSimulationsDB))

		// Check if we need to add any database simulations that weren't in our active map
		for _, runDB := range runningSimulationsDB {
			// Extract strategy ID from parameters
			params := runDB.SimulationParameters
			if strategyIDParam, ok := params["strategyID"]; ok {
				strategyID := int64(strategyIDParam.(float64))

				// Check if this simulation is already in our list
				found := false
				for _, simDTO := range runningSimulations {
					if simDTO.StrategyID == strategyID {
						found = true
						break
					}
				}

				// If not in our list, try to add it from the database
				if !found {
					s.logger.Info("Adding simulation for strategy ID=%d from database", strategyID)
					strategy, err := s.strategyRepo.GetByID(strategyID)
					if err != nil {
						s.logger.Error("Error getting strategy %d: %v", strategyID, err)
						continue
					}

					// Create a minimal DTO
					simDTO := &dto.SimulationStatusDTO{
						StrategyID:       strategyID,
						StrategyName:     strategy.Name,
						IsRunning:        true,
						StartTime:        runDB.StartTime.Unix(),
						ExecutionTimeSec: time.Since(runDB.StartTime).Seconds(),
						// Set minimal values for other fields
						TotalTrades:      0,
						ActiveTrades:     0,
						ProfitableTrades: 0,
						LosingTrades:     0,
						WinRate:          0,
						TotalProfit:      0,
						TotalLoss:        0,
						AvgProfit:        0,
						AvgLoss:          0,
						MaxDrawdown:      0,
						NetPnL:           0,
						InitialBalance:   1000, // Default value
						CurrentBalance:   1000, // Default value
						ROI:              0,
						SimConfig:        &dto.SimConfigDTO{},
					}
					runningSimulations = append(runningSimulations, simDTO)
				}
			}
		}
	}

	s.logger.Info("Found %d running simulations", len(runningSimulations))
	return runningSimulations
}

// sendSimulationStatusUpdate sends current simulation status via WebSocket
// and also saves the current metrics to the database
func (s *SimulationService) sendSimulationStatusUpdate(ctx *SimulationContext) {
	// Calculate active trades count
	ctx.tokensMu.RLock()
	activeTrades := 0
	profitableTrades := 0
	losingTrades := 0
	totalTrades := 0
	var totalProfit, totalLoss float64

	for _, trade := range ctx.Trades {
		if trade.Status == "active" {
			activeTrades++
		} else if trade.Status == "completed" || trade.Status == "closed" {
			totalTrades++
			if trade.ProfitLoss != nil {
				if *trade.ProfitLoss > 0 {
					profitableTrades++
					totalProfit += *trade.ProfitLoss
				} else {
					losingTrades++
					totalLoss += *trade.ProfitLoss
				}
			}
		}
	}
	ctx.tokensMu.RUnlock()

	// Get current balance and initial balance
	ctx.mu.RLock()
	currentBalance := ctx.CurrentBalance
	initialBalance := ctx.InitialBalance
	ctx.mu.RUnlock()

	// Calculate ROI
	roi := 0.0
	if initialBalance > 0 {
		roi = ((currentBalance / initialBalance) - 1.0) * 100.0
	}

	// Calculate win rate
	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(profitableTrades) / float64(totalTrades) * 100.0
	}

	// Calculate average profit per trade
	avgProfit := 0.0
	if profitableTrades > 0 {
		avgProfit = totalProfit / float64(profitableTrades)
	}

	//Calculate average loss per trade
	avgLoss := 0.0
	if losingTrades > 0 {
		avgLoss = totalLoss / float64(losingTrades)
	}

	// Send status event
	s.sendSimulationEvent(ctx, "simulation_status", map[string]interface{}{
		"total_trades":      totalTrades,
		"active_trades":     activeTrades,
		"profitable_trades": profitableTrades,
		"losing_trades":     losingTrades,
		"win_rate":          winRate,
		"roi":               roi,
		"current_balance":   currentBalance,
		"initial_balance":   initialBalance,
	})

	// Save or update current metrics to database for running simulations
	if s.strategyMetricRepo != nil && totalTrades > 0 {
		// Calculate max drawdown (simplified version for running simulations)
		maxDrawdown := 0.0
		if initialBalance > currentBalance {
			maxDrawdown = (initialBalance - currentBalance) / initialBalance * 100.0
		}

		// Create strategy metric object
		strategyMetric := &models.StrategyMetric{
			StrategyID:       ctx.StrategyID,
			SimulationRunID:  &ctx.SimulationRunID,
			WinRate:          winRate,
			AvgProfit:        avgProfit,
			AvgLoss:          avgLoss,
			MaxDrawdown:      maxDrawdown,
			TotalTrades:      totalTrades,
			SuccessfulTrades: profitableTrades,
			RiskScore:        ctx.Strategy.RiskScore,
			ROI:              roi,
			CurrentBalance:   currentBalance,
			InitialBalance:   initialBalance,
			CreatedAt:        time.Now(),
		}

		// Use UpdateLatestByStrategy to update existing metric or create new one
		err := s.strategyMetricRepo.UpdateLatestByStrategy(strategyMetric)
		if err != nil {
			s.logger.Error("Error updating real-time strategy metrics: %v", err)
		} else {
			s.logger.Info("Updated real-time strategy metrics for strategy %d", ctx.StrategyID)
		}
	}
	// 		// Also save a simulation result for periodic analysis
	// 		// First check if we already have a simulation result for this strategy and run
	// 		existingResults, err := s.simulationResultRepo.GetBySimulationRun(ctx.SimulationRunID)
	// 		if err != nil {
	// 			s.logger.Error("Error checking for existing simulation results: %v", err)
	// 		} else {
	// 			strategyHasResult := false
	// 			for _, result := range existingResults {
	// 				if result.StrategyID == ctx.StrategyID {
	// 					strategyHasResult = true
	// 					break
	// 				}
	// 			}

	// 			// If no existing result, or we have enough trades for meaningful results, create/update one
	// 			if !strategyHasResult && totalTrades > 5 {
	// 				// Create a simulation result using current metrics
	// 				_, err := s.saveSimulationResult(strategyMetric, ctx.SimulationRunID, ctx.StrategyID)
	// 				if err != nil {
	// 					s.logger.Error("Error saving periodic simulation result: %v", err)
	// 				} else {
	// 					s.logger.Info("Saved periodic simulation result for strategy %d during active simulation", ctx.StrategyID)
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	s.logger.Info("Sent simulation status update: Balance=%.6f, ROI=%.2f%%, Win Rate=%.2f%%",
		currentBalance, roi, winRate)
}

// GetWSHub returns the WebSocket hub
func (s *SimulationService) GetWSHub() *websocket.WSHub {
	return s.wsHub
}
