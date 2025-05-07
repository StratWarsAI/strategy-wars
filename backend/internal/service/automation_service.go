// internal/service/automation_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/config"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
)

// AutomationConfig holds configuration for the automation service
type AutomationConfig struct {
	StrategyGenerationInterval time.Duration
	PerformanceAnalysisInterval time.Duration
	StrategiesPerInterval int
	MaxConcurrentSimulations int
}

// AutomationService handles the automation of strategy generation, simulation, and analysis
type AutomationService struct {
	config               AutomationConfig
	ctx                  context.Context
	cancelFunc           context.CancelFunc
	strategyRepo         repository.StrategyRepositoryInterface
	simulationRunRepo    repository.SimulationRunRepositoryInterface
	aiService            *AIService
	simulationService    *SimulationService
	performanceAnalyzer  *AIPerformanceAnalyzer
	logger               *logger.Logger
	runningSimulations   map[int64]bool
	simulationQueue      chan int64
	runningSimulationsMu sync.RWMutex
	isRunning            bool
	lastStrategyGenTime  time.Time
}

// NewAutomationService creates a new automation service
func NewAutomationService(
	strategyRepo repository.StrategyRepositoryInterface,
	simulationRunRepo repository.SimulationRunRepositoryInterface,
	aiService *AIService,
	simulationService *SimulationService,
	performanceAnalyzer *AIPerformanceAnalyzer,
	logger *logger.Logger,
	cfg *config.Config, // Add config parameter
) *AutomationService {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Convert minutes to time.Duration
	strategyGenInterval := time.Duration(cfg.Automation.StrategyGenerationInterval) * time.Minute
	perfAnalysisInterval := time.Duration(cfg.Automation.PerformanceAnalysisInterval) * time.Minute
	
	logger.Info("Initializing AutomationService with config: StrategyGenInterval=%v, AnalysisInterval=%v, StrategiesPerInterval=%d, MaxConcurrentSims=%d",
		strategyGenInterval, 
		perfAnalysisInterval,
		cfg.Automation.StrategiesPerInterval,
		cfg.Automation.MaxConcurrentSimulations)
	
	return &AutomationService{
		config: AutomationConfig{
			StrategyGenerationInterval: strategyGenInterval,
			PerformanceAnalysisInterval: perfAnalysisInterval,
			StrategiesPerInterval: cfg.Automation.StrategiesPerInterval,
			MaxConcurrentSimulations: cfg.Automation.MaxConcurrentSimulations,
		},
		ctx:                 ctx,
		cancelFunc:          cancel,
		strategyRepo:        strategyRepo,
		simulationRunRepo:   simulationRunRepo,
		aiService:           aiService,
		simulationService:   simulationService,
		performanceAnalyzer: performanceAnalyzer,
		logger:              logger,
		runningSimulations:  make(map[int64]bool),
		simulationQueue:     make(chan int64, 100), // Buffer for queued strategies
		isRunning:           false,
		lastStrategyGenTime: time.Now(),
	}
}

