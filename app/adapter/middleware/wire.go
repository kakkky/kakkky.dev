//go:build wireinject
// +build wireinject

package middleware

import "github.com/google/wire"

var Set = wire.NewSet(NewMiddleware)
