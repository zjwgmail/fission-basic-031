package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.PushEvent4UserRepo = (*PushEvent4User)(nil)

type PushEvent4User struct {
	data *Data
	l    *log.Helper
}

func NewPushEvent4User(d *Data, logger log.Logger) biz.PushEvent4UserRepo {
	return &PushEvent4User{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (p PushEvent4User) InsertIgnore(ctx context.Context, waId string) (int64, error) {
	return model.PushEvent4UserInsertIgnore(ctx, p.data.db, waId)
}
