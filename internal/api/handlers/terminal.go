package handlers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/internal/services/terminal"
	"github.com/deep-agent/sandbox/pkg/logger"
	"github.com/hertz-contrib/websocket"
)

type TerminalHandler struct {
	workspace string
	upgrader  *websocket.HertzUpgrader
}

func NewTerminalHandler(workspace string) *TerminalHandler {
	return &TerminalHandler{
		workspace: workspace,
		upgrader: &websocket.HertzUpgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(ctx *app.RequestContext) bool {
				return true
			},
		},
	}
}

type wsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (h *TerminalHandler) HandleWebSocket(ctx context.Context, c *app.RequestContext) {
	err := h.upgrader.Upgrade(c, func(conn *websocket.Conn) {
		defer conn.Close()

		term, err := terminal.New("/bin/bash", h.workspace, nil)
		if err != nil {
			logger.Printf("Failed to create terminal: %v", err)
			if writeErr := conn.WriteJSON(map[string]string{"type": "error", "data": err.Error()}); writeErr != nil {
				logger.Printf("WebSocket write error: %v", writeErr)
			}
			return
		}
		defer term.Close()

		var writeMu sync.Mutex
		wsCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		go h.readFromTerminal(wsCtx, conn, term, &writeMu)

		h.readFromWebSocket(wsCtx, conn, term, cancel, &writeMu)

		if err := term.Wait(); err != nil {
			logger.Printf("Terminal wait error: %v", err)
		}
	})

	if err != nil {
		logger.Printf("WebSocket upgrade error: %v", err)
		return
	}
}

func (h *TerminalHandler) readFromTerminal(ctx context.Context, conn *websocket.Conn, term *terminal.Terminal, writeMu *sync.Mutex) {
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := term.Read(buf)
			if err != nil {
				logger.Printf("Terminal read error: %v", err)
				return
			}
			if n > 0 {
				msg := map[string]string{
					"type": "output",
					"data": string(buf[:n]),
				}
				writeMu.Lock()
				err := conn.WriteJSON(msg)
				writeMu.Unlock()
				if err != nil {
					logger.Printf("WebSocket write error: %v", err)
					return
				}
			}
		}
	}
}

func (h *TerminalHandler) readFromWebSocket(ctx context.Context, conn *websocket.Conn, term *terminal.Terminal, cancel context.CancelFunc, writeMu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
				logger.Printf("WebSocket set read deadline error: %v", err)
				cancel()
				return
			}
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Printf("WebSocket read error: %v", err)
				}
				cancel()
				return
			}

			var msg wsMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				logger.Printf("JSON unmarshal error: %v", err)
				continue
			}

			switch msg.Type {
			case "input":
				var input string
				if err := json.Unmarshal(msg.Data, &input); err != nil {
					logger.Printf("Input unmarshal error: %v", err)
					continue
				}
				if _, err := term.Write([]byte(input)); err != nil {
					logger.Printf("Terminal write error: %v", err)
					cancel()
					return
				}
			case "resize":
				var size terminal.Size
				if err := json.Unmarshal(msg.Data, &size); err != nil {
					logger.Printf("Resize unmarshal error: %v", err)
					continue
				}
				if err := term.Resize(size); err != nil {
					logger.Printf("Terminal resize error: %v", err)
				}
			case "ping":
				writeMu.Lock()
				err := conn.WriteJSON(map[string]string{"type": "pong"})
				writeMu.Unlock()
				if err != nil {
					logger.Printf("WebSocket write error: %v", err)
					cancel()
					return
				}
			}
		}
	}
}
