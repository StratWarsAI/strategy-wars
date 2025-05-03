// internal/api/handlers/trigger_handler.go
package handlers

import (
	"fmt"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/service"
	"github.com/gofiber/fiber/v2"
)

// TriggerHandler handles manual triggers for automated services
type TriggerHandler struct {
	aiService           *service.AIService
	simulationService   *service.SimulationService
	performanceAnalyzer *service.AIPerformanceAnalyzer
	logger              *logger.Logger
}

// NewTriggerHandler creates a new trigger handler
func NewTriggerHandler(
	aiService *service.AIService,
	simulationService *service.SimulationService,
	performanceAnalyzer *service.AIPerformanceAnalyzer,
	logger *logger.Logger,
) *TriggerHandler {
	return &TriggerHandler{
		aiService:           aiService,
		simulationService:   simulationService,
		performanceAnalyzer: performanceAnalyzer,
		logger:              logger,
	}
}

// TriggerStrategyCreation manually triggers AI strategy creation
func (h *TriggerHandler) TriggerStrategyCreation(c *fiber.Ctx) error {
	h.logger.Info("Manual trigger for strategy creation received")

	// Get prompt from request body if provided, otherwise use default
	var body struct {
		Prompt string `json:"prompt"`
	}

	if err := c.BodyParser(&body); err != nil {
		// Use default prompt if body parsing fails
		body.Prompt = "Generate a profitable trading strategy for cryptocurrency tokens"
	}

	// Use empty prompt if none provided
	if body.Prompt == "" {
		body.Prompt = "Generate a profitable trading strategy for cryptocurrency tokens"
	}

	// Get top performing strategies to learn from
	topStrategies, err := h.aiService.GetTopPerformingStrategies()
	if err != nil {
		h.logger.Error("Error getting top strategies: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error getting top strategies: %v", err),
		})
	}

	// Create metadata with top strategies
	metadata := map[string]interface{}{
		"top_strategies": topStrategies,
	}

	// Generate new strategy
	strategy, err := h.aiService.GenerateStrategy(body.Prompt, metadata)
	if err != nil {
		h.logger.Error("Error generating strategy: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error generating strategy: %v", err),
		})
	}

	// Save the strategy
	id, err := h.aiService.SaveStrategy(strategy)
	if err != nil {
		h.logger.Error("Error saving generated strategy: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error saving generated strategy: %v", err),
		})
	}

	h.logger.Info("Successfully generated and saved new strategy with ID: %d", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":     true,
		"message":     fmt.Sprintf("Strategy created with ID: %d", id),
		"strategy_id": id,
	})
}

// TriggerSimulation manually triggers a simulation for a strategy
func (h *TriggerHandler) TriggerSimulation(c *fiber.Ctx) error {
	h.logger.Info("Manual trigger for simulation received")

	// Get strategy ID from params
	strategyID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	// First, try to stop any existing simulation for this strategy
	// Ignore any errors since there might not be a simulation running
	_ = h.simulationService.StopSimulation(int64(strategyID))

	// Give it a moment to clean up (if there was a simulation)
	time.Sleep(1 * time.Second)

	// Start simulation
	if err := h.simulationService.StartSimulation(int64(strategyID)); err != nil {
		h.logger.Error("Error starting simulation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error starting simulation: %v", err),
		})
	}

	h.logger.Info("Successfully started simulation for strategy %d", strategyID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Simulation started for strategy ID: %d", strategyID),
	})
}

// TriggerAnalysis manually triggers performance analysis
func (h *TriggerHandler) TriggerAnalysis(c *fiber.Ctx) error {
	h.logger.Info("Manual trigger for performance analysis received")

	// Run the analysis cycle
	if err := h.performanceAnalyzer.RunAnalysisCycle(); err != nil {
		h.logger.Error("Error running performance analysis: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error running performance analysis: %v", err),
		})
	}

	h.logger.Info("Successfully ran performance analysis")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Performance analysis completed",
	})
}

// TriggerStopSimulation stops a running simulation
func (h *TriggerHandler) TriggerStopSimulation(c *fiber.Ctx) error {
	h.logger.Info("Manual trigger to stop simulation received")

	// Get strategy ID from params
	strategyID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	// Stop simulation
	if err := h.simulationService.StopSimulation(int64(strategyID)); err != nil {
		h.logger.Error("Error stopping simulation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error stopping simulation: %v", err),
		})
	}

	h.logger.Info("Successfully stopped simulation for strategy %d", strategyID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Simulation stopped for strategy ID: %d", strategyID),
	})
}

// TriggerGetSimulationStatus gets the status of a simulation
func (h *TriggerHandler) TriggerGetSimulationStatus(c *fiber.Ctx) error {
	h.logger.Info("Getting simulation status")

	// Get strategy ID from params
	strategyID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	// Get simulation status
	status, err := h.simulationService.GetSimulationStatus(int64(strategyID))
	if err != nil {
		// If no simulation is found, it's not really an error
		if err.Error() == fmt.Sprintf("no simulation found for strategy %d", strategyID) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"success": true,
				"running": false,
				"message": fmt.Sprintf("No active simulation for strategy ID: %d", strategyID),
			})
		}

		h.logger.Error("Error getting simulation status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error getting simulation status: %v", err),
		})
	}

	// Check if simulation is running using the IsActive() method
	isRunning := false
	if status != nil {
		isRunning = status.IsActive()
	}

	h.logger.Info("Got simulation status for strategy %d: running=%v", strategyID, isRunning)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"running": isRunning,
		"message": fmt.Sprintf("Simulation status for strategy ID: %d, Running: %v", strategyID, isRunning),
	})
}

// RegisterRoutes registers all trigger routes
func (h *TriggerHandler) RegisterRoutes(app fiber.Router) {
	triggers := app.Group("/trigger")
	triggers.Post("/create-strategy", h.TriggerStrategyCreation)
	triggers.Post("/simulate/:id", h.TriggerSimulation)
	triggers.Post("/stop/:id", h.TriggerStopSimulation)
	triggers.Get("/status/:id", h.TriggerGetSimulationStatus)
	triggers.Post("/analyze", h.TriggerAnalysis)
}
