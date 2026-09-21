package logging

import (
	"log/slog"
	"os"

	"github.com/kakkky/kakkky.dev/sentry"
)

func InitLogger() {
	handlers := []slog.Handler{
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}
	if h := sentry.SlogHandler(); h != nil {
		handlers = append(handlers, h)
	}
	slog.SetDefault(slog.New(slog.NewMultiHandler(handlers...)))
}
