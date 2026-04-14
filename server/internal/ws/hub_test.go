package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockConn is not needed for hub tests since we test at the channel level.

func newTestHub() *Hub {
	h := NewHub(zap.NewNop())
	go h.Run()
	return h
}

func TestHub_RegisterUnregister(t *testing.T) {
	h := newTestHub()

	client := &Client{
		hub:    h,
		userID: "user-1",
		send:   make(chan []byte, 256),
	}

	// Register
	h.register <- client
	time.Sleep(50 * time.Millisecond)

	assert.True(t, h.IsConnected("user-1"))
	assert.False(t, h.IsConnected("user-2"))

	// Unregister
	h.unregister <- client
	time.Sleep(50 * time.Millisecond)

	assert.False(t, h.IsConnected("user-1"))
}

func TestHub_MultipleClientsPerUser(t *testing.T) {
	h := newTestHub()

	client1 := &Client{
		hub:    h,
		userID: "user-1",
		send:   make(chan []byte, 256),
	}
	client2 := &Client{
		hub:    h,
		userID: "user-1",
		send:   make(chan []byte, 256),
	}

	h.register <- client1
	h.register <- client2
	time.Sleep(50 * time.Millisecond)

	assert.True(t, h.IsConnected("user-1"))

	// Unregister one client, user should still be connected
	h.unregister <- client1
	time.Sleep(50 * time.Millisecond)

	assert.True(t, h.IsConnected("user-1"))

	// Unregister second client
	h.unregister <- client2
	time.Sleep(50 * time.Millisecond)

	assert.False(t, h.IsConnected("user-1"))
}

func TestHub_SendToUser(t *testing.T) {
	h := newTestHub()

	client1 := &Client{
		hub:    h,
		userID: "user-1",
		send:   make(chan []byte, 256),
	}
	client2 := &Client{
		hub:    h,
		userID: "user-2",
		send:   make(chan []byte, 256),
	}

	h.register <- client1
	h.register <- client2
	time.Sleep(50 * time.Millisecond)

	// Send to user-1 only
	h.SendToUser("user-1", &WSMessage{
		Type: "message",
		Data: map[string]string{"text": "hello user-1"},
	})

	// user-1 should receive the message
	select {
	case msg := <-client1.send:
		var wsMsg WSMessage
		err := json.Unmarshal(msg, &wsMsg)
		require.NoError(t, err)
		assert.Equal(t, "message", wsMsg.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("user-1 did not receive the message")
	}

	// user-2 should NOT receive the message
	select {
	case <-client2.send:
		t.Fatal("user-2 should not have received the message")
	case <-time.After(100 * time.Millisecond):
		// Expected: no message
	}
}

func TestHub_SendToUsers(t *testing.T) {
	h := newTestHub()

	client1 := &Client{
		hub:    h,
		userID: "user-1",
		send:   make(chan []byte, 256),
	}
	client2 := &Client{
		hub:    h,
		userID: "user-2",
		send:   make(chan []byte, 256),
	}
	client3 := &Client{
		hub:    h,
		userID: "user-3",
		send:   make(chan []byte, 256),
	}

	h.register <- client1
	h.register <- client2
	h.register <- client3
	time.Sleep(50 * time.Millisecond)

	// Send to user-1 and user-2
	h.SendToUsers([]string{"user-1", "user-2"}, &WSMessage{
		Type: "notification",
		Data: map[string]string{"text": "hello both"},
	})

	// user-1 should receive the message
	select {
	case msg := <-client1.send:
		var wsMsg WSMessage
		err := json.Unmarshal(msg, &wsMsg)
		require.NoError(t, err)
		assert.Equal(t, "notification", wsMsg.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("user-1 did not receive the message")
	}

	// user-2 should receive the message
	select {
	case msg := <-client2.send:
		var wsMsg WSMessage
		err := json.Unmarshal(msg, &wsMsg)
		require.NoError(t, err)
		assert.Equal(t, "notification", wsMsg.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("user-2 did not receive the message")
	}

	// user-3 should NOT receive the message
	select {
	case <-client3.send:
		t.Fatal("user-3 should not have received the message")
	case <-time.After(100 * time.Millisecond):
		// Expected: no message
	}
}

func TestHub_SendToDisconnectedUser(t *testing.T) {
	h := newTestHub()

	// Send to a user that doesn't exist - should not panic
	h.SendToUser("nonexistent", &WSMessage{
		Type: "message",
		Data: map[string]string{"text": "hello"},
	})

	// Give the hub time to process
	time.Sleep(50 * time.Millisecond)
	// If we get here without panicking, the test passes
}
