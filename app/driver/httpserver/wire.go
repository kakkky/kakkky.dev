//go:build wireinject
// +build wireinject

package httpserver

import "github.com/google/wire"

var Set = wire.NewSet(NewMux, NewHTTPServer)
