// internal/api/handlers/simulation_handler.go
package handlers

import (
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/service"
	"github.com/gofiber/fiber/v2"
)

// SimulationHandler handles simulation related requests
type SimulationHandler struct {
	simulationService *service.SimulationService
	logger            *logger.Logger
}

// NewSimulationHandler creates a new simulation handler
func NewSimulationHandler(
	simulationService *service.SimulationService,
	logger *logger.Logger,
) *SimulationHandler {
	return &SimulationHandler{
		simulationService: simulationService,
		logger:            logger,
	}
}

// GetRunningSimulations returns all currently running simulations
func (h *SimulationHandler) GetRunningSimulations(c *fiber.Ctx) error {
	h.logger.Info("Getting all running simulations")
	
	// Get active simulations from the simulation service
	runningSimulations := h.simulationService.GetRunningSimulations()
	
	return c.Status(fiber.StatusOK).JSON(runningSimulations)
}

// GetSimulationSummary gets the simulation summary for a strategy
func (h *SimulationHandler) GetSimulationSummary(c *fiber.Ctx) error {
	h.logger.Info("Getting simulation summary")
	
	// Get strategy ID from params
	strategyID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}
	
	// Get simulation summary
	summary, err := h.simulationService.GetSimulationSummary(int64(strategyID))
	if err != nil {
		if err.Error() == "no simulation found for strategy" {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "No simulation data found for this strategy",
			})
		}
		
		h.logger.Error("Error getting simulation summary: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(summary)
}

// RegisterRoutes registers all simulation routes
func (h *SimulationHandler) RegisterRoutes(app fiber.Router) {
	simulations := app.Group("/simulations")
	simulations.Get("/running", h.GetRunningSimulations)
	simulations.Get("/summary/:id", h.GetSimulationSummary)
}