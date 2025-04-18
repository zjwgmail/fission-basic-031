//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"fission-basic/internal/conf"
	"fission-basic/internal/server"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(
		wire.Build(
			server.StaticProviderSet,
			// data.ConsumerProviderSet,
			// biz.ConsumerProviderSet,
			// service.ConsumerProviderSet,
			newApp,
		),
	)
}
