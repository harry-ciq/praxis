package ws

import (
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

// WSMessage represents a message sent over WebSocket connections.
type WSMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	UserIDs []string    `json:"-"`
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients    map[string]map[*Client]bool // userID -> set of clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan *WSMessage
	mu         sync.RWMutex
	logger     *zap.Logger
}

// NewHub creates a new Hub instance.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *WSMessage, 256),
		logger:     logger,
	}
}

// Run starts the hub's event loop processing register, unregister, and broadcast
// channels. It should be started as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()
			h.logger.Debug("client registered", zap.String("userID", client.userID))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.userID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()
			h.logger.Debug("client unregistered", zap.String("userID", client.userID))

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				h.logger.Error("failed to marshal ws message", zap.Error(err))
				continue
			}

			if len(msg.UserIDs) > 0 {
				h.mu.RLock()
				for _, userID := range msg.UserIDs {
					if clients, ok := h.clients[userID]; ok {
						h.logger.Info("delivering ws message to connected user",
							zap.String("type", msg.Type),
							zap.String("targetUserID", userID),
							zap.Int("clientCount", len(clients)),
						)
						for client := range clients {
							select {
							case client.send <- data:
							default:
								h.mu.RUnlock()
								h.mu.Lock()
								delete(clients, client)
								close(client.send)
								if len(clients) == 0 {
									delete(h.clients, userID)
								}
								h.mu.Unlock()
								h.mu.RLock()
							}
						}
					}
				}
				h.mu.RUnlock()
			} else {
				// Broadcast to all connected clients
				h.mu.RLock()
				for _, clients := range h.clients {
					for client := range clients {
						select {
						case client.send <- data:
						default:
							// Skip slow clients during broadcast; they'll be cleaned up later
						}
					}
				}
				h.mu.RUnlock()
			}
		}
	}
}

// SendToUser sends a message to all connections of a specific user.
func (h *Hub) SendToUser(userID string, msg *WSMessage) {
	msg.UserIDs = []string{userID}
	h.broadcast <- msg
}

// SendToUsers sends a message to all connections of multiple users.
func (h *Hub) SendToUsers(userIDs []string, msg *WSMessage) {
	msg.UserIDs = userIDs
	h.broadcast <- msg
}

// Register queues a client for registration with the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister queues a client for removal from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// IsConnected returns whether a user has any active connections.
func (h *Hub) IsConnected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}
