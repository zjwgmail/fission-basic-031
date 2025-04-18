package biz

import (
	"context"
	"database/sql"
	"time"
)

type InitRepo interface {
	InitDB(ctx context.Context, record *InitDBRecord) error
	QuerySql(ctx context.Context, sql string) ([]map[string]interface{}, error)
	ExeSql(ctx context.Context, sql string) (sql.Result, error)
}

type InitDBRecord struct {
	Name      string
	CreatedAt time.Time
}
