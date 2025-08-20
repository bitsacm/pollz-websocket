package server

import (
	"fmt"
	"net/http"

	"github.com/pollz/websocket-server/internal/config"
	"github.com/pollz/websocket-server/internal/handlers"
	"github.com/pollz/websocket-server/internal/middleware"
	"github.com/rs/cors"
)

type Server struct {
	config    *config.Config
	wsHandler *handlers.WebSocketHandler
	apiHandler *handlers.APIHandler
}

func New(cfg *config.Config, wsHandler *handlers.WebSocketHandler, apiHandler *handlers.APIHandler) *Server {
	return &Server{
		config:     cfg,
		wsHandler:  wsHandler,
		apiHandler: apiHandler,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	// WebSocket endpoints - Temporarily disabled (no new message sending)
	// mux.HandleFunc("/ws/chat/live", s.wsHandler.HandleConnection)
	
	// API endpoints - Keep read-only endpoints for existing messages
	mux.HandleFunc("/api/messages/search", s.apiHandler.SearchMessages)
	mux.HandleFunc("/api/messages/date", s.apiHandler.GetMessagesByDate)
	mux.HandleFunc("/api/stats", s.apiHandler.GetStats)
	mux.HandleFunc("/health", s.apiHandler.HealthCheck)
	
	// Apply middleware
	handler := middleware.Logging(mux)
	handler = middleware.Recovery(handler)
	
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   s.config.AllowedOrigins,
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	})
	
	handler = c.Handler(handler)
	
	addr := fmt.Sprintf(":%s", s.config.Port)
	return http.ListenAndServe(addr, handler)
}