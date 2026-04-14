package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/service"
	"github.com/praxis-social/praxis/server/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for WebSocket connections.
		// In production, this should be restricted to the frontend origin.
		return true
	},
}

type WebSocketHandler struct {
	hub         *ws.Hub
	authService *service.AuthService
	logger      *zap.Logger
}

func NewWebSocketHandler(hub *ws.Hub, authService *service.AuthService, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authService: authService,
		logger:      logger,
	}
}

// HandleWS upgrades an HTTP connection to a WebSocket connection.
// The JWT token is passed via the "token" query parameter.
// GET /ws?token=<JWT>
func (h *WebSocketHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing token")
		return
	}

	claims, err := h.authService.ValidateAccessToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := ws.NewClient(h.hub, conn, claims.UserID)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
