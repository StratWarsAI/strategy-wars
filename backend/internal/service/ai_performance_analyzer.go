// internal/service/ai_performance_analyzer.go
package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/config"
	"github.com/StratWarsAI/strategy-wars/internal/models"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
)

// PerformanceReport represents a structured analysis of a strategy's performance
type PerformanceReport struct {
	StrategyID      int64                  `json:"strategy_id"`
	StrategyName    string                 `json:"strategy_name"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	TotalTrades     int                    `json:"total_trades"`
	WinRate         float64                `json:"win_rate"`
	ROI             float64                `json:"roi"`
	MaxDrawdown     float64                `json:"max_drawdown"`
	NetPnL          float64                `json:"net_pnl"`
	AvgTradeProfit  float64                `json:"avg_trade_profit"`
	Analysis        string                 `json:"analysis"`
	Rating          string                 `json:"rating"`
	Metrics         map[string]interface{} `json:"metrics"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// AIPerformanceAnalyzer analyzes the performance of AI-generated strategies
type AIPerformanceAnalyzer struct {
	strategyRepo         repository.StrategyRepositoryInterface
	strategyMetricRepo   repository.StrategyMetricRepositoryInterface
	simulatedTradeRepo   repository.SimulatedTradeRepositoryInterface
	simulationRunRepo    repository.SimulationRunRepositoryInterface
	simulationEventRepo  repository.SimulationEventRepositoryInterface
	simulationResultRepo repository.SimulationResultRepositoryInterface
	simulationService    *SimulationService
	aiService            *AIService
	logger               *logger.Logger
	analysisInterval     time.Duration
	lastAnalysisTime     time.Time
	activeAnalysis       bool
	mu                   sync.Mutex
}

// NewAIPerformanceAnalyzer creates a new AI performance analyzer
func NewAIPerformanceAnalyzer(
	strategyRepo repository.StrategyRepositoryInterface,
	strategyMetricRepo repository.StrategyMetricRepositoryInterface,
	simulatedTradeRepo repository.SimulatedTradeRepositoryInterface,
	simulationRunRepo repository.SimulationRunRepositoryInterface,
	simulationEventRepo repository.SimulationEventRepositoryInterface,
	simulationResultRepo repository.SimulationResultRepositoryInterface,
	simulationService *SimulationService,
	aiService *AIService,
	logger *logger.Logger,
	cfg *config.Config,
) *AIPerformanceAnalyzer {
	// Convert minutes to time.Duration
	analysisInterval := time.Duration(cfg.Automation.PerformanceAnalysisInterval) * time.Minute
	logger.Info("Initializing Performance Analyzer with analysis interval: %v", analysisInterval)

	return &AIPerformanceAnalyzer{
		strategyRepo:         strategyRepo,
		strategyMetricRepo:   strategyMetricRepo,
		simulatedTradeRepo:   simulatedTradeRepo,
		simulationRunRepo:    simulationRunRepo,
		simulationEventRepo:  simulationEventRepo,
		simulationResultRepo: simulationResultRepo,
		simulationService:    simulationService,
		aiService:            aiService,
		logger:               logger,
		analysisInterval:     analysisInterval,
		lastAnalysisTime:     time.Now(),
		activeAnalysis:       false,
	}
}

// StartAutomatedAnalysis starts the automated performance analysis cycle
func (a *AIPerformanceAnalyzer) StartAutomatedAnalysis(ctx context.Context) {
	a.logger.Info("Starting automated performance analysis service")

	ticker := time.NewTicker(1 * time.Minute) // Check every minute if it's time to run an analysis
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Stopping automated performance analysis service")
			return
		case <-ticker.C:
			a.mu.Lock()
			if time.Since(a.lastAnalysisTime) >= a.analysisInterval && !a.activeAnalysis {
				a.activeAnalysis = true
				a.mu.Unlock()

				a.logger.Info("Starting scheduled performance analysis")
				go func() {
					if err := a.RunAnalysisCycle(); err != nil {
						a.logger.Error("Error in analysis cycle: %v", err)
					}

					a.mu.Lock()
					a.lastAnalysisTime = time.Now()
					a.activeAnalysis = false
					a.mu.Unlock()
				}()
			} else {
				a.mu.Unlock()
			}
		}
	}
}

