// Package sentry は Sentry client の init と、logging / errors から使う通知・handler を集約する。
// process 内での Sentry 接続の唯一の owner。
package sentry

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	sentrysdk "github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"

	"github.com/kakkky/kakkky.dev/config"
)

func Init(cfg *config.Config) (cleanup func(), err error) {
	if cfg.SentryDSN == "" {
		return func() {}, nil
	}
	if err := sentrysdk.Init(sentrysdk.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.Env,
	}); err != nil {
		return nil, err
	}
	return func() { sentrysdk.Flush(2 * time.Second) }, nil
}

func NewContext(ctx context.Context, r *http.Request) context.Context {
	if sentrysdk.CurrentHub().Client() == nil {
		return ctx
	}
	hub := sentrysdk.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentrysdk.CurrentHub().Clone()
		ctx = sentrysdk.SetHubOnContext(ctx, hub)
	}
	hub.Scope().SetRequest(r)
	return ctx
}

func Notify(ctx context.Context, err error) {
	if sentrysdk.CurrentHub().Client() == nil {
		return
	}
	hub := sentrysdk.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentrysdk.CurrentHub()
	}
	hub.CaptureException(err)
}

func Recover(ctx context.Context, recovered any) {
	if recovered == nil {
		return
	}
	if sentrysdk.CurrentHub().Client() == nil {
		return
	}
	hub := sentrysdk.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentrysdk.CurrentHub()
	}
	hub.RecoverWithContext(ctx, recovered)
}

func SlogHandler() slog.Handler {
	if sentrysdk.CurrentHub().Client() == nil {
		return nil
	}
	return sentryslog.Option{
		LogLevel: []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError},
	}.NewSentryHandler(context.Background())
}
