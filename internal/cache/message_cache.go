package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pollz/websocket-server/internal/models"
	"github.com/redis/go-redis/v9"
)

type MessageCache struct {
	client *redis.Client
	key    string
	maxLen int64
}

func NewMessageCache(client *redis.Client) *MessageCache {
	return &MessageCache{
		client: client,
		key:    "chat_messages",
		maxLen: 100,
	}
}

func (c *MessageCache) Push(msg models.Message) error {
	ctx := context.Background()
	
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Push to Redis list
	if err := c.client.LPush(ctx, c.key, data).Err(); err != nil {
		return fmt.Errorf("failed to push message to cache: %w", err)
	}
	
	// Trim to maintain max length
	if err := c.client.LTrim(ctx, c.key, 0, c.maxLen-1).Err(); err != nil {
		return fmt.Errorf("failed to trim cache: %w", err)
	}
	
	return nil
}

func (c *MessageCache) GetRecent(limit int64) ([]models.Message, error) {
	ctx := context.Background()
	
	if limit > c.maxLen {
		limit = c.maxLen
	}
	
	data, err := c.client.LRange(ctx, c.key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get messages from cache: %w", err)
	}
	
	messages := make([]models.Message, 0, len(data))
	for i := len(data) - 1; i >= 0; i-- {
		var msg models.Message
		if err := json.Unmarshal([]byte(data[i]), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func (c *MessageCache) Clear() error {
	ctx := context.Background()
	return c.client.Del(ctx, c.key).Err()
}

func (c *MessageCache) Populate(messages []models.Message) error {
	ctx := context.Background()
	
	// Clear existing cache
	if err := c.Clear(); err != nil {
		return err
	}
	
	// Add messages to cache
	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		c.client.RPush(ctx, c.key, data)
	}
	
	// Trim to max length
	return c.client.LTrim(ctx, c.key, 0, c.maxLen-1).Err()
}