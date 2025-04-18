package data

import (
	"context"
	"database/sql"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.InitRepo = (*InitDB)(nil)

type InitDB struct {
	data *Data
	l    *log.Helper
}

func NewInitDB(d *Data, logger log.Logger) biz.InitRepo {
	return &InitDB{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (i *InitDB) InitDB(ctx context.Context, record *biz.InitDBRecord) error {
	err := model.InitDB(ctx, i.data.db)
	if err != nil {
		return err
	}
	i.l.Info("InitDB done")
	return nil
}

func (init *InitDB) QuerySql(ctx context.Context, sql string) ([]map[string]interface{}, error) {
	querySql, err := model.QuerySql(ctx, init.data.db, sql)
	if err != nil {
		return nil, err
	}
	return querySql, nil
}

// ExeSql implements biz.InitRepo.
func (i *InitDB) ExeSql(ctx context.Context, sql string) (sql.Result, error) {
	return model.ExeSql(ctx, i.data.db, sql)
}
