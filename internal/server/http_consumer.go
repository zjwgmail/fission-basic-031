package server

import (
	"context"
	v1f "fission-basic/api/fission/v1"
	"fission-basic/internal/conf"
	"fission-basic/internal/service"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/pprof"
)

// NewHTTPServer new an HTTP server.
func NewHTTPConsumerServer(
	c *conf.Server,
	logger log.Logger,
	initService *service.InitService,

) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(logger),
		),
	}

	if c.ConsumerHttp.Network != "" {
		opts = append(opts, http.Network(c.ConsumerHttp.Network))
	}
	if c.ConsumerHttp.Addr != "" {
		opts = append(opts, http.Address(c.ConsumerHttp.Addr))
	}
	if c.ConsumerHttp.Timeout != nil {
		opts = append(opts, http.Timeout(c.ConsumerHttp.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)

	srv.HandleFunc("/events/mlbb25031consumer/ping", func(w http.ResponseWriter, r *http.Request) {
		return
	})
	r := srv.Route("/")
	r.POST("/events/mlbb25031consumer/activity/sql-query", func(ctx http.Context) error {
		if ctx.Request().Header.Get("Content-Type") == "" {
			ctx.Request().Header.Set("Content-Type", "application/json")
		}

		var in v1f.QuerySqlRequest
		if err := ctx.Bind(&in); err != nil {
			fmt.Println("bind", err)
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			fmt.Println("bindQuery", err)
			return err
		}

		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return initService.QuerySql1(ctx, req.(*v1f.QuerySqlRequest))
		})

		reply, err := h(ctx, &in)
		if err != nil {
			return err
		}

		_, err = ctx.Response().Write(reply.([]byte))

		return err
	})
	srv.Handle("/debug/pprof/", pprof.NewHandler())
	return srv
}
