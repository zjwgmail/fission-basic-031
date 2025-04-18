package server

import (
	"context"
	v1f "fission-basic/api/fission/v1"
	v1 "fission-basic/api/helloworld/v1"
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
func NewHTTPServer(c *conf.Server,
	greeter *service.GreeterService,
	student *service.StudentService,
	nxCloudService *service.NxCloudService,
	cdkService *service.CDKService,
	initService *service.InitService,
	imageService *service.ImageService,
	uploadService *service.UploadService,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(logger),
		),
	}

	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	srv.HandleFunc("/events/mlbb25031gateway/ping", func(w http.ResponseWriter, r *http.Request) {
		return
	})

	srv.Handle("/debug/pprof/", pprof.NewHandler())

	r := srv.Route("/")
	r.GET("/events/mlbb25031gateway/invite", func(ctx http.Context) error {
		if ctx.Request().Header.Get("Content-Type") == "" {
			ctx.Request().Header.Set("Content-Type", "application/json")
		}

		var in v1.InvitationRequest
		if err := ctx.Bind(&in); err != nil {
			fmt.Println("bind", err)
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			fmt.Println("bindQuery", err)
			return err
		}

		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return student.Invitation(ctx, req.(*v1.InvitationRequest))
		})

		reply, err := h(ctx, &in)
		if err != nil {
			return err
		}
		resp := reply.(*v1.InvitationResponse)

		_, err = ctx.Response().Write([]byte(resp.HtmlText))

		return err
	})

	r.POST("/events/mlbb25031gateway/activity/sql-query", func(ctx http.Context) error {
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

	v1.RegisterGreeterHTTPServer(srv, greeter)
	v1.RegisterStudentHTTPServer(srv, student)
	v1f.RegisterNXCloudHTTPServer(srv, nxCloudService)
	v1f.RegisterCDKHTTPServer(srv, cdkService)
	v1f.RegisterInitDBHTTPServer(srv, initService)
	v1f.RegisterImageGenerateHTTPServer(srv, imageService)
	v1f.RegisterUploadHTTPServer(srv, uploadService)
	return srv
}
