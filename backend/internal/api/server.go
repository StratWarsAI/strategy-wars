// internal/api/server.go
package api

import (
	"context"
	"database/sql"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/api/handlers"
	"github.com/StratWarsAI/strategy-wars/internal/config"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
	"github.com/StratWarsAI/strategy-wars/internal/service"
	"github.com/StratWarsAI/strategy-wars/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberWebsocket "github.com/gofiber/websocket/v2"
)

// Server represents the API server
type Server struct {
	app                 *fiber.App
	logger              *logger.Logger
	dataService         *service.DataService
	wsHub               *websocket.WSHub
	wsClientHandler     *websocket.ClientWSHandler
	strategyService     service.StrategyServiceInterface
	strategyHandler     *handlers.StrategyHandler
	dashboardService    service.DashboardServiceInterface
	dashboardHandler    *handlers.DashboardHandler
	aiService           *service.AIService
	simulationService   *service.SimulationService
	automationService   *service.AutomationService
	triggerHandler      *handlers.TriggerHandler
	simulationHandler   *handlers.SimulationHandler
	performanceAnalyzer *service.AIPerformanceAnalyzer
}

// NewServer creates a new API server
func NewServer(port int, db *sql.DB, cfg *config.Config, logger *logger.Logger) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	wsHub := websocket.NewWSHub(logger)
	go wsHub.Run()

	// Create a CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	// Create WebSocket client handler
	wsClientHandler := websocket.NewClientWSHandler(wsHub, logger)

	// Create repositories
	strategyRepo := repository.NewStrategyRepository(db)
	strategyMetricRepo := repository.NewStrategyMetricRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	tradeRepo := repository.NewTradeRepository(db)
	simulatedTradeRepo := repository.NewSimulatedTradeRepository(db)
	simulationRunRepo := repository.NewSimulationRunRepository(db)
	simulationEventRepo := repository.NewSimulationEventRepository(db)
	simulationResultRepo := repository.NewSimulationResultRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)

	// Create basic services
	dataService := service.NewDataService(db, logger)
	strategyService := service.NewStrategyService(strategyRepo, strategyMetricRepo, logger)
	dashboardService := service.NewDashboardService(dashboardRepo, simulatedTradeRepo, logger)

	// Create handlers
	strategyHandler := handlers.NewStrategyHandler(strategyService, logger, strategyMetricRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService, logger)

	// Create AI and automation services
	aiService := service.NewAIService(
		cfg.AI.APIKey,
		cfg.AI.Endpoint,
		strategyRepo,
		logger,
	)

	simulationService := service.NewSimulationService(
		db,
		strategyRepo,
		tokenRepo,
		tradeRepo,
		simulatedTradeRepo,
		strategyMetricRepo,
		simulationRunRepo,
		wsHub,
		logger,
	)

	performanceAnalyzer := service.NewAIPerformanceAnalyzer(
		strategyRepo,
		strategyMetricRepo,
		simulatedTradeRepo,
		simulationRunRepo,
		simulationEventRepo,
		simulationResultRepo,
		simulationService,
		aiService,
		logger,
		cfg, // Pass configuration to use the correct analysis interval
	)

	automationService := service.NewAutomationService(
		strategyRepo,
		simulationRunRepo,
		aiService,
		simulationService,
		performanceAnalyzer,
		logger,
		cfg, // Pass the configuration object
	)

	// Create trigger handler
	triggerHandler := handlers.NewTriggerHandler(
		aiService,
		simulationService,
		performanceAnalyzer,
		logger,
	)
	
	// Create simulation handler
	simulationHandler := handlers.NewSimulationHandler(
		simulationService,
		logger,
	)

	server := &Server{
		app:                 app,
		logger:              logger,
		dataService:         dataService,
		wsHub:               wsHub,
		wsClientHandler:     wsClientHandler,
		strategyService:     strategyService,
		strategyHandler:     strategyHandler,
		dashboardService:    dashboardService,
		dashboardHandler:    dashboardHandler,
		aiService:           aiService,
		simulationService:   simulationService,
		automationService:   automationService,
		triggerHandler:      triggerHandler,
		simulationHandler:   simulationHandler,
		performanceAnalyzer: performanceAnalyzer,
	}

	// Create AI handler
	aiHandler := handlers.NewAIHandler(
		aiService,
		performanceAnalyzer,
		logger,
	)

	// Register routes
	server.registerRoutes(aiHandler)

	return server
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes(aiHandler *handlers.AIHandler) {

	// Add middleware
	s.app.Use(s.loggingMiddleware())

	// WebSocket endpoint
	s.app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberWebsocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	s.app.Get("/ws", s.wsClientHandler.ServeWS())

	// Add health check endpoint
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := s.app.Group("/api")

	// Only register strategy routes if handler exists
	if s.strategyHandler != nil {
		s.strategyHandler.RegisterRoutes(api)
	} else {
		s.logger.Warn("Strategy handler is nil, routes not registered")
	}

	// Register trigger routes
	if s.triggerHandler != nil {
		s.triggerHandler.RegisterRoutes(api)
	} else {
		s.logger.Warn("Trigger handler is nil, routes not registered")
	}
	
	// Register simulation routes
	if s.simulationHandler != nil {
		s.simulationHandler.RegisterRoutes(api)
	} else {
		s.logger.Warn("Simulation handler is nil, routes not registered")
	}
	
	// Register AI routes
	if aiHandler != nil {
		aiHandler.RegisterRoutes(api)
	} else {
		s.logger.Warn("AI handler is nil, routes not registered")
	}
	
	// Register dashboard routes
	if s.dashboardHandler != nil {
		s.dashboardHandler.RegisterRoutes(api)
	} else {
		s.logger.Warn("Dashboard handler is nil, routes not registered")
	}
}

// loggingMiddleware logs API requests
func (s *Server) loggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		s.logger.Info("API Request: %s %s", c.Method(), c.Path())

		err := c.Next()

		s.logger.Info("API Response: %s %s - %v", c.Method(), c.Path(), time.Since(start))
		return err
	}
}

// Start starts the API server
func (s *Server) Start(automation_enabled bool) error {
	// Start the automation service if enabled in config
	if automation_enabled && s.automationService != nil {
		if err := s.automationService.Start(); err != nil {
			s.logger.Error("Failed to start automation service: %v", err)
		} else {
			s.logger.Info("Automation service started successfully")
		}
	} else if s.automationService != nil {
		s.logger.Info("Automation service is disabled in configuration")
	}

	s.logger.Info("Starting API server on %s", ":8080")
	return s.app.Listen(":8080")
}

// Stop stops the API server gracefully
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server and services")

	// Stop the automation service if it exists
	if s.automationService != nil {
		if err := s.automationService.Stop(); err != nil {
			s.logger.Error("Error stopping automation service: %v", err)
		} else {
			s.logger.Info("Automation service stopped successfully")
		}
	}

	// Stop the simulation service if it exists
	if s.simulationService != nil {
		s.simulationService.Shutdown()
		s.logger.Info("Simulation service stopped successfully")
	}

	// Shutdown the API server
	return s.app.Shutdown()
}
