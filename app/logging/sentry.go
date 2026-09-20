package logging

import (
	"context"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"

	"github.com/kakkky/kakkky.dev/config"
)

func newSentryHandler(cfg *config.Config) (slog.Handler, func(), error) {
	if cfg.SentryDSN == "" {
		return nil, func() {}, nil
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.Env,
	}); err != nil {
		return nil, nil, err
	}
	cleanup := func() { sentry.Flush(2 * time.Second) }

	handler := sentryslog.Option{
		LogLevel: []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError},
	}.NewSentryHandler(context.Background())

	return handler, cleanup, nil
}
