//go:build wireinject
// +build wireinject

package config

import (
	"github.com/google/wire"
)

var Set = wire.NewSet(
	NewConfig,
)