// Start starts the automation service
func (s *AutomationService) Start() error {
	if s.isRunning {
		return fmt.Errorf("automation service is already running")
	}

	s.logger.Info("Starting automation service with the following configuration:")
	s.logger.Info("- Strategy Generation Interval: %v", s.config.StrategyGenerationInterval)
	s.logger.Info("- Performance Analysis Interval: %v", s.config.PerformanceAnalysisInterval)
	s.logger.Info("- Strategies Per Interval: %d", s.config.StrategiesPerInterval)
	s.logger.Info("- Max Concurrent Simulations: %d", s.config.MaxConcurrentSimulations)
	
	// Check if there are any running simulations at startup
	activeRuns, err := s.simulationRunRepo.GetByStatus("running", 10)
	if err != nil {
		s.logger.Error("Error checking for running simulations at startup: %v", err)
	} else {
		if len(activeRuns) > 0 {
			s.logger.Info("Found %d running simulations at startup. Will wait for these to complete before starting new ones.", len(activeRuns))
			for _, run := range activeRuns {
				s.logger.Info("Active simulation found: ID=%d, Started=%v", run.ID, run.StartTime)
			}
		} else {
			s.logger.Info("No running simulations found at startup")
			
			// First check if there are any strategies in the database
			strategies, err := s.getAllAIStrategies()
			if err != nil {
				s.logger.Error("Error getting AI strategies at startup: %v", err)
			} else {
				if len(strategies) == 0 {
					// No strategies exist, generate them immediately without waiting for the timer
					s.logger.Info("No strategies found in database. Generating initial strategies immediately...")
					go s.generateInitialStrategies() // Use a special method for initial generation
					s.lastStrategyGenTime = time.Now() // Reset the timer
				} else {
					// Automatically run simulations for existing strategies on startup
					s.logger.Info("Starting automatic simulation of existing strategies...")
					go func() {
						// Only take MaxConcurrentSimulations strategies to avoid overloading
						count := 0
						for _, strategy := range strategies {
							if count >= s.config.MaxConcurrentSimulations {
								break
							}
							
							s.logger.Info("Queuing existing strategy ID=%d for automatic simulation at startup", strategy.ID)
							s.queueStrategyForSimulation(strategy.ID)
							count++
						}
						
						s.logger.Info("Queued %d existing strategies for simulation at startup", count)
					}()
				}
			}
		}
	}
	
	s.isRunning = true

	// Start the simulation queue processor
	go s.processSimulationQueue()

	// Start the main automation loop
	go s.runAutomationLoop()

	// Start the performance analyzer
	go s.performanceAnalyzer.StartAutomatedAnalysis(s.ctx)

	s.logger.Info("All automation service components started successfully")
	return nil
}

// Stop stops the automation service
func (s *AutomationService) Stop() error {
	if !s.isRunning {
		return fmt.Errorf("automation service is not running")
	}

	s.logger.Info("Stopping automation service")
	s.isRunning = false
	s.cancelFunc()
	
	return nil
}

// runAutomationLoop is the main loop that coordinates all automated tasks
func (s *AutomationService) runAutomationLoop() {
	s.logger.Info("Starting automation loop")

	checkInterval := 1 * time.Minute
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// Periodically log the number of active simulations for monitoring
	logTicker := time.NewTicker(5 * time.Minute)
	defer logTicker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Automation loop stopped")
			return
		case <-logTicker.C:
			// Log status of active simulations periodically
			activeRuns, err := s.simulationRunRepo.GetByStatus("running", 10)
			if err != nil {
				s.logger.Error("Error checking for active simulations in periodic check: %v", err)
			} else {
				s.logger.Info("Periodic status check: %d active simulations running", len(activeRuns))
				if len(activeRuns) > 0 {
					for _, run := range activeRuns {
						runningDuration := time.Since(run.StartTime)
						s.logger.Info("  - Simulation ID=%d, Running for %v", run.ID, runningDuration)
					}
				}
			}
		case <-ticker.C:
			// Check if it's time to generate new strategies
			if time.Since(s.lastStrategyGenTime) >= s.config.StrategyGenerationInterval {
				s.logger.Info("Time for scheduled strategy generation (last gen: %v ago)", time.Since(s.lastStrategyGenTime))
				go s.generateStrategies()
				s.lastStrategyGenTime = time.Now()
			} else {
				remaining := s.config.StrategyGenerationInterval - time.Since(s.lastStrategyGenTime)
				s.logger.Info("Next strategy generation in %v", remaining)
			}

			// Check for pending strategies that need simulation
			go s.checkPendingStrategies()
		}
	}
}

