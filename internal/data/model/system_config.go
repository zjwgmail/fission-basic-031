package model

import (
	"context"
	"fission-basic/kit/sqlx"
)

const tableSystemConfig = `system_config`

type SystemConfigEntity struct {
	Key   string `db:"param_key"`
	Value string `db:"param_value"`
}

func InsertSystemConfig(ctx context.Context, db sqlx.DB, scEntity *SystemConfigEntity) error {
	_, err := sqlx.InsertIgnoreContext(ctx, db, tableSystemConfig, scEntity)
	return err
}

func GetSystemConfigByKey(ctx context.Context, db sqlx.DB, key string) (string, error) {
	where := map[string]interface{}{
		"param_key": key,
	}
	var scEntity SystemConfigEntity
	err := sqlx.GetContext(ctx, db, &scEntity, tableSystemConfig, where)
	if err != nil {
		return "", err
	}
	return scEntity.Value, nil
}

func UpdateSystemConfigByKey(ctx context.Context, db sqlx.DB, param *SystemConfigEntity) error {
	where := map[string]interface{}{
		"param_key": param.Key,
	}
	update := map[string]interface{}{
		"param_value": param.Value,
	}
	return sqlx.UpdateContext(ctx, db, tableSystemConfig, where, update)
}
