//go:build wireinject
// +build wireinject

package db

import (
	"github.com/google/wire"
)

var Set = wire.NewSet(
	NewDB,
)