// RunAnalysisCycle runs a full cycle of performance analysis
func (a *AIPerformanceAnalyzer) RunAnalysisCycle() error {
	a.logger.Info("Running performance analysis cycle")

	// Get all AI-enhanced strategies
	strategies, err := a.getAllAIStrategies()
	if err != nil {
		return fmt.Errorf("error getting AI strategies: %v", err)
	}

	a.logger.Info("Found %d AI strategies to analyze", len(strategies))

	if len(strategies) == 0 {
		return nil // Nothing to analyze
	}

	// Check if there are any active simulations running
	activeRuns, err := a.simulationRunRepo.GetByStatus("running", 5)
	if err != nil {
		a.logger.Error("Error checking for active simulations during analysis: %v", err)
	} else if len(activeRuns) > 0 {
		a.logger.Info("Found %d active simulations during analysis cycle. Will include these in the analysis.", len(activeRuns))
		for _, run := range activeRuns {
			a.logger.Info("Active simulation found: ID=%d, Started=%v", run.ID, run.StartTime)
		}
	}

	// Analyze each strategy's performance
	var reports []*PerformanceReport
	for _, strategy := range strategies {
		report, err := a.AnalyzeStrategyPerformance(strategy.ID)
		if err != nil {
			a.logger.Error("Error analyzing strategy %d: %v", strategy.ID, err)
			continue
		}
		reports = append(reports, report)

		// Save the analysis report to the database
		if err := a.saveAnalysisReport(report); err != nil {
			a.logger.Error("Error saving analysis report for strategy %d: %v", strategy.ID, err)
		}
	}

	// Sort reports by performance (ROI)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ROI > reports[j].ROI
	})

	// Log performance summary
	a.logPerformanceSummary(reports)

	return nil
}

// saveAnalysisReport saves an analysis report to the database
func (a *AIPerformanceAnalyzer) saveAnalysisReport(report *PerformanceReport) error {
	// Get the current active simulation run
	currentRun, err := a.simulationRunRepo.GetCurrent()
	if err != nil {
		return fmt.Errorf("error getting current simulation run: %v", err)
	}

	var simulationRunID int64
	if currentRun != nil {
		simulationRunID = currentRun.ID
	} else {
		// If there's no active simulation, get the most recent one
		runs, err := a.simulationRunRepo.GetByStatus("completed", 1)
		if err != nil {
			return fmt.Errorf("error getting latest completed simulation run: %v", err)
		}

		if len(runs) > 0 {
			simulationRunID = runs[0].ID
		} else {
			return fmt.Errorf("no simulation run found to associate with the analysis")
		}
	}

	// Create and save a simulation result
	result := &models.SimulationResult{
		SimulationRunID:   simulationRunID,
		StrategyID:        report.StrategyID,
		ROI:               report.ROI,
		TradeCount:        report.TotalTrades,
		WinRate:           report.WinRate,
		MaxDrawdown:       report.MaxDrawdown,
		PerformanceRating: report.Rating,
		Analysis:          report.Analysis,
		CreatedAt:         time.Now(),
	}

	_, err = a.simulationResultRepo.Save(result)
	if err != nil {
		return fmt.Errorf("error saving simulation result: %v", err)
	}

	// Also create a simulation event to record the analysis
	event := &models.SimulationEvent{
		StrategyID:      report.StrategyID,
		SimulationRunID: simulationRunID,
		EventType:       "ai_analysis", // Create a specific event type for AI analysis
		EventData: models.JSONB{
			"analysis":         report.Analysis,
			"rating":           report.Rating,
			"roi":              report.ROI,
			"win_rate":         report.WinRate,
			"total_trades":     report.TotalTrades,
			"max_drawdown":     report.MaxDrawdown,
			"net_pnl":          report.NetPnL,
			"avg_trade_profit": report.AvgTradeProfit,
		},
		Timestamp: report.GeneratedAt,
		CreatedAt: time.Now(),
	}

	_, err = a.simulationEventRepo.Save(event)
	if err != nil {
		return fmt.Errorf("error saving simulation event: %v", err)
	}

	a.logger.Info("Saved analysis report for strategy %d (%s) with rating %s",
		report.StrategyID, report.StrategyName, report.Rating)

	return nil
}