// generateStrategies generates new AI strategies
func (s *AutomationService) generateStrategies() {
	// First, check if there are any active simulations running
	activeRuns, err := s.simulationRunRepo.GetByStatus("running", 5)
	if err != nil {
		s.logger.Error("Error checking for active simulations before generation: %v", err)
		// Continue with caution
	} else if len(activeRuns) > 0 {
		s.logger.Info("Found %d active simulations running. Will skip strategy generation this cycle.", len(activeRuns))
		return
	}

	s.logger.Info("Generating %d new AI strategies", s.config.StrategiesPerInterval)

	for i := 0; i < s.config.StrategiesPerInterval; i++ {
		// Get top performing strategies to learn from
		topStrategies, err := s.aiService.GetTopPerformingStrategies()
		if err != nil {
			s.logger.Error("Error getting top strategies: %v", err)
			continue
		}

		// Create metadata with top strategies
		metadata := map[string]interface{}{
			"top_strategies": topStrategies,
		}

		// Generate strategy prompts based on the iteration
		var prompt string
		if i == 0 {
			prompt = "Generate a profitable trading strategy for cryptocurrency tokens that focuses on early entry and quick profit-taking"
		} else {
			prompt = "Generate a diversified trading strategy for cryptocurrency tokens that focuses on longer holds and larger market caps"
		}

		// Generate new strategy
		strategy, err := s.aiService.GenerateStrategy(prompt, metadata)
		if err != nil {
			s.logger.Error("Error generating strategy %d: %v", i+1, err)
			continue
		}

		// Make sure strategy name is unique by adding timestamp
		strategy.Name = fmt.Sprintf("%s (%s)", strategy.Name, time.Now().Format("20060102-1504"))

		// Save the strategy
		id, err := s.strategyRepo.Save(strategy)
		if err != nil {
			s.logger.Error("Error saving generated strategy: %v", err)
			continue
		}

		// Create initial metrics record for the strategy
		s.createInitialStrategyMetric(id)

		s.logger.Info("Successfully generated and saved new strategy %d with ID: %d", i+1, id)
		
		// Add strategy to simulation queue
		s.queueStrategyForSimulation(id)
	}
}

// checkPendingStrategies checks for strategies that need to be simulated
func (s *AutomationService) checkPendingStrategies() {
	// First, check if there are any active simulations running
	activeRuns, err := s.simulationRunRepo.GetByStatus("running", 5)
	if err != nil {
		s.logger.Error("Error checking for active simulations in checkPendingStrategies: %v", err)
		// Continue with caution
	} else if len(activeRuns) > 0 {
		s.logger.Info("Found %d active simulations running. Will skip checking pending strategies this cycle.", len(activeRuns))
		return
	}

	// Get AI strategies that don't have metrics yet (never been simulated)
	strategies, err := s.getUnsimulatedStrategies()
	if err != nil {
		s.logger.Error("Error getting unsimulated strategies: %v", err)
		return
	}

	if len(strategies) == 0 {
		return // No unsimulated strategies
	}

	s.logger.Info("Found %d strategies that need simulation", len(strategies))

	// Queue them for simulation
	for _, strategy := range strategies {
		s.queueStrategyForSimulation(strategy.ID)
	}
}

// getUnsimulatedStrategies gets AI strategies that have not been simulated yet
func (s *AutomationService) getUnsimulatedStrategies() ([]*models.Strategy, error) {
	// Get all AI strategies
	strategies, err := s.getAllAIStrategies()
	if err != nil {
		return nil, fmt.Errorf("error getting AI strategies: %v", err)
	}

	var unsimulated []*models.Strategy
	for _, strategy := range strategies {
		// Check if this strategy has metrics
		metric, err := s.performanceAnalyzer.strategyMetricRepo.GetLatestByStrategy(strategy.ID)
		if err != nil {
			s.logger.Error("Error checking metrics for strategy %d: %v", strategy.ID, err)
			continue
		}

		// If no metrics, the strategy needs simulation
		if metric == nil {
			// Check if it's already in the simulation queue or running
			s.runningSimulationsMu.RLock()
			isRunning := s.runningSimulations[strategy.ID]
			s.runningSimulationsMu.RUnlock()

			if !isRunning {
				unsimulated = append(unsimulated, strategy)
			}
		}
	}

	return unsimulated, nil
}

