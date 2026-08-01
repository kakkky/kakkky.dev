package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	srv, err := InitServer()
	if err != nil {
		return err
	}
	return srv.Run(context.Background())
}
