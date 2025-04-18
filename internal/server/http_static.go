package server

import (
	"fission-basic/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/pprof"
)

// NewHTTPServer new an HTTP server.
func NewHTTPStaticServer(c *conf.Server,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(logger),
		),
	}

	if c.StaticHttp.Network != "" {
		opts = append(opts, http.Network(c.StaticHttp.Network))
	}
	if c.StaticHttp.Addr != "" {
		opts = append(opts, http.Address(c.StaticHttp.Addr))
	}
	if c.StaticHttp.Timeout != nil {
		opts = append(opts, http.Timeout(c.StaticHttp.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)
	srv.HandleFunc("/events/mlbb25031static/ping", func(w http.ResponseWriter, r *http.Request) {
		return
	})
	srv.Handle("/debug/pprof/", pprof.NewHandler())
	return srv
}
