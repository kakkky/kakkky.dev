//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"

	"github.com/kakkky/kakkky.dev/adapter/handler"
	"github.com/kakkky/kakkky.dev/adapter/middleware"
	"github.com/kakkky/kakkky.dev/adapter/repository"
	"github.com/kakkky/kakkky.dev/config"
	"github.com/kakkky/kakkky.dev/driver/db"
	"github.com/kakkky/kakkky.dev/driver/httpserver"
)

func InitServer(ctx context.Context) (*Server, func(), error) {
	wire.Build(
		config.Set,
		db.Set,
		repository.Set,
		handler.Set,
		middleware.Set,
		httpserver.Set,
		NewServer,
	)
	return nil, nil, nil
}
