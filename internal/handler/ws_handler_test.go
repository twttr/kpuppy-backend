package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ws "github.com/twttr/kpuppy-backend/internal/websocket"
)

func TestNewWSHandler(t *testing.T) {
	hub := ws.NewHub()
	handler := NewWSHandler(hub, nil)

	assert.NotNil(t, handler)
	assert.Equal(t, hub, handler.hub)
}

func TestWSHandler_HandleWebSocket_MissingRoomID(t *testing.T) {
	hub := ws.NewHub()
	handler := NewWSHandler(hub, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws/content/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("")

	err := handler.HandleWebSocket(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestWSHandler_HandleWebSocket_Success(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	// Pass empty allowed origins — CheckOrigin falls back to same-origin (Origin == Host)
	// In tests, no Origin header is sent by the default dialer, so it passes.
	handler := NewWSHandler(hub, nil)

	e := echo.New()
	e.GET("/ws/content/:kinopubItemId", handler.HandleWebSocket)

	server := httptest.NewServer(e)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/content/123"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
}

func TestWSHandler_HandleWebSocket_UpgradeFailure(t *testing.T) {
	hub := ws.NewHub()
	handler := NewWSHandler(hub, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws/content/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("123")

	err := handler.HandleWebSocket(c)

	assert.Error(t, err)
}
