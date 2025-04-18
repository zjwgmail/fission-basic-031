package service

import (
	"context"
	"fmt"

	v1 "fission-basic/api/helloworld/v1"
	"fission-basic/internal/biz"

	"github.com/go-kratos/kratos/v2/transport"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer

	uc *biz.GreeterUsecase
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

// SayHello implements helloworld.GreeterServer.
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		fmt.Println("获取header失败")
	} else {
		header := tr.RequestHeader()
		fmt.Println(header.Get("User-Agent"))
	}

	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name})
	if err != nil {
		return nil, err
	}
	return &v1.HelloReply{Message: "Hello " + g.Hello}, nil
}
