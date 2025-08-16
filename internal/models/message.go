package models

import (
	"time"
)

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"message"`
	Type      MessageType `json:"type"`
	UserID    string    `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageType string

const (
	TextMessage    MessageType = "text"
	StickerMessage MessageType = "sticker"
	SystemMessage  MessageType = "system"
	SuperChat      MessageType = "superchat"
)

type RecentMessagesResponse struct {
	Type     string    `json:"type"`
	Messages []Message `json:"messages"`
}