package biz

import "context"

type PushEvent4UserRepo interface {
	InsertIgnore(ctx context.Context, waId string) (int64, error)
}

type PushEvent4UserDTO struct {
	ID   int
	WaId string
}
