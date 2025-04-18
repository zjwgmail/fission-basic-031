package biz

import (
	"context"
	"database/sql"

	"github.com/go-kratos/kratos/v2/log"
)

type Init struct {
	initRepo InitRepo
	l        *log.Helper
}

func NewInit(initRepo InitRepo, l log.Logger) *Init {
	return &Init{
		initRepo: initRepo,
		l:        log.NewHelper(l),
	}
}

func (init *Init) InitDB(ctx context.Context, name string) error {
	err := init.initRepo.InitDB(ctx, &InitDBRecord{})
	if err != nil {
		return err
	}
	return nil
}

func (init *Init) ExeSql(ctx context.Context, sql string) (sql.Result, error) {
	querySql, err := init.initRepo.ExeSql(ctx, sql)
	if err != nil {
		return nil, err
	}
	return querySql, nil
}

func (init *Init) QuerySql(ctx context.Context, sql string) ([]map[string]interface{}, error) {
	querySql, err := init.initRepo.QuerySql(ctx, sql)
	if err != nil {
		return nil, err
	}

	return querySql, nil
}
