package ws

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan []byte
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, 256),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub.
// It runs in its own goroutine per connection.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.hub.logger.Warn("websocket read error", zap.Error(err), zap.String("userID", c.userID))
			}
			break
		}

		// Parse incoming message for typing indicators
		var incoming struct {
			Type           string `json:"type"`
			ConversationID string `json:"conversationId"`
		}
		if err := json.Unmarshal(message, &incoming); err != nil {
			continue
		}

		if incoming.Type == "typing" && incoming.ConversationID != "" {
			c.hub.broadcast <- &WSMessage{
				Type: "typing",
				Data: map[string]string{
					"conversationId": incoming.ConversationID,
					"userId":         c.userID,
				},
				// Typing indicators are handled by the hub to all participants
				// except the sender. The service layer would need to resolve
				// participants, but for simplicity we broadcast to all and let
				// clients filter.
			}
		}
	}
}

// WritePump pumps messages from the send channel to the WebSocket connection.
// It runs in its own goroutine per connection.
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
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain queued messages into the current write.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
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
