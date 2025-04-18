package model

import (
	"context"
	"fission-basic/kit/sqlx"
	"time"
)

const tableUserJoinGroup = `user_join_group`

// UserJoinGroup 对应 user_join_group 表的结构体
type UserJoinGroup struct {
	ID            int64     `db:"id"`              // 自增主键，唯一标识每条记录
	JoinWaID      string    `db:"join_wa_id"`      // 助力人ID
	HelpCode      string    `db:"help_code"`       // 被助力码
	JoinGroupTime int64     `db:"join_group_time"` // 助力时间
	CreateTime    time.Time `db:"create_time"`     // 记录创建时间，插入时自动记录当前时间
	UpdateTime    time.Time `db:"update_time"`     // 记录更新时间，更新记录时自动更新为当前时间
	Del           int8      `db:"del"`             // 标记记录是否删除，0 表示未删除，1 表示已删除
}

func InsertUserJoinGroup(
	ctx context.Context, db sqlx.DB,
	userJoinGroup *UserJoinGroup,
) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableUserJoinGroup, userJoinGroup)
}

func SelectUserJoinGroupsByHelpCode(
	ctx context.Context, db sqlx.DB,
	helpCode string,
) ([]*UserJoinGroup, error) {
	where := map[string]interface{}{
		"help_code": helpCode,
	}

	var rets []*UserJoinGroup
	err := sqlx.SelectContext(ctx, db, &rets, tableUserJoinGroup, where)

	return rets, err
}

func UserJoinGroupListGtIdBetweenJoinGroupTime(ctx context.Context, db sqlx.DB, minId int64,
	startTimestamp int64, endTimestamp int64, limit uint) ([]*UserJoinGroup, error) {
	where := map[string]interface{}{
		"id > ":              minId,
		"join_group_time >=": startTimestamp,
		"join_group_time <":  endTimestamp,
		"_orderby":           "id",
		"_limit":             []uint{limit},
	}

	var userJoinGroups []*UserJoinGroup
	err := sqlx.SelectContext(ctx, db, &userJoinGroups, tableUserJoinGroup, where)
	if err != nil {
		return nil, err
	}

	return userJoinGroups, nil
}

func UserJoinGroupGetFirstGtJoinGroupTime(ctx context.Context, db sqlx.DB, timestamp int64) (*UserJoinGroup, error) {
	where := map[string]interface{}{
		"join_group_time >= ": timestamp,
		"_orderby":            "id",
	}

	var userJoinGroup UserJoinGroup
	err := sqlx.GetContext(ctx, db, &userJoinGroup, tableUserJoinGroup, where)
	if err != nil {
		return nil, err
	}

	return &userJoinGroup, nil
}

func FindUserJoinGroupByWaID(ctx context.Context, db sqlx.DB, waID string) (*UserJoinGroup, error) {
	where := map[string]interface{}{
		"join_wa_id": waID,
	}

	var userJoinGroup UserJoinGroup
	err := sqlx.GetContext(ctx, db, &userJoinGroup, tableUserJoinGroup, where)
	if err != nil {
		return nil, err
	}

	return &userJoinGroup, nil
}