// AnalyzeStrategyPerformance analyzes a single strategy's performance
func (a *AIPerformanceAnalyzer) AnalyzeStrategyPerformance(strategyID int64) (*PerformanceReport, error) {
	a.logger.Info("Analyzing performance for strategy %d", strategyID)

	// Get strategy details
	strategy, err := a.strategyRepo.GetByID(strategyID)
	if err != nil {
		return nil, fmt.Errorf("error getting strategy: %v", err)
	}

	if strategy == nil {
		return nil, fmt.Errorf("strategy not found: %d", strategyID)
	}

	// Get all trades for this strategy (both completed and active)
	allTrades, err := a.simulatedTradeRepo.GetByStrategyID(strategyID)
	if err != nil {
		a.logger.Error("Error getting all trades for strategy %d: %v", strategyID, err)
		// Continue with empty trades if we can't get trades
		allTrades = []*models.SimulatedTrade{}
	}

	// Get active trades separately for logging/debugging
	activeTrades, err := a.simulatedTradeRepo.GetActiveByStrategyID(strategyID)
	if err != nil {
		a.logger.Error("Error getting active trades for strategy %d: %v", strategyID, err)
		activeTrades = []*models.SimulatedTrade{}
	}

	a.logger.Info("Found %d total trades (%d active) for strategy %d",
		len(allTrades), len(activeTrades), strategyID)

	// Use all trades for analysis, not just active ones
	trades := allTrades

	// Get performance metrics
	metrics, err := a.simulatedTradeRepo.GetSummaryByStrategyID(strategyID)
	if err != nil {
		return nil, fmt.Errorf("error getting performance metrics: %v", err)
	}

	// Get the latest metric record
	latestMetric, err := a.strategyMetricRepo.GetLatestByStrategy(strategyID)
	if err != nil {
		a.logger.Error("Error getting latest metric for strategy %d: %v", strategyID, err)
		// Continue without latest metric
	}

	// Build performance report
	report := &PerformanceReport{
		StrategyID:   strategyID,
		StrategyName: strategy.Name,
		TotalTrades:  len(trades),
		Metrics:      metrics,
		GeneratedAt:  time.Now(),
	}

	// Include total trades from metrics which correctly counts all trade types
	if totalTradesIncludingActive, ok := metrics["total_trades_including_active"].(int); ok {
		report.TotalTrades = totalTradesIncludingActive
	}

	// Extract metrics from summary
	if winRate, ok := metrics["win_rate"].(float64); ok {
		report.WinRate = winRate
	}
	if roi, ok := metrics["roi"].(float64); ok {
		report.ROI = roi
	}
	if maxDrawdown, ok := metrics["max_drawdown"].(float64); ok {
		report.MaxDrawdown = maxDrawdown
	}
	if netPnL, ok := metrics["net_pnl"].(float64); ok {
		report.NetPnL = netPnL
	}
	if avgProfit, ok := metrics["avg_profit"].(float64); ok {
		report.AvgTradeProfit = avgProfit
	}

	// Generate performance rating
	report.Rating = a.calculatePerformanceRating(report)

	// Generate textual analysis
	report.Analysis = a.generatePerformanceAnalysis(report, strategy, latestMetric)

	return report, nil
}

// calculatePerformanceRating assigns a performance rating based on metrics
func (a *AIPerformanceAnalyzer) calculatePerformanceRating(report *PerformanceReport) string {
	// No trades means we can't evaluate
	if report.TotalTrades == 0 {
		return "not_rated"
	}

	// Base rating on ROI and other factors
	if report.ROI >= 50 {
		return "excellent"
	} else if report.ROI >= 20 {
		return "good"
	} else if report.ROI >= 0 {
		return "average"
	} else if report.ROI >= -20 {
		return "poor"
	} else {
		return "very_poor"
	}
}

