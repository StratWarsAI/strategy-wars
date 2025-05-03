// internal/service/automation_service.go
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

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
) *AutomationService {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AutomationService{
		config: AutomationConfig{
			StrategyGenerationInterval: 1 * time.Hour,
			PerformanceAnalysisInterval: 15 * time.Minute,
			StrategiesPerInterval: 2, // Generate 2 strategies each hour
			MaxConcurrentSimulations: 3, // Run up to 3 simulations concurrently
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

	s.logger.Info("Starting automation service")
	s.isRunning = true

	// Start the simulation queue processor
	go s.processSimulationQueue()

	// Start the main automation loop
	go s.runAutomationLoop()

	// Start the performance analyzer
	go s.performanceAnalyzer.StartAutomatedAnalysis(s.ctx)

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

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Automation loop stopped")
			return
		case <-ticker.C:
			// Check if it's time to generate new strategies
			if time.Since(s.lastStrategyGenTime) >= s.config.StrategyGenerationInterval {
				s.logger.Info("Triggering scheduled strategy generation")
				go s.generateStrategies()
				s.lastStrategyGenTime = time.Now()
			}

			// Check for pending strategies that need simulation
			go s.checkPendingStrategies()
		}
	}
}

// generateStrategies generates new AI strategies
func (s *AutomationService) generateStrategies() {
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

		s.logger.Info("Successfully generated and saved new strategy %d with ID: %d", i+1, id)
		
		// Add strategy to simulation queue
		s.queueStrategyForSimulation(id)
	}
}

// checkPendingStrategies checks for strategies that need to be simulated
func (s *AutomationService) checkPendingStrategies() {
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

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Simulation queue processor stopped")
			return
		case strategyID := <-s.simulationQueue:
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

				// Run the simulation
				s.logger.Info("Starting simulation for strategy %d", id)
				if err := s.simulationService.StartSimulation(id); err != nil {
					s.logger.Error("Error starting simulation for strategy %d: %v", id, err)
					return
				}

				// Wait for a reasonable time for the simulation to complete (e.g., 10 minutes)
				simulationTimeout := 10 * time.Minute
				simulationDone := make(chan bool)

				go func() {
					// Check periodically if the simulation is still running
					ticker := time.NewTicker(30 * time.Second)
					defer ticker.Stop()

					for {
						select {
						case <-ticker.C:
							// Check if the simulation is still running
							_, err := s.simulationService.GetSimulationStatus(id)
							if err != nil {
								// If error, the simulation is probably done
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
				}

				// After simulation, trigger performance analysis
				if _, err := s.performanceAnalyzer.AnalyzeStrategyPerformance(id); err != nil {
					s.logger.Error("Error analyzing strategy %d: %v", id, err)
				}
			}(strategyID)
		}
	}
}