package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

type CountLimit struct {
	data *Data
	l    *log.Helper
}

func NewCountLimit(d *Data, logger log.Logger) biz.CountLimitRepo {
	return &CountLimit{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (c CountLimit) AddKey(ctx context.Context, key string) error {
	return model.CountLimitInsert(ctx, c.data.db, key)
}

func (c CountLimit) AddOne(ctx context.Context, key string) error {
	return model.CountLimitAddOne(ctx, c.data.db, key)
}

func (c CountLimit) Get(ctx context.Context, key string) (int, error) {
	count, err := model.CountLimitGet(ctx, c.data.db, key)
	if err != nil {
		return 0, err
	}
	return count, nil
}
