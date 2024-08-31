package ws

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pietjan/events"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

type Event interface {
	On(event string, handler func(event *events.Event)) func()
}

func Handler(event Event) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("ws", "client", "upgrade")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("ws", "upgrade", err)
			return
		}
		defer conn.Close()

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		slog.Debug("ws", "client", "connect")

		unregister := event.On("message", func(e *events.Event) {
			message, ok := e.Data.(string)
			if !ok {
				slog.Error("invalid message data type", "data", e.Data)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				slog.Debug("failed to write message", "error", err)
				cancel()
				return
			}
		})
		defer unregister()

		<-ctx.Done()
		slog.Debug("ws", "client", "disconnect")
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}
}
