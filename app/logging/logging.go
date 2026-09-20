package logging

import (
	"log/slog"
	"os"

	"github.com/kakkky/kakkky.dev/config"
)

func InitLogger(cfg *config.Config) (func(), error) {
	handlers := []slog.Handler{
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}

	sentryH, cleanup, err := newSentryHandler(cfg)
	if err != nil {
		return nil, err
	}
	if sentryH != nil {
		handlers = append(handlers, sentryH)
	}

	slog.SetDefault(slog.New(slog.NewMultiHandler(handlers...)))
	return cleanup, nil
}
