package server

import (
	"context"
	v1f "fission-basic/api/fission/v1"
	"fission-basic/internal/conf"
	"fission-basic/internal/service"
	"fmt"
	"mime/multipart"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/pprof"
)

// NewHTTPServer new an HTTP server.
func NewHTTPJobServer(c *conf.Server,
	logger log.Logger,
	helpCode *service.HelpCodeService,
	initService *service.InitService,
	uploadService *service.UploadService,

) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(logger),
		),
	}

	if c.JobHttp.Network != "" {
		opts = append(opts, http.Network(c.JobHttp.Network))
	}
	if c.JobHttp.Addr != "" {
		opts = append(opts, http.Address(c.JobHttp.Addr))
	}
	if c.JobHttp.Timeout != nil {
		opts = append(opts, http.Timeout(c.JobHttp.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)
	r := srv.Route("/")
	r.POST("/events/mlbb25031job/activity/sql-query", func(ctx http.Context) error {
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

	r.POST("/events/mlbb25031job/activity/upload", func(ctx http.Context) error {
		req := ctx.Request()
		resourceFile, resourceHeader, err := req.FormFile("file")
		if err != nil {
			_ = fmt.Errorf("error retrieving file: %v", err)
			return err
		}
		defer func(resourceFile multipart.File) {
			_ = resourceFile.Close()
		}(resourceFile)

		if resourceHeader.Size > 50*1024*1024 || resourceHeader.Size <= 0 {
			fmt.Println("上传文件的大小超过了限定值")
			return nil
		}
		type FormData struct {
			FileContent []byte `form:"file"` // 假设前端传递的文件字段名为 file
			FileName    string `form:"file_name"`
		}
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			form := (*req.(**http.Request)).MultipartForm
			file := form.File["file"][0]
			return uploadService.UploadFileV1(ctx, file)
		})

		reply, err := h(ctx, &req)
		if err != nil {
			return err
		}
		return ctx.Result(200, reply.(*v1f.UploadResponse))
	})
	srv.HandleFunc("/events/mlbb25031job/ping", func(w http.ResponseWriter, r *http.Request) {
		return
	})
	v1f.RegisterHelpCodeHTTPServer(srv, helpCode)
	v1f.RegisterUploadHTTPServer(srv, uploadService)
	srv.Handle("/debug/pprof/", pprof.NewHandler())
	return srv
}
