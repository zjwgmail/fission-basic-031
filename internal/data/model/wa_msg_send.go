package model

import (
	"context"
	"fission-basic/util"
	"time"

	"fission-basic/kit/sqlx"
)

const tableWaMsgSend = `wa_msg_send`

// WaMsgSend 对应 wa_msg_send 表的结构体
type WaMsgSend struct {
	ID            int64     `db:"id"`              // 自增主键，唯一标识每条记录
	WaMsgID       string    `db:"wa_msg_id"`       // wa 消息的唯一标识
	WaID          string    `db:"wa_id"`           // 可能是用户或会话的 WhatsApp ID
	MsgType       string    `db:"msg_type"`        // 消息类型
	State         int       `db:"state"`           // 发送状态 1:发送成功 2:发送失败
	Content       string    `db:"content"`         // 发送消息的具体内容
	BuildMsgParam string    `db:"build_msg_param"` // 发送消息的具体内容
	SendRes       string    `db:"send_res"`        // 发送返回结果
	Pt            string    `db:"pt"`              // 分区字段
	CreateTime    time.Time `db:"create_time"`     // 记录创建的时间，插入记录时自动记录当前时间
	UpdateTime    time.Time `db:"update_time"`     // 记录更新的时间，插入或更新记录时自动更新为当前时间
	Del           int8      `db:"del"`             // 标记记录是否删除，0 表示未删除，1 表示已删除
}

func GetWaMsgSend(ctx context.Context, db sqlx.DB, waMsgID string) (*WaMsgSend, error) {
	where := map[string]interface{}{
		"wa_msg_id": waMsgID,
	}

	var wms WaMsgSend
	err := sqlx.GetContext(ctx, db, &wms, tableWaMsgSend, where)
	if err != nil {
		return nil, err
	}

	return &wms, nil
}

func InsertWaMsgSend(ctx context.Context, db sqlx.DB, msg *WaMsgSend) (int64, error) {
	now := time.Now()
	// 格式化时间为 "2006-01-02" 格式
	formattedDate := now.Format("20060102")
	msg.Pt = formattedDate
	msg.CreateTime = now
	msg.UpdateTime = now
	return sqlx.InsertContext(ctx, db, tableWaMsgSend, msg)
}

func UpdateWaMsgSend(ctx context.Context, db sqlx.DB, msg *WaMsgSend) error {
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

	return sqlx.UpdateContext(ctx, db, tableWaMsgSend, where, updateEntity)
}

func UpdateWaMsgSendState(ctx context.Context, db sqlx.DB, waMsgID string, state int) error {
	update := map[string]interface{}{
		"state": state,
	}
	where := map[string]interface{}{
		"wa_msg_id": waMsgID,
	}

	return sqlx.UpdateContext(ctx, db, tableWaMsgSend, where, update)
}

func ListMsgSendGtId(ctx context.Context, db sqlx.DB, minId int64, limit uint) ([]*WaMsgSend, error) {
	where := map[string]interface{}{
		"id > ":    minId,
		"_orderby": "id",
		"_limit":   []uint{limit},
	}
	var waMsgSends []*WaMsgSend
	err := sqlx.SelectContext(ctx, db, &waMsgSends, tableWaMsgSend, where)
	if err != nil {
		return nil, err
	}
	return waMsgSends, nil
}

func ListMsgSendGtIdInPts(ctx context.Context, db sqlx.DB, minId int64, pts []string, limit uint) ([]*WaMsgSend, error) {
	where := map[string]interface{}{
		"id > ":    minId,
		"pt in":    pts,
		"_orderby": "id",
		"_limit":   []uint{limit},
	}
	var waMsgSends []*WaMsgSend
	err := sqlx.SelectContext(ctx, db, &waMsgSends, tableWaMsgSend, where)
	if err != nil {
		return nil, err
	}
	return waMsgSends, nil
}

func ListMsgSendByWaIdAndState(ctx context.Context, db sqlx.DB, state []int, waId string, ptList []string) ([]*WaMsgSend, error) {
	where := map[string]interface{}{
		"state in": state,
		"wa_id":    waId,
	}
	if len(ptList) > 0 {
		where["pt in"] = ptList
	}
	var waMsgSends []*WaMsgSend

	err := sqlx.SelectContext(ctx, db, &waMsgSends, tableWaMsgSend, where)
	if err != nil {
		return nil, err
	}
	return waMsgSends, nil
}

func ListWaIdByState(ctx context.Context, db sqlx.DB, minWaId string, limit uint, state []int) ([]string, error) {

	where := map[string]interface{}{
		"state in": state,
		"wa_id >":  minWaId,
		"pt in":    util.GetPtTimeList(),
		"_limit":   []uint{limit},
		"_orderby": "wa_id asc",
	}
	var waIdList []string

	err := sqlx.SelectContext(ctx, db, &waIdList, tableWaMsgSend, where, "distinct(wa_id)")
	if err != nil {
		return nil, err
	}
	return waIdList, nil
}
