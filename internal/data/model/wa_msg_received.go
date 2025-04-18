package model

import (
	"context"
	"fission-basic/kit/sqlx"
	"fmt"
	"time"
)

const tableWaMsgReceived = `wa_msg_received`

// WaMsgReceived 对应 wa_msg_received 表的结构体
type WaMsgReceived struct {
	ID              int       `db:"id"`                // 自增主键，唯一标识每条记录
	WaMsgID         string    `db:"wa_msg_id"`         // wa 消息的唯一标识
	WaID            string    `db:"wa_id"`             // 可能是用户或会话的 WhatsApp ID
	Content         string    `db:"content"`           // 发送消息的具体内容
	MsgReceivedTime int64     `db:"msg_received_time"` // 消息发送的时间，以大整数表示
	CreateTime      time.Time `db:"create_time"`       // 记录创建的时间，插入记录时自动记录当前时间
	UpdateTime      time.Time `db:"update_time"`       // 记录更新的时间，插入或更新记录时自动更新为当前时间
	Del             int8      `db:"del"`               // 标记记录是否删除，0 表示未删除，1 表示已删除
}

func InsertWaMsgReceived(ctx context.Context, db sqlx.DB, msg *WaMsgReceived) (int64, error) {
	realTable := fmt.Sprintf("%s_%s", tableWaMsgReceived, msg.CreateTime.Format("2006_01_02"))
	return sqlx.InsertContext(ctx, db, realTable, msg)
}

func WaMsgReceivedListGtIdReceivedTime(ctx context.Context, db sqlx.DB, suffix string, startTimeStamp int64, endTimeStamp int64, minId int, limit uint) ([]*WaMsgReceived, error) {
	realTable := fmt.Sprintf("%s_%s", tableWaMsgReceived, suffix)
	where := map[string]interface{}{
		"id >":                 minId,
		"msg_received_time >":  startTimeStamp,
		"msg_received_time <=": endTimeStamp,
		"_orderby":             "id",
		"_limit":               []uint{limit},
	}
	var waMsgReceivedList []*WaMsgReceived
	err := sqlx.SelectContext(ctx, db, &waMsgReceivedList, realTable, where)
	return waMsgReceivedList, err
}
