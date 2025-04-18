package biz

import "context"

type CountLimitRepo interface {
	AddKey(ctx context.Context, key string) error
	AddOne(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (int, error)
}

type CountLimitDTO struct {
	Key   string
	Count int
}
