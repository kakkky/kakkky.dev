//go:build wireinject
// +build wireinject

package driver

import (
	"github.com/google/wire"

	"github.com/kakkky/kakkky.dev/driver/db"
	"github.com/kakkky/kakkky.dev/driver/httpserver"
)

var Set = wire.NewSet(
	db.NewDB,
	httpserver.NewMux,
	httpserver.NewHTTPServer,
)
