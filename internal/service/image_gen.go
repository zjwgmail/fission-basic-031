package service

import (
	"context"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/biz"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
)

type ImageService struct {
	imageGenerate *biz.ImageGenerate
	l             *log.Helper
}

func NewImageService(imageGenerate *biz.ImageGenerate, l log.Logger) *ImageService {
	return &ImageService{
		imageGenerate: imageGenerate,
		l:             log.NewHelper(l),
	}
}

func (i *ImageService) ImageGenerate(ctx context.Context, req *v1.SynthesisParamRequest) (*v1.SynthesisResponse, error) {
	if req == nil {
		return nil, nil
	}

	res, err := i.imageGenerate.GetInteractiveImageUrl(ctx, req, "")

	if err != nil {
		i.l.Error(fmt.Sprintf("方法[%s]，err:%v", "InitDB", err))
		return nil, err
	}
	return &v1.SynthesisResponse{
		Url: res,
	}, nil
}

func (i *ImageService) ImagedDowngrade(ctx context.Context, req *v1.SynthesisParamRequest) (*v1.SynthesisResponse, error) {
	if req == nil {
		return nil, nil
	}
	res, err := i.imageGenerate.ImageDowngrade(ctx, req, "")
	if err != nil {
		i.l.Error(fmt.Sprintf("方法[%s]，err:%v", "InitDB", err))
		return nil, err
	}
	return &v1.SynthesisResponse{
		Url: res,
	}, nil
}
