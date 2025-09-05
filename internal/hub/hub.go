package hub

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pollz/websocket-server/internal/cache"
	"github.com/pollz/websocket-server/internal/models"
	"github.com/pollz/websocket-server/internal/repository"
	"github.com/redis/go-redis/v9"
)
type TrieNode struct {
    children map[rune]*TrieNode
    isEnd    bool
}

// Trie structure
type Trie struct {
    root *TrieNode
}

// Create new Trie
func NewTrie() *Trie {
    return &Trie{root: &TrieNode{children: make(map[rune]*TrieNode)}}
}

// Insert a word into the Trie
func (t *Trie) Insert(word string) {
    w := strings.ToLower(strings.TrimSpace(word))
    if w == "" {
        return
    }
    node := t.root
    for _, ch := range w {
        if node.children[ch] == nil {
            node.children[ch] = &TrieNode{children: make(map[rune]*TrieNode)}
        }
        node = node.children[ch]
    }
    node.isEnd = true
}

// Search checks if the word exists in the Trie
func (t *Trie) Search(word string) bool {
    node := t.root
    for _, ch := range word {
        if node.children[ch] == nil {
            return false
        }
        node = node.children[ch]
    }
    return node.isEnd
}
type Hub struct {
	clients    map[*models.Client]bool
	broadcast  chan models.Message
	register   chan *models.Client
	unregister chan *models.Client
	mu         sync.RWMutex
	tri       *Trie
	messageRepo  *repository.MessageRepository
	messageCache *cache.MessageCache
}


func New(redisClient *redis.Client, db *sql.DB) *Hub {
	words := []string{
        "aad", "aand", "bahenchod", "behenchod", "bhenchod", "bhenchodd", "b.c.", "bc",
        "bakchod", "bakchodd", "bakchodi", "bevda", "bewda", "bevdey", "bewday", "bevakoof",
        "bevkoof", "bevkuf", "bewakoof", "bewkoof", "bewkuf", "bhadua", "bhaduaa", "bhadva",
        "bhadvaa", "bhadwa", "bhadwaa", "bhosada", "bhosda", "bhosdaa", "bhosdike", "bhonsdike",
        "bsdk", "b.s.d.k", "bhosdiki", "bhosdiwala", "bhosdiwale", "bhosadchodal", "bhosadchod",
        "babbe", "babbey", "bube", "bubey", "bur", "burr", "buurr", "buur", "charsi", "chooche",
        "choochi", "chuchi", "chhod", "chod", "chodd", "chudne", "chudney", "chudwa", "chudwaa",
        "chudwane", "chudwaane", "choot", "chut", "chute", "chutia", "chutiya", "chutiye",
        "chuttad", "chutad", "dalaal", "dalal", "dalle", "dalley", "fattu", "gadha", "gadhe",
        "gadhalund", "gaand", "gand", "gandu", "gandfat", "gandfut", "gandiya", "gandiye", "goo",
        "gu", "gote", "gotey", "gotte", "hag", "haggu", "hagne", "hagney", "harami", "haramjada",
        "haraamjaada", "haramzyada", "haraamzyaada", "haraamjaade", "haraamzaade", "haraamkhor",
        "haramkhor", "jhat", "jhaat", "jhaatu", "jhatu", "kutta", "kutte", "kuttey", "kutia",
        "kutiya", "kuttiya", "kutti", "landi", "landy", "laude", "laudey", "laura", "lora",
        "lauda", "ling", "loda", "lode", "lund", "launda", "lounde", "laundey", "laundi", "loundi",
        "laundiya", "loundiya", "lulli", "maar", "maro", "marunga", "madarchod", "madarchodd",
        "madarchood", "madarchoot", "madarchut", "m.c.", "mc", "mamme", "mammey", "moot", "mut",
        "mootne", "mutne", "mooth", "muth", "nunni", "nunnu", "paaji", "paji", "pesaab", "pesab",
        "peshaab", "peshab", "pilla", "pillay", "pille", "pilley", "pisaab", "pisab", "pkmkb",
        "porkistan", "raand", "rand", "randi", "randy", "suar", "tatte", "tatti", "tatty", "ullu",
		"anuj","wagh","sajal","yadav","aditya","khandelwal","nepali","daksh","tyagi","apoorv","singh","gurgaon","tarang","agrawal",
    }
	trie := NewTrie()

    // Insert words into Trie
    for _, w := range words {
        trie.Insert(w)
    }


	return &Hub{
		clients:      make(map[*models.Client]bool),
		broadcast:    make(chan models.Message, 256),
		register:     make(chan *models.Client),
		unregister:   make(chan *models.Client),
		tri:			trie,
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

func (h *Hub) removeBad(content string) string {
    if h.tri == nil || content == "" {
        return content
    }
    var b strings.Builder
    token := make([]rune, 0, 32)
    flush := func() {
        if len(token) == 0 {
            return
        }
        word := string(token)
        norm := strings.ToLower(word)
        if h.tri.Search(norm) {
            b.WriteString("***")
        } else {
            b.WriteString(word)
        }
        token = token[:0]
    }
    for _, r := range content {
        if unicode.IsLetter(r) || unicode.IsDigit(r) {
            token = append(token, r)
        } else {
            flush()
            b.WriteRune(r) // preserve original punctuation/whitespace
        }
    }
    flush()
    return b.String()
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
	message.Content=h.removeBad(message.Content)
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