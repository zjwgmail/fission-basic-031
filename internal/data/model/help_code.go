package model

import (
	"context"
	"database/sql"
	"fission-basic/internal/biz"
	"fission-basic/kit/sqlx"
	"time"
)

const tableHelpCode = `help_code`

type HelpCodeEntity struct {
	Id          int64     `db:"id"`
	Del         string    `db:"del"`
	CreateTime  time.Time `db:"create_time"`
	UpdateTime  time.Time `db:"update_time"`
	HelpCode    string    `db:"help_code"`
	ShortLinkV0 string    `db:"short_link_v0"`
	ShortLinkV1 string    `db:"short_link_v1"`
	ShortLinkV2 string    `db:"short_link_v2"`
	ShortLinkV3 string    `db:"short_link_v3"`
	ShortLinkV4 string    `db:"short_link_v4"`
	ShortLinkV5 string    `db:"short_link_v5"`
}

func HelpCodeInsert(ctx context.Context, db sqlx.DB, hcEntity *HelpCodeEntity) (int64, error) {
	return sqlx.InsertIgnoreContext(ctx, db, tableHelpCode, hcEntity)
}

func HelpCodeUpdateById(ctx context.Context, db sqlx.DB, hcEntity *HelpCodeEntity) error {
	where := map[string]interface{}{
		"id": hcEntity.Id,
	}
	update := map[string]interface{}{
		"help_code": hcEntity.HelpCode,
	}
	return sqlx.UpdateContext(ctx, db, tableHelpCode, where, update)
}

func UpdateShortLinkByHelpCode(ctx context.Context, db sqlx.DB, hcParam *biz.HelpCodeParam) error {
	where := map[string]interface{}{
		"help_code": hcParam.HelpCode,
	}
	update := map[string]interface{}{
		hcParam.ShortLinkVersion: hcParam.ShortLink,
	}
	return sqlx.UpdateContext(ctx, db, tableHelpCode, where, update)
}

func ListShortLinkByHelpCode(ctx context.Context, db sqlx.DB, hcParam *biz.HelpCodeParam) ([]string, error) {
	where := map[string]interface{}{
		"help_code": hcParam.HelpCode,
	}
	var helpCodeEntityList HelpCodeEntity
	err := sqlx.GetContext(ctx, db, &helpCodeEntityList, tableHelpCode, where)
	if err != nil {
		return nil, err
	}

	return []string{
		helpCodeEntityList.ShortLinkV0,
		helpCodeEntityList.ShortLinkV1,
		helpCodeEntityList.ShortLinkV2,
		helpCodeEntityList.ShortLinkV3,
		helpCodeEntityList.ShortLinkV4,
		helpCodeEntityList.ShortLinkV5,
	}, nil
}

func HelpCodeGetById(ctx context.Context, db sqlx.DB, id int64) (string, map[int]string, error) {
	where := map[string]interface{}{
		"id": id,
	}
	var helpCodeEntity HelpCodeEntity
	err := sqlx.GetContext(ctx, db, &helpCodeEntity, tableHelpCode, where)
	if err != nil {
		return "", nil, err
	}
	return helpCodeEntity.HelpCode, map[int]string{
		0: helpCodeEntity.ShortLinkV0,
		1: helpCodeEntity.ShortLinkV1,
		2: helpCodeEntity.ShortLinkV2,
		3: helpCodeEntity.ShortLinkV3,
		4: helpCodeEntity.ShortLinkV4,
		5: helpCodeEntity.ShortLinkV5,
	}, nil
}

func HelpCodeDeleteById(ctx context.Context, db sqlx.DB, id int64) error {
	return sqlx.DeleteContext(ctx, db, tableHelpCode, map[string]interface{}{
		"id": id,
	})
}

func HelpCodeGetMaxId(ctx context.Context, db sqlx.DB) (int64, error) {
	where := map[string]interface{}{}
	var maxId sql.NullInt64
	err := sqlx.GetContext(ctx, db, &maxId, tableHelpCode, where, "max(id)")
	return maxId.Int64, err
}

func HelpCodeListGtId(ctx context.Context, db sqlx.DB, id int64, limit uint) ([]*HelpCodeEntity, error) {
	where := map[string]interface{}{
		"id >":     id,
		"_orderby": "id",
		"_limit":   []uint{limit},
	}
	var rets []*HelpCodeEntity

	err := sqlx.SelectContext(ctx, db, &rets, tableHelpCode, where)
	if err != nil {
		return nil, err
	}
	return rets, nil
}
