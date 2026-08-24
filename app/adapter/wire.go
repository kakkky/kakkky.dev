//go:build wireinject
// +build wireinject

package adapter

import (
	"github.com/google/wire"

	"github.com/kakkky/kakkky.dev/adapter/client"
	"github.com/kakkky/kakkky.dev/adapter/handler"
	"github.com/kakkky/kakkky.dev/adapter/middleware"
	"github.com/kakkky/kakkky.dev/adapter/query"
	"github.com/kakkky/kakkky.dev/adapter/repository"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

var Set = wire.NewSet(
	client.NewOGPFetcher,
	handler.NewHandler,
	middleware.NewMiddleware,
	repository.NewRepository,
	query.NewQueryService,
	usecase.NewUseCase,
	wire.Bind(new(domain.Repository), new(*repository.Repository)),
	wire.Bind(new(domain.QueryService), new(*query.QueryService)),
)