// getAllAIStrategies gets all AI-enhanced strategies
func (s *AutomationService) getAllAIStrategies() ([]*models.Strategy, error) {
	// In a real implementation, you would modify the repository to support this query directly
	// For now, we'll get all public strategies and filter
	strategies, err := s.strategyRepo.ListPublic(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("error listing strategies: %v", err)
	}

	var aiStrategies []*models.Strategy
	for _, strategy := range strategies {
		if strategy.AIEnhanced {
			aiStrategies = append(aiStrategies, strategy)
		}
	}

	return aiStrategies, nil
}

// generateInitialStrategies generates the initial set of strategies when the database is empty
// This is a special method that ensures at least 2 strategies are created at startup
func (s *AutomationService) generateInitialStrategies() {
	s.logger.Info("Generating initial set of strategies (ensuring %d are created)", s.config.StrategiesPerInterval)
	
	// Define prompts for both strategies upfront
	prompts := []string{
		"Generate a profitable trading strategy for cryptocurrency tokens that focuses on early entry and quick profit-taking",
		"Generate a diversified trading strategy for cryptocurrency tokens that focuses on longer holds and larger market caps",
	}
	
	// Create a wait group to ensure both strategies are generated
	var wg sync.WaitGroup
	wg.Add(s.config.StrategiesPerInterval)
	
	// Channel to collect generated strategy IDs
	successfulIDs := make(chan int64, s.config.StrategiesPerInterval)
	
	// Generate strategies concurrently
	for i := 0; i < s.config.StrategiesPerInterval; i++ {
		go func(index int) {
			defer wg.Done()
			
			// Get prompt for this strategy
			prompt := prompts[index]
			
			// Make multiple attempts to generate a strategy
			for attempt := 1; attempt <= 3; attempt++ {
				// Get top performing strategies to learn from
				topStrategies, err := s.aiService.GetTopPerformingStrategies()
				if err != nil {
					s.logger.Error("Attempt %d: Error getting top strategies: %v", attempt, err)
					continue
				}
				
				// Create metadata with top strategies
				metadata := map[string]interface{}{
					"top_strategies": topStrategies,
				}
				
				// Generate new strategy
				strategy, err := s.aiService.GenerateStrategy(prompt, metadata)
				if err != nil {
					s.logger.Error("Attempt %d: Error generating strategy %d: %v", attempt, index+1, err)
					if attempt < 3 {
						s.logger.Info("Retrying strategy generation (attempt %d of 3)...", attempt+1)
						time.Sleep(2 * time.Second)
					}
					continue
				}
				
				// Make sure strategy name is unique by adding timestamp
				strategy.Name = fmt.Sprintf("%s (%s)", strategy.Name, time.Now().Format("20060102-1504"))
				
				// Save the strategy
				id, err := s.strategyRepo.Save(strategy)
				if err != nil {
					s.logger.Error("Attempt %d: Error saving generated strategy: %v", attempt, err)
					if attempt < 3 {
						s.logger.Info("Retrying strategy save (attempt %d of 3)...", attempt+1)
						time.Sleep(2 * time.Second)
					}
					continue
				}
				
				// Create initial metrics even before simulation
				s.createInitialStrategyMetric(id)
				
				s.logger.Info("Successfully generated and saved initial strategy %d with ID: %d", index+1, id)
				successfulIDs <- id
				return // Success - break out of the retry loop
			}
			
			s.logger.Error("Failed to generate strategy %d after multiple attempts", index+1)
		}(i)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(successfulIDs)
	
	// Count successful strategies
	var successIDs []int64
	for id := range successfulIDs {
		successIDs = append(successIDs, id)
	}
	
	s.logger.Info("Initial strategy generation complete. Generated %d of %d strategies successfully.", 
		len(successIDs), s.config.StrategiesPerInterval)
	
	// Queue successful strategies for simulation
	for _, id := range successIDs {
		s.queueStrategyForSimulation(id)
	}
}

// queueStrategyForSimulation adds a strategy to the simulation queue
func (s *AutomationService) queueStrategyForSimulation(strategyID int64) {
	s.runningSimulationsMu.RLock()
	isRunning := s.runningSimulations[strategyID]
	s.runningSimulationsMu.RUnlock()

	if isRunning {
		s.logger.Info("Strategy %d is already queued or running simulation", strategyID)
		return
	}

	// Mark as running before adding to queue to prevent duplicates
	s.runningSimulationsMu.Lock()
	s.runningSimulations[strategyID] = true
	s.runningSimulationsMu.Unlock()

	// Add to queue
	select {
	case s.simulationQueue <- strategyID:
		s.logger.Info("Added strategy %d to simulation queue", strategyID)
	default:
		s.logger.Warn("Simulation queue is full, could not add strategy %d", strategyID)
		// If queue is full, remove from running list
		s.runningSimulationsMu.Lock()
		delete(s.runningSimulations, strategyID)
		s.runningSimulationsMu.Unlock()
	}
}

// processSimulationQueue processes the queue of strategies to simulate
func (s *AutomationService) processSimulationQueue() {
	s.logger.Info("Starting simulation queue processor")

	// Create a semaphore channel to limit concurrent simulations
	semaphore := make(chan struct{}, s.config.MaxConcurrentSimulations)

	// Create a ticker to periodically check for active simulations
	checkTicker := time.NewTicker(30 * time.Second)
	defer checkTicker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Simulation queue processor stopped")
			return
		case <-checkTicker.C:
			// Periodic check for any orphaned simulations or cleanup
			continue
		case strategyID := <-s.simulationQueue:
			// First, check if there are any active simulations running
			activeRuns, err := s.simulationRunRepo.GetByStatus("running", 5)
			if err != nil {
				s.logger.Error("Error checking for active simulations: %v", err)
				// Continue with caution
			} else if len(activeRuns) > 0 {
				s.logger.Info("Found %d active simulations running. Will not start new simulation for strategy %d yet.", 
					len(activeRuns), strategyID)
				
				// Put the strategy back in the queue after a delay
				go func(id int64) {
					time.Sleep(1 * time.Minute)
					s.queueStrategyForSimulation(id)
				}(strategyID)
				
				// Skip starting a new simulation
				continue
			}
			
			s.logger.Info("No active simulations running, proceeding with strategy %d", strategyID)
			
			// Acquire a slot in the semaphore
			semaphore <- struct{}{}

			go func(id int64) {
				defer func() {
					// Release the semaphore slot when done
					<-semaphore
					
					// Remove from running simulations
					s.runningSimulationsMu.Lock()
					delete(s.runningSimulations, id)
					s.runningSimulationsMu.Unlock()
				}()

				// Verify strategy exists
				strategy, err := s.strategyRepo.GetByID(id)
				if err != nil || strategy == nil {
					s.logger.Error("Strategy %d not found or error retrieving: %v", id, err)
					return
				}

				// Run the simulation
				s.logger.Info("Starting simulation for strategy %d: %s", id, strategy.Name)
				if err := s.simulationService.StartSimulation(id); err != nil {
					s.logger.Error("Error starting simulation for strategy %d: %v", id, err)
					
					// If simulation failed to start, create initial metrics anyway
					// This ensures metrics exist even if simulation doesn't run
					s.createInitialStrategyMetric(id)
					return
				}

				// Wait for a reasonable time for the simulation to complete (e.g., 10 minutes)
				simulationTimeout := 60 * time.Minute
				simulationDone := make(chan bool)

				go func() {
					// Check periodically if the simulation is still running
					ticker := time.NewTicker(30 * time.Second)
					defer ticker.Stop()

					var completedOrTimeoutTriggered bool
					for {
						select {
						case <-ticker.C:
							if completedOrTimeoutTriggered {
								return
							}

							// Check if the simulation is still running
							status, err := s.simulationService.GetSimulationStatus(id)
							
							// If error or status shows not running anymore
							if err != nil || (status != nil && !status.IsActive()) {
								// Check if there are metrics for this strategy
								metric, err := s.performanceAnalyzer.strategyMetricRepo.GetLatestByStrategy(id)
								if err != nil || metric == nil {
									// No metrics exist, create initial metrics
									s.logger.Info("No metrics found for completed simulation of strategy %d, creating initial metrics", id)
									s.createInitialStrategyMetric(id)
								}
								completedOrTimeoutTriggered = true
								simulationDone <- true
								return
							}
						}
					}
				}()

				// Wait for simulation to complete or timeout
				select {
				case <-simulationDone:
					s.logger.Info("Simulation for strategy %d completed", id)
				case <-time.After(simulationTimeout):
					s.logger.Warn("Simulation for strategy %d timed out, stopping", id)
					if err := s.simulationService.StopSimulation(id); err != nil {
						s.logger.Error("Error stopping simulation for strategy %d: %v", id, err)
					}
					
					// Check if metrics exist after timeout
					metric, err := s.performanceAnalyzer.strategyMetricRepo.GetLatestByStrategy(id)
					if err != nil || metric == nil {
						// No metrics exist, create initial metrics
						s.logger.Info("No metrics found after timeout for strategy %d, creating initial metrics", id)
						s.createInitialStrategyMetric(id)
					}
				}

				// After simulation, trigger performance analysis
				if _, err := s.performanceAnalyzer.AnalyzeStrategyPerformance(id); err != nil {
					s.logger.Error("Error analyzing strategy %d: %v", id, err)
				}
			}(strategyID)
		}
	}
}

