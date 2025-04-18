package model

import (
	"context"
	"fission-basic/kit/sqlx"
)

const tableCountLimit = `count_limit`

type CountLimitEntity struct {
	Key   string `db:"ckey"`
	Count int    `db:"count"`
}

func CountLimitInsert(ctx context.Context, db sqlx.DB, key string) error {
	_, err := sqlx.InsertIgnoreContext(ctx, db, tableCountLimit, &CountLimitEntity{
		Key: key,
	})
	return err
}

func CountLimitGet(ctx context.Context, db sqlx.DB, key string) (int, error) {
	where := map[string]interface{}{
		"ckey": key,
	}
	var entity CountLimitEntity
	err := sqlx.GetContext(ctx, db, &entity, tableCountLimit, where)
	if err != nil {
		return 0, err
	}
	return entity.Count, nil
}

func CountLimitAddOne(ctx context.Context, db sqlx.DB, key string) error {
	return sqlx.ExecContext(ctx, db, "update count_limit set count = count + 1 where ckey = ?", key)
}
