//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/kakkky/kakkky.dev/adapter/handler"
	"github.com/kakkky/kakkky.dev/adapter/middleware"
	"github.com/kakkky/kakkky.dev/config"
	"github.com/kakkky/kakkky.dev/driver/httpserver"
)

func InitServer() (*Server, error) {
	wire.Build(
		config.Set,
		handler.Set,
		middleware.Set,
		httpserver.Set,
		NewServer,
	)
	return nil, nil
}
