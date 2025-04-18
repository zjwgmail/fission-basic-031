package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type SystemConfigUsecase struct {
	repo SystemConfigRepo
	l    *log.Helper
}

func NewSystemConfigUsecase(repo SystemConfigRepo, l log.Logger) *SystemConfigUsecase {
	return &SystemConfigUsecase{
		repo: repo,
		l:    log.NewHelper(l),
	}
}

func (sc *SystemConfigUsecase) AddOne(ctx context.Context, key string, value string) error {
	return sc.repo.AddOne(ctx, &SystemConfigParam{
		Key:   key,
		Value: value,
	})
}

func (sc *SystemConfigUsecase) UpdateByKey(ctx context.Context, key string, value string) error {
	return sc.repo.UpdateByKey(ctx, &SystemConfigParam{
		Key:   key,
		Value: value,
	})
}

func (sc *SystemConfigUsecase) GetByKey(ctx context.Context, key string) (string, error) {
	return sc.repo.GetByKey(ctx, key)
}
