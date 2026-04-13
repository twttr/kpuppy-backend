package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type Message struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	RoomID string
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	broadcast  chan *RoomMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	// done is closed to signal Hub.Run() to stop
	done chan struct{}
}

type RoomMessage struct {
	RoomID  string
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan *RoomMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

// Stop signals Hub.Run() to exit, preventing goroutine leak on shutdown.
func (h *Hub) Stop() {
	close(h.done)
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			// Graceful shutdown: close all client channels
			h.mu.Lock()
			for _, room := range h.rooms {
				for client := range room {
					close(client.Send)
					delete(room, client)
				}
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.RoomID] == nil {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.RoomID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)
					if len(room) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()

		case roomMsg := <-h.broadcast:
			// Fix #2: collect slow clients under RLock, then remove under Lock
			// to avoid the RLock→Lock upgrade (deadlock/panic) pattern.
			h.mu.RLock()
			room, ok := h.rooms[roomMsg.RoomID]
			if !ok {
				h.mu.RUnlock()
				continue
			}
			var slowClients []*Client
			for client := range room {
				select {
				case client.Send <- roomMsg.Message:
				default:
					slowClients = append(slowClients, client)
				}
			}
			h.mu.RUnlock()

			if len(slowClients) > 0 {
				h.mu.Lock()
				for _, client := range slowClients {
					if room, ok := h.rooms[roomMsg.RoomID]; ok {
						if _, ok := room[client]; ok {
							delete(room, client)
							close(client.Send)
						}
					}
				}
				h.mu.Unlock()
			}
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToRoom(roomID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.broadcast <- &RoomMessage{
		RoomID:  roomID,
		Message: data,
	}
}

// Fix #3: WritePump with write deadline and ping/pong support.
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
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

// Fix #3: ReadPump with pong handler and read deadline.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
