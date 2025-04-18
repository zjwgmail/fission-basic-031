package model

import (
	"context"
	"fission-basic/api/constants"
	"time"

	"fission-basic/kit/sqlx"
)

const tableUserRemind = `user_remind`

// UserRemind 对应 user_remind 表的结构体
type UserRemind struct {
	ID           int64     `db:"id"`             // 自增主键，唯一标识每条记录
	WaID         string    `db:"wa_id"`          // 用户的 WhatsApp ID
	LastSendTime int64     `db:"last_send_time"` // 最后消息发送时间
	SendTimeV0   int64     `db:"send_time_v0"`   // V0 信息发送时间
	StatusV0     int       `db:"status_v0"`      // 提醒消息发送状态，1:未发送 2:已发送
	SendTimeV22  int64     `db:"send_time_v22"`  // v22 信息发送时间
	StatusV22    int       `db:"status_v22"`     // 提醒消息发送状态，1:未发送 2:已发送
	SendTimeV3   int64     `db:"send_time_v3"`   // v3 信息发送时间
	StatusV3     int       `db:"status_v3"`      // 提醒消息发送状态，1:未发送 2:已发送
	SendTimeV36  int64     `db:"send_time_v36"`  // v36 信息发送时间
	StatusV36    int       `db:"status_v36"`     // 提醒消息发送状态，1:未发送 2:已发送 3:不发送
	CreateTime   time.Time `db:"create_time"`    // 记录创建的时间，插入记录时自动记录当前时间
	UpdateTime   time.Time `db:"update_time"`    // 记录更新的时间，插入或更新记录时自动更新为当前时间
	Del          int8      `db:"del"`            // 标记记录是否删除，0 表示未删除，1 表示已删除
}

func InsertUserRemind(ctx context.Context, db sqlx.DB, userRemind *UserRemind) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableUserRemind, userRemind)
}

func GetUserRemindByWaID(ctx context.Context, db sqlx.DB, waID string) (*UserRemind, error) {
	where := map[string]interface{}{
		"wa_id": waID,
	}

	var ur UserRemind
	err := sqlx.GetContext(ctx, db, &ur, tableUserRemind, where)
	if err != nil {
		return nil, err
	}

	return &ur, nil
}

func UpdateUserRemindLastSentTime(ctx context.Context, db sqlx.DB, waID string, lastSendTime int64) error {
	update := map[string]interface{}{
		"last_send_time": lastSendTime,
	}
	where := map[string]interface{}{
		"wa_id": waID,
	}

	return sqlx.UpdateContext(ctx, db, tableUserRemind, where, update)
}

func UpdateUserRemindV22SendTime(ctx context.Context, db sqlx.DB, waID string, lastSendTime, sentTimeV22 int64) error {
	update := map[string]interface{}{
		"last_send_time": lastSendTime,
		"send_time_v22":  sentTimeV22,
		"status_v22":     constants.UserRemindVXStatusMsgNotSend,
	}
	where := map[string]interface{}{
		"wa_id": waID,
	}

	return sqlx.UpdateContext(ctx, db, tableUserRemind, where, update)
}

func UpdateUserRemindV3Status(
	ctx context.Context, db sqlx.DB,
	waID string, oldStatus, newStatus int,
) error {
	where := map[string]interface{}{
		"wa_id":     waID,
		"status_v3": oldStatus,
	}
	update := map[string]interface{}{
		"status_v3": newStatus,
	}

	return sqlx.UpdateContext(ctx, db, tableUserRemind, where, update)
}

func SelectUserRemindsTODOV3(
	ctx context.Context, db sqlx.DB,
	offset, length uint,
	minID, minSendTime int64) ([]*UserRemind, error) {
	where := map[string]interface{}{
		"id > ":           minID,
		"send_time_v3 < ": minSendTime,
		"status_v3":       1,
		"_limit":          []uint{offset, length},
		"_orderby":        "id asc",
	}

	var userReminds []*UserRemind
	err := sqlx.SelectContext(ctx, db, &userReminds, tableUserRemind, where)
	if err != nil {
		return nil, err
	}

	return userReminds, nil
}

func SelectUserRemindsTODOV22(
	ctx context.Context, db sqlx.DB,
	offset, length uint,
	minID, minSendTime int64) ([]*UserRemind, error) {
	where := map[string]interface{}{
		"id > ":            minID,
		"send_time_v22 < ": minSendTime,
		"status_v22":       1,
		"_limit":           []uint{offset, length},
		"_orderby":         "id asc",
	}

	var userReminds []*UserRemind
	err := sqlx.SelectContext(ctx, db, &userReminds, tableUserRemind, where)
	if err != nil {
		return nil, err
	}

	return userReminds, nil
}

func SelectUserRemindsTODOV0(
	ctx context.Context, db sqlx.DB,
	offset, length uint,
	minID, minSendTime int64) ([]*UserRemind, error) {
	where := map[string]interface{}{
		"id > ":           minID,
		"send_time_v0 < ": minSendTime,
		"status_v0":       1,
		"_limit":          []uint{offset, length},
		"_orderby":        "id asc",
	}

	var userReminds []*UserRemind
	err := sqlx.SelectContext(ctx, db, &userReminds, tableUserRemind, where)
	if err != nil {
		return nil, err
	}

	return userReminds, nil
}

func UpdateUserRemindV22Status(
	ctx context.Context, db sqlx.DB,
	waID string, oldStatus, newStatus int,
) error {
	where := map[string]interface{}{
		"wa_id":      waID,
		"status_v22": oldStatus,
	}
	update := map[string]interface{}{
		"status_v22": newStatus,
	}

	return sqlx.UpdateContext(ctx, db, tableUserRemind, where, update)
}

func UpdateUserRemindV0Status(
	ctx context.Context, db sqlx.DB,
	waID string, oldStatus, newStatus int,
) error {
	where := map[string]interface{}{
		"wa_id":     waID,
		"status_v0": oldStatus,
	}
	update := map[string]interface{}{
		"status_v0": newStatus,
	}

	return sqlx.UpdateContext(ctx, db, tableUserRemind, where, update)
}
