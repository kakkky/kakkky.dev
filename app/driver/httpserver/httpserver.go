package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kakkky/kakkky.dev/config"
	"github.com/kakkky/scope"
)

type HTTPServer struct {
	srv *http.Server
}

func NewHTTPServer(cfg *config.Config, mux http.Handler) *HTTPServer {
	return &HTTPServer{
		srv: &http.Server{
			Addr:              net.JoinHostPort("", cfg.HTTPPort),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 16, // 64 KB
			ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		},
	}
}

func (s *HTTPServer) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	s.srv.BaseContext = func(net.Listener) context.Context { return ctx }

	return scope.Run(ctx, func(sc *scope.Scope) error {
		sc.Go(func(ctx context.Context) error {
			slog.Info("http server is running", "addr", s.srv.Addr)
			if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		})
		sc.Go(func(ctx context.Context) error {
			<-ctx.Done()
			slog.Info("http server is shutting down")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := s.srv.Shutdown(shutdownCtx); err != nil {
				slog.Error("http server shutdown error", "error", err)
				return err
			}
			slog.Info("http server was shut down gracefully")
			return nil
		})
		return nil
	})
}