// generatePerformanceAnalysis creates a textual analysis of the strategy's performance
func (a *AIPerformanceAnalyzer) generatePerformanceAnalysis(
	report *PerformanceReport,
	strategy *models.Strategy,
	latestMetric *models.StrategyMetric,
) string {
	// Start with a basic analysis
	analysis := fmt.Sprintf("Strategy '%s' ", strategy.Name)

	// Check if there are any active trades for this strategy
	activeTrades, err := a.simulatedTradeRepo.GetActiveByStrategyID(strategy.ID)
	hasActiveTrades := err == nil && len(activeTrades) > 0

	// Also check if there are any active simulations running
	activeRuns, err := a.simulationRunRepo.GetByStatus("running", 1)
	isActiveSimulation := err == nil && len(activeRuns) > 0

	// Check if this specific strategy is currently being simulated
	isThisStrategyActive := false
	if isActiveSimulation && len(activeRuns) > 0 {
		// Get trades from the active simulation run to see if they include this strategy
		simulationTrades, err := a.simulatedTradeRepo.GetBySimulationRun(activeRuns[0].ID)
		if err == nil {
			for _, trade := range simulationTrades {
				if trade.StrategyID == strategy.ID {
					isThisStrategyActive = true
					break
				}
			}
		}
	}

	// Get all trades for this strategy
	allTrades, err := a.simulatedTradeRepo.GetByStrategyID(strategy.ID)
	totalTradesCount := 0
	if err == nil {
		totalTradesCount = len(allTrades)
	}

	// Get completed trades count
	var completedTradesCount int
	if totalTrades, ok := report.Metrics["total_trades"].(int); ok {
		completedTradesCount = totalTrades
	}

	// No trades at all
	if totalTradesCount == 0 {
		if isThisStrategyActive {
			return analysis + "is currently running but has no positions yet."
		} else {
			return analysis + "has not executed any trades yet, so performance cannot be evaluated."
		}
	}

	// Strategy has trades (active, completed, or both)
	if hasActiveTrades {
		analysis += fmt.Sprintf("has %d active positions", len(activeTrades))
		if completedTradesCount > 0 {
			analysis += fmt.Sprintf(" and %d completed trades. ", completedTradesCount)
		} else {
			analysis += ". "
		}
	} else if completedTradesCount > 0 {
		analysis += fmt.Sprintf("has completed %d trades. ", completedTradesCount)
	}

	// Add performance metrics
	if completedTradesCount > 0 {
		// Add ROI information
		if report.ROI > 0 {
			analysis += fmt.Sprintf("It has performed positively with an ROI of %.2f%%. ", report.ROI)
		} else {
			analysis += fmt.Sprintf("It has performed negatively with an ROI of %.2f%%. ", report.ROI)
		}

		// Add win rate details
		analysis += fmt.Sprintf("The strategy won %.2f%% of its completed trades. ", report.WinRate)
	} else if hasActiveTrades {
		// Only has active trades, no completed ones
		analysis += "Since all positions are still active, final performance metrics cannot be calculated yet. "
	}

	// Add active simulation disclaimer if this strategy is currently running
	if isThisStrategyActive {
		analysis += "This strategy is currently being simulated. Metrics may change as trades complete. "
	}

	// Add drawdown analysis
	if report.MaxDrawdown > 0 {
		analysis += fmt.Sprintf("Maximum drawdown was %.2f%%. ", report.MaxDrawdown)

		if report.MaxDrawdown > 50 {
			analysis += "This indicates high volatility and risk. "
		} else if report.MaxDrawdown > 25 {
			analysis += "This indicates moderate volatility and risk. "
		}
	}

	// Recommendation based on performance
	switch report.Rating {
	case "excellent":
		analysis += "This strategy is performing exceptionally well and should be maintained."
	case "good":
		analysis += "This strategy is performing well and should be considered for further optimization."
	case "average":
		analysis += "This strategy is performing adequately but could benefit from optimization."
	case "poor":
		analysis += "This strategy is underperforming and should be reviewed for potential improvements."
	case "very_poor":
		analysis += "This strategy is performing poorly and should be reconsidered or replaced."
	}

	return analysis
}

// getAllAIStrategies gets all AI-enhanced strategies
func (a *AIPerformanceAnalyzer) getAllAIStrategies() ([]*models.Strategy, error) {
	// In a real implementation, you would modify the repository to support this query directly
	// For now, we'll get all public strategies and filter
	strategies, err := a.strategyRepo.ListPublic(1000, 0)
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

// logPerformanceSummary logs a summary of all strategy performance
func (a *AIPerformanceAnalyzer) logPerformanceSummary(reports []*PerformanceReport) {
	a.logger.Info("=== Strategy Performance Summary ===")

	if len(reports) == 0 {
		a.logger.Info("No strategy reports available")
		return
	}

	// Log top 3 strategies
	topCount := min(3, len(reports))
	a.logger.Info("Top %d Strategies:", topCount)
	for i := 0; i < topCount; i++ {
		a.logger.Info("%d. %s (ID: %d): ROI: %.2f%%, Win Rate: %.2f%%, Trades: %d, Rating: %s",
			i+1, reports[i].StrategyName, reports[i].StrategyID,
			reports[i].ROI, reports[i].WinRate, reports[i].TotalTrades, reports[i].Rating)
	}

	// Calculate overall statistics
	var totalROI, totalWinRate float64
	totalTrades := 0
	goodPerformerCount := 0

	for _, report := range reports {
		totalROI += report.ROI
		totalWinRate += report.WinRate
		totalTrades += report.TotalTrades

		if report.Rating == "excellent" || report.Rating == "good" {
			goodPerformerCount++
		}
	}

	// Log overall stats
	a.logger.Info("Overall Statistics:")
	a.logger.Info("- Total Strategies: %d", len(reports))
	a.logger.Info("- Total Trades: %d", totalTrades)

	if len(reports) > 0 {
		a.logger.Info("- Average ROI: %.2f%%", totalROI/float64(len(reports)))
		a.logger.Info("- Average Win Rate: %.2f%%", totalWinRate/float64(len(reports)))
		a.logger.Info("- Good Performers: %d (%.2f%%)",
			goodPerformerCount, float64(goodPerformerCount)/float64(len(reports))*100)
	}

	a.logger.Info("=====================================")
}

// helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
