//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"fission-basic/internal/biz"
	"fission-basic/internal/conf"
	"fission-basic/internal/data"
	"fission-basic/internal/pkg/feishu"
	"fission-basic/internal/pkg/queue"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/server"
	"fission-basic/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Bootstrap, *conf.Server, *conf.Data, *conf.Business, log.Logger) (*kratos.App, func(), error) {
	panic(
		wire.Build(
			server.ConsumerProviderSet,
			data.ConsumerProviderSet,
			biz.ConsumerProviderSet,
			service.ConsumerProviderSet,
			queue.ProviderSet,
			redis.ConsumerProviderSet,
			feishu.ProviderSet,
			newApp,
		),
	)
}
