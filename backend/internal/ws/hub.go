package ws

import (
	"encoding/json"
	"sync"

	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to rooms
type Hub struct {
	// Room-keyed client registry
	rooms    map[string]map[*Client]bool
	mu       sync.RWMutex

	// Channels
	Register   chan *Client
	Unregister chan *Client

	// Game session manager
	GameManager *GameManager

	// Message handler callback (set by the application)
	OnMessage func(client *Client, msg WSMessage)
}

// NewHub creates and returns a new Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client, 256),
		Unregister: make(chan *Client, 256),
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.rooms[client.RoomCode] == nil {
				h.rooms[client.RoomCode] = make(map[*Client]bool)
			}
			h.rooms[client.RoomCode][client] = true
			h.mu.Unlock()

			logrus.Infof("Client %s joined room %s", client.DisplayName, client.RoomCode)

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.RoomCode]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.rooms, client.RoomCode)
					}
				}
			}
			h.mu.Unlock()

			logrus.Infof("Client %s left room %s", client.DisplayName, client.RoomCode)

			// Notify game manager of disconnect
			if h.GameManager != nil {
				h.GameManager.HandleDisconnect(client)
			}
		}
	}
}

// ProcessMessage routes an incoming client message
func (h *Hub) ProcessMessage(client *Client, msg WSMessage) {
	if h.OnMessage != nil {
		h.OnMessage(client, msg)
	}
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(roomCode string, msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		logrus.Errorf("Failed to marshal broadcast: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomCode]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				logrus.Warnf("Broadcast: client %s buffer full", client.ID)
			}
		}
	}
}

// SendToClient sends a message to a specific client in a room by seat
func (h *Hub) SendToSeat(roomCode string, seat int, msg WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomCode]; ok {
		for client := range clients {
			if client.Seat == seat {
				client.SendMessage(msg)
				return
			}
		}
	}
}

// GetRoomClients returns all clients in a room
func (h *Hub) GetRoomClients(roomCode string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if room, ok := h.rooms[roomCode]; ok {
		for client := range room {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetClientByUserID finds a client in a room by user ID
func (h *Hub) GetClientByUserID(roomCode string, userID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomCode]; ok {
		for client := range clients {
			if client.UserID.String() == userID {
				return client
			}
		}
	}
	return nil
}

// IsConnected checks if a user has an active connection in the room
func (h *Hub) IsConnected(roomCode string, userID string) bool {
	return h.GetClientByUserID(roomCode, userID) != nil
}

// RoomClientCount returns the number of connected clients in a room
func (h *Hub) RoomClientCount(roomCode string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomCode]; ok {
		return len(clients)
	}
	return 0
}
