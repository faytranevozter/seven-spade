package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 4096
)

// Client represents a single WebSocket connection
type Client struct {
	ID       string
	UserID   uuid.UUID
	DisplayName string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	RoomCode string
	Seat     int
	mu       sync.Mutex
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub, userID uuid.UUID, displayName, roomCode string) *Client {
	return &Client{
		ID:          uuid.NewString(),
		UserID:      userID,
		DisplayName: displayName,
		Conn:        conn,
		Hub:         hub,
		Send:        make(chan []byte, 256),
		RoomCode:    roomCode,
		Seat:        -1,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMsgSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logrus.Warnf("WebSocket unexpected close: %v", err)
			}
			break
		}

		// Parse the message
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.SendMessage(WSMessage{Type: MsgTypeError, Payload: PayloadError{Message: "invalid message format"}})
			continue
		}

		// Route to hub for processing
		c.Hub.ProcessMessage(c, msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a typed message to this client
func (c *Client) SendMessage(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		logrus.Errorf("Failed to marshal message: %v", err)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case c.Send <- data:
	default:
		logrus.Warnf("Client %s send buffer full, dropping message", c.ID)
	}
}
