package biz

import "context"

type SystemConfigRepo interface {
	AddOne(ctx context.Context, param *SystemConfigParam) error
	GetByKey(ctx context.Context, key string) (string, error)
	UpdateByKey(ctx context.Context, param *SystemConfigParam) error
}

type SystemConfigParam struct {
	Key   string
	Value string
}