// createInitialStrategyMetric creates an initial metric record for a strategy
// This ensures that even if a simulation doesn't run or fails, there's always a metric record
func (s *AutomationService) createInitialStrategyMetric(strategyID int64) {
	s.logger.Info("Creating initial metric record for strategy %d", strategyID)
	
	// Get the strategy to retrieve its risk score
	strategy, err := s.strategyRepo.GetByID(strategyID)
	if err != nil || strategy == nil {
		s.logger.Error("Failed to get strategy %d for creating initial metrics: %v", strategyID, err)
		return
	}
	
	// Get current simulation run if any
	var simulationRunID *int64
	currentRun, err := s.simulationRunRepo.GetCurrent()
	if err == nil && currentRun != nil {
		simulationRunID = &currentRun.ID
	}
	
	// Parse strategy config to get the correct balance values
	var config models.StrategyConfig
	configData, err := json.Marshal(strategy.Config)
	if err != nil {
		s.logger.Error("Error marshaling strategy config: %v", err)
		return
	}

	if err := json.Unmarshal(configData, &config); err != nil {
		s.logger.Error("Error unmarshaling strategy config: %v", err)
		return
	}

	// Use actual balance from strategy config
	initialBalance := config.InitialBalance

	// Create initial metric with values from strategy config
	metric := &models.StrategyMetric{
		StrategyID:       strategyID,
		SimulationRunID:  simulationRunID,
		WinRate:          0,
		AvgProfit:        0,
		AvgLoss:          0,
		MaxDrawdown:      0,
		TotalTrades:      0,
		SuccessfulTrades: 0,
		RiskScore:        strategy.RiskScore,
		ROI:              0,
		CurrentBalance:   initialBalance, // Use the actual initial balance from config
		InitialBalance:   initialBalance, // Use the actual initial balance from config
		CreatedAt:        time.Now(),
	}
	
	// Save the metric
	_, err = s.performanceAnalyzer.strategyMetricRepo.Save(metric)
	if err != nil {
		s.logger.Error("Failed to save initial metrics for strategy %d: %v", strategyID, err)
	} else {
		s.logger.Info("Successfully saved initial metrics for strategy %d", strategyID)
	}
}