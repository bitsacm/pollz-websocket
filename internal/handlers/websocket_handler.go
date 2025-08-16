package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pollz/websocket-server/internal/models"
	ws "github.com/pollz/websocket-server/internal/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type connectionInfo struct {
	count     int
	lastReset time.Time
}

type WebSocketHandler struct {
	hub         models.Hub
	connections map[string]*connectionInfo
	mutex       sync.RWMutex
}

func NewWebSocketHandler(hub models.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		connections: make(map[string]*connectionInfo),
	}
}

func (h *WebSocketHandler) getClientIP(r *http.Request) string {
	// Get real IP from headers (for reverse proxy setups)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func (h *WebSocketHandler) checkRateLimit(ip string) bool {
	const maxConnectionsPerMinute = 10
	const resetInterval = time.Minute

	h.mutex.Lock()
	defer h.mutex.Unlock()

	now := time.Now()
	info, exists := h.connections[ip]

	if !exists {
		h.connections[ip] = &connectionInfo{
			count:     1,
			lastReset: now,
		}
		return true
	}

	// Reset counter if enough time has passed
	if now.Sub(info.lastReset) >= resetInterval {
		info.count = 1
		info.lastReset = now
		return true
	}

	// Check if under limit
	if info.count < maxConnectionsPerMinute {
		info.count++
		return true
	}

	return false
}

func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Get client IP for rate limiting
	clientIP := h.getClientIP(r)
	
	// Check rate limit
	if !h.checkRateLimit(clientIP) {
		log.Printf("Rate limit exceeded for IP: %s", clientIP)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Extract user info from query params or headers (implement your auth logic here)
	userID := r.URL.Query().Get("user_id")
	username := r.URL.Query().Get("username")
	
	// If no username provided, use anonymous
	if username == "" {
		username = "Anonymous"
	}

	// Create new client
	client := ws.NewClient(h.hub, conn, userID, username)
	
	// Start client
	client.Start()
}