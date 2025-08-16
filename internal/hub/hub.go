package hub

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pollz/websocket-server/internal/cache"
	"github.com/pollz/websocket-server/internal/models"
	"github.com/pollz/websocket-server/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	clients    map[*models.Client]bool
	broadcast  chan models.Message
	register   chan *models.Client
	unregister chan *models.Client
	mu         sync.RWMutex
	
	messageRepo  *repository.MessageRepository
	messageCache *cache.MessageCache
}

func New(redisClient *redis.Client, db *sql.DB) *Hub {
	return &Hub{
		clients:      make(map[*models.Client]bool),
		broadcast:    make(chan models.Message, 256),
		register:     make(chan *models.Client),
		unregister:   make(chan *models.Client),
		messageRepo:  repository.NewMessageRepository(db),
		messageCache: cache.NewMessageCache(redisClient),
	}
}

func (h *Hub) Run() {
	// Start cleanup routine
	go h.startCleanupRoutine()
	
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)
			
		case client := <-h.unregister:
			h.handleUnregister(client)
			
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

func (h *Hub) Register(client *models.Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *models.Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(message models.Message) {
	h.broadcast <- message
}

func (h *Hub) handleRegister(client *models.Client) {
	h.mu.Lock()
	h.clients[client] = true
	clientCount := len(h.clients)
	h.mu.Unlock()
	
	// Send recent messages to new client
	messages, err := h.getRecentMessages()
	if err != nil {
		log.Printf("Error getting recent messages: %v", err)
		messages = []models.Message{}
	}
	
	response := models.RecentMessagesResponse{
		Type:     "recent_messages",
		Messages: messages,
	}
	
	// Send as a special message type
	select {
	case client.Send <- models.Message{
		Type:    "recent_messages",
		Content: mustMarshalString(response),
	}:
	default:
		// Client's send channel is full, close it
		close(client.Send)
		delete(h.clients, client)
	}
	
	log.Printf("Client %s connected. Total: %d", client.ID, clientCount)
}

func (h *Hub) handleUnregister(client *models.Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
		clientCount := len(h.clients)
		h.mu.Unlock()
		log.Printf("Client %s disconnected. Total: %d", client.ID, clientCount)
	} else {
		h.mu.Unlock()
	}
}

func (h *Hub) handleBroadcast(message models.Message) {
	// Ensure message has an ID
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	
	// Set timestamp if not set
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}
	
	// Save message asynchronously
	go h.saveMessage(message)
	
	// Broadcast to all connected clients
	h.mu.RLock()
	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, close it
			close(client.Send)
			delete(h.clients, client)
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) saveMessage(msg models.Message) {
	// Save to cache
	if err := h.messageCache.Push(msg); err != nil {
		log.Printf("Error saving to cache: %v", err)
	}
	
	// Save to database
	if err := h.messageRepo.Save(msg); err != nil {
		log.Printf("Error saving to database: %v", err)
	}
}

func (h *Hub) getRecentMessages() ([]models.Message, error) {
	// Try cache first
	messages, err := h.messageCache.GetRecent(100)
	if err == nil && len(messages) > 0 {
		return messages, nil
	}
	
	// Fallback to database
	messages, err = h.messageRepo.GetRecent(100)
	if err != nil {
		return nil, err
	}
	
	// Repopulate cache
	if len(messages) > 0 {
		go h.messageCache.Populate(messages)
	}
	
	return messages, nil
}

func (h *Hub) SearchMessages(query string, limit int) ([]models.Message, error) {
	return h.messageRepo.Search(query, limit)
}

func (h *Hub) GetMessagesByDateRange(start, end time.Time) ([]models.Message, error) {
	return h.messageRepo.GetByDateRange(start, end)
}

func (h *Hub) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) startCleanupRoutine() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// Clean messages older than 30 days
		if err := h.messageRepo.DeleteOlderThan(30 * 24 * time.Hour); err != nil {
			log.Printf("Error cleaning old messages: %v", err)
		}
	}
}

func mustMarshalString(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}