package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/kakkky/kakkky.dev/config"
	"github.com/kakkky/kakkky.dev/logging"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() (err error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}
	logCleanup, err := logging.InitLogger(cfg)
	if err != nil {
		return err
	}
	defer logCleanup()
	defer func() {
		if err != nil {
			slog.Error("server failed", slog.Any("err", err))
		}
	}()

	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv, dbCleanup, err := InitServer(initCtx, cfg)
	if err != nil {
		return err
	}
	defer dbCleanup()

	return srv.Run(context.Background())
}
