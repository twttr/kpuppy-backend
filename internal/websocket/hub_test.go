package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.rooms)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.rooms["room1"]
	hub.mu.RUnlock()
	assert.True(t, exists)

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	_, existsAfter := hub.rooms["room1"]
	hub.mu.RUnlock()
	assert.False(t, existsAfter)
}

func TestHub_MultipleClientsInRoom(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}
	client2 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	roomSize := len(hub.rooms["room1"])
	hub.mu.RUnlock()
	assert.Equal(t, 2, roomSize)

	hub.Unregister(client1)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	roomSizeAfter := len(hub.rooms["room1"])
	hub.mu.RUnlock()
	assert.Equal(t, 1, roomSizeAfter)
}

func TestHub_BroadcastToRoom(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}
	client2 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}
	clientOtherRoom := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room2",
	}

	hub.Register(client1)
	hub.Register(client2)
	hub.Register(clientOtherRoom)
	time.Sleep(10 * time.Millisecond)

	hub.BroadcastToRoom("room1", Message{
		Event: "test_event",
		Data:  map[string]string{"message": "hello"},
	})
	time.Sleep(10 * time.Millisecond)

	select {
	case msg := <-client1.Send:
		var received Message
		json.Unmarshal(msg, &received)
		assert.Equal(t, "test_event", received.Event)
	default:
		t.Error("client1 should have received message")
	}

	select {
	case msg := <-client2.Send:
		var received Message
		json.Unmarshal(msg, &received)
		assert.Equal(t, "test_event", received.Event)
	default:
		t.Error("client2 should have received message")
	}

	select {
	case <-clientOtherRoom.Send:
		t.Error("clientOtherRoom should NOT have received message")
	default:
	}
}

func TestHub_DifferentRooms(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}
	client2 := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room2",
	}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	numRooms := len(hub.rooms)
	hub.mu.RUnlock()
	assert.Equal(t, 2, numRooms)
}

func TestMessage_JSONSerialization(t *testing.T) {
	msg := Message{
		Event: "comment_created",
		Data: map[string]interface{}{
			"id":   "123",
			"text": "test comment",
		},
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var decoded Message
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "comment_created", decoded.Event)
}

func TestHub_BroadcastToEmptyRoom(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	hub.BroadcastToRoom("nonexistent", Message{
		Event: "test",
		Data:  nil,
	})
	time.Sleep(10 * time.Millisecond)
}

func TestHub_ClientUnregisterCleanup(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	roomExists := hub.rooms["room1"] != nil
	hub.mu.RUnlock()
	assert.True(t, roomExists)

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	_, roomStillExists := hub.rooms["room1"]
	hub.mu.RUnlock()
	assert.False(t, roomStillExists)
}

func TestHub_BroadcastDropsSlowClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	slowClient := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 1),
		RoomID: "room1",
	}

	hub.Register(slowClient)
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < 5; i++ {
		hub.BroadcastToRoom("room1", Message{
			Event: "test",
			Data:  map[string]int{"count": i},
		})
	}
	time.Sleep(50 * time.Millisecond)
}

func TestHub_UnregisterNonExistentClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)
}

func TestHub_UnregisterFromNonExistentRoom(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "nonexistent",
	}

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)
}
