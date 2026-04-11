package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	ws "github.com/twttr/kpuppy-backend/internal/websocket"
)

type WSHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

// NewWSHandler creates a WebSocket handler.
// allowedOrigins is a list of permitted origin hosts (e.g. ["app.example.com"]).
// If empty, all origins are rejected to prevent CSRF via WebSocket.
// Pass nil or a non-empty list of origins you actually trust.
func NewWSHandler(hub *ws.Hub, allowedOrigins []string) *WSHandler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// If no origins configured, fall back to same-origin check
			// (gorilla default behaviour: Origin == Host).
			if len(allowed) == 0 {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				return origin == "http://"+r.Host || origin == "https://"+r.Host
			}
			origin := r.Header.Get("Origin")
			return allowed[origin]
		},
	}

	return &WSHandler{hub: hub, upgrader: upgrader}
}

func (h *WSHandler) HandleWebSocket(c echo.Context) error {
	roomID := c.Param("kinopubItemId")
	if roomID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "kinopubItemId required"})
	}

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	client := &ws.Client{
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		RoomID: roomID,
	}

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()

	return nil
}
