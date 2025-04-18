package model

import (
	"context"
	"time"

	"fission-basic/kit/sqlx"
)

const tableUserCreteGroup = `user_create_group`

// UserCreateGroup 对应 user_create_group 表的结构体
type UserCreateGroup struct {
	ID              int       `db:"id"`                // 自增主键，唯一标识每条记录
	CreateWaID      string    `db:"create_wa_id"`      // 开团人ID
	HelpCode        string    `db:"help_code"`         // 助力码
	Generation      int       `db:"generation"`        // 代次
	CreateGroupTime int64     `db:"create_group_time"` // 开团时间
	CreateTime      time.Time `db:"create_time"`       // 记录创建时间，插入时自动记录当前时间
	UpdateTime      time.Time `db:"update_time"`       // 记录更新时间，更新记录时自动更新为当前时间
	Del             int8      `db:"del"`               // 标记记录是否删除，0 表示未删除，1 表示已删除
}

func InsertUserCreateGroup(ctx context.Context,
	db sqlx.DB, userCreateGroup *UserCreateGroup) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableUserCreteGroup, userCreateGroup)
}

func GetUserCreateGroup(ctx context.Context,
	db sqlx.DB, waID string) (*UserCreateGroup, error) {
	where := map[string]interface{}{
		"create_wa_id": waID,
	}

	var userCreateGroup UserCreateGroup
	err := sqlx.GetContext(ctx, db, &userCreateGroup, tableUserCreteGroup, where)
	if err != nil {
		return nil, err
	}

	return &userCreateGroup, nil
}

func GetUserCreateGroupByHelpCode(ctx context.Context,
	db sqlx.DB, helpCode string) (*UserCreateGroup, error) {
	where := map[string]interface{}{
		"help_code": helpCode,
	}

	var userCreateGroup UserCreateGroup
	err := sqlx.GetContext(ctx, db, &userCreateGroup, tableUserCreteGroup, where)
	if err != nil {
		return nil, err
	}

	return &userCreateGroup, nil
}
