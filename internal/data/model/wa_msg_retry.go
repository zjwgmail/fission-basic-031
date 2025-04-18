package model

import (
	"context"
	"fission-basic/kit/sqlx"
	"time"
)

const tableWaMsgRetry = `wa_msg_retry`

// WaMsgRetry 对应 wa_msg_retry 表的结构体
type WaMsgRetry struct {
	ID            int64     `db:"id"`              // 自增主键，唯一标识记录
	WaID          string    `db:"wa_id"`           // wa用户ID
	WaMsgID       string    `db:"wa_msg_id"`       // wa消息id
	MsgType       string    `db:"msg_type"`        // 消息类型
	State         int       `db:"state"`           // 发送状态 0:初始化 1:发送成功 2:发送失败
	Content       string    `db:"content"`         // 发送消息内容
	CreateTime    time.Time `db:"create_time"`     // 记录创建时间，插入时自动记录当前时间
	UpdateTime    time.Time `db:"update_time"`     // 记录更新时间，更新记录时自动更新为当前时间
	Del           int8      `db:"del"`             // 标记记录是否删除，0 表示未删除，1 表示已删除
	BuildMsgParam string    `db:"build_msg_param"` // 发送消息的具体内容
	SendRes       string    `db:"send_res"`        // 发送返回结果
}

func InsertWaMsgRetry(ctx context.Context, db sqlx.DB, msg *WaMsgRetry) (int64, error) {
	now := time.Now()
	msg.CreateTime = now
	msg.UpdateTime = now
	return sqlx.InsertContext(ctx, db, tableWaMsgRetry, msg)
}

func GetWaMsgRetry(ctx context.Context, db sqlx.DB, waMsgID string) (*WaMsgRetry, error) {
	where := map[string]interface{}{
		"wa_msg_id": waMsgID,
	}

	var wmr WaMsgRetry
	err := sqlx.GetContext(ctx, db, &wmr, tableWaMsgRetry, where)
	if err != nil {
		return nil, err
	}
	return &wmr, nil
}

func UpdateWaMsgRetryState(ctx context.Context, db sqlx.DB, waMsgID string, newState int) error {
	where := map[string]interface{}{
		"wa_msg_id": waMsgID,
	}
	update := map[string]interface{}{
		"state": newState,
	}

	return sqlx.UpdateContext(ctx, db, tableWaMsgRetry, where, update)
}

func UpdateWaMsgRetry(ctx context.Context, db sqlx.DB, msg *WaMsgRetry) error {
	where := map[string]interface{}{
		"id": msg.ID,
	}
	updateEntity := map[string]interface{}{}
	if "" != msg.WaMsgID {
		updateEntity["wa_msg_id"] = msg.WaMsgID
	}
	if 0 != msg.State {
		updateEntity["state"] = msg.State
	}
	if "" != msg.Content {
		updateEntity["content"] = msg.Content
	}
	if "" != msg.BuildMsgParam {
		updateEntity["build_msg_param"] = msg.BuildMsgParam
	}

	if "" != msg.SendRes {
		updateEntity["send_res"] = msg.SendRes
	}

	return sqlx.UpdateContext(ctx, db, tableWaMsgRetry, where, updateEntity)
}

func ListMsgRetryByWaIdAndState(ctx context.Context, db sqlx.DB, state []int, waId string) ([]*WaMsgRetry, error) {
	where := map[string]interface{}{
		"state in": state,
		"wa_id":    waId,
	}

	var waMsgSends []*WaMsgRetry

	err := sqlx.SelectContext(ctx, db, &waMsgSends, tableWaMsgRetry, where)
	if err != nil {
		return nil, err
	}
	return waMsgSends, nil
}

func ListWaIdOfRetryByState(ctx context.Context, db sqlx.DB, minWaId string, limit uint, state []int) ([]string, error) {
	where := map[string]interface{}{
		"state in": state,
		"wa_id > ": minWaId,
		"_limit":   []uint{limit},
		"_orderby": "wa_id asc",
	}
	var waIdList []string

	err := sqlx.SelectContext(ctx, db, &waIdList, tableWaMsgRetry, where, "distinct(wa_id)")
	if err != nil {
		return nil, err
	}
	return waIdList, nil
}
