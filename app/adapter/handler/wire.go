//go:build wireinject
// +build wireinject

package handler

import "github.com/google/wire"

var Set = wire.NewSet(NewHandler)
