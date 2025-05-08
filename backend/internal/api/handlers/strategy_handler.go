package handlers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/StratWarsAI/strategy-wars/internal/api/dto"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
	"github.com/StratWarsAI/strategy-wars/internal/service"
)

type StrategyHandler struct {
	service            service.StrategyServiceInterface
	logger             *logger.Logger
	validator          *validator.Validate
	strategyMetricRepo repository.StrategyMetricRepositoryInterface
}

func NewStrategyHandler(
	strategyService service.StrategyServiceInterface,
	logger *logger.Logger,
	strategyMetricRepo repository.StrategyMetricRepositoryInterface,
) *StrategyHandler {
	return &StrategyHandler{
		service:            strategyService,
		logger:             logger,
		validator:          validator.New(),
		strategyMetricRepo: strategyMetricRepo,
	}
}

func (h *StrategyHandler) Create(c *fiber.Ctx) error {
	// Log the raw request body for debugging
	body := c.Body()
	h.logger.Info("Received raw request body: %s", string(body))
	
	// Parse request body into DTO
	var createDto dto.StrategyCreateDto
	if err := c.BodyParser(&createDto); err != nil {
		h.logger.Error("Error decoding strategy: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}
	
	// Log the parsed DTO for debugging
	h.logger.Info("Parsed DTO: %+v", createDto)

	// Validate DTO
	if err := h.validator.Struct(createDto); err != nil {
		// Try to cast to ValidationErrors, but handle the case where it might be a different error type
		var errorMessages []string
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages = make([]string, 0, len(validationErrors))
			for _, e := range validationErrors {
				errorMessages = append(errorMessages,
					fmt.Sprintf("%s validation failed: %s (value: %v)", e.Field(), e.Tag(), e.Value()),
				)
				h.logger.Error("Validation error: Field '%s', Tag '%s', Value '%v', Namespace '%s'", 
					e.Field(), e.Tag(), e.Value(), e.Namespace())
			}
		} else {
			// Generic error handling for non-validation errors
			errorMessages = []string{err.Error()}
			h.logger.Error("Non-validation error during validation: %v", err)
		}
		
		// Log the actual content of the config for debugging
		h.logger.Error("Config content: %+v", createDto.Config)
		
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errorMessages,
		})
	}

	// Extract user ID (you'll replace this with actual auth middleware)
	userID := int64(1)

	// Convert DTO to model
	strategy := createDto.ToModel(userID)

	// Create strategy
	id, err := h.service.CreateStrategy(strategy)
	if err != nil {
		h.logger.Error("Error creating strategy: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Fetch the created strategy to return full details
	createdStrategy, err := h.service.GetStrategyByID(id)
	if err != nil {
		h.logger.Error("Error fetching created strategy: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to response DTO
	responseDto := dto.NewStrategyResponseDto(createdStrategy)

	// Return created strategy
	return c.Status(fiber.StatusCreated).JSON(responseDto)
}

// GetByID handles retrieving a strategy by its ID
func (h *StrategyHandler) GetByID(c *fiber.Ctx) error {
	// Parse strategy ID from URL parameter
	strategyID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	// Fetch strategy
	strategy, err := h.service.GetStrategyByID(int64(strategyID))
	if err != nil {
		h.logger.Error("Error fetching strategy: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if strategy exists
	if strategy == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Strategy not found",
		})
	}

	// Get latest metrics for this strategy
	latestMetric, err := h.strategyMetricRepo.GetLatestByStrategy(int64(strategyID))
	
	// Convert to response DTO
	responseDto := dto.NewStrategyResponseDto(strategy)
	
	// Add metrics data if available
	if err == nil && latestMetric != nil {
		// Transform the metric data into the expected frontend format
		responseDto.Metrics = &dto.StrategyMetricsDto{
			TotalTrades:      latestMetric.TotalTrades,
			WinningTrades:    latestMetric.SuccessfulTrades,
			LosingTrades:     latestMetric.TotalTrades - latestMetric.SuccessfulTrades,
			WinRate:          latestMetric.WinRate,
			AverageProfitPct: latestMetric.AvgProfit,
			AverageLossPct:   latestMetric.AvgLoss,
			Balance:          latestMetric.CurrentBalance,
			InitialBalance:   latestMetric.InitialBalance,
			ROI:              latestMetric.ROI,
			// Default values for fields not directly in the database
			LargestWinPct:    0,
			LargestLossPct:   0, 
			ProfitFactor:     1.0,
		}
	}

	return c.JSON(responseDto)
}

// GetPublicStrategies handles retrieving public strategies
func (h *StrategyHandler) GetPublicStrategies(c *fiber.Ctx) error {
	// Parse optional query parameters
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	// Fetch public strategies
	strategies, err := h.service.GetPublicStrategies(limit, offset)
	if err != nil {
		h.logger.Error("Error fetching public strategies: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to response DTOs with metrics
	responseDtos := make([]dto.StrategyResponseDto, len(strategies))
	for i, strategy := range strategies {
		responseDtos[i] = dto.NewStrategyResponseDto(strategy)
		
		// Get latest metrics for this strategy
		latestMetric, err := h.strategyMetricRepo.GetLatestByStrategy(strategy.ID)
		if err == nil && latestMetric != nil {
			// Transform the metric data into the expected frontend format
			responseDtos[i].Metrics = &dto.StrategyMetricsDto{
				TotalTrades:      latestMetric.TotalTrades,
				WinningTrades:    latestMetric.SuccessfulTrades,
				LosingTrades:     latestMetric.TotalTrades - latestMetric.SuccessfulTrades,
				WinRate:          latestMetric.WinRate,
				AverageProfitPct: latestMetric.AvgProfit,
				AverageLossPct:   latestMetric.AvgLoss,
				Balance:          latestMetric.CurrentBalance,
				InitialBalance:   latestMetric.InitialBalance,
				ROI:              latestMetric.ROI,
				LargestWinPct:    0,
				LargestLossPct:   0,
				ProfitFactor:     1.0,
			}
		}
	}

	return c.JSON(responseDtos)
}

// GetTopStrategies handles retrieving top performing strategies
func (h *StrategyHandler) GetTopStrategies(c *fiber.Ctx) error {
	// Parse optional query parameters
	limit := c.QueryInt("limit", 12)
	criteria := c.Query("criteria", "performance")

	// Fetch top strategies
	strategies, err := h.service.GetTopStrategies(criteria, limit)
	if err != nil {
		h.logger.Error("Error fetching top strategies: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Fetch latest metrics for each strategy and prepare response
	responseDtos := make([]dto.StrategyResponseDto, len(strategies))
	for i, strategy := range strategies {
		// Convert to response DTO
		responseDtos[i] = dto.NewStrategyResponseDto(strategy)
		
		// Get latest metrics for this strategy
		latestMetric, err := h.strategyMetricRepo.GetLatestByStrategy(strategy.ID)
		if err == nil && latestMetric != nil {
			// Transform the metric data into the expected frontend format
			responseDtos[i].Metrics = &dto.StrategyMetricsDto{
				TotalTrades:      latestMetric.TotalTrades,
				WinningTrades:    latestMetric.SuccessfulTrades,
				LosingTrades:     latestMetric.TotalTrades - latestMetric.SuccessfulTrades,
				WinRate:          latestMetric.WinRate,
				AverageProfitPct: latestMetric.AvgProfit,
				AverageLossPct:   latestMetric.AvgLoss,
				Balance:          latestMetric.CurrentBalance,
				InitialBalance:   latestMetric.InitialBalance,
				ROI:              latestMetric.ROI,
				LargestWinPct:    0,
				LargestLossPct:   0,
				ProfitFactor:     1.0,
			}
		}
	}

	return c.JSON(responseDtos)
}

func (h *StrategyHandler) RegisterRoutes(app fiber.Router) {
	route := app.Group("")
	route.Post("/strategies", h.Create)
	// Order matters for routes, more specific routes should come first
	route.Get("/strategies/top", h.GetTopStrategies)
	route.Get("/strategies/:id", h.GetByID)
	route.Get("/strategies", h.GetPublicStrategies)
}
