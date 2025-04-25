// internal/api/server.go
package api

import (
	"context"
	"database/sql"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/api/handlers"
	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/repository"
	"github.com/StratWarsAI/strategy-wars/internal/service"
	"github.com/StratWarsAI/strategy-wars/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Server represents the API server
type Server struct {
	app             *fiber.App
	logger          *logger.Logger
	dataService     *service.DataService
	wsHub           *websocket.WSHub
	wsClientHandler *websocket.ClientWSHandler
	strategyService service.StrategyServiceInterface
	strategyHandler *handlers.StrategyHandler
}

// NewServer creates a new API server
func NewServer(port int, db *sql.DB, logger *logger.Logger) *Server {
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
	userRepo := repository.NewUserRepository(db)
	userScoreRepo := repository.NewUserScoreRepository(db)

	// Create services
	dataService := service.NewDataService(db, logger)
	strategyService := service.NewStrategyService(strategyRepo, userRepo, userScoreRepo, logger)

	// Create handlers
	strategyHandler := handlers.NewStrategyHandler(strategyService, logger)

	server := &Server{
		app:             app,
		logger:          logger,
		dataService:     dataService,
		wsHub:           wsHub,
		wsClientHandler: wsClientHandler,
		strategyService: strategyService,
		strategyHandler: strategyHandler,
	}

	// Register routes
	server.registerRoutes()

	return server
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {

	// Add middleware
	s.app.Use(s.loggingMiddleware())

	// WebSocket endpoint
	//s.router.HandleFunc("/ws", s.wsClientHandler.ServeWS)

	// Add health check endpoint
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := s.app.Group("/api")
	s.strategyHandler.RegisterRoutes(api)
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
func (s *Server) Start() error {
	s.logger.Info("Starting API server on %s", ":8080")
	return s.app.Listen(":8080")
}

// Stop stops the API server gracefully
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")
	return s.app.Shutdown()
}
