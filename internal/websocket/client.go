package websocket

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pollz/websocket-server/internal/models"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

type Client struct {
	ID       string
	hub      models.Hub
	conn     *websocket.Conn
	send     chan models.Message
	userID   string
	username string
	joinedAt time.Time
}

func NewClient(hub models.Hub, conn *websocket.Conn, userID, username string) *Client {
	return &Client{
		ID:       uuid.New().String(),
		hub:      hub,
		conn:     conn,
		send:     make(chan models.Message, 256),
		userID:   userID,
		username: username,
		joinedAt: time.Now(),
	}
}

// GetClient returns the models.Client representation
func (c *Client) GetClient() *models.Client {
	return &models.Client{
		ID:       c.ID,
		Hub:      c.hub,
		Conn:     c.conn,
		Send:     c.send,
		UserID:   c.userID,
		Username: c.username,
		JoinedAt: c.joinedAt,
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c.GetClient())
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg models.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// Set default message type if not specified
		if msg.Type == "" {
			msg.Type = models.TextMessage
		}

		// Add user info to message
		msg.UserID = c.userID
		msg.Username = c.username
		msg.CreatedAt = time.Now()

		c.hub.Broadcast(msg)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Start begins the read and write pumps
func (c *Client) Start() {
	c.hub.Register(c.GetClient())
	go c.WritePump()
	go c.ReadPump()
}