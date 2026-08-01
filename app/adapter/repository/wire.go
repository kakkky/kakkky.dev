//go:build wireinject
// +build wireinject

package repository

import (
	"github.com/google/wire"
	"github.com/kakkky/kakkky.dev/domain"
)

var Set = wire.NewSet(
	NewRepository,
	wire.Bind(new(domain.Repository), new(*Repository)),
)
