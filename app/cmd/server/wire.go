//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"

	"github.com/kakkky/kakkky.dev/adapter"
	"github.com/kakkky/kakkky.dev/config"
	"github.com/kakkky/kakkky.dev/driver"
)

func InitServer(ctx context.Context) (*Server, func(), error) {
	wire.Build(
		config.Set,
		driver.Set,
		adapter.Set,
		NewServer,
	)
	return nil, nil, nil
}
