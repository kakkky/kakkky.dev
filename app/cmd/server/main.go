package main

import (
	"context"
	"log/slog"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv, dbCleanup, err := InitServer(initCtx)
	if err != nil {
		return err
	}
	defer dbCleanup()

	return srv.Run(context.Background())
}
