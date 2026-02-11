package handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
)

type WSHandler struct {
	upgrader *websocket.HertzUpgrader
}

func NewWSHandler() *WSHandler {
	return &WSHandler{
		upgrader: &websocket.HertzUpgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(ctx *app.RequestContext) bool {
				return true
			},
		},
	}
}

type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

func (h *WSHandler) HandleWebSocket(ctx context.Context, c *app.RequestContext) {
	err := h.upgrader.Upgrade(c, func(conn *websocket.Conn) {
		defer conn.Close()

		wsCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		for {
			select {
			case <-wsCtx.Done():
				return
			default:
				conn.SetReadDeadline(time.Now().Add(60 * time.Second))
				_, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket read error: %v", err)
					}
					return
				}

				var msg WSMessage
				if err := json.Unmarshal(message, &msg); err != nil {
					log.Printf("JSON unmarshal error: %v", err)
					conn.WriteJSON(WSMessage{Type: "error", Data: json.RawMessage(`"invalid message format"`)})
					continue
				}

				h.handleMessage(conn, &msg)
			}
		}
	})

	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
}

func (h *WSHandler) handleMessage(conn *websocket.Conn, msg *WSMessage) {
	switch msg.Type {
	case "ping":
		conn.WriteJSON(WSMessage{Type: "pong"})
	default:
		conn.WriteJSON(WSMessage{Type: "error", Data: json.RawMessage(`"unknown message type"`)})
	}
}
