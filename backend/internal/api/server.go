// internal/api/server.go
package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/StratWarsAI/strategy-wars/internal/pkg/logger"
	"github.com/StratWarsAI/strategy-wars/internal/service"
	"github.com/StratWarsAI/strategy-wars/internal/websocket"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Server represents the API server
type Server struct {
	router          *mux.Router
	server          *http.Server
	logger          *logger.Logger
	dataService     *service.DataService
	wsHub           *websocket.WSHub
	wsClientHandler *websocket.ClientWSHandler
}

// NewServer creates a new API server
func NewServer(port int, db *sql.DB, logger *logger.Logger) *Server {
	router := mux.NewRouter()

	// Create a CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true, // Enable for debugging, remove in production
	})

	wsHub := websocket.NewWSHub(logger)
	go wsHub.Run()

	// Create WebSocket client handler
	wsClientHandler := websocket.NewClientWSHandler(wsHub, logger)

	// Create services
	dataService := service.NewDataService(db, logger)

	server := &Server{
		router: router,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      corsMiddleware.Handler(router), // Apply CORS middleware
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger:          logger,
		dataService:     dataService,
		wsHub:           wsHub,
		wsClientHandler: wsClientHandler,
	}

	// Register routes
	server.registerRoutes()

	return server
}

// CorsMiddleware is a middleware function that adds CORS headers to the response
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // In production, specify exact origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {

	// Add middleware
	s.router.Use(s.loggingMiddleware)

	// WebSocket endpoint
	s.router.HandleFunc("/ws", s.wsClientHandler.ServeWS)

	// Add CORS middleware
	s.router.Use(corsMiddleware)

	// Add health check endpoint
	s.router.HandleFunc("/health", s.healthCheckHandler).Methods("GET")
}

// loggingMiddleware logs API requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.logger.Info("API Request: %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		s.logger.Info("API Response: %s %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// healthCheckHandler handles health check requests
func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(`{"status":"ok"}`))
	if err != nil {
		// Log the error or handle it appropriately
		s.logger.Error("Failed to write response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Start starts the API server
func (s *Server) Start() error {
	s.logger.Info("Starting API server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop stops the API server gracefully
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")
	return s.server.Shutdown(ctx)
}
