package model

import (
	"context"
	"fission-basic/kit/sqlx"
	"time"
)

const tableOfficialMsgRecord = `official_msg_record`

// OfficialMsgRecord 对应 official_msg_record 表的结构体
type OfficialMsgRecord struct {
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

func InsertOfficialMsgRecord(ctx context.Context, db sqlx.DB, officialMsgRecord *OfficialMsgRecord) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableOfficialMsgRecord, officialMsgRecord)
}

func UpdateOfficialMsgRecordLanguageAndState(
	ctx context.Context, db sqlx.DB,
	waID, rallycode string,
	language string,
	newState int,
) error {
	where := map[string]interface{}{
		"wa_id":      waID,
		"rally_code": rallycode,
	}
	update := map[string]interface{}{
		"language": language,
		"state":    newState,
	}

	return sqlx.UpdateContext(ctx, db, tableOfficialMsgRecord, where, update)
}

func UpdateOfficialMsgRecordState(
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

	return sqlx.UpdateContext(ctx, db, tableOfficialMsgRecord, where, update)
}

func GetOfficialMsgRecord(
	ctx context.Context, db sqlx.DB,
	waID, rallyCode string,
) (*OfficialMsgRecord, error) {
	where := map[string]interface{}{
		"wa_id":      waID,
		"rally_code": rallyCode,
	}

	var officialMsgRecord OfficialMsgRecord
	err := sqlx.GetContext(ctx, db,
		&officialMsgRecord, tableOfficialMsgRecord, where)
	if err != nil {
		return nil, err
	}

	return &officialMsgRecord, nil
}

func SelectOfficialMsgRecords(
	ctx context.Context, db sqlx.DB,
	minID int, offset, length uint,
	status int,
	maxTime time.Time,
) ([]*OfficialMsgRecord, error) {
	where := map[string]interface{}{
		"id >":          minID,
		"create_time <": maxTime,
		"state":         status,

		"_limit":   []uint{offset, length},
		"_orderby": "id asc",
	}

	var officialMsgRecords []*OfficialMsgRecord
	err := sqlx.SelectContext(ctx, db, &officialMsgRecords, tableOfficialMsgRecord, where)
	if err != nil {
		return nil, err
	}

	return officialMsgRecords, nil
}
