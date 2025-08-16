package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pollz/websocket-server/internal/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(msg models.Message) error {
	query := `
		INSERT INTO chat_messages (id, content, type, user_id, username, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := r.db.Exec(query, msg.ID, msg.Content, msg.Type, msg.UserID, msg.Username, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	
	return nil
}

func (r *MessageRepository) GetRecent(limit int) ([]models.Message, error) {
	query := `
		SELECT id, content, type, COALESCE(user_id, ''), COALESCE(username, ''), created_at 
		FROM chat_messages 
		ORDER BY created_at DESC 
		LIMIT $1`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()
	
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.Content, &msg.Type, &msg.UserID, &msg.Username, &msg.CreatedAt)
		if err != nil {
			continue
		}
		messages = append([]models.Message{msg}, messages...) // Prepend to reverse order
	}
	
	return messages, nil
}

func (r *MessageRepository) Search(query string, limit int) ([]models.Message, error) {
	sqlQuery := `
		SELECT id, content, type, COALESCE(user_id, ''), COALESCE(username, ''), created_at 
		FROM chat_messages 
		WHERE content ILIKE $1
		ORDER BY created_at DESC 
		LIMIT $2`
	
	rows, err := r.db.Query(sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	defer rows.Close()
	
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.Content, &msg.Type, &msg.UserID, &msg.Username, &msg.CreatedAt)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func (r *MessageRepository) GetByDateRange(start, end time.Time) ([]models.Message, error) {
	query := `
		SELECT id, content, type, COALESCE(user_id, ''), COALESCE(username, ''), created_at 
		FROM chat_messages 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at ASC`
	
	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by date range: %w", err)
	}
	defer rows.Close()
	
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.Content, &msg.Type, &msg.UserID, &msg.Username, &msg.CreatedAt)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func (r *MessageRepository) DeleteOlderThan(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.Exec("DELETE FROM chat_messages WHERE created_at < $1", cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete old messages: %w", err)
	}
	return nil
}