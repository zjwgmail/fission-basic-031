package cron

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel/trace"

	"fission-basic/contrib/internel"
)

type Server struct {
	*cron.Cron
	Ctx context.Context
	sig chan struct{}
}

func NewServer() *Server {
	c := cron.New()
	return &Server{Cron: c, sig: make(chan struct{})}
}

func (c *Server) Start(ctx context.Context) error {
	c.Ctx = ctx
	c.Cron.Run()
	<-c.sig
	return nil
}

func (c *Server) Stop(ctx context.Context) error {
	c.Cron.Stop()
	c.sig <- struct{}{}
	return nil
}

func (c *Server) AddFunc(spec string, f func(ctx context.Context) error) (cron.EntryID, error) {
	return c.Cron.AddFunc(spec, func() {
		ctx := c.Ctx

		tracer := tracing.NewTracer(trace.SpanKindServer)
		ctx, span := tracer.Start(ctx, "consumer", internel.NewTextMap())
		defer func() {
			tracer.End(ctx, span, "", nil)
		}()

		_ = f(ctx)
	})
}
