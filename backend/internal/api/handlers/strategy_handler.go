package handlers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/StratWarsAI/strategy-wars/internal/api/dto"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/service"
)

type StrategyHandler struct {
	service   service.StrategyServiceInterface
	logger    *logger.Logger
	validator *validator.Validate
}

func NewStrategyHandler(
	strategyService service.StrategyServiceInterface,
	logger *logger.Logger,
) *StrategyHandler {
	return &StrategyHandler{
		service:   strategyService,
		logger:    logger,
		validator: validator.New(),
	}
}

func (h *StrategyHandler) Create(c *fiber.Ctx) error {
	// Parse request body into DTO
	var createDto dto.StrategyCreateDto
	if err := c.BodyParser(&createDto); err != nil {
		h.logger.Error("Error decoding strategy: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate DTO
	if err := h.validator.Struct(createDto); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make([]string, 0, len(validationErrors))
		for _, e := range validationErrors {
			errorMessages = append(errorMessages,
				fmt.Sprintf("%s validation failed: %s", e.Field(), e.Tag()),
			)
		}
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

	// Convert to response DTO
	responseDto := dto.NewStrategyResponseDto(strategy)

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

	// Convert to response DTOs
	responseDtos := make([]dto.StrategyResponseDto, len(strategies))
	for i, strategy := range strategies {
		responseDtos[i] = dto.NewStrategyResponseDto(strategy)
	}

	return c.JSON(responseDtos)
}

func (h *StrategyHandler) RegisterRoutes(app fiber.Router) {
	route := app.Group("")
	route.Post("/strategies", h.Create)
	route.Get("/strategies/:id", h.GetByID)
	route.Get("/strategies", h.GetPublicStrategies)

}
