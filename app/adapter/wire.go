//go:build wireinject
// +build wireinject

package adapter

import (
	"github.com/google/wire"

	"github.com/kakkky/kakkky.dev/adapter/handler"
	"github.com/kakkky/kakkky.dev/adapter/middleware"
	"github.com/kakkky/kakkky.dev/adapter/repository"
	"github.com/kakkky/kakkky.dev/domain"
)

var Set = wire.NewSet(
	handler.NewHandler,
	middleware.NewMiddleware,
	repository.NewRepository,
	wire.Bind(new(domain.Repository), new(*repository.Repository)),
)
