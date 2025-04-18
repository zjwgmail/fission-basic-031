package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

type SystemConfig struct {
	data *Data
	l    *log.Helper
}

func NewSystemConfig(d *Data, logger log.Logger) biz.SystemConfigRepo {
	return &SystemConfig{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (sc *SystemConfig) AddOne(ctx context.Context, param *biz.SystemConfigParam) error {
	return model.InsertSystemConfig(ctx, sc.data.db, &model.SystemConfigEntity{
		Key:   param.Key,
		Value: param.Value,
	})
}

func (sc *SystemConfig) GetByKey(ctx context.Context, key string) (string, error) {
	return model.GetSystemConfigByKey(ctx, sc.data.db, key)
}

func (sc *SystemConfig) UpdateByKey(ctx context.Context, param *biz.SystemConfigParam) error {
	return model.UpdateSystemConfigByKey(ctx, sc.data.db, &model.SystemConfigEntity{
		Key:   param.Key,
		Value: param.Value,
	})
}
