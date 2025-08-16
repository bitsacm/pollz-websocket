package models

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID         string
	Hub        Hub
	Conn       *websocket.Conn
	Send       chan Message
	UserID     string
	Username   string
	JoinedAt   time.Time
}

type Hub interface {
	Register(client *Client)
	Unregister(client *Client)
	Broadcast(message Message)
}