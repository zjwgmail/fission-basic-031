package model

import (
	"context"
	"time"

	"fission-basic/kit/sqlx"
)

const tableReceiptMsgRecord = `receipt_msg_record`

// ReceiptMsgRecord 对应 receipt_msg_record 表的结构体
type ReceiptMsgRecord struct {
	ID         int       `db:"id"`          // 自增主键，唯一标识每条记录
	MsgID      string    `db:"msg_id"`      // wa消息id
	MsgState   int       `db:"msg_state"`   // wa消息状态
	State      int       `db:"state"`       // 状态表 1:未完成 2:已完成
	CreateTime time.Time `db:"create_time"` // 记录创建时间，插入时自动记录当前时间
	UpdateTime time.Time `db:"update_time"` // 记录更新时间，更新记录时自动更新为当前时间
	Del        int8      `db:"del"`         // 标记记录是否删除，0 表示未删除，1 表示已删除
	CostInfo   string    `db:"cost_info"`   // 花费实体json数据
	WaId       string    `db:"wa_id"`       // waId
	Pt         string    `db:"pt"`          // 分区字段
}

func InsertReceiptMsgRecord(ctx context.Context, db sqlx.DB, receiptMsgRecord *ReceiptMsgRecord) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableReceiptMsgRecord, receiptMsgRecord)
}

func GetReceiptMsgRecord(ctx context.Context, db sqlx.DB, msgID string) (*ReceiptMsgRecord, error) {
	where := map[string]interface{}{
		"msg_id": msgID,
	}

	var receiptMsgRecord ReceiptMsgRecord
	err := sqlx.GetContext(ctx, db, &receiptMsgRecord, tableReceiptMsgRecord, where)
	if err != nil {
		return nil, err
	}

	return &receiptMsgRecord, nil
}

func UpdateReceiptMsgRecordState(ctx context.Context, db sqlx.DB, msgID string, oldState, newState int) error {
	where := map[string]interface{}{
		"msg_id": msgID,
		"state":  oldState,
	}
	update := map[string]interface{}{
		"state": newState,
	}

	return sqlx.UpdateContext(ctx, db, tableReceiptMsgRecord, where, update)
}

func SelectReceiptMsgRecords(ctx context.Context,
	db sqlx.DB, state int,
	minID int, offset, length uint,
	maxCreateTime time.Time,
) ([]*ReceiptMsgRecord, error) {
	today := maxCreateTime.Format("20060102")
	yesterday := maxCreateTime.AddDate(0, 0, -1).Format("20060102")
	where := map[string]interface{}{
		"state":         state,
		"id >":          minID,
		"create_time <": maxCreateTime,
		"pt in":         []string{today, yesterday},

		"_limit":   []uint{offset, length},
		"_orderby": "id asc",
	}

	var receiptMsgRecords []*ReceiptMsgRecord
	err := sqlx.SelectContext(ctx, db, &receiptMsgRecords, tableReceiptMsgRecord, where)
	if err != nil {
		return nil, err
	}

	return receiptMsgRecords, nil
}
