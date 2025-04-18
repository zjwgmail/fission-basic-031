package model

import (
	"context"
	"fission-basic/kit/sqlx"
	"time"
)

const tableUnOfficialMsgRecord = `unofficial_msg_record`

// UnOfficialMsgRecord 对应 unofficial_msg_record 表的结构体
type UnOfficialMsgRecord struct {
	ID         int       `db:"id"`          // 自增主键，唯一标识每条记录
	WaID       string    `db:"wa_id"`       // 开团人标识
	RallyCode  string    `db:"rally_code"`  // 助力码，用于识别助力活动
	State      int       `db:"state"`       // 状态表 1:未完成 2:已完成
	Channel    string    `db:"channel"`     // 用户来源的渠道
	Language   string    `db:"language"`    // 用户使用的语言
	Generation int       `db:"generation"`  // 用户参与活动的代数
	Nickname   string    `db:"nickname"`    // 用户的昵称
	SendTime   int64     `db:"send_time"`   // 消息发送时间
	CreateTime time.Time `db:"create_time"` // 记录创建时间，插入时自动记录当前时间
	UpdateTime time.Time `db:"update_time"` // 记录更新时间，更新记录时自动更新为当前时间
	Del        int8      `db:"del"`         // 标记记录是否删除，0 表示未删除，1 表示已删除
}

func InsertUnOfficialMsgRecord(ctx context.Context, db sqlx.DB, unOfficialMsgRecord *UnOfficialMsgRecord) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableUnOfficialMsgRecord, unOfficialMsgRecord)
}

func UpdateUnOfficialMsgRecordState(
	ctx context.Context, db sqlx.DB,
	waID, rallycode string,
	oldState, newState int,
) error {
	where := map[string]interface{}{
		"wa_id":      waID,
		"rally_code": rallycode,
		"state":      oldState,
	}
	update := map[string]interface{}{
		"state": newState,
	}

	return sqlx.UpdateContext(ctx, db, tableUnOfficialMsgRecord, where, update)
}

func GetUnOfficialMsgRecord(
	ctx context.Context, db sqlx.DB,
	waID, rallyCode string,
) (*UnOfficialMsgRecord, error) {
	where := map[string]interface{}{
		"wa_id":      waID,
		"rally_code": rallyCode,
	}

	var unOfficialMsgRecord UnOfficialMsgRecord
	err := sqlx.GetContext(ctx, db,
		&unOfficialMsgRecord, tableUnOfficialMsgRecord, where)
	if err != nil {
		return nil, err
	}

	return &unOfficialMsgRecord, nil
}

func SelectUnOfficialMsgRecords(
	ctx context.Context, db sqlx.DB,
	minID int, offset, length uint,
	status int,
	maxTime time.Time,
) ([]*UnOfficialMsgRecord, error) {
	where := map[string]interface{}{
		"id >":          minID,
		"create_time <": maxTime,
		"state":         status,
		"_limit":        []uint{offset, length},
		"_orderby":      "id asc",
	}

	var unOfficialMsgRecords []*UnOfficialMsgRecord
	err := sqlx.SelectContext(ctx, db, &unOfficialMsgRecords, tableUnOfficialMsgRecord, where)
	if err != nil {
		return nil, err
	}

	return unOfficialMsgRecords, nil
}
