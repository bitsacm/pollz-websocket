package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/pollz/websocket-server/internal/config"
	"github.com/pollz/websocket-server/internal/database"
	"github.com/pollz/websocket-server/internal/handlers"
	"github.com/pollz/websocket-server/internal/hub"
	"github.com/pollz/websocket-server/internal/redis"
	"github.com/pollz/websocket-server/internal/server"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize Redis
	redisClient, err := redis.Connect(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Create message hub
	messageHub := hub.New(redisClient, db)
	go messageHub.Run()

	// Create handlers
	wsHandler := handlers.NewWebSocketHandler(messageHub)
	apiHandler := handlers.NewAPIHandler(messageHub)

	// Start server
	srv := server.New(cfg, wsHandler, apiHandler)

	log.Printf("Starting WebSocket server on port %s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
