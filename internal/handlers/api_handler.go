package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pollz/websocket-server/internal/models"
)

type APIHandler struct {
	hub interface {
		SearchMessages(query string, limit int) ([]models.Message, error)
		GetMessagesByDateRange(start, end time.Time) ([]models.Message, error)
		GetConnectedClients() int
	}
}

func NewAPIHandler(hub interface {
	SearchMessages(query string, limit int) ([]models.Message, error)
	GetMessagesByDateRange(start, end time.Time) ([]models.Message, error)
	GetConnectedClients() int
}) *APIHandler {
	return &APIHandler{
		hub: hub,
	}
}

// SearchMessages handles GET /api/messages/search?q=query&limit=50
func (h *APIHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	messages, err := h.hub.SearchMessages(query, limit)
	if err != nil {
		h.sendError(w, "Failed to search messages", http.StatusInternalServerError)
		return
	}
	
	h.sendJSON(w, messages)
}

// GetMessagesByDate handles GET /api/messages/date?start=2024-01-01&end=2024-01-31
func (h *APIHandler) GetMessagesByDate(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		h.sendError(w, "Invalid start date format", http.StatusBadRequest)
		return
	}
	
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		h.sendError(w, "Invalid end date format", http.StatusBadRequest)
		return
	}
	
	messages, err := h.hub.GetMessagesByDateRange(start, end.Add(24*time.Hour))
	if err != nil {
		h.sendError(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}
	
	h.sendJSON(w, messages)
}

// GetStats handles GET /api/stats
func (h *APIHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"connected_clients": h.hub.GetConnectedClients(),
		"server_time":      time.Now(),
	}
	
	h.sendJSON(w, stats)
}

// Health check endpoint
func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *APIHandler) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *APIHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}