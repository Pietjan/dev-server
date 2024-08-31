package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func New(debug bool) *slog.Logger {
	level := slog.LevelInfo

	if debug {
		level = slog.LevelDebug
	}

	return slog.New(&Handler{
		tint.NewHandler(os.Stdout, &tint.Options{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}),
	})
}

type Handler struct {
	slog.Handler
}

func (h *Handler) Handle(ctx context.Context, e slog.Record) error {
	return h.Handler.Handle(ctx, e)
}
